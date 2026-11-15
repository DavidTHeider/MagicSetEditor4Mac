// Die Antworten der Übersicht — Werkstück U11 von Phase 5.1
// (#1422).
// Die Gestalt kommt aus `internal/app/weiterarbeiten.go`.
import type { Umschlag } from './umschlag'

export type Arbeitsart = 'set' | 'deck' | 'paket'
export type Arbeitsvorgang = 'geoeffnet' | 'eingefuehrt' | 'ausgegeben'

export interface ArbeitsobjektDTO {
  art: Arbeitsart
  kennung: string
  name: string
  angeheftet: boolean
}

export interface ArbeitsschrittDTO {
  objekt: ArbeitsobjektDTO
  vorgang: Arbeitsvorgang
  zeitpunkt: string
}

export interface WeiterarbeitenAntwort extends Umschlag {
  zuletzt: ArbeitsschrittDTO | null
  geoeffnet: ArbeitsschrittDTO[]
  eingefuehrt: ArbeitsschrittDTO[]
  ausgegeben: ArbeitsschrittDTO[]
  angeheftet: ArbeitsobjektDTO[]
}

export interface Weiterarbeitsapi {
  Weiterarbeiten: () => Promise<WeiterarbeitenAntwort>
  Anheften: (art: string, kennung: string) => Promise<Umschlag>
  Loesen: (art: string, kennung: string) => Promise<Umschlag>
  DeckBenutzt: (deckID: string) => Promise<Umschlag>
}
