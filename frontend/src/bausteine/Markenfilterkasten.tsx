// **Der Markenfilter samt seinen gespeicherten Filtern** — UX-12.AK-03
// (#1684).
//
// Verbindet `Markenfilter` mit der Einstellung `filter.gespeichert`, damit
// eine Sicht nur den Stand des Haken hereinreicht. **Ein Kasten und nicht
// jede Seite für sich**: die Sammlung, die Sets und die Decks lesen und
// schreiben dieselbe Liste.
import { useState } from 'react'

import type { Markenart } from '../apiobjektmarken'
import { useEinstellungen } from '../haken/useEinstellungen'
import type { Markenfilterstand } from '../haken/useMarkenfilter'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'
import { Markenfilter } from './Markenfilter'

// gespeicherteFilter ist der Schlüssel aus `domain.EinstellungGespeicherteFilter`.
export const gespeicherteFilter = 'filter.gespeichert'

export function Markenfilterkasten({
  art,
  stand,
}: {
  art: Markenart
  stand: Markenfilterstand
}) {
  const einstellungen = useEinstellungen()

  return (
    <Markenfilter
      art={art}
      je={stand.je}
      filter={stand.filter}
      setzeFilter={stand.setzeFilter}
      gespeichert={einstellungen.wert(gespeicherteFilter)}
      setzeGespeichert={(wert) => {
        einstellungen.setze(gespeicherteFilter, wert)
      }}
      getroffen={stand.getroffen}
    />
  )
}

// Markenfilterknopf zeigt den Kasten erst auf Wunsch — eine Liste, deren
// Kopf zwei Überschriften mehr trägt, liest sich schlechter, solange niemand
// nach Marken sucht.
export function Markenfilterknopf({
  art,
  stand,
}: {
  art: Markenart
  stand: Markenfilterstand
}) {
  const eigene = useTexte().bausteine.objektmarken
  const [offen, setzeOffen] = useState(false)

  return (
    <div className="flex flex-col gap-2 text-sm">
      <div>
        <Knopf
          gedrueckt={offen}
          aufKlick={() => {
            setzeOffen(!offen)
          }}
        >
          {eigene.werkzeugeUeberschrift}
        </Knopf>
      </div>
      {offen && <Markenfilterkasten art={art} stand={stand} />}
    </div>
  )
}
