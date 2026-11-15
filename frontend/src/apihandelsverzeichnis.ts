// **Die Liste der Handelsvorgänge und ihre Entwürfe** — TRADE-01
// (#1715).
//
// Dieselbe Gestalt wie `internal/app/handelsverzeichnis.go`.
import type {
  Handelsvorgang,
  Handelsvorgangantwort,
  Handelswunsch,
} from './apihandelsleitung'
import type { Umschlag } from './umschlag'

// Handelslage sagt, ob ich an einem Vorgang etwas tun muss.
export type Handelslage = 'dran' | 'wartet' | 'abgeschlossen' | 'abgebrochen'

// Handelsordnung ist die Ordnung der Liste.
export type Handelsordnung = '' | 'geaendert' | 'angelegt' | 'lage' | 'partner'

// Handelsfilter ist, was die Liste zeigen soll — leer heißt „alles außer
// dem Archiv".
export interface Handelsfilter {
  partner: string
  lagen: Handelslage[]
  suche: string
  marke: string
  seit: string
  bis: string
  ordnung: Handelsordnung
  archiv: boolean
  sprache: string
}

export interface Handelspartner {
  kennung: string
  name: string
}

// Handelsvorgangzeile ist ein Vorgang in der Liste.
export interface Handelsvorgangzeile {
  vorgang: Handelsvorgang
  lage: Handelslage
  naechster: string
  gegenueber: Handelspartner[]
  karten: string[]
  marken: string[]
  angelegt: string
  geaendert: string
  archiviert: boolean
  unbestaetigt: boolean
}

export interface Handelsuebersichtantwort extends Umschlag {
  zeilen: Handelsvorgangzeile[]
}

// Handelsverzeichnisapi sind die fünf Wege der Liste und der Entwürfe.
export interface Handelsverzeichnisapi {
  HandelUebersicht: (filter: Handelsfilter) => Promise<Handelsuebersichtantwort>
  HandelEntwurfAendern: (
    wunsch: Handelswunsch,
  ) => Promise<Handelsvorgangantwort>
  HandelKopieren: (vorgang: string) => Promise<Handelsvorgangantwort>
  HandelEntwurfLoeschen: (vorgang: string) => Promise<Umschlag>
  HandelArchivieren: (vorgang: string, an: boolean) => Promise<Umschlag>
}
