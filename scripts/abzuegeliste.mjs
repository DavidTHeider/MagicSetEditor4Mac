// **Welche Abzüge das README bekommt** — und in welcher Ordnung.
//
// Seit dem Abschluss von 0.6.0 (#1660) nimmt sie das native Fenster auf
// (`scripts/huellenprobe/readmebilder.mjs`), nicht mehr ein Browser: die
// Oberfläche ist seit Phase 6.13 eine andere, und ein Bild ohne Fensterrahmen
// und Menüleiste ist nicht, was jemand auf dem Mac sieht.
//
// **Die Ordnung ist die der Oberfläche**: die Arbeitsbereiche der Leiste der
// Reihe nach, darin ihre Seiten, dann das Programm. Das README folgt ihr.
//
// Je Weg:
//
//   - `name` ist der Dateiname unter `docs/bilder/` (dazu `-dunkel`),
//   - `seite` der Name der Seite in der Befehlspalette — über die Palette und
//     nicht über eine Stelle in der Leiste, denn die Leiste führt nur auf die
//     erste Seite eines Bereichs,
//   - `lasche`, `klick`, `texte`, `vielleicht` und `fertig` führen weiter und sagen, wann das Bild
//     steht — **auf die Wirkung gewartet**, nicht auf eine Frist (§3).

// mitNamen findet den sichtbaren Knopf mit diesem **zugänglichen Namen** —
// „Öffnen: Blaue Flut" steht im `aria-label`, sichtbar ist nur „Öffnen".
const mitNamen = (name) =>
  `[...document.querySelectorAll('main button')].find((e) => e.checkVisibility() && (e.getAttribute('aria-label') ?? e.innerText.trim()) === ${JSON.stringify(name)})`

export const wege = [
  // **Start** (⌘1)
  { name: 'oberflaeche-start', seite: 'Start' },

  // **Gestalten** (⌘2): Sets, der Kartenarbeitsplatz und was am offenen Set hängt.
  { name: 'oberflaeche-sets', seite: 'Sets' },
  {
    name: 'oberflaeche-karten',
    seite: 'Karten',
    fertig: `[...document.querySelectorAll('img')].some((e) => /vorschau/.test(e.src) && e.complete && e.naturalWidth > 0)`,
    ruhe: 1500,
  },
  { name: 'oberflaeche-stil', seite: 'Stil' },
  { name: 'oberflaeche-setangaben', seite: 'Setangaben' },
  { name: 'oberflaeche-symbol', seite: 'Symbol' },
  { name: 'oberflaeche-schluesselwoerter', seite: 'Schlüsselwörter' },
  { name: 'oberflaeche-statistik', seite: 'Statistik', ruhe: 1500 },
  {
    name: 'oberflaeche-zufallspack',
    seite: 'Zufallspack',
    // **Der fertige Pack, nicht sein Laden** — der erste Aufruf rechnet über
    // zehn Sekunden, und das Bild zeigte bis zum 2026-10-17 „Wird geladen …“.
    fertig: `![...document.querySelectorAll('[role=dialog] *')].some((e) => e.checkVisibility() && e.childElementCount === 0 && e.textContent.trim().startsWith('Wird geladen'))`,
    fertigFrist: 180000,
    ruhe: 1500,
  },
  { name: 'oberflaeche-konsole', seite: 'Konsole' },
  { name: 'oberflaeche-suche', seite: 'Suche' },
  { name: 'oberflaeche-ersetzen', seite: 'Ersetzen' },
  { name: 'oberflaeche-ausgeben', seite: 'Ausgeben' },
  { name: 'oberflaeche-bilder', seite: 'Bilder' },
  { name: 'oberflaeche-druckbogen', seite: 'Druckbogen', ruhe: 2500 },

  // **Sammeln** (⌘3)
  { name: 'oberflaeche-sammlung', seite: 'Sammlung', ruhe: 1500 },

  // **Spielen** (⌘4): Decks, Turniere, Partien und der Tisch.
  { name: 'oberflaeche-decks', seite: 'Decks', klick: mitNamen('Öffnen: Blaue Flut'), klickName: 'das Deck »Blaue Flut«', ruhe: 1500 },
  { name: 'oberflaeche-turniere', seite: 'Turniere', klick: mitNamen('Öffnen: Freitagsrunde'), klickName: 'das Turnier', ruhe: 1500 },
  { name: 'oberflaeche-spielen', seite: 'Spielen' },
  {
    name: 'oberflaeche-tisch',
    seite: 'Spielen',
    // **Wie ein Mensch**: »Weiterspielen« öffnet ein Menü der eigenen Sitze
    // (dieselbe Folge wie die Hüllenprobe `bildrate`), dann die Starthand —
    // sonst zeigte das Bild einen Tisch, auf dem nichts liegt.
    texte: ['Weiterspielen', 'Weiterspielen als Anna'],
    // **Nur, wenn es dasteht** — im zweiten Erscheinungsbild ist die Hand
    // schon gezogen, es ist dieselbe Partie.
    vielleicht: ['Starthand ziehen', 'Hand behalten'],
    fertig: `!!document.querySelector('[role=application] canvas')`,
    // **Die Kartenbilder der Hand kommen über den Bildweg nach** — gerechnet
    // je Karte; ein Bild nach 2,5 s zeigte leere Flächen.
    ruhe: 8000,
  },

  // **Handeln** (⌘5)
  { name: 'oberflaeche-handeln', seite: 'Handeln', texte: ['Turfer'], ruhe: 1500 },

  // **Ressourcen** (⌘6)
  { name: 'oberflaeche-pakete', seite: 'Pakete', ruhe: 1500 },
  { name: 'oberflaeche-vorlagen', seite: 'Vorlagen' },

  // **Programm**: Einstellungen in ihren Gruppen, Nutzer, Hilfe und die Palette.
  { name: 'oberflaeche-einstellungen', seite: 'Einstellungen' },
  { name: 'oberflaeche-einstellungen-bereiche', seite: 'Einstellungen', lasche: 'bereiche' },
  { name: 'oberflaeche-einstellungen-erscheinungsbild', seite: 'Einstellungen', lasche: 'erscheinungsbild' },
  { name: 'oberflaeche-einstellungen-ki', seite: 'Einstellungen', lasche: 'ki', ruhe: 1500 },
  { name: 'oberflaeche-einstellungen-dienste', seite: 'Einstellungen', lasche: 'dienste' },
  {
    name: 'oberflaeche-einstellungen-werkzeuge',
    seite: 'Einstellungen',
    // **Seit AP-10 (#2367) unter *Verbindungen*** und nicht mehr unter
    // *Diagnose*.
    lasche: 'verbindungen',
    // **Ein Abschnitt, den sein Knopf aufklappt** (UXR-19, #1927).
    texte: ['Werkzeuge ohne Fenster'],
    zeige: 'Werkzeuge ohne Fenster',
  },
  { name: 'oberflaeche-nutzer', seite: 'Nutzer und Profile' },
  { name: 'oberflaeche-hilfe', seite: 'Hilfe' },
  { name: 'oberflaeche-befehlspalette', palette: true, tippen: 'Karte' },
];
