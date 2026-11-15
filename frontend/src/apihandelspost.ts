// **Was zu einem Gegenüber noch aussteht** — Werkstück T5 von Phase 5.6
// (#1470).
//
// **Zahlen und keine Nachrichten**: die Diagnose sagt, ob etwas klemmt, und
// nicht, worum es geht.
import type { Umschlag } from './umschlag'

// Poststand sagt, was noch auf seine Bestätigung wartet.
export interface Poststand {
  kennung: string

  // offen ist die Zahl der Nachrichten, die noch warten.
  offen: number

  // hoechsteVersuche ist, wie oft die hartnäckigste schon hinausging —
  // **die höchste und nicht die Summe**.
  hoechsteVersuche: number
}

// Poststandantwort trägt den Stand samt seinem Fehler.
export interface Poststandantwort extends Umschlag {
  stand: Poststand
}

// Handelspostapi ist der eine Weg zum Poststand.
export interface Handelspostapi {
  Handelspoststand: (kennung: string) => Promise<Poststandantwort>
}
