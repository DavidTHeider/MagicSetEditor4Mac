// **Ob die letzte Sitzung sauber endete** — F04
// (#2343).
import type { Umschlag } from './umschlag'

// AbbruchDTO ist eine Sitzung, deren Startmarke beim nächsten Start noch
// lag.
export interface AbbruchDTO {
  // gestartet ist der Beginn der Sitzung — die Nullzeit von Go, wenn ihre
  // Marke sich nicht lesen ließ.
  gestartet: string
  fassung: string

  // mitSpur heißt: Go schrieb beim Absturz etwas. Ohne Spur war es ein Ende
  // von außen — ein Abschuss, ein leerer Akku.
  mitSpur: boolean
}

export interface WiederanlaufAntwort extends Umschlag {
  abbrueche: AbbruchDTO[]
}

export interface Wiederanlaufapi {
  Wiederanlauf: () => Promise<WiederanlaufAntwort>
}
