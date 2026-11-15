// **Die Deckwahl beim Beitritt an einen fremden Tisch** — P9c
// (#2377),
// Artboard `Netzpartie09`: „Mit welchem Deck“.
//
// **Drei Arten**: ein Deck des Wirts (nichts mitbringen), ein eigenes Deck
// und eine angenommene Leihe, die sich spielen lässt. Vorgewählt ist das
// eigene Deck, mit dem zuletzt gearbeitet wurde (`spiel/mitbringen.ts`) —
// **erst, wenn beides geladen ist**; davor gibt es keine Vorwahl, und wer
// dann schon beitritt, bringt nichts mit.
import { api, auspacken } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { mitbringvorgabe } from '../spiel/mitbringen'
import { spielbar } from '../spiel/leihwahl'
import { leihwert } from '../spiel/sitzwahl'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorbelegung } from './useVorbelegung'

export interface Mitbringwahl {
  eintraege: { wert: string; name: string }[]
  vorgabe: string | undefined
}

export function useMitbringwahl(): Mitbringwahl {
  const eigene = useTexte().seiten.spielen

  const decks = useLaden(async () => {
    const antwort = await api().Decks('')

    return auspacken(antwort, () => antwort.decks)
  }, [])

  const leihen = useLaden(async () => {
    const antwort = await api().Deckleihen()

    return auspacken(antwort, () => antwort.leihen).filter(spielbar)
  }, [])

  const quellen = useVorbelegung()
  const eigeneDecks = decks.daten ?? []

  return {
    eintraege: [
      { wert: '', name: eigene.gastMitbringenWirt },
      ...eigeneDecks.map((deck) => ({ wert: deck.id, name: deck.name })),
      ...(leihen.daten ?? []).map((leihe) => ({
        wert: leihwert(leihe.kennung),
        name: einsetzen(eigene.sitzLeihe, {
          von: leihe.verleiher.slice(-8),
          karten: String(leihe.karten),
        }),
      })),
    ],
    vorgabe:
      decks.daten === undefined || quellen === undefined
        ? undefined
        : mitbringvorgabe(eigeneDecks, quellen.zuletzt),
  }
}
