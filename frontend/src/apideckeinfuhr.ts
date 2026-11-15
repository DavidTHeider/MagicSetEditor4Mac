// **Eine Deckliste wird ein Deck** — Werkstück T8 von Phase 5.6
// (#1473).
//
// **Eine eigene Datei neben `apidecks.ts`**, weil jene mit 360 Zeilen an der
// Grenze aus §2 steht — und weil es eine andere Frage ist: dort, was ein Deck
// ist; hier, wie eine fremde Liste hineinkommt.
import type { Umschlag } from './api'
import type { Listentreffer } from './antwortenlisteneinfuhr'

// derSideboardbereich ist der Bereich, den `domain.BereichSideboard` schickt.
//
// **Nur dieser eine**: das Hauptdeck ist der Normalfall, und ein zweiter
// Name dafür stünde nur da, damit er dasteht.
export const derSideboardbereich = 'sideboard'

// Deckvorschauzeile ist eine Zeile, wie sie ins Deck käme.
export interface Deckvorschauzeile {
  // bereich ist `hauptdeck` oder `sideboard` — leer, wenn die Zeile gar nicht
  // ins Deck gehört.
  bereich: string

  menge: number

  // name, setcode und nummer sind, was dastand — nicht, was gefunden wurde.
  name: string
  setcode: string
  nummer: string

  // printing ist der Druck — **leer, solange die Zeile nicht eindeutig ist**.
  printing: string

  treffer: Listentreffer[]
  drucke: number

  // befund sagt, warum diese Zeile nicht übernommen wird.
  befund: string

  uebernehmbar: boolean
}

// Deckvorschau ist, was aus einer Liste würde — ohne dass etwas geschieht.
export interface Deckvorschau {
  deckID: number

  // stand ist der Abdruck des Decks, aus dem diese Vorschau entstand. Er geht
  // beim Übernehmen zurück; hat sich das Deck geändert, wird nichts gelegt.
  stand: string

  zeilen: Deckvorschauzeile[]
  befunde: string[]

  hauptdeck: number
  sideboard: number
}

// Deckvorschauantwort trägt die Vorschau samt ihrem Fehler.
export interface Deckvorschauantwort extends Umschlag {
  vorschau: Deckvorschau | null
}

// Deckeinfuhrzeile ist eine Zeile, für die entschieden ist.
export interface Deckeinfuhrzeile {
  printing: string
  menge: number
  bereich: string
}

// Deckeinfuhrbericht sagt, was gelegt wurde.
export interface Deckeinfuhrbericht {
  gelegt: number
  stueck: number

  // stand ist der Abdruck **nach** dem Legen.
  stand: string
}

// Deckeinfuhrantwort trägt den Bericht samt seinem Fehler.
export interface Deckeinfuhrantwort extends Umschlag {
  bericht: Deckeinfuhrbericht | null
}

// Deckeinfuhrapi sind die zwei Schritte.
export interface Deckeinfuhrapi {
  // DecklisteVorschau liest die Liste und zeigt, was ins Deck käme —
  // **ohne zu schreiben**.
  DecklisteVorschau: (
    deckID: number,
    text: string,
  ) => Promise<Deckvorschauantwort>

  // DecklisteUebernehmen legt die bestätigten Zeilen ins Deck.
  //
  // **`stand` kommt aus der Vorschau** und ist die Bedingung: hat sich das
  // Deck seitdem geändert, wird nichts gelegt.
  DecklisteUebernehmen: (
    deckID: number,
    stand: string,
    zeilen: Deckeinfuhrzeile[],
  ) => Promise<Deckeinfuhrantwort>
}
