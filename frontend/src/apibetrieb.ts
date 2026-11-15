// **Was die Anwendung über sich selbst sagt** — OPS-07
// (#1724).
// Die Gestalt kommt aus `internal/app/betriebsauskunft.go` und
// `internal/app/pruefstand.go`.
import type { Umschlag } from './umschlag'
import type { Wiederanlaufapi } from './apiwiederanlauf'

export interface BetriebsortDTO {
  name: string
  pfad: string
}

// PruefpunktDTO ist ein Gegenstand mit Stand und Beleg — der Beleg ist ein
// Pfad beim Quelltext oder die Adresse eines Issues.
export interface PruefpunktDTO {
  bereich: string
  gegenstand: string
  stand: string
  beleg: string
}

export interface BetriebsauskunftAntwort extends Umschlag {
  fassung: string
  stand: string
  gebaut: string
  veraendert: boolean
  migrationen: number
  letzteMigration: string
  signatur: string
  notarisiert: boolean
  aktualisierung: string
  orte: BetriebsortDTO[]
  pruefstand: PruefpunktDTO[]
}

// **Und der Wiederanlauf gehört dazu** (F04, #2343): ob die letzte Sitzung
// sauber endete, ist eine Auskunft über den Betrieb, und `api.ts` steht an
// der Zeilengrenze.
export interface Betriebsapi extends Wiederanlaufapi {
  Betriebsauskunft: () => Promise<BetriebsauskunftAntwort>
}
