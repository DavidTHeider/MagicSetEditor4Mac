// **Was die Pergamenttextur mit den Kontrasten macht**
// (#1325,
// Werkstück J).
//
// # Wozu ein eigener Lauf
//
// `docs/KONTRASTE.md` führt **Paare aus zwei Farben**, und `kontrastlauf.mjs`
// misst sie am gerenderten Element über `getComputedStyle`. Beide lesen eine
// **Farbe**. Seit dem 2026-09-03 liegt über Grund und Fläche eine Textur, und
// damit hat der Hintergrund keine eine Farbe mehr: er hat eine hellste und
// eine dunkelste Stelle, und die Zahl in der Tabelle gilt für keine von
// beiden.
//
// Dieser Lauf schließt die Lücke: er rastert die Kachel, legt sie über jede
// Grundfarbe und rechnet den Kontrast am **ungünstigsten** Bildpunkt.
//
// ```bash
// node scripts/pergamentmass.mjs
// ```
//
// # Wie gemessen wird
//
// Die Kachel steht als `data:`-SVG in `frontend/src/stil.css`. Der Lauf liest
// sie **von dort**, nicht aus einer Kopie hier — sonst misst er eine Textur,
// die die Anwendung nicht benutzt. Chrome rastert sie, ein Canvas legt sie
// über die Grundfarbe, und `getImageData` gibt jeden Bildpunkt zurück.
//
// # Was er **nicht** misst
//
//   - **Nicht, ob die Textur schön ist.** Das entscheidet ein Auge.
//   - **Nicht Text auf einem Bild**, und nicht die Kartenvorschau: sie trägt
//     die Textur nicht, denn sie zeigt, was gedruckt würde.
//   - **Nicht eine Vignette am Fensterrand** — es gibt keine. Sie stand im
//     Entwurf und ist an dieser Messung gescheitert: der dunkelste Punkt der
//     Kachel schöpft im Hellen das Budget bis auf 0,37 aus (`kante` auf
//     `grund`: 3,37 bei einer Schranke von 3), und ein Verlauf darüber nähme
//     genau den Rest. Ein Fensterrand ist ohnehin kein Blattrand: links liegt
//     die Seitenleiste.

import { spawn } from "node:child_process";
import { readFile, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { verbinden, warte, werte } from "./abzuegecdp.mjs";

// vordergruende sind die Farben, die auf einer Fläche stehen.
const vordergruende = {
  text: "--mse-text",
  leise: "--mse-leise",
  kante: "--mse-kante",
  marke: "--mse-marke",
  ring: "--mse-ring",
  marketext: "--mse-marketext",
};

// hintergruende sind die Flächen — und ob sie die Faser tragen.
//
// **`betont` und `marke` tragen sie nicht**: die Regel in `stil.css` hängt das
// Bild an `.bg-grund` und `.bg-flaeche`, sonst nirgends. Wer das ändert, ändert
// diese Tabelle mit — sonst misst der Lauf eine Fläche ohne die Textur, die
// darauf liegt.
const hintergruende = {
  grund: { merkmal: "--mse-grund", faser: true },
  flaeche: { merkmal: "--mse-flaeche", faser: true },
  betont: { merkmal: "--mse-betont", faser: false },
  marke: { merkmal: "--mse-marke", faser: false },
};

// sinnvoll sagt, welche Klassen die Anwendung auf welchem Grund wirklich malt.
//
// **Ohne sie stünden 48 Zeilen da, von denen die Hälfte niemand benutzt** —
// und ein Wert unter der Schranke für ein Paar, das es nicht gibt, ist kein
// Befund, sondern Rauschen. Die Liste kommt aus dem Quelltext: gezählt am
// 2026-09-03 über `grep` nach den Klassenpaaren.
//
// **Die Klasse steht dabei, nicht nur die Farbe.** `docs/KONTRASTE.md` führt
// Klassen, und der Wächter daneben liest sie — `divide-kante` und
// `border-kante` sind dieselbe Farbe und zwei Zeilen.
//
// **`outline-ring` auf `bg-marke` fehlt mit Absicht.** Das Verhältnis wäre
// 1,00 — Ring und Marke sind dieselbe Farbe. Der Ring liegt dort aber nie:
// der Hauptknopf ist die einzige Fläche in Markenfarbe, und er trägt
// `outline-offset-2`, also liegt sein Ring auf dem Grund dahinter (5,87 hell
// / 7,12 dunkel). Nachgezählt am 2026-09-03: die Stellen **ohne** Offset sind
// Eingabefelder auf `bg-flaeche`.
const sinnvoll = {
  grund: [
    ["text-text", "text"],
    ["text-leise", "leise"],
    ["border-kante", "kante"],
    ["divide-kante", "kante"],
    ["text-marke", "marke"],
    ["border-marke", "marke"],
    ["outline-ring", "ring"],
  ],
  flaeche: [
    ["text-text", "text"],
    ["text-leise", "leise"],
    ["border-kante", "kante"],
    ["divide-kante", "kante"],
    ["text-marke", "marke"],
    ["border-marke", "marke"],
    ["outline-ring", "ring"],
  ],
  betont: [
    ["text-text", "text"],
    ["text-leise", "leise"],
    ["border-kante", "kante"],
    ["outline-ring", "ring"],
  ],
  marke: [["text-marketext", "marketext"]],
};

// leseStil zieht die Merkmale und die beiden Kacheln aus `stil.css`.
//
// **Aus der Datei und nicht aus einer Kopie** — eine Textur, die hier
// nachgebaut steht, misst sich selbst.
async function leseStil() {
  const quelle = await readFile("frontend/src/stil.css", "utf8");
  const bloecke = {};

  // **Der Block ist der mit `--mse-grund` darin, nicht der erste.** Beide
  // Selektoren kommen in `stil.css` mehrfach vor — `html[…='dunkel']` steht
  // auch über einem `color-scheme: dark;`. Der erste Anlauf nahm ihn, maß acht
  // Paare gegen Schwarz und meldete überall 1,00. Eine Messung, die den
  // falschen Gegenstand erwischt, sieht wie ein Befund aus.
  for (const [name, muster] of [
    ["hell", /:root\s*\{([^}]*)\}/g],
    ["dunkel", /html\[data-erscheinungsbild='dunkel'\]\s*\{([^}]*)\}/g],
  ]) {
    const treffer = [...quelle.matchAll(muster)]
      .map((m) =>
        Object.fromEntries(
          [...m[1].matchAll(/(--mse-[a-z]+):\s*([^;]+);/g)].map((e) => [
            e[1],
            e[2].trim(),
          ]),
        ),
      )
      .find((werte) => "--mse-grund" in werte);

    if (!treffer) throw new Error(`kein Farbblock für ${name} in stil.css`);

    bloecke[name] = treffer;
  }

  // **Die Kachel kommt aus der Datei, die `stil.css` nennt** — nicht aus
  // einem Pfad, der hier steht. Wer den Verweis umhängt und diese Messung
  // vergisst, misst sonst weiter die alte Textur.
  const kacheln = {};

  for (const [name, merkmal] of [
    ["hell", "--mse-faser-hell"],
    ["dunkel", "--mse-faser-dunkel"],
  ]) {
    const treffer = quelle.match(
      new RegExp(`${merkmal}:\\s*url\\('([^']+)'\\)`),
    );

    if (!treffer) throw new Error(`keine Kachel ${merkmal} in stil.css`);

    const svg = await readFile(
      join("frontend/src", treffer[1].replace(/^\.\//, "")),
      "utf8",
    );

    kacheln[name] = `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
  }

  return { bloecke, kacheln };
}

// messung läuft in der Seite: Kachel über Grundfarbe, dann jeder Bildpunkt.
const messung = (kachel, farbe) => `(async () => {
  const bild = new Image();
  bild.src = ${JSON.stringify(kachel)};
  await bild.decode();

  const flaeche = document.createElement('canvas');
  flaeche.width = bild.naturalWidth;
  flaeche.height = bild.naturalHeight;

  const stift = flaeche.getContext('2d', { willReadFrequently: true });
  stift.fillStyle = ${JSON.stringify(farbe)};
  stift.fillRect(0, 0, flaeche.width, flaeche.height);
  stift.drawImage(bild, 0, 0);

  const punkte = stift.getImageData(0, 0, flaeche.width, flaeche.height).data;
  const kanal = (v) => {
    const s = v / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };

  let dunkelste = 2;
  let hellste = -1;
  let summe = 0;

  for (let i = 0; i < punkte.length; i += 4) {
    const l =
      0.2126 * kanal(punkte[i]) +
      0.7152 * kanal(punkte[i + 1]) +
      0.0722 * kanal(punkte[i + 2]);

    if (l < dunkelste) dunkelste = l;
    if (l > hellste) hellste = l;
    summe += l;
  }

  return { dunkelste, hellste, mittel: summe / (punkte.length / 4) };
})()`;

// leuchtdichte einer Farbangabe, gerechnet über denselben Canvas.
const leuchtdichte = (farbe) => `(() => {
  const flaeche = document.createElement('canvas');
  flaeche.width = 1;
  flaeche.height = 1;

  const stift = flaeche.getContext('2d');
  stift.fillStyle = ${JSON.stringify(farbe)};
  stift.fillRect(0, 0, 1, 1);

  const [r, g, b] = stift.getImageData(0, 0, 1, 1).data;
  const kanal = (v) => {
    const s = v / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };

  return 0.2126 * kanal(r) + 0.7152 * kanal(g) + 0.0722 * kanal(b);
})()`;

const verhaeltnis = (a, b) =>
  (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);

async function lauf() {
  const { bloecke, kacheln } = await leseStil();
  const profil = await mkdtemp(join(tmpdir(), "pergament-"));
  const chrome = spawn(
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    [
      "--headless",
      "--disable-gpu",
      "--remote-debugging-port=0",
      `--user-data-dir=${profil}`,
      "about:blank",
    ],
    { stdio: "ignore" },
  );

  try {
    const draht = await verbinden(profil);
    const zeilen = [];

    for (const bild of ["hell", "dunkel"]) {
      for (const [ruecken, art] of Object.entries(hintergruende)) {
        const farbe = bloecke[bild][art.merkmal];
        const spanne = art.faser
          ? await werte(draht, messung(kacheln[bild], farbe))
          : null;

        for (const [klasse, vorn] of sinnvoll[ruecken]) {
          const vmerkmal = vordergruende[vorn];
          const l = await werte(draht, leuchtdichte(bloecke[bild][vmerkmal]));
          const ohne = await werte(draht, leuchtdichte(farbe));

          zeilen.push({
            bild,
            ruecken,
            klasse,
            vorn,
            ohneTextur: verhaeltnis(l, ohne),
            schlechtester: spanne
              ? Math.min(
                  verhaeltnis(l, spanne.dunkelste),
                  verhaeltnis(l, spanne.hellste),
                )
              : verhaeltnis(l, ohne),
            spanne,
          });
        }
      }
    }

    for (const z of zeilen) {
      console.log(
        `${z.bild.padEnd(7)} ${z.klasse.padEnd(14)} auf ${z.ruecken.padEnd(8)}` +
          ` ohne ${z.ohneTextur.toFixed(2).padStart(6)}` +
          ` mit ${z.schlechtester.toFixed(2).padStart(6)}` +
          ` (Δ ${(z.schlechtester - z.ohneTextur).toFixed(2).padStart(6)})`,
      );
    }

    // **Ein Rand hält 3, ein Text 4,5** — WCAG 1.4.11 gegen 1.4.3.
    const rand = (v) => v === "kante" || v === "ring";
    const unter = zeilen.filter(
      (z) => !rand(z.vorn) && z.schlechtester < 4.5,
    );
    const unterRand = zeilen.filter(
      (z) => rand(z.vorn) && z.schlechtester < 3,
    );

    console.log(
      `\n${zeilen.length} Paare, ${unter.length + unterRand.length} unter der Schranke`,
    );

    console.log("\n### Zeilen für docs/KONTRASTE.md\n");

    for (const z of zeilen) {
      console.log(
        `| \`${z.klasse}\` | \`bg-${z.ruecken}\` | ` +
          `${z.schlechtester.toFixed(2).replace(".", ",")} | ${z.bild} |`,
      );
    }

    await writeFile(
      "/tmp/pergamentmass.json",
      JSON.stringify(zeilen, null, 2),
      "utf8",
    );

    process.exitCode = unter.length + unterRand.length > 0 ? 1 : 0;
  } finally {
    chrome.kill();
    await warte(200);
    await rm(profil, { recursive: true, force: true });
  }
}

await lauf();
