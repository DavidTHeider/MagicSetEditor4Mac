// **Die Angaben einer eigenen Übersetzung** — LOC-02.AK-03 und AK-04
// (#1688).
//
// Name, Stand, Lizenz und Herkunft — was mit der Datei hinausgeht. **Die
// Lizenz ist Pflicht zum Ausgeben**; der Hinweis darunter sagt es, bevor es
// jemand versucht.
import { useState } from 'react'

import { useTexte } from '../sprache/kontext'
import { Aktionsknopf } from './Aktionsknopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { SprachpaketDTO, SprachpaketangabenDTO } from '../apisprache'
import type { Sprachbefund } from '../haken/useSprachpakete'

// **Eine Höhe für Feld und Auswahl** — WebKit zeichnet ein `select` sonst
// niedriger, und seine Beschriftung rutschte aus der Reihe.
const feld =
  'h-9 rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

// angabenfelder sind die Felder hier — ein Befund dazu steht darunter, und
// einer ohne Feld auch: im Editor gibt es keinen anderen Ort für ihn.
const angabenfelder = ['name', 'stand', 'lizenz', 'herkunft']

export function Uebersetzungsangaben({
  paket,
  angehalten,
  befund,
  sichern,
}: {
  paket: SprachpaketDTO
  angehalten: boolean
  befund: Sprachbefund | undefined
  sichern: (angaben: SprachpaketangabenDTO) => void
}) {
  const eigene = useTexte().seiten.einstellungen.eigenuebersetzung
  const [name, setzeName] = useState(paket.name)
  const [stand, setzeStand] = useState(paket.stand)
  const [lizenz, setzeLizenz] = useState(paket.lizenz)
  const [herkunft, setzeHerkunft] = useState(paket.herkunft)

  return (
    <div className="flex flex-col gap-2 rounded-md border border-kante p-3">
      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.name}
          <input
            value={name}
            onChange={(ereignis) => {
              setzeName(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          {eigene.stand}
          <select
            value={stand}
            onChange={(ereignis) => {
              setzeStand(ereignis.target.value)
            }}
            className={feld}
          >
            {Object.entries(eigene.staende).map(([wert, bezeichnung]) => (
              <option key={wert} value={wert}>
                {bezeichnung}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-sm">
          {eigene.lizenz}
          <input
            value={lizenz}
            onChange={(ereignis) => {
              setzeLizenz(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          {eigene.herkunft}
          <input
            value={herkunft}
            onChange={(ereignis) => {
              setzeHerkunft(ereignis.target.value)
            }}
            className={feld}
          />
        </label>
        <Aktionsknopf
          aktion="sichern"
          angehalten={angehalten}
          aufKlick={() => {
            sichern({ name, stand, lizenz, herkunft })
          }}
        >
          {eigene.angabenSichern}
        </Aktionsknopf>
      </div>
      <Erklaerungsabsatz
        text={eigene.lizenzHinweis}
        klasse="text-sm text-leise"
      />
      {befund !== undefined && [...angabenfelder, ''].includes(befund.feld) && (
        <p role="alert" className="text-sm text-fehler">
          {befund.text}
        </p>
      )}
    </div>
  )
}
