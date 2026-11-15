// **Aus einem Set ein Deck** — FMT-04
// (#1693).
//
// Neben dem leeren Deck steht der zweite Anfang: die Karten eines Sets, jede
// mit derselben Menge. **Verwiesen, nicht kopiert** — das sagt der Satz über
// dem Knopf, denn es ist der Unterschied zu einer Deckliste.
import { useCallback, useId, useState } from 'react'

import { api } from '../fassade'
import { ganzzahlAus } from '../ganzzahl'
import { einsetzen } from '../sprache/einsetzen'
import { auspacken } from '../umschlag'
import { Aktionsknopf } from './Aktionsknopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { useTexte } from '../sprache/kontext'
import { useLaden } from '../haken/useLaden'
import { feldbefund } from '../meldung'
import { zaehlwort } from '../sprache/zaehlwort'

import type { DeckDTO } from '../apidecks'
import type { Deckstand } from '../haken/useDecks'

const feld =
  'rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Deckausset({ stand }: { stand: Deckstand }) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const befundID = useId()

  const [setID, setzeSetID] = useState('')
  const [menge, setzeMenge] = useState('1')
  const [befund, setzeBefund] = useState<string>()

  // **Nur Sets mit Karten** — aus einem leeren entstünde ein leeres Deck,
  // und das legt der Knopf daneben schon an.
  const sets = useLaden(
    useCallback(async () => {
      const antwort = await api().Sets()

      return auspacken(antwort, () => antwort.sets)
    }, []),
    [],
  )
  const mitKarten = (sets.daten ?? []).filter((satz) => satz.karten > 0)
  const gewaehlt = setID === '' ? (mitKarten[0]?.id ?? '') : setID

  const anlegen = async () => {
    const angelegt: { deck: DeckDTO | null } = { deck: null }

    // **Streng gelesen** (F03, #2245): „2.5" ging als Kommazahl an ein `int`
    // in Go und kam als Satz über JSON zurück, ein leeres Feld wurde null.
    // Ob die Zahl reicht, sagt weiter Go — am Feld.
    const stueck = ganzzahlAus(menge)

    if (stueck === undefined) {
      setzeBefund(einsetzen(texte.allgemein.ganzeZahlAb, { kleinstens: '1' }))

      return
    }

    setzeBefund(undefined)

    await stand.fuehre(
      'deck-aus-set',
      async () => {
        const antwort = await api().DeckAusSet(gewaehlt, '', stueck)

        angelegt.deck = auspacken(antwort, () => antwort.deck)
      },
      (ursache) => {
        const amFeld = feldbefund(ursache, 'menge', texte)
        setzeBefund(amFeld)

        return amFeld !== undefined
      },
    )

    // **Das neue Deck geht auf** — wie beim leeren (Befund 3 des Zweitblicks
    // über Phase 5.1), und danach, nicht darin (§7, Regel 3).
    if (angelegt.deck !== null) {
      await stand.waehle(angelegt.deck.id)
    }
  }

  return (
    <section
      aria-label={eigene.ausSetTitel}
      className="mt-4 flex flex-col gap-3 border-t border-kante pt-4"
    >
      <h3 className="font-medium">{eigene.ausSetTitel}</h3>
      <Erklaerungsabsatz
        text={eigene.ausSetErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {sets.daten !== undefined && mitKarten.length === 0 ? (
        <p className="text-sm text-leise">{eigene.ausSetKeinSet}</p>
      ) : (
        <div className="flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-sm">{eigene.ausSetSet}</span>
            <select
              value={gewaehlt}
              onChange={(ereignis) => {
                setzeSetID(ereignis.target.value)
              }}
              className={feld}
            >
              {mitKarten.map((satz) => (
                <option key={satz.id} value={satz.id}>
                  {`${satz.name} · ${zaehlwort(satz.karten, texte.allgemein.zaehlwoerter.karten)}`}
                </option>
              ))}
            </select>
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-sm">{eigene.ausSetMenge}</span>
            <input
              type="number"
              min={1}
              value={menge}
              aria-invalid={befund !== undefined}
              aria-describedby={befund === undefined ? undefined : befundID}
              onChange={(ereignis) => {
                setzeMenge(ereignis.target.value)
                setzeBefund(undefined)
              }}
              className="w-24 rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
          </label>

          <Aktionsknopf
            aktion="anlegen"
            angehalten={stand.vorgang.laeuft('deck-aus-set') || gewaehlt === ''}
            aufKlick={() => {
              void anlegen()
            }}
          >
            {eigene.ausSetAnlegen}
          </Aktionsknopf>

          {befund !== undefined && (
            <div role="alert" className="w-full">
              <p id={befundID} className="text-sm text-fehler">
                {befund}
              </p>
            </div>
          )}
        </div>
      )}
    </section>
  )
}
