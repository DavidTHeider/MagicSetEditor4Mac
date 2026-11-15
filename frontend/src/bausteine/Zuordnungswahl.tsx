// **Die Wahl eines Exemplars für einen Deckplatz** — Werkstück U9 von
// Phase 5.1 (#1420).
//
// **Gewählt wird aus der Sammlung, nicht getippt**: eine Kennung von Hand
// träfe jede Zeile oder keine. Und **ein anderer Druck ist erlaubt** — er
// wird als Abweichung geführt, weil Namensgleichheit keine Identität ist.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { Zahlenfeld } from './Zahlenfeld'
import { useTexte } from '../sprache/kontext'

import type { DeckDTO, ZuordnungszeileDTO } from '../apidecks'
import type { ExemplarDTO } from '../api'

export function Zuordnungswahl({
  deck,
  zeile,
  exemplare,
  angehalten,
  aufZuordnen,
}: {
  deck: DeckDTO
  zeile: ZuordnungszeileDTO
  exemplare: readonly ExemplarDTO[]
  angehalten: boolean
  aufZuordnen: (platzID: string, exemplarID: string, menge: number) => void
}) {
  const eigene = useTexte().seiten.decks
  const [exemplarID, setzeExemplar] = useState('')
  const [menge, setzeMenge] = useState(1)
  const [mengeGilt, setzeMengeGilt] = useState(true)

  // **Nur die Plätze dieser Zeile** — sie nennen denselben Druck.
  const plaetze = deck.plaetze.filter((platz) =>
    platz.printing === ''
      ? platz.druckID !== '' && platz.druckID === zeile.druckID
      : platz.printing === zeile.printing,
  )

  const [platzID, setzePlatz] = useState('')

  // **Über den ersten Platz und nicht über die Länge**: so weiß der Typ, dass
  // es ihn gibt, und `plaetze[0]?.id ?? ''` wäre ein Zweig, den kein Lauf
  // betreten kann (§3).
  const ersterPlatz = plaetze[0]

  if (ersterPlatz === undefined) {
    return null
  }

  const gewaehlterPlatz = platzID === '' ? ersterPlatz.id : platzID

  return (
    <div className="flex flex-wrap items-end gap-2">
      {plaetze.length > 1 && (
        <label className="flex flex-col gap-1">
          <span className="text-sm font-medium">{eigene.zuordnungPlatz}</span>
          <select
            value={gewaehlterPlatz}
            onChange={(ereignis) => {
              setzePlatz(ereignis.target.value)
            }}
            className="rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {plaetze.map((platz) => (
              <option key={platz.id} value={platz.id}>
                {platz.bereich === 'sideboard'
                  ? eigene.sideboard
                  : eigene.hauptdeck}
              </option>
            ))}
          </select>
        </label>
      )}

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.zuordnungExemplar}</span>
        <select
          value={exemplarID}
          onChange={(ereignis) => {
            setzeExemplar(ereignis.target.value)
          }}
          className="rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          <option value="">{eigene.zuordnungExemplarWaehlen}</option>
          {exemplare.map((exemplar) => (
            <option key={exemplar.id} value={String(exemplar.id)}>
              {exemplar.bezeichnung}
            </option>
          ))}
        </select>
      </label>

      {/* **Eine ganze Zahl ab eins, und leer ist keine** (F03, #2245): bis
          dahin wurden ein leeres Feld und „0" still zu einer Eins, und „2.5"
          ging an ein `int`. */}
      <Zahlenfeld
        beschriftung={eigene.zuordnungMenge}
        wert={menge}
        kleinstens={1}
        ganzzahlig
        geaendert={setzeMenge}
        gilt={setzeMengeGilt}
      />

      <Knopf
        bezeichnung={`${eigene.zuordnen}: ${zeile.bezeichnung}`}
        angehalten={angehalten || exemplarID === '' || !mengeGilt}
        aufKlick={() => {
          aufZuordnen(gewaehlterPlatz, exemplarID, menge)
        }}
      >
        {eigene.zuordnen}
      </Knopf>
    </div>
  )
}
