// **Was der erste Weg kostet — von der leeren Anwendung zur ersten Karte**
// (#1325,
// Werkstück D von Phase 4.10).
//
// # Wozu
//
// Die Abnahme der Phase verlangt eine Zahl: *„Der Weg von der leeren Anwendung
// zur ersten fertigen Karte ist gezählt und steht in `docs/SERIENARBEIT.md`."*
// Sie ist die Zahl, an der spätere Änderungen an der Oberfläche gemessen
// werden.
//
// `kernwegmass.mjs` daneben zählt den Weg von der **zweiten** Karte zur
// dritten — was jemand hundertmal tut. Dieser Lauf zählt, was jemand
// **einmal** tut, und das ist die schwierigere Zahl: beim ersten Mal weiß man
// nichts.
//
// ```bash
// node scripts/erstlaufmass.mjs
// ```
//
// # Was er voraussetzt — und was das für die Aussage heißt
//
// **Die Saat ist installiert.** Beim ersten Start bringt die Anwendung acht
// Vorlagenpakete und ein Beispielset mit; ohne sie führte der Weg zuerst über
// die Paketseite und einen Dateidialog, den ein Lauf ohne Fenster nicht
// bedienen kann. Der Zettel im `.dmg` und `README.md` sagen dasselbe.
//
// **Der Bestand darf schon Sets tragen.** Das ändert die Zahl nicht: gezählt
// wird der Weg, nicht die Liste, durch die er führt. Was es ändert, ist die
// Startseite — sie zeigt dann „Zuletzt benutzt" statt eines leeren Zustands.
// **Die Reichweite dieser Messung ist der Weg, nicht der erste Eindruck.**
//
// # Was eine Nutzerhandlung ist
//
// Dieselbe Zählung wie in `kernwegmass.mjs`: ein Klick oder ein Tastendruck,
// und Tippen zählt als eine, gleich wie lang der Text ist. Zwei Zahlen, die
// verschieden gezählt wären, ließen sich nicht nebeneinanderlegen.

import { spawn } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";
import { handgriffe } from "./kernweggriffe.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

// Der Setname trägt die Kennung dieses Laufs — derselbe Grund wie im
// Kernwegmaß: Setnamen sind im Bestand nicht eindeutig.
const setname = `Erstlauf ${String(process.pid)}`;
const kartenname = "Meine erste Karte";

// erwartet ist die Abnahme, gegen die dieser Lauf prüft.
//
// **Er weist ab und zeigt nicht nur an.** Ein Messlauf, der eine falsche Zahl
// druckt und mit 0 endet, ist eine Anzeige.
// **Sieben, gemessen am 2026-09-03** — nicht geschätzt. Der erste Anlauf
// erwartete sechs; die Zahl steht hier, weil sie **erhoben** wurde und nicht,
// weil sie schön wäre. Wer sie senkt, senkt sie sichtbar.
const erwartet = { handlungen: 7 };

const profil = await mkdtemp(join(tmpdir(), "erstlauf-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

const { griffe, gezaehlt } = handgriffe();
const { bis, bisRuhe, klick, tippe } = griffe;

// setzeFeld schreibt in ein Feld über sein Label — und zählt zwei Handlungen:
// hinzeigen und tippen.
async function setzeFeld(ws, beschriftung, wert) {
  const getroffen = await werte(
    ws,
    `(() => {
      const l = [...document.querySelectorAll('label')]
        .find((x) => x.textContent.trim().startsWith(${JSON.stringify(beschriftung)}));
      const f = l?.control ?? document.getElementById(l?.getAttribute('for') ?? '');
      if (!f) return false;
      f.focus();
      return true;
    })()`,
  );

  if (getroffen !== true) {
    throw new Error(`kein Feld ${beschriftung}`);
  }

  // Der Klick ins Feld ist die eine Handlung, das Tippen die zweite.
  await tippe(ws, "");
  await tippe(ws, wert);
}

const stationen = [];

// **Er räumt sein Set weg** — dieselbe Regel wie im Kernwegmaß: ein Messlauf,
// der schreibt, hinterlässt nichts im Bestand eines Menschen.
let ws = null;
let aufraeumfehler = false;

try {
  ws = await verbinden(profil);

  await schicke(ws, "Emulation.setDeviceMetricsOverride", {
    width: 1280,
    height: 800,
    deviceScaleFactor: 1,
    mobile: false,
  });

  await schicke(ws, "Page.navigate", { url: adresse });
  await warte(3500);

  const bruecke = await werte(
    ws,
    `typeof window.go === 'object' && typeof window.go.app === 'object'`,
  );

  if (bruecke !== true) {
    throw new Error(`unter ${adresse} antwortet keine Anwendung — läuft \`make dev-browser\`?`);
  }

  await bisRuhe(ws);
  stationen.push(["Die Anwendung steht", gezaehlt()]);

  // **1 — in die Sets.** Der Weg beginnt in der Leiste; die Startseite bietet
  // ihn zusätzlich als Kachel an, und beides ist ein Klick.
  await klick(ws, "Sets⌘2");
  await bisRuhe(ws);
  stationen.push(["Die Setliste steht", gezaehlt()]);

  // **2 und 3 — der Name.** Spiel und Stil sind vorgewählt (wie in MSE2,
  // `src/gui/new_window.cpp`, Zeile 66) und kosten nichts.
  await setzeFeld(ws, "Name des Sets", setname);

  // **Nachgesehen, nicht angenommen.** „Set anlegen" ist ohne Namen
  // angehalten, und ein Klick darauf tut nichts — der Lauf lief dann in das
  // Warten auf die Laschen und meldete „kam in 20 Sekunden nicht", wo der
  // Name fehlte.
  const nameSteht = await werte(
    ws,
    `[...document.querySelectorAll('input')].some((f) => f.value === ${JSON.stringify(setname)})`,
  );

  if (nameSteht !== true) {
    throw new Error("der Setname kam nicht im Feld an");
  }

  stationen.push(["Der Name steht", gezaehlt()]);

  // **4 — anlegen.**
  await klick(ws, "Set anlegen");
  await bis(
    ws,
    `[...document.querySelectorAll('[role="tab"]')].some((b) => b.textContent.trim() === 'Karten')`,
    "die Laschen des neuen Sets",
  );
  await bisRuhe(ws);
  stationen.push(["Das Set ist offen", gezaehlt()]);

  // **5 — die Karte.**
  await klick(ws, "Karte anlegen");
  await bisRuhe(ws);
  stationen.push(["Die Karte steht", gezaehlt()]);

  // **6 und 7 — ihr Name.** Das erste Kartenfeld heißt in jedem Spielpaket
  // anders; gesucht wird deshalb das erste **Textfeld des Editors**.
  //
  // **Nicht das erste `input[aria-label]`** — das war der erste Anlauf, und er
  // traf das Ankreuzfeld der Kartenliste („Ohne Namen — Für die
  // Massenänderung anhaken"). Getippt wurde dann in ein Kästchen, und der Lauf
  // meldete „der Name steht nicht im Feld", ohne zu sagen, warum.
  const feldname = await werte(
    ws,
    `(() => {
      const f = [...document.querySelectorAll('input[aria-label]')]
        .find((x) => (x.type === 'text' || !x.hasAttribute('type'))
          && x.getAttribute('aria-label') !== ''
          && !/anhaken|Massenänderung/.test(x.getAttribute('aria-label')));
      return f ? f.getAttribute('aria-label') : '';
    })()`,
  );

  if (feldname === "") {
    throw new Error("der Editor zeigt kein Kartenfeld");
  }

  await griffe.insFeld(ws, feldname);
  await tippe(ws, kartenname);

  // **Der Wert wird geschrieben, wenn das Feld den Fokus verliert** — genau
  // wie beim Menschen, der weiterklickt.
  await werte(ws, `document.activeElement?.blur()`);
  await bisRuhe(ws);

  stationen.push([`Die Karte heißt „${kartenname}"`, gezaehlt()]);

  const steht = await werte(
    ws,
    `[...document.querySelectorAll('input')].some((f) => f.value === ${JSON.stringify(kartenname)})`,
  );

  process.stdout.write(`| Station | Handlungen bis hierher |\n| --- | ---: |\n`);
  for (const [was, zahl] of stationen) {
    process.stdout.write(`| ${was} | ${String(zahl)} |\n`);
  }

  const befunde = [];

  if (steht !== true) {
    befunde.push("der Name steht nicht im Feld");
  }

  if (gezaehlt() !== erwartet.handlungen) {
    befunde.push(
      `${String(gezaehlt())} Handlungen statt ${String(erwartet.handlungen)}`,
    );
  }

  if (befunde.length > 0) {
    process.stderr.write(`\n✗ ${befunde.join("; ")}\n`);
    process.exitCode = 1;
  } else {
    process.stdout.write(
      `\n✓ ${String(gezaehlt())} Nutzerhandlungen von der Anwendung zur ersten Karte\n`,
    );
  }
} finally {
  if (ws !== null) {
    try {
      // **Über den Namen und nicht über eine gemerkte Kennung.** Der Lauf
      // kann vor dem Anlegen abbrechen, und dann gibt es keine — er kann aber
      // auch danach abbrechen, und dann liegt das Set da. Der Name trägt die
      // Prozesskennung und ist damit für diesen Lauf eindeutig.
      const weg = JSON.parse(
        await werte(
          ws,
          `(async () => {
            const s = await window.go.app.Bindungen.Sets();
            const meine = (s.sets ?? []).filter((e) => e.name === ${JSON.stringify(setname)});
            const fehler = [];
            for (const eines of meine) {
              const antwort = await window.go.app.Bindungen.SetLoeschen(eines.id, eines.name);
              if (antwort.fehler) fehler.push(antwort.fehler);
            }
            return JSON.stringify({ geloescht: meine.length, fehler });
          })()`,
        ),
      );

      if (weg.fehler.length > 0) {
        aufraeumfehler = true;
        process.stderr.write(`✗ Aufräumen: ${JSON.stringify(weg.fehler)}\n`);
      } else if (weg.geloescht > 0) {
        process.stdout.write(`  aufgeräumt — ${String(weg.geloescht)} Messset(s)\n`);
      }
    } catch (fehler) {
      aufraeumfehler = true;
      process.stderr.write(`✗ Aufräumen fehlgeschlagen: ${String(fehler)}\n`);
    }
  }

  chrome.kill();
  await warte(300);
  await rm(profil, { recursive: true, force: true }).catch(() => undefined);

  if (aufraeumfehler) {
    process.exitCode = 1;
  }
}
