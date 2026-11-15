// **Was einem Deck fehlt, auf die Wunschliste** — DOM-03.AK-02
// (#1698).
//
// **Neben „Die N eintauschen“**: beides beantwortet dieselbe Frage — was
// fehlt, und wie komme ich daran. Der Tausch reist zum Handel, der Wunsch
// bleibt in der eigenen Sammlung.
//
// **Unter dem Schlüssel der Zuordnung** (§7, Regel 4): der Datensatz ist das
// Deck, und `key={deck.id}` an der Werkbank räumt den Baustein beim Wechsel
// ab — samt einer Antwort, die noch unterwegs ist (Regel 1 und 2).
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

import type { DeckDTO } from '../apidecks'
import type { Deckstand } from '../haken/useDecks'

export function Wunschlistenknopf({
  stand,
  deck,
  fehlend,
}: {
  stand: Deckstand
  deck: DeckDTO
  fehlend: number
}) {
  const eigene = useTexte().seiten.decks
  const [ergebnis, setzeErgebnis] = useState<{ neu: number; schon: number }>()

  const schluessel = `zuordnung:${deck.id}`

  // **Wenn nichts fehlt, gibt es nichts zu wünschen** — ein Knopf, der dann
  // dastünde, täte nichts.
  if (fehlend === 0) {
    return null
  }

  const wuenschen = async () => {
    await stand.fuehre(schluessel, async () => {
      const antwort = await api().DeckFehlkartenWuenschen(deck.id)
      const gewuenscht = auspacken(antwort, () => antwort)

      setzeErgebnis({
        neu: gewuenscht.neu.length,
        schon: gewuenscht.vorhanden.length,
      })
    })
  }

  return (
    <div className="flex flex-col items-start gap-1">
      <Knopf
        angehalten={stand.vorgang.laeuft(schluessel)}
        aufKlick={() => {
          void wuenschen()
        }}
      >
        {eigene.aufWunschliste}
      </Knopf>

      {ergebnis !== undefined && (
        <p role="status" className="text-sm text-leise">
          {einsetzen(eigene.wunschlisteErgebnis, {
            neu: String(ergebnis.neu),
            schon: String(ergebnis.schon),
          })}
        </p>
      )}
    </div>
  )
}
