// **Der Rückweg von einem Exemplar zu seinen Decks** — Werkstück U9 von
// Phase 5.1 (#1420).
//
// **Wer ein Stück Papier aus der Sammlung nimmt, soll sehen, was daran
// hängt.** Die Zuordnung ist eine Entscheidung über dieses Exemplar, und sie
// steht sonst nur beim Deck — von dort aus ist sie nicht zu finden.
//
// **Über `useLaden` und keinen eigenen Effekt** (§7, Regel 2): der Haken
// verwirft eine überholte Antwort und **fängt den Fehler**. Der erste Anlauf
// hatte einen eigenen Effekt mit `void laden()`; wo die Bindung nicht belegt
// war, blieb eine abgewiesene Zusage ohne Empfänger stehen — die Läufe waren
// grün, und `vitest` endete trotzdem rot.
import { useCallback } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'

export function Exemplardecks({ exemplarID }: { exemplarID: number }) {
  const eigene = useTexte().seiten.sammlung

  const holen = useCallback(async () => {
    const antwort = await api().ExemplarDecks(String(exemplarID))

    return auspacken(antwort, () => antwort.decks)
  }, [exemplarID])

  const geladen = useLaden(holen, [exemplarID])

  if (geladen.fehler !== undefined) {
    return (
      <p role="status" className="text-sm text-leise">
        {geladen.fehler}
      </p>
    )
  }

  const decks = geladen.daten

  // **Kein Kasten über null Decks** — ein leerer Titel sähe aus wie eine
  // Aussage und wäre keine.
  if (decks === undefined || decks.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-1">
      <span className="text-sm font-medium">{eigene.inDecks}</span>
      <ul aria-label={eigene.inDecks} className="flex flex-col gap-1">
        {decks.map((deck) => (
          <li key={deck.deckID} className="text-sm text-leise tabular-nums">
            {deck.menge} × {deck.name}
          </li>
        ))}
      </ul>
    </div>
  )
}
