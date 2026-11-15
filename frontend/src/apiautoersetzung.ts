// **Die Autoersetzung beim Tippen** — herausgelöst aus `api.ts`, das sonst
// über die Zeilengrenze liefe (§2). Der Schnitt folgt der Frage: dort steht,
// **wie** das Frontend das Backend erreicht, hier der eine Weg, der einen
// getippten Text unterwegs ändert.
import type { Umschlag } from './api'

// AutoersetzungAntwort trägt Text und Schreibmarke zurück.
//
// **Auch im Fehlerfall unverändert** — ein fehlendes Spielpaket darf nicht den
// Text kosten, den jemand gerade tippt (§7 Regel 3).
export type AutoersetzungAntwort = Umschlag & {
  text: string
  stelle: number
}

// Autoersetzungsapi ist der eine Weg dorthin.
export interface Autoersetzungsapi {
  // AutoErsetzen wendet die `auto replace:`-Regeln des Spiels an der
  // Schreibmarke an — `->` wird zu `→`, `(C)` zu `©`.
  //
  // **Die Stelle ist eine UTF-16-Einheit**, so wie `selectionStart` sie zählt.
  // Umgerechnet wird auf der Go-Seite; zwei Zählweisen an zwei Orten liefen
  // auseinander, sobald jemand ein Zeichen außerhalb der Grundebene tippt.
  AutoErsetzen: (
    setID: string,
    text: string,
    stelle: number,
  ) => Promise<AutoersetzungAntwort>
}
