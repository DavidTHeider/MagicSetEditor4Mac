// Der Weg für MTG JSON — Spielwahl und die beiden Dialoge.
//
// **Eine eigene Komponente**, weil sie eine eigene Frage beantwortet: *für
// welches Spiel* eine Datei gelesen wird. Die Setliste beantwortet eine andere
// (§2), und beides in einer Datei ließe die eine in der anderen verschwinden.
import { useCallback, useState } from 'react'

import { api } from '../api'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface MTGJSONOeffnenEigenschaften {
  laeuft: boolean
  oeffnen: (verzeichnis: boolean, spiel: string) => void
}

export function MTGJSONOeffnen({
  laeuft,
  oeffnen,
}: MTGJSONOeffnenEigenschaften) {
  const texte = useTexte()

  // **Die Liste der Spiele kommt aus dem Backend**, nicht aus einer Kopie hier.
  // Sie ist die Menge der Zuordnungen, und die wächst dort.
  const holen = useCallback(async () => {
    const spiele = await api().MTGJSONSpiele()

    return spiele
  }, [])

  const { daten, fehler } = useLaden<string[]>(holen, [])

  // **Die Wahl liegt hier und nicht in der Setliste.** Die Vorbelegung ergibt
  // sich erst, wenn die Liste da ist; wer sie oben hielte, reichte bis dahin
  // einen leeren Spielnamen ans Backend — und bekäme „für »« gibt es keine
  // Zuordnung".
  const [gewaehltesSpiel, setzeSpiel] = useState('')

  // Ein Fehler beim Nachsehen nimmt nur diesen Weg. Die Seite bleibt: eine
  // `.mse-set` lässt sich weiterhin öffnen, und eine halbe Seite wäre eine
  // schlechtere Auskunft als eine ganze mit einem fehlenden Angebot.
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
        {texte.seiten.sets.mtgjsonUeberschrift}
      </h2>

      <Erklaerungsabsatz
        text={texte.seiten.sets.mtgjsonErklaerung}
        klasse="mt-1 max-w-prose text-leise"
      />

      {spiele.length === 0 ? (
        <p className="mt-3 max-w-prose text-leise">
          {texte.seiten.sets.mtgjsonKeinSpiel}
        </p>
      ) : (
        <div className="mt-3 flex flex-wrap items-end gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-sm">{texte.seiten.sets.mtgjsonSpiel}</span>
            <select
              value={gewaehlt}
              onChange={(ereignis) => {
                setzeSpiel(ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {spiele.map((name) => (
                <option key={name} value={name}>
                  {name}
                </option>
              ))}
            </select>
          </label>

          <button
            type="button"
            disabled={laeuft}
            onClick={() => {
              oeffnen(false, gewaehlt)
            }}
            className="rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          >
            {texte.seiten.sets.mtgjsonDatei}
          </button>

          <button
            type="button"
            disabled={laeuft}
            onClick={() => {
              oeffnen(true, gewaehlt)
            }}
            className="rounded-sm border border-kante px-3 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          >
            {texte.seiten.sets.mtgjsonVerzeichnis}
          </button>
        </div>
      )}
    </section>
  )
}
