// **Was die Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).
//
// **Gefragt wird mit den Zeilen, die dastehen** — auch mit denen eines
// Entwurfs, der noch nirgends gespeichert ist. Der Schlüssel ist ihr Inhalt:
// eine neue Zeile ist eine neue Frage, dasselbe Zeichnen keine.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { Handelsposition, Preisanhalt } from '../apihandelsleitung'

export function useHandelsanhalt(positionen: Handelsposition[]) {
  const schluessel = JSON.stringify(positionen)

  return useLaden<Preisanhalt | undefined>(async () => {
    if (positionen.length === 0) {
      return undefined
    }

    const antwort = await api().HandelPreisanhalt(positionen)

    return auspacken(antwort, () => antwort.anhalt)
    // **Der Inhalt ist der Schlüssel**, nicht die Liste: sie ist bei jedem
    // Zeichnen eine neue.
  }, [schluessel])
}
