// **Was fünftausend Exemplare im Fenster kosten** — Werkstück 1 von Phase 4.5
// (#1292).
//
// # Wozu
//
// Die Abnahme von #1292 verlangt, dass die Sammlungsseite bei mehreren tausend
// Exemplaren **bedienbar bleibt** — mit einer benannten Schranke für Aufbau
// und Suche.
//
// `internal/app/messsammlung_bench_intern_test.go` misst den Go-Weg: 4,5 ms
// für die ganze Liste. Was er nicht sieht, ist die Strecke dahinter, und die
// ist bei diesem Projekt schon einmal die teure gewesen: bei
// #1270 hielt
// der Go-Weg 53 ms, und im Fenster wartete man 0,4 bis 1,0 Sekunden — eine
// 242-KB-Antwort durch die JSON-Naht kostete 550 ms.
//
// **Fünftausend Exemplare sind 2,1 MB.** Das ist das Neunfache.
//
// # Wie er gefahren wird
//
// Er braucht ein laufendes `make dev-browser` **gegen einen eigenen Bestand** — nicht
// gegen den der nutzenden Person. Verlegt wird das Zuhause — der Bestand liegt
// darunter am Vorgabeort, und mit ihm Pakete und Bilder
// (`internal/app/bestandsort.go`):
//
// ```bash
// HOME=/tmp/mse45heim make dev          # in einem eigenen Fenster
// node scripts/sammlungsmass.mjs
// ```
//
// Der Bestand darunter wird von `scripts/sammlungssaat.mjs` gelegt.
//
// # Was er **nicht** misst
//
//   - **Nicht, ob die Seite gut aussieht.** Dafür gibt es die Sichtprüfung.
//   - **Nicht den Bestand dieses Rechners.** Er misst den, der unter dem
//     verlegten Zuhause liegt — und sagt zu Beginn, wie viele Zeilen das sind.
//   - **Nicht die Zeit auf einer anderen Maschine.** Eine Dauer ist eine
//     Aussage über diesen Rechner (0007).

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

const breite = 1280;
const hoehe = 800;

// Die Schranken — gemessen am 2026-08-30 auf einem Apple M5, 1280 × 800,
// gegen 5.000 Exemplare, je drei Läufe:
//
//	Aufbau der Seite   321–358 ms → 600 ms   (Faktor 1,7)
//	Suche, 4 Tasten      67– 69 ms → 120 ms   (Faktor 1,7)
//
// **Beide fangen den Rückschritt, gegen den es sie gibt.** Ohne die Grenze aus
// `sammlungssicht.ts` steht dieselbe Seite bei **791 ms** und **150 ms** —
// nachgestellt am selben Tag, mit demselben Messgerät und demselben Bestand.
//
// **Sie gelten für diese Maschine** (0007). Eine langsamere reißt sie, ohne
// dass am Code etwas falsch wäre.
const schrankeAufbau = 600;
const schrankeSuche = 120;

// saat ist die Zahl der Exemplare, für die gemessen wird.
//
// **Sie wird geprüft, nicht angenommen** (Befund 8): der Lauf fragt die
// laufende Bindung, wie viele wirklich im Bestand liegen, und bricht ab, wenn
// es andere sind. `scripts/sammlungssaat.mjs` legt sie.
const saat = Number(process.argv[3] ?? 5000);

// grenze ist die Zahl der Zeilen, die die Seite zeichnet.
//
// **Dieselbe wie `zeilengrenze` in `frontend/src/seiten/sammlungssicht.ts`.**
// Sie hier zu prüfen fängt eine Seite, die plötzlich alles zeichnet — langsam
// und ohne dass die Zeilenzahl es verriete.
const grenze = Number(process.argv[4] ?? 200);

// stehtDieAnwendung prüft, ob unter der Adresse überhaupt etwas antwortet.
async function stehtDieAnwendung(ws) {
  const bruecke = await werte(
    ws,
    `typeof window.go === 'object' && typeof window.go.app === 'object'`,
  );

  if (bruecke !== true) {
    throw new Error(
      `unter ${adresse} antwortet keine Anwendung — läuft \`make dev-browser\`?`,
    );
  }
}

// ruheImFenster ist der Ausdruck, der **im Fenster** auf Ruhe wartet.
//
// # Warum nicht die Schleife aus `serienmass.mjs`
//
// Weil sie alle 500 ms nachsieht und zwei gleiche Zahlen verlangt. Für eine
// Zählung ist das gleichgültig — für eine **Zeitmessung** ist es der ganze
// Messwert: der Aufbau kam damit auf 1030 ms, gleich ob 5.000 Zeilen im DOM
// standen oder 200. Das war die Auflösung des Messgeräts, nicht die Seite.
//
// **Deshalb wird im Fenster gewartet und nicht über die Leitung.** Ein
// `Runtime.evaluate` je Blick zählte den Weg durch WebSocket und JSON mit;
// diese Schleife läuft in der Seite, sieht alle 16 ms nach und verlangt drei
// gleiche Zahlen — 48 ms Auflösung statt 1000.
//
// **Und sie hat eine Frist.** Ohne die hinge der Lauf an einer Seite, die
// nicht zur Ruhe kommt, statt es zu sagen.
// **Und sie wartet erst darauf, dass überhaupt etwas dasteht.** Ohne diese
// Bedingung war der erste Anlauf nach drei Bildern fertig — die Zahl der
// Knoten steht still, solange die Antwort noch unterwegs ist, und der Lauf maß
// eine **leere** Seite. Er hat es selbst gesagt („es stehen nur 0 Zeilen da"),
// weil die Wächterzeile darunter zählt; ohne sie hätte er eine plausible Zahl
// ausgegeben.
const ruheImFenster = `(async () => {
  const zaehlen = () => document.querySelectorAll('*').length;
  const zeilen = () => document.querySelectorAll('main ul > li').length;
  const schlaf = () => new Promise((f) => requestAnimationFrame(() => f()));

  let gleich = 0;
  let vorher = zaehlen();

  for (let blick = 0; blick < 1200; blick += 1) {
    await schlaf();

    const jetzt = zaehlen();
    if (jetzt === vorher && zeilen() > 0) {
      gleich += 1;
      if (gleich >= 3) return true;
    } else {
      gleich = 0;
      vorher = jetzt;
    }
  }

  return false;
})()`;

// zurSammlung klickt die Sammlungsseite auf und stoppt, wie lange es dauert.
//
// **Gestoppt wird im Fenster und nicht hier** (`performance.now()`): eine
// Messung über die CDP-Leitung zählte den Weg durch WebSocket und JSON mit.
async function zurSammlung(ws) {
  const gestartet = await werte(
    ws,
    `(() => {
      const e = [...document.querySelectorAll('button')]
        .find((b) => b.textContent.trim().startsWith('Sammlung'));
      if (!e) return false;
      window.__massBeginn = performance.now();
      e.click();
      return true;
    })()`,
  );

  if (gestartet !== true) {
    throw new Error("kein Knopf, der mit Sammlung beginnt");
  }

  return await bisRuhe(ws);
}

// zeilenZaehlen sagt, wie viele Exemplare wirklich dastehen.
//
// **`main ul > li` und keine Tabellenzeile.** Die Sammlung ist eine Liste aus
// `<li>`, nicht ein `<table>` — ein Ausdruck über `tbody tr` gäbe hier
// verlässlich **null** und liefe in die Wächterzeile darunter.
//
// **Das ist die Zeile, ohne die der Rest nichts wert wäre.** Ein Aufbau über
// eine leere Seite ist schnell; die Reichweite einer Messung ist die
// Reichweite ihrer Aussage (§10).
async function zeilenZaehlen(ws) {
  return Number(
    await werte(
      ws,
      `document.querySelectorAll('main ul > li').length`,
    ),
  );
}

// tippen schreibt Text in das Suchfeld — Taste für Taste, wie ein Mensch.
//
// # Das Suchfeld ist `input[type="search"]` und nichts sonst
//
// Der erste Anlauf nahm `input[type="search"], input[type="text"]` und traf
// damit das **Aufnahmeformular**, das über der Sicht steht: vier Tastendrücke
// in das Feld „Druckkennung", 5.000 Zeilen davor und 5.000 danach. Das sah aus
// wie eine kaputte Suche und war eine kaputte Messung (§10).
//
// **Über `Input.dispatchKeyEvent` und nicht über `value =`.** Eine Zuweisung
// löst kein `input`-Ereignis aus; React sähe sie nicht, und die Messung wäre
// die einer Seite, die nichts tut.
async function tippen(ws, text) {
  const getroffen = await werte(
    ws,
    `(() => {
      const feld = document.querySelector('input[type="search"]');
      if (!feld) return false;
      feld.focus();
      return true;
    })()`,
  );

  if (getroffen !== true) {
    throw new Error("kein Suchfeld auf der Sammlungsseite");
  }

  await werte(ws, `window.__massBeginn = performance.now()`);

  for (const zeichen of text) {
    await schicke(ws, "Input.dispatchKeyEvent", {
      type: "keyDown",
      text: zeichen,
      key: zeichen,
    });
    await schicke(ws, "Input.dispatchKeyEvent", { type: "keyUp", key: zeichen });
  }

  return await bisRuhe(ws);
}

// bisRuhe wartet im Fenster auf Ruhe und gibt die Dauer seit `__massBeginn`.
//
// **Die drei gleichen Blicke werden abgezogen.** Sie sind die Frist des
// Messgeräts und nicht die Arbeit der Seite; bei 60 Bildern je Sekunde sind
// das 48 ms, und ohne den Abzug stünden sie in jeder Zahl.
async function bisRuhe(ws) {
  const ruhig = await werte(ws, ruheImFenster);

  if (ruhig !== true) {
    throw new Error("die Seite kam in zwanzig Sekunden nicht zur Ruhe");
  }

  return Number(
    await werte(
      ws,
      `Math.max(0, Math.round(performance.now() - window.__massBeginn - 48))`,
    ),
  );
}

const profil = await mkdtemp(join(tmpdir(), "sammlungsmass-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  [
    "--headless=new",
    "--remote-debugging-port=0",
    `--user-data-dir=${profil}`,
    "about:blank",
  ],
  { stdio: ["ignore", "ignore", "ignore"] },
);

try {
  const ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: breite,
    height: hoehe,
    deviceScaleFactor: 1,
    mobile: false,
  });

  await schicke(ws, "Page.navigate", { url: adresse });
  await warte(3500);
  await stehtDieAnwendung(ws);

  // **Erst den Bestand prüfen, dann messen.**
  //
  // Eine Zählung der **DOM-Zeilen** genügte nicht: von denen werden
  // absichtlich 200 gezeichnet, und ein Bestand mit 100 Exemplaren erfüllte
  // den Wächter und meldete die für 5.000 behaupteten Schranken **grün**. Die
  // Zahl, um die es geht, steht hinter der Bindung.
  const imBestand = Number(
    await werte(
      ws,
      `(async () => (await window.go.app.Bindungen.Sammlung()).exemplare.length)()`,
    ),
  );

  if (imBestand !== saat) {
    throw new Error(
      `im Bestand liegen ${String(imBestand)} Exemplare, gemessen wird für ` +
        `${String(saat)} — erst \`node scripts/sammlungssaat.mjs\` fahren und ` +
        `\`make dev-browser\` mit demselben HOME starten`,
    );
  }

  const aufbau = await zurSammlung(ws);
  const zeilen = await zeilenZaehlen(ws);

  // **Und die DOM-Zahl ist die Zeichengrenze**, nicht der Bestand. Sie hier
  // zu prüfen ist die zweite Hälfte von Befund 8: eine Seite, die plötzlich
  // alles zeichnete, wäre langsam **und** unauffällig.
  if (zeilen !== grenze) {
    throw new Error(
      `${String(zeilen)} Zeilen im DOM, erwartet ${String(grenze)} — die ` +
        `Zeichengrenze aus sammlungssicht.ts greift nicht`,
    );
  }

  const suche = await tippen(ws, "Bolt");
  const nachSuche = await zeilenZaehlen(ws);

  // **Die Schranken werden gehalten und nicht nur berichtet.** Ein Messlauf,
  // der nur ausgibt, fängt keinen Rückschritt: seine Zahlen liest niemand
  // zweimal (dieselbe Regel wie bei `messschwellen_test.go`).
  const gerissen = [];

  if (aufbau > schrankeAufbau) {
    gerissen.push(
      `der Aufbau braucht ${String(aufbau)} ms und reißt die Schranke von ` +
        `${String(schrankeAufbau)} ms`,
    );
  }

  if (suche > schrankeSuche) {
    gerissen.push(
      `die Suche braucht ${String(suche)} ms und reißt die Schranke von ` +
        `${String(schrankeSuche)} ms`,
    );
  }

  process.stdout.write(
    `| Messung | Wert |\n| --- | ---: |\n` +
      `| Zeilen im DOM | ${String(zeilen)} |\n` +
      `| Aufbau der Seite | ${String(aufbau)} ms |\n` +
      `| Suche nach »Bolt« (4 Tastendrücke) | ${String(suche)} ms |\n` +
      `| Zeilen danach | ${String(nachSuche)} |\n`,
  );

  if (gerissen.length > 0) {
    throw new Error(
      `${gerissen.join("; ")} — entweder ist etwas langsamer geworden, oder ` +
        `diese Maschine ist eine andere als die, an der gemessen wurde`,
    );
  }
} finally {
  chrome.kill();
  await warte(200);
  await rm(profil, { recursive: true, force: true });
}
