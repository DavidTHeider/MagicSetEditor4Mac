// **Wer am Tisch saß** — Werkstück T7 von Phase 5.6
// (#1472).
//
// **Es öffnet nichts und fragt nichts an.** Es beantwortet die eine Frage,
// die nach einer Partie offen ist: mit wem habe ich gerade gespielt? Was ein
// Gegenüber sehen darf, entscheidet weiterhin seine Freigabe — eine Partie
// ist keine.
import type { Peerkontakt } from './apipeerkontakt'
import type { Umschlag } from './umschlag'

// Handelspartnerantwort ist das Gegenüber zu einem Sitz.
export interface Handelspartnerantwort extends Umschlag {
  sitz: string
  kontakt: Peerkontakt

  // bekannt sagt, ob zu diesem Sitz ein bekanntes Gegenüber gehört.
  //
  // **Eine eigene Angabe und keine leere Kennung**: „niemand saß da" und „da
  // saß jemand, den ich nicht kenne" sind zwei Auskünfte.
  bekannt: boolean
}

// Handelspartnerapi ist der eine Weg dorthin.
export interface Handelspartnerapi {
  Handelspartner: (sitz: string) => Promise<Handelspartnerantwort>
}
