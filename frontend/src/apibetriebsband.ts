// **Das Betriebsband** — OPS02
// (#2348).
// Die Gestalt kommt aus `internal/app/betriebsband.go`; wann eine Lage ein
// Hinweis ist, entscheidet `internal/domain/betriebsband.go`.
import type { Umschlag } from './umschlag'

export type Betriebshinweisart =
  'platzKnapp' | 'sicherungAlt' | 'dienstBeendet' | 'webprozessNeu'

export interface BetriebshinweisDTO {
  art: Betriebshinweisart

  // ziel ist, wohin der Hinweis führt — `einstellungen/<gruppe>`.
  ziel: string

  // werte füllen den Satz: `frei` in Bytes, `tage`, `ausgang`, `anzahl`.
  werte: Record<string, string>
}

export interface BetriebsbandAntwort extends Umschlag {
  hinweise: BetriebshinweisDTO[]

  // ungemessen nennt, was sich nicht fragen ließ — `platz`, `sicherung`,
  // `dienst`.
  ungemessen: string[]
  aufwandMikros: number
}

export interface Betriebsbandapi {
  Betriebsband: () => Promise<BetriebsbandAntwort>
}
