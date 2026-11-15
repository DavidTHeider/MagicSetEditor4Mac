// **Die Antworten der Autoersetzungsverwaltung** —
// #1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
import type { Umschlag } from './api'

// AutoersetzungsregelDTO ist eine Regel, wie die Seite sie zeigt.
export interface AutoersetzungsregelDTO {
  muster: string
  ersatz: string

  // ganzesWort sagt, ob nur ein ganzes Wort ersetzt wird.
  ganzesWort: boolean

  // aktiv sagt, ob die Regel greift.
  aktiv: boolean

  // eigen sagt, dass die Regel hier angelegt wurde — nur eine solche lässt
  // sich löschen.
  eigen: boolean
}

// Autoersetzungsantwort trägt die Regeln und den Schalter.
export interface Autoersetzungsantwort extends Umschlag {
  regeln: AutoersetzungsregelDTO[]

  // an ist der globale Schalter dieses **Spiels**.
  an: boolean
}
