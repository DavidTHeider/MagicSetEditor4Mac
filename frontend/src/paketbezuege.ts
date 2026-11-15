// **Die Stände der bezogenen Bäume** — G04
// (#2251):
// die Gestalt, in der `internal/app/paketbezuege.go` sie schickt.

import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Bezugslage sagt, wie ein bezogener Stand zu seiner Quelle steht.
export type Bezugslage =
  'ungefragt' | 'aktuell' | 'neuer' | 'abweichend' | 'unbekannt'

export interface PaketbezugDTO {
  name: string
  adresse: string
  stand: string
  geprueft: string
  pakete: number
  vorige: boolean
  vorigerStand: string
  lage: Bezugslage
  spitze: string
  spitzeZeit: string
  vorsprung: number
  fehler: FehlerDTO | null
}

export interface PaketbezuegeAntwort extends Umschlag {
  bezuege: PaketbezugDTO[]
}

export interface PaketaenderungDTO {
  paket: string
  neu: number
  geaendert: number
  entfernt: number
}

// Bezugsgrund ist, warum eine Aktualisierung eine Bestätigung verlangt.
export type Bezugsgrund = 'setVerliert' | 'nichtVoraus' | 'gekuerzt'

export interface BezugsvorschauAntwort extends Umschlag {
  name: string
  adresse: string
  von: string
  zu: string
  lage: string
  vorsprung: number
  rueckstand: number
  pakete: PaketaenderungDTO[]
  uebrige: number
  gekuerzt: boolean
  sets: string[]
  bestaetigungNoetig: boolean
  gruende: Bezugsgrund[]
  vorigeErsetzt: boolean
}
