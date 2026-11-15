// **Die eigenen Übersetzungen** — LOC-02
// (#1688).
//
// Was fehlt, veraltet oder unbekannt ist und ob ein Satz gilt, entscheidet
// Go gegen den deutschen Katalog (`service.Sprachpakete`); hier steht die
// Gestalt, in der es ankommt.

import type { Umschlag } from './umschlag'

// SprachpaketDTO ist eine Übersetzung mit ihren Zahlen.
export interface SprachpaketDTO {
  kennung: string
  name: string
  basis: string
  stand: string
  lizenz: string
  herkunft: string
  geaendert: string
  gesamt: number
  uebersetzt: number
  fehlend: number
  veraltet: number
  unbekannt: number
}

// UebersetzungstextDTO ist ein Satz und der deutsche, gegen den er
// übersetzt wurde.
export interface UebersetzungstextDTO {
  text: string
  quelle: string
}

export interface SprachpaketlisteAntwort extends Umschlag {
  pakete: SprachpaketDTO[]
}

export interface SprachpaketAntwort extends Umschlag {
  paket: SprachpaketDTO
  texte: Record<string, UebersetzungstextDTO>
  fehlendePfade: string[]
  veraltetePfade: string[]
  unbekanntePfade: string[]
  pfad: string
  abgebrochen: boolean
}

// SprachpaketangabenDTO sind die Angaben, die sich ändern — `null` bleibt.
export interface SprachpaketangabenDTO {
  name: string | null
  stand: string | null
  lizenz: string | null
  herkunft: string | null
}

export interface Sprachapi {
  Sprachpakete: () => Promise<SprachpaketlisteAntwort>
  Sprachpaket: (kennung: string) => Promise<SprachpaketAntwort>
  SprachpaketAnlegen: (
    kennung: string,
    name: string,
    basis: string,
  ) => Promise<SprachpaketAntwort>

  // SprachpaketAendern setzt Angaben und Sätze — ein leerer Satz nimmt die
  // Übersetzung zurück.
  SprachpaketAendern: (
    kennung: string,
    angaben: SprachpaketangabenDTO,
    texte: Record<string, string>,
  ) => Promise<SprachpaketAntwort>
  SprachpaketLoeschen: (kennung: string) => Promise<Umschlag>

  // Aus- und Eingeben fangen mit einem Dateidialog an; eine vorhandene
  // ersetzt nur `ersetzen`.
  SprachpaketAusgabeSpeichern: (kennung: string) => Promise<SprachpaketAntwort>
  SprachpaketEinlesenOeffnen: (ersetzen: boolean) => Promise<SprachpaketAntwort>
}
