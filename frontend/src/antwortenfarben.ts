// **Die Gestalten der Quelltextfarben** — was eine Bindung zurückgibt, wenn
// sie einen Text in seine Merkmale zerlegt.
//
// **Aus `antworten.ts` herausgelöst**, als es über 400 Zeilen ging (§2). Der
// Schnitt folgt der Frage: dort stehen die Antworten über **Karten, Sets und
// Pakete**, hier die über **Text**.

// Merkmalstelle ist ein Stück Quelltext mit seiner Art und seiner Stelle
// (#1164).
export interface Merkmalstelle {
  art: string
  zeile: number
  spalte: number

  // breite zählt die Zeichen im **Quelltext** — bei einer Zeichenkette also
  // samt ihren Anführungszeichen.
  breite: number
}

// Farbantwort trägt die Merkmale eines Ausdrucks.
//
// **Ohne Umschlag**: es gibt nichts, was beim Zerlegen schiefgehen könnte —
// ein unlesbarer Ausdruck färbt so weit, wie er lesbar ist.
export interface Farbantwort {
  merkmale: Merkmalstelle[]
}

// Werkzeugzeile ist ein Befehl samt dem Schlüssel seines Satzes (#1164).
