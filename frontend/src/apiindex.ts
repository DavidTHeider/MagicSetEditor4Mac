// **Der lokale Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
// Die Gestalt kommt aus `internal/app/retrievalindex.go`.
import type { Umschlag } from './umschlag'

// KorpusstandDTO ist der Stand **einer** Quelle im Index.
//
// **Je Quelle einer**: es gibt keinen gemeinsamen „aktuellen Regelstand".
export interface KorpusstandDTO {
  quelle: string
  revision: string

  // datum ist das Gültigkeitsdatum — leer, wo die Quelle keines nennt.
  datum: string
}

export interface IndexmanifestDTO {
  modell: string
  modellrevision: string
  laufzeit: string
  abstandsmass: string
  dimension: number
  zerteilfassung: string
  suchkonfiguration: string
  korpusstaende: KorpusstandDTO[]
}

export interface IndexaufbaustandDTO {
  laeuft: boolean
  generation: number
  geschrieben: number
  fehler: string
}

export interface IndexstandAntwort extends Umschlag {
  aktiv: boolean
  manifest: IndexmanifestDTO
  aufbau: IndexaufbaustandDTO

  // suchbar sagt, ob **semantisch** gesucht werden kann; grund sagt, warum
  // nicht.
  suchbar: boolean
  grund: string
}

// IndexsuchfilterDTO ist die geprüfte Filterstruktur — **kein Suchtext**.
export interface IndexsuchfilterDTO {
  quellen: string[]
  sprachen: string[]
  sets: string[]

  // farben ist die Farbe der Karte, farbidentitaet ihre Identität nach
  // Regel 903.4 — nicht dasselbe.
  farben: string[]
  farbidentitaet: string[]
  typen: string[]

  manawertVon: number | null
  manawertBis: number | null

  format: string
  formatlage: string
  formatstand: string
}

export interface IndextrefferDTO {
  rang: number
  kennung: string
  quelle: string
  bezugsadresse: string
  titel: string
  text: string

  // ausschnitt ist die Textstelle — leer, wo nur der Vektor passte.
  ausschnitt: string

  sprache: string
  regelnummer: string
  urheber: string

  quellrevision: string
  quelldatum: string
  abgerufen: string
}

export interface IndexsucheAntwort extends Umschlag {
  treffer: IndextrefferDTO[]

  // modus ist `hybrid`, `lexikalisch` oder `ohne-index`.
  modus: string
  grund: string
}

export interface IndexaufbauAntwort extends Umschlag {
  generation: number
  geschrieben: number

  // dauer ist die Bauzeit in Millisekunden.
  dauer: number

  // fortgeschrieben sagt, ob auf der aktiven Generation aufgebaut wurde.
  fortgeschrieben: boolean

  manifest: IndexmanifestDTO
}

export interface Indexapi {
  Indexstand: () => Promise<IndexstandAntwort>
  Indexsuche: (
    frage: string,
    grenze: number,
    filter: IndexsuchfilterDTO,
  ) => Promise<IndexsucheAntwort>
  IndexAufbauen: (
    modell: string,
    quelle: string,
    revision: string,
    datum: string,
    fortschreiben: boolean,
  ) => Promise<IndexaufbauAntwort>
  IndexAbbrechen: () => Promise<boolean>
}
