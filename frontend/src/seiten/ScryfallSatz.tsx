// **Ein ganzer Satz von Scryfall**
// (#1033).
//
// **Eine eigene Komponente**, wie der Weg für MTG JSON daneben, und aus
// demselben Grund: sie beantwortet eine eigene Frage — *welcher Satz, für
// welches Spiel*. Beides in der Setliste ließe die eine in der anderen
// verschwinden (§2).
import { useCallback, useState } from 'react'

import { api } from '../api'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface ScryfallSatzEigenschaften {
  laeuft: boolean
  einlesen: (code: string, spiel: string) => void
}

export function ScryfallSatz({ laeuft, einlesen }: ScryfallSatzEigenschaften) {
  const texte = useTexte().seiten.sets

  // **Dieselbe Spieleliste wie bei MTG JSON.** Sie ist die Menge der
  // Zuordnungen; eine zweite Kopie liefe beim ersten Nachziehen auseinander.
  const holen = useCallback(async () => api().MTGJSONSpiele(), [])

  const { daten, fehler } = useLaden<string[]>(holen, [])

  const [code, setzeCode] = useState('')
  const [gewaehltesSpiel, setzeSpiel] = useState('')

  // Ein Fehler beim Nachsehen nimmt nur diesen Weg — die Seite bleibt.
  if (fehler !== undefined) {
    return null
  }

  const spiele = daten ?? []
  const gewaehlt = spiele.includes(gewaehltesSpiel)
    ? gewaehltesSpiel
    : (spiele[0] ?? '')

  return (
    <section className="mt-6 border-t border-kante pt-4">
      <h2 className="text-lg font-semibold">
        {texte.scryfallsatzUeberschrift}
      </h2>

      <Erklaerungsabsatz
        text={texte.scryfallsatzErklaerung}
        klasse="mt-1 max-w-prose text-leise"
      />

      {spiele.length === 0 ? (
        <p className="mt-3 max-w-prose text-leise">
          {texte.scryfallsatzKeinSpiel}
        </p>
      ) : (
        <div className="mt-3 flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-sm">{texte.scryfallsatzCode}</span>
            <input
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              onChange={(ereignis) => {
                setzeCode(ereignis.target.value)
              }}
              placeholder={texte.scryfallsatzPlatzhalter}
              type="text"
              value={code}
            />
          </label>

          <label className="flex flex-col gap-1">
            <span className="text-sm">{texte.scryfallsatzSpiel}</span>
            <select
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              onChange={(ereignis) => {
                setzeSpiel(ereignis.target.value)
              }}
              value={gewaehlt}
            >
              {spiele.map((name) => (
                <option key={name} value={name}>
                  {name}
                </option>
              ))}
            </select>
          </label>

          <button
            className="rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
            disabled={laeuft || code.trim() === ''}
            onClick={() => {
              einlesen(code.trim(), gewaehlt)
            }}
            type="button"
          >
            {texte.scryfallsatzEinlesen}
          </button>
        </div>
      )}
    </section>
  )
}
