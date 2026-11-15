// **Was vom offenen Deck in der Sammlung liegt** — AP-9
// (#2366).
//
// # Gelesen, nicht auf Knopfdruck
//
// Bis AP-9 kam der Abgleich erst, wenn jemand in der Lasche „Prüfung“
// „Was davon da ist“ drückte. **Die Leiste über dem Deck nennt jetzt, was
// fehlt**, und führt in diese Lasche — dort muss es dann schon stehen. Der
// Abgleich liest nur den Bestand und geht nicht ins Netz; er fragt neu, wenn
// sich das Deck ändert (seine Fassung).
//
// # Einmal gefragt, an zwei Stellen gelesen
//
// **Die Seite fragt, Leiste und Lasche lesen dieselbe Antwort** — sonst
// könnten zwei Zahlen über dasselbe Deck auseinanderlaufen, wenn sich die
// Sammlung zwischen zwei Fragen ändert.
import { useCallback } from 'react'

import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'

// bestandsstand zählt fremde Schreibvorgänge — **auch eine geänderte
// Sammlung**, die die Fassung des Decks nicht ändert (Zweitblick über Phase 6
// von 0.9.0, zweiter Teil, Befund 3). Ein neuer Stand überholt eine Antwort,
// die noch unterwegs ist (`useLaden`).
export function useDeckabgleich(
  deck: DeckDTO | null,
  bestandsstand = 0,
): Ladezustand<HabenstandAntwort | null> {
  const deckID = deck?.id
  const fassung = deck?.fassung

  return useLaden(
    useCallback(async () => {
      if (deckID === undefined) {
        return null
      }

      const antwort = await api().DeckAbgleichen(deckID)

      return auspacken(antwort, () => antwort)
    }, [deckID]),
    [deckID, fassung, bestandsstand],
  )
}
