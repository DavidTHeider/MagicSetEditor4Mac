// **Die zweite Kürzelreihe an der laufenden Brücke** —
// #1291,
// Werkstück E.
//
// # Warum es diesen Lauf gibt
//
// `zweitereihe.test.ts` prüft `deuten()` — eine **reine Funktion**. Ob ⌥⌘1 im
// Fenster wirklich auf die Vorlagen führt, sagt er nicht. Das ist die Grenze,
// vor der der Auftrag dieser Runde warnt: **zwei grüne Seiten einer Grenze
// sind kein Nachweis für die Grenze.**
//
// ```
// make dev                      # in einem zweiten Fenster
// node scripts/kuerzellauf.mjs
// ```
//
// # Zwei Messfehler, die er gekostet hat
//
// **Der erste**: er prüfte `window.location.pathname`. Acht rote Zeilen sahen
// wie ein Befund aus — und die Gegenprobe klärte es: ⌘K öffnet die Palette
// (der Horcher liegt also), und ein **Klick** auf „Sets" ändert die Anzeige,
// ohne die Adresse anzufassen. Diese Anwendung führt ihre Navigation im
// Zustand; `pfad` in `navigation.ts` ist ein Schlüssel und keine URL.
//
// **Der zweite**: er prüfte die Überschrift. Sechs der acht Seiten brauchen ein
// **offenes Set** und zeigen sonst „Kein Set offen" — ohne `<h1>` und ohne
// Eintrag in der Navigationsleiste. Wieder sechs rote Zeilen, wieder war die
// Messung schuld.
//
// **Beide Male hat erst eine Kontrolle es entschieden** — nicht das Lesen des
// Quelltextes (§10).
//
// # Was er misst und was nicht
//
// **Er misst die beiden Seiten, die ohne offenes Set stehen**: ⌥⌘1 auf
// *Vorlagen* und ⌥⌘2 auf *Sammlung* — genau die zwei, die ausdrücklich
// verlangt waren. Die übrigen sechs verlangen ein offenes Set, und ein
// Messlauf, der eines öffnet, fasste einen fremden Bestand an (siehe
// `kernwegmass.mjs`).
//
// **Was er damit nicht sagt**: dass ⌥⌘3 bis ⌥⌘8 auf die **richtige** Seite
// führen. Dass sie die Anzeige wechseln, ist gesehen; welche es wird, sagt nur
// `zweitereihe.test.ts`.

import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";

// **Ein relativer Pfad wie überall sonst** — ein absoluter trüge einen
// Benutzernamen und liefe auf keinem anderen Rechner.
import { schicke, verbinden, warte, werte } from "./abzuegecdp.mjs";

const adresse = process.argv[2] ?? "http://localhost:34115";

const profil = await mkdtemp(join(tmpdir(), "kuerzellauf-chrome-"));
const chrome = spawn(
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
  ["--headless=new", "--remote-debugging-port=0", `--user-data-dir=${profil}`, "about:blank"],
  { stdio: ["ignore", "ignore", "ignore"] },
);

// **Alle, nicht drei.** Eine Stichprobe sagt über die Reihe nichts; was hier
// geprüft wird, ist die Reihe.
// **Erwartet wird der Navigationseintrag mit `aria-current` — nicht der Pfad
// und nicht die Überschrift.**
//
// Der erste Anlauf prüfte `window.location.pathname` — und **acht rote Zeilen**
// sahen wie ein Befund aus. Die Gegenprobe hat es geklärt: ⌘K öffnet die
// Palette (der Horcher liegt also), und ein **Klick** auf „Sets" ändert die
// Überschrift, ohne den Pfad anzufassen. Diese Anwendung führt ihre Navigation
// im Zustand, nicht in der Adresse; `pfad` in `navigation.ts` ist ein
// Schlüssel. Die Messung war falsch, nicht die Anwendung (§10).
//
// **Und die Überschrift war der zweite Fehlversuch.** Sechs der acht Seiten
// brauchen ein offenes Set und zeigen sonst „Kein Set offen" — ohne `<h1>`.
// Sechs rote Zeilen, und wieder war die Messung schuld. Der
// Navigationseintrag trägt `aria-current` in **beiden** Fällen.
//
// **Er prüft die erste Reihe** (Werkstück C von Phase 4.10,
// #1325):
// sie gehört ganz der Seitenleiste, und ihre Ziele stehen ohne offenes Set da.
//
// **Was er nicht misst**: die zweite Reihe. Sie führt ausschließlich auf
// **Laschen eines Sets**, und die verlangen ein offenes — ein Messlauf, der
// eines öffnet, fasste einen fremden Bestand an (siehe `kernwegmass.mjs`).
// `zweitereihe.test.ts` prüft ihre Deutung als Funktion; dass sie im Fenster
// ankommt, ist damit **nicht** gemessen.
// # Die Fälle kommen aus der Leiste, nicht von Hand
//
// **Eine Liste von Hand bliebe hinter den Phasen zurück** — und ein Lauf, der
// in der Freigabeliste steht und nicht in der Pipeline, wäre rot, ohne dass es
// jemandem auffällt.
//
// Die Leiste **zeigt jede Taste an** — jeder Eintrag trägt seinen Namen und
// sein Kürzel nebeneinander (`rahmen/Leistenzone.tsx`). Der Lauf liest sie aus
// der laufenden Anwendung und prüft, was sie behauptet: **steht ⌘4 an einem
// Eintrag, muss ⌘4 dorthin führen.**
//
// **Damit gibt es keine zweite Stelle, an der dieselbe Wahrheit steht** — und
// der Lauf misst genau das, was ein Mensch am Bildschirm liest.
const leiste = `[...document.querySelectorAll('nav button')]
  .map(b => ({
    name: (b.firstElementChild?.textContent ?? '').trim(),
    kuerzel: (b.lastElementChild?.textContent ?? '').trim().replace('⌘', ''),
  }))
  .filter(e => e.name && /^[0-9]$/.test(e.kuerzel))`;

// aktuell liest den Navigationseintrag, der als aktueller markiert ist.
//
// **Der Name steht im ersten Kind, nicht im ganzen Knopf** — und das ist der
// **dritte** Messfehler dieser Datei (2026-09-03). Seit die Leiste an jedem
// Eintrag seine Taste zeigt, lautet sein `textContent` „Vorlagen⌘4"; der
// Vergleich mit „Vorlagen" schlug bei allen vier Fällen fehl, und vier rote
// Zeilen sahen aus, als führte kein Kürzel irgendwohin.
//
// Wieder war die Messung schuld, und wieder hat erst die **Kontrolle** es
// entschieden: „Vorlagen⌘4" ist der Beleg, dass ⌘4 auf die Vorlagen geführt
// hat — er stand in der Fehlerzeile.
const aktuell = `(() => {
  const e = document.querySelector('[aria-current=page]');
  if (!e) return '(keiner)';
  return (e.firstElementChild?.textContent ?? e.textContent ?? '').trim();
})()`;

let ws = null;
let schlecht = 0;

// **Wie viele es wurden, steht erst nach dem Lesen der Leiste fest.**
let gemessen = 0;

try {
  ws = await verbinden(profil);

  await schicke(ws, "Page.navigate", { url: adresse + "/" });
  await warte(2500);

  const faelle = (await werte(ws, leiste))
    // **`Karten` ist dabei**, obwohl es eine Setlasche ist: die Leiste zeigt
    // sie seit Werkstück C von Phase 4.10 in der ersten Reihe, und ⌘3 führt
    // auch ohne offenes Set dorthin — nachgemessen am 2026-09-07.
    .map(({ name, kuerzel }) => ({
      ziffer: kuerzel,
      code: `Digit${kuerzel}`,
      zeichen: kuerzel,
      erwartet: name,
    }))
    .sort((a, b) => Number(a.ziffer) - Number(b.ziffer));

  gemessen = faelle.length;

  // **Die Gegenprobe zur Messung**: findet sie nichts, ist sie grün und sagt
  // nichts (§10). Am 2026-09-07 waren es acht.
  if (faelle.length < 5) {
    process.stderr.write(
      `nur ${String(faelle.length)} Ziffern in der Leiste gefunden — ` +
        "zeigt sie ihre Kürzel noch an?\n",
    );
    process.exit(1);
  }

  for (const fall of faelle) {
    // **Erst irgendwohin, nur nicht ans Ziel.** Sonst prüfte der Fall, ob die
    // Seite dort bleibt, wo sie schon war — und das ist immer wahr.
    //
    // **Und nicht bloß auf die Startseite**: dann wäre genau ⌘1 nicht
    // messbar, weil `/` schon dorthin führt — eine Lücke in der Messung, die
    // wie ein Befund aussähe.
    //
    // Zurückgesetzt wird deshalb über einen **Klick** auf einen anderen Eintrag
    // der Leiste. Ein Klick ist der Weg, den auch ein Mensch nimmt, und er
    // ändert die Anzeige, ohne eine Taste zu benutzen — sonst prüfte der
    // Rücksetzer schon, was gemessen werden soll.
    await schicke(ws, "Page.navigate", { url: adresse + "/" });
    await warte(1200);

    await schicke(ws, "Runtime.evaluate", {
      expression: `[...document.querySelectorAll('nav button')]
        .find(b => (b.firstElementChild?.textContent ?? '').trim()
          !== ${JSON.stringify(fall.erwartet)})?.click()`,
    });

    await warte(800);

    const vorher = await werte(ws, aktuell);
    if (vorher === fall.erwartet) {
      console.log(`  ⌘${fall.ziffer}: der Rücksetzer landete schon auf ${vorher}`);
      schlecht += 1;

      continue;
    }

    // **modifiers 4 = Meta.** Die erste Reihe wird über `event.key` gedeutet
    // — ohne ⌥ schreibt nichts die Taste um, und `key` trägt die Ziffer
    // selbst. (Die zweite Reihe braucht `event.code`, weil ⌥1 auf einer
    // Mac-Tastatur `"¡"` schickt; der Grund steht in `zweitereihe.test.ts`.)
    await schicke(ws, "Input.dispatchKeyEvent", {
      type: "keyDown",
      modifiers: 4,
      key: fall.zeichen,
      code: fall.code,
      windowsVirtualKeyCode: 48 + Number(fall.ziffer),
    });

    await warte(800);

    const pfad = await werte(ws, aktuell);
    const gut = pfad === fall.erwartet;
    if (!gut) schlecht += 1;

    console.log(
      `  ⌘${fall.ziffer} (${fall.code}) → ${pfad}` +
        (gut ? "  ✓" : `  ✗ erwartet ${fall.erwartet}`),
    );
  }
} finally {
  if (ws) ws.close();
  chrome.kill();
  // **Ein Aufräumen, das scheitert, hält den Lauf nicht auf.** Chrome legt
  // seinen Zwischenspeicher noch weg, während wir schon räumen; ein
  // `ENOTEMPTY` darüber machte aus einer geglückten Messung einen Absturz.
  await rm(profil, { recursive: true, force: true }).catch(() => {});
}

console.log(
  schlecht === 0
    ? `\nalle ${String(gemessen)} an der laufenden Brücke ✓`
    : `\n${String(schlecht)} von ${String(gemessen)} ✗`,
);
process.exit(schlecht === 0 ? 0 : 1);
