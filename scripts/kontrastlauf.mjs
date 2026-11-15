// **Kontraste am gerenderten Element, in beiden Erscheinungsbildern**
// (#1290).
//
// # Wozu, wo es doch `docs/KONTRASTE.md` gibt
//
// Weil dort **Paare** stehen — „`text-slate-600` auf `bg-white`: 7,58" — und
// ein Wächter hält die Datei gegen den Quelltext. Was **keiner** von beiden
// sagt: ob dieses Paar auf der Seite auch wirklich so übereinander liegt. Eine
// Klasse steht im Quelltext; was ein Auge sieht, entscheidet der Baum darüber.
//
// §3 sagt es: *„Unter jsdom meldet axe `color-contrast` als unentschieden. Die
// Sichtprüfung im Browser findet, was die Tests durchlassen."* Dieser Lauf ist
// die maschinelle Hälfte davon.
//
// ```bash
// node scripts/kontrastlauf.mjs
// ```
//
// # Wie gemessen wird
//
// **Über einen Canvas, nicht über ein Textmuster** — dieselbe Begründung wie
// in `docs/KONTRASTE.md`: Tailwind 4 gibt `oklch()`, und eine Regex, die
// Ziffern aus einer Farbangabe zieht, liest daraus Unsinn. Wer die Farbe in
// eine 1×1-Fläche malt und zurückliest, bekommt sRGB.
//
// **Der Hintergrund kommt vom nächsten Vorfahren, der einen hat.** Ein
// `<span>` ist fast immer durchsichtig; was darunter liegt, entscheidet.
//
// # Was er **nicht** misst
//
//   - **Nicht, ob die Farbe die richtige ist.** Er misst, was dasteht.
//   - **Nicht Ränder ohne Fläche.** WCAG 1.4.11 gilt auch für sie, und dafür
//     hat weder axe noch dieser Lauf eine Regel — `docs/KONTRASTE.md` führt
//     sie von Hand.
//   - **Nicht Text auf einem Bild.** Ein Kartenbild hat keine eine Farbe.
//   - **Nicht die Kartenseite.** Sie braucht ein geöffnetes Set und eine
//     aufgeschlagene Karte; der Lauf geht die Seiten, die die Seitenleiste
//     führt. Das ist eine Lücke und keine Aussage über sie — angestrichen vom
//     Zweitblick über Phase 4.3, Befund 4.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

// schranke ist, was WCAG 2.2 AA für gewöhnlichen Text verlangt.
//
// **4,5 und nicht 3.** Die 3 gilt für großen Text ab 18,66 px fett oder 24 px;
// dieser Lauf misst die Größe mit und wendet sie an.
const schranke = 4.5;
const schrankeGross = 3;

// messung läuft in der Seite und gibt jede Textstelle mit ihrem Kontrast.
const messung = `(() => {
  const malflaeche = document.createElement('canvas');
  malflaeche.width = 1;
  malflaeche.height = 1;
  const stift = malflaeche.getContext('2d', { willReadFrequently: true });

  // **Über den Canvas und nicht über ein Textmuster**: Tailwind 4 gibt
  // \`oklch()\`, und wer daraus Ziffern zieht, liest Unsinn.
  // **Mit Alphakanal** (Befund 4 des Zweitblicks über Phase 4.3). Der erste
  // Anlauf warf ihn weg — ein halbdurchsichtiger Grund galt damit als
  // undurchsichtig, und was darunter lag, ging nicht ein.
  const alsRGBA = (farbe) => {
    stift.clearRect(0, 0, 1, 1);
    stift.fillStyle = farbe;
    stift.fillRect(0, 0, 1, 1);
    const [r, g, b, a] = stift.getImageData(0, 0, 1, 1).data;
    return [r, g, b, a / 255];
  };

  // ueber legt eine Farbe über eine andere — die Rechnung, die ein Browser
  // beim Zeichnen macht.
  const ueber = ([r, g, b, a], [gr, gg, gb]) => [
    r * a + gr * (1 - a),
    g * a + gg * (1 - a),
    b * a + gb * (1 - a),
  ];

  const kanal = (wert) => {
    const anteil = wert / 255;
    return anteil <= 0.03928 ? anteil / 12.92 : Math.pow((anteil + 0.055) / 1.055, 2.4);
  };

  const helligkeit = ([r, g, b]) =>
    0.2126 * kanal(r) + 0.7152 * kanal(g) + 0.0722 * kanal(b);

  const kontrast = (vorne, hinten) => {
    const a = helligkeit(vorne);
    const b = helligkeit(hinten);
    const hell = Math.max(a, b);
    const dunkel = Math.min(a, b);
    return (hell + 0.05) / (dunkel + 0.05);
  };

  // **Der Grund wird geschichtet, nicht gesucht** (Befund 4).
  //
  // Ein \`span\` ist fast immer durchsichtig; was darunter liegt, entscheidet.
  // Aber auch der erste Vorfahre **mit** einer Farbe kann halbdurchsichtig
  // sein — dann zählt, was unter ihm liegt, anteilig mit. Der erste Anlauf
  // nahm die erste Farbe, die nicht ganz durchsichtig war, und war damit bei
  // jedem \`bg-white/80\` daneben.
  const grundVon = (knoten) => {
    const schichten = [];

    // Beim Knoten selbst beginnen, nicht beim Vater: ein Hauptknopf traegt
    // seine Flaeche selbst, und wer beim Vater anfaengt, misst weisse Schrift
    // gegen den weissen Grund dahinter — 1,05:1. Der Lauf hat diesen Fehler in
    // sich selbst gefunden.
    for (let auf = knoten; auf !== null; auf = auf.parentElement) {
      const farbe = alsRGBA(getComputedStyle(auf).backgroundColor);
      if (farbe[3] === 0) continue;

      schichten.push(farbe);
      if (farbe[3] === 1) break;
    }

    // **Von unten nach oben.** Weiß ist der letzte Grund, den ein Fenster hat.
    let grund = [255, 255, 255];
    for (const schicht of schichten.reverse()) {
      grund = ueber(schicht, grund);
    }

    return grund;
  };

  const befunde = [];

  for (const knoten of document.querySelectorAll('*')) {
    // **Nur Knoten mit eigenem Text.** Ein \`div\` um zehn Absätze hat den Text
    // seiner Kinder und nicht seinen eigenen.
    const eigener = [...knoten.childNodes]
      .filter((k) => k.nodeType === Node.TEXT_NODE)
      .map((k) => k.textContent.trim())
      .join('');

    if (eigener === '') continue;

    const kasten = knoten.getBoundingClientRect();
    if (kasten.width === 0 || kasten.height === 0) continue;

    const stil = getComputedStyle(knoten);
    if (stil.visibility === 'hidden') continue;

    // **Abgeschaltete Bedienelemente sind ausgenommen**, und das steht so in
    // der Norm: WCAG 2.2, 1.4.3 „Contrast (Minimum)", Ausnahme *Incidental* —
    // „Text or images of text that are part of an inactive user interface
    // component ... have no contrast requirement."
    //
    // Der Lauf fand hier zuerst „Set anlegen" mit 3,08:1 und meldete es. Der
    // Knopf ist abgeschaltet, solange kein Name dasteht; sein blasser Zustand
    // **ist** die Auskunft.
    if (knoten.closest('[disabled], [aria-disabled="true"]') !== null) continue;

    // **Durchsichtigkeit zählt anteilig** (Befund 4). Der erste Anlauf sprang
    // nur bei genau null ab — ein Text bei \`opacity: 0.5\` ging als voll
    // gedeckt durch.
    const sicht = [...(() => {
      const werte = [];
      for (let auf = knoten; auf !== null; auf = auf.parentElement) {
        werte.push(parseFloat(getComputedStyle(auf).opacity));
      }
      return werte;
    })()].reduce((a, b) => a * b, 1);

    if (sicht === 0) continue;

    const groesse = parseFloat(stil.fontSize);
    const fett = parseInt(stil.fontWeight, 10) >= 700;
    const gross = groesse >= 24 || (fett && groesse >= 18.66);

    const grund = grundVon(knoten);
    const farbe = alsRGBA(stil.color);

    // **Die Schriftfarbe wird auf ihren Grund gelegt** — mit ihrem eigenen
    // Alpha und der geerbten Durchsichtigkeit.
    const vorne = ueber([farbe[0], farbe[1], farbe[2], farbe[3] * sicht], grund);

    befunde.push({
      text: eigener.slice(0, 40),
      kontrast: Math.round(kontrast(vorne, grund) * 100) / 100,
      gross,
    });
  }

  return JSON.stringify(befunde);
})()`;

// seiten sind die Wege, die gemessen werden.
//
// **Die Liste muss mit den Seiten mitwachsen.** §3 sagt es für die
// Barrierefreiheitsprüfung, und es gilt hier genauso: **eine zu kurze
// Aufzählung ist grün, und niemand sieht es.**
//
// **Warum sie trotzdem von Hand steht und nicht aus `navigation.ts` kommt**:
// dieser Lauf ist ein Node-Skript neben der Anwendung, kein Test in ihr — es
// gibt keinen Weg, das Modul zu laden, ohne den Bau nachzustellen. Der Preis
// ist diese Liste; sie wird beim Phasenabschluss durchgesehen, so wie die
// Abzugsliste daneben.
//
// # Was er nicht misst
//
// **Die dreizehn Setlaschen.** Sie brauchen ein offenes Set, und der Weg
// dorthin ist ein anderer als ein Klick in der Leiste. Der Serienabzug fährt
// sie (`scripts/serienabzug.mjs`); die Kontraste dort sind ungemessen und
// stehen unter der Sichtprüfung von Hand.
const seiten = [
  { name: "Start", weg: [] },
  { name: "Sets", weg: ["Sets"] },
  { name: "Pakete", weg: ["Pakete"] },
  { name: "Vorlagen", weg: ["Vorlagen"] },
  { name: "Sammlung", weg: ["Sammlung"] },
  { name: "Decks", weg: ["Decks"] },
  { name: "Turniere", weg: ["Turniere"] },
  { name: "Nutzer und Profile", weg: ["Nutzer und Profile"] },
  { name: "Einstellungen", weg: ["Einstellungen"] },
  { name: "Hilfe", weg: ["Hilfe"] },
];

// hinNavigieren klickt einen Weg über die Beschriftungen der Seitenleiste.
//
// **Sie bricht ab, wenn ein Knopf fehlt** — Befund 3 des Zweitblicks über
// Phase 4.3. Der erste Anlauf warf das `false` weg und wartete stur weiter;
// eine fehlgeschlagene Navigation gab damit plausible Zahlen **für die falsche
// Seite**. Ein Messgerät, das bei einem Fehler weitermisst, ist schlimmer als
// keines.
async function hinNavigieren(ws, weg) {
  for (const beschriftung of weg) {
    const getroffen = await werte(
      ws,
      `(() => {
        const e = [...document.querySelectorAll('button')]
          .find((b) => b.textContent.trim().startsWith(${JSON.stringify(beschriftung)}));
        if (!e) return false;
        e.click();
        return true;
      })()`,
    );

    if (getroffen !== true) {
      throw new Error(`kein Knopf, der mit ${beschriftung} beginnt`);
    }

    await warte(1500);
  }
}

// stehtDieAnwendung prüft, ob unter der Adresse überhaupt etwas antwortet.
//
// **Weil dieser Lauf keinen Server startet** (Befund 3): er braucht ein
// laufendes `make dev-browser`. Ohne diese Prüfung misst er eine Fehlerseite und gibt
// Zahlen aus, die wie welche aussehen.
async function stehtDieAnwendung(ws) {
  const bruecke = await werte(
    ws,
    `typeof window.go === 'object' && typeof window.go.app === 'object'`,
  );

  if (bruecke !== true) {
    throw new Error(
      `unter ${adresse} antwortet keine Anwendung — l\u00e4uft \`make dev-browser\`?`,
    );
  }
}

// bisRuhe wartet, bis die Seite fertig geladen ist.
//
// **Weil eine halbe Seite plausible Zahlen gibt** (Befund 3 des Zweitblicks
// über Phase 4.3). Die Paketseite zählte je nach Lauf 51 oder 1.833
// Textstellen — dieselbe Seite, einmal vor und einmal nach dem Laden ihrer
// Liste. Beide Läufe waren grün, und einer davon sagte nichts.
//
// **Zweimal dieselbe Zahl heißt fertig.** Ein fester Wartewert müsste raten,
// wie lange — und ein Test darüber prüfte den Rechner.
async function bisRuhe(ws) {
  const zaehlen = async () =>
    Number(await werte(ws, `document.querySelectorAll('*').length`));

  let vorher = await zaehlen();

  for (let versuch = 0; versuch < 20; versuch += 1) {
    await warte(500);

    const jetzt = await zaehlen();
    if (jetzt === vorher) {
      return;
    }

    vorher = jetzt;
  }

  throw new Error("die Seite kam in zehn Sekunden nicht zur Ruhe");
}

const profil = await mkdtemp(join(tmpdir(), "kontrastlauf-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

let zuSchwach = 0;
let gemessen = 0;

try {
  const ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: 1280,
    height: 800,
    deviceScaleFactor: 1,
    mobile: false,
  });

  for (const schema of ["light", "dark"]) {
    await schicke(ws, "Emulation.setEmulatedMedia", {
      features: [{ name: "prefers-color-scheme", value: schema }],
    });

    process.stdout.write(`\n── ${schema === "light" ? "hell" : "dunkel"}\n`);

    for (const seite of seiten) {
      await schicke(ws, "Page.navigate", { url: adresse });
      await warte(3500);

      if (schema === "light" && seite.weg.length === 0) {
        await stehtDieAnwendung(ws);
      }

      await hinNavigieren(ws, seite.weg);
      await bisRuhe(ws);

      const befunde = JSON.parse(await werte(ws, messung));
      const schwach = befunde.filter(
        (b) => b.kontrast < (b.gross ? schrankeGross : schranke),
      );

      gemessen += befunde.length;
      zuSchwach += schwach.length;

      if (schwach.length === 0) {
        process.stdout.write(`  ✓ ${seite.name} — ${String(befunde.length)} Textstellen\n`);

        continue;
      }

      process.stdout.write(
        `  ✗ ${seite.name} — ${String(schwach.length)} von ${String(befunde.length)} zu schwach\n`,
      );

      for (const befund of schwach.slice(0, 8)) {
        process.stdout.write(
          `      ${String(befund.kontrast)}:1${befund.gross ? " (groß)" : ""} — ${befund.text}\n`,
        );
      }
    }
  }
} finally {
  chrome.kill();
  await warte(200);
  await rm(profil, { recursive: true, force: true });
}

process.stdout.write(
  `\n${String(gemessen)} Textstellen gemessen, ${String(zuSchwach)} unter der Schranke\n`,
);

if (zuSchwach > 0) {
  process.exit(1);
}
