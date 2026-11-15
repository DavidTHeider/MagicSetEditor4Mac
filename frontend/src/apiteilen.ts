// **Teilen über die Ziele des Systems** — NET-04
// (#1708).
//
// Die Gestalt steht in `internal/app/teilen.go`; was mitgeht und was fehlt,
// sind Schlüssel, die der Baustein `Teilen` in die Sprache des Fensters setzt.
import type { FehlerDTO } from './formen'

// Teilart ist, was geteilt wird.
export type Teilart = 'einladung' | 'kartenbild' | 'set' | 'deck' | 'sammlung'

// TeilgutDTO ist, was vor dem Teilen gezeigt wird.
export interface TeilgutDTO {
  kennung: string
  art: Teilart
  titel: string
  dateiname: string
  groesse: number

  // text ist, was als Text geht — bei einer Einladung der Code.
  text: string

  mitgehendes: string[]
  verluste: string[]

  // zuGrossFuerMail empfiehlt einen anderen Weg als Mail.
  zuGrossFuerMail: boolean
}

// TeilenAntwort trägt das Teilgut oder den Fehler.
export interface TeilenAntwort {
  teilgut: TeilgutDTO | null
  fehler: FehlerDTO | null
}

// DecklisteAntwort trägt ein Deck als Liste.
export interface DecklisteAntwort {
  liste: string
  fehler: FehlerDTO | null
}

// Teilenapi sind die Wege des Teilens und die Deckliste, auf der es baut.
export interface Teilenapi {
  TeilenVorbereiten: (art: Teilart, bezug: string) => Promise<TeilenAntwort>
  TeilenAnbieten: (kennung: string) => Promise<TeilenAntwort>
  TeilenVerwerfen: (kennung: string) => Promise<TeilenAntwort>
  DecklisteAusgeben: (deckID: string) => Promise<DecklisteAntwort>
}
