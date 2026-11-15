// Die Antwort des Dienstkatalogs — Werkstück A1 von Phase 5.3
// (#1440).
//
// **Die Gestalt kommt aus `internal/app/dienstkatalog.go`**, Feld für Feld;
// was die Werte bedeuten, steht dort und in `internal/domain/dienstlage.go`.
import type { Umschlag } from './umschlag'

export interface DienstvertragDTO {
  basisadresse: string
  endpunkte: string[]
  zugang: string
  tarif: string
  grenze: string
  frist: string
  bedingungen: string
  abgerufen: string
  beleg: string
}

export interface NachweisDTO {
  lauf: string
  datum: string
  fassung: string
  endpunkte: string[]
  giltFuerDieseFassung: boolean
}

export type Dienststatus =
  'bereit' | 'aus' | 'zugangFehlt' | 'gestoert' | 'gesperrt' | 'nichtGewaehlt'

export interface DienstfaehigkeitDTO {
  dienst: string
  frage: string
  status: Dienststatus
  grund: string
  angabe: string
  hatWahl: boolean
  gewaehlt: boolean
  wirkung: string
  vertrag: DienstvertragDTO
  nachweis: NachweisDTO | null
}

export interface DienstkatalogAntwort extends Umschlag {
  eintraege: DienstfaehigkeitDTO[]
}
