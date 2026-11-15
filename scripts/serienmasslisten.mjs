// **Welche Seiten das Serienmaß fährt** — Werkstück A von Phase 4.10
// (#1325).
//
// # Warum das eine eigene Datei ist
//
// Zwei Fragen, zwei Listen: *was kostet eine Anwendungsseite* und *was kostet
// eine Lasche eines offenen Sets*. Sie werden verschieden erreicht — jene über
// die Seitenleiste, diese erst, nachdem ein Set offen ist —, und der Lauf
// darüber steht in `serienmass.mjs`. Hier steht nur, **worüber** gemessen
// wird.
//
// **Und ein Wächter hält beide gegen `navigation.ts`**
// (`internal/werkzeuge/serienmassseiten_test.go`). Ohne ihn ist eine
// Aufzählung von Hand grün, sobald sie zu kurz ist — genau das war sie bis
// zum 2026-09-02, und Decks und Turniere fehlten seit ihrer Entstehung.

// anwendungsseiten sind die neun Seiten, die an keinem Set hängen.
//
// **Jede nennt ihre Überschrift**, und der Lauf hält sie dagegen: so steht
// eine Zeile der Tabelle für die Seite, deren Namen sie trägt, und nicht für
// die vorige.
export const anwendungsseiten = [
  { schluessel: "start", name: "Start", weg: [], ueberschrift: "MagicSetEditor4Mac" },
  { schluessel: "sets", name: "Sets", weg: ["Sets"], ueberschrift: "Sets" },
  { schluessel: "pakete", name: "Pakete", weg: ["Pakete"], ueberschrift: "Pakete" },
  { schluessel: "vorlagen", name: "Vorlagen", weg: ["Vorlagen"], ueberschrift: "Vorlagen" },
  { schluessel: "sammlung", name: "Sammlung", weg: ["Sammlung"], ueberschrift: "Sammlung" },
  // **Decks und Turniere kamen am 2026-09-02 dazu** — sie fehlten hier, seit
  // es sie gibt. Die Phasen 4.6 und 4.7 haben die Liste nicht mitgezogen.
  { schluessel: "decks", name: "Decks", weg: ["Decks"], ueberschrift: "Decks" },
  { schluessel: "turniere", name: "Turniere", weg: ["Turniere"], ueberschrift: "Turniere" },
  { schluessel: "spielen", name: "Spielen", weg: ["Spielen"], ueberschrift: "Spielen" },
  // **Handeln kam mit T9 dazu** (#1474) — ⌘0, die letzte Ziffer der Reihe.
  { schluessel: "handeln", name: "Handeln", weg: ["Handeln"], ueberschrift: "Handeln" },
  { schluessel: "einstellungen", name: "Einstellungen", weg: ["Einstellungen"], ueberschrift: "Einstellungen" },
  { schluessel: "nutzer", name: "Nutzer und Profile", weg: ["Nutzer und Profile"], ueberschrift: "Nutzer und Profile" },
  { schluessel: "hilfe", name: "Hilfe", weg: ["Hilfe"], ueberschrift: "Hilfe" },
];

// setlaschen sind die dreizehn Laschen eines offenen Sets.
//
// **Sieben davon stehen seit UX-08 in einer Schublade** (#1680) — `schublade`
// nennt dann den Knopf im Kopf des Sets, der sie öffnet. Der Lauf drückt ihn
// vor dem Reiter; ohne ihn gäbe es den Reiter gar nicht.
//
// **Sie haben keine Überschrift zum Vergleichen.** Eine Lasche trägt ihren
// Namen im Reiter, nicht in einer `h1` — geprüft wird deshalb, dass der
// Reiter danach `aria-selected="true"` trägt. Das ist dieselbe Zusage in der
// Gestalt, die die Seite wirklich hat.
//
// **Die Beschriftungen sind die deutschen** aus `frontend/src/sprache/de/`;
// gemessen wird an der deutschen Oberfläche, weil sie die Ausgangssprache ist
// (§9).
export const setlaschen = [
  { schluessel: "karten", name: "Karten", reiter: "Karten" },
  { schluessel: "setangaben", name: "Setangaben", reiter: "Setangaben", schublade: "Eigenschaften" },
  { schluessel: "schluesselwoerter", name: "Schlüsselwörter", reiter: "Schlüsselwörter", schublade: "Eigenschaften" },
  { schluessel: "konsole", name: "Konsole", reiter: "Konsole", schublade: "Werkzeuge" },
  { schluessel: "pack", name: "Zufallspack", reiter: "Zufallspack", schublade: "Werkzeuge" },
  { schluessel: "symbol", name: "Symbol", reiter: "Symbol", schublade: "Eigenschaften" },
  { schluessel: "stil", name: "Stil", reiter: "Stil", schublade: "Eigenschaften" },
  { schluessel: "suche", name: "Suche", reiter: "Suche" },
  { schluessel: "ersetzen", name: "Ersetzen", reiter: "Ersetzen" },
  { schluessel: "statistik", name: "Statistik", reiter: "Statistik", schublade: "Werkzeuge" },
  // **Seit UXR-09 (#1917) Ziele des Ausgabeassistenten** — der Knopf
  // *Ausgeben* im Kopf öffnet ihn, das Ziel ist ein Optionsfeld darin.
  { schluessel: "bilder", name: "Bilder", reiter: "Bilder", schublade: "Ausgeben" },
  { schluessel: "druckbogen", name: "Druckbogen", reiter: "Druckbogen", schublade: "Ausgeben" },
  { schluessel: "ausgeben", name: "Ausgeben", reiter: "Setdatei", schublade: "Ausgeben" },
];
