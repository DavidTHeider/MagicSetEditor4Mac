// **Das Filterfeld über der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416),
// nach MSE2s `FilterCtrl` (`src/gui/control/filter_ctrl.cpp`, Fassung
// f9d9356): ein Textfeld mit Platzhalter, ein Löschknopf, der nur bei Inhalt
// da ist, und Escape leert. Die Hilfe zur Sprache ist die Beschreibung des
// Feldes und steht seit AP-11 (#2368, B-27) hinter „?“ — dauerhaft
// vierzeilig unter dem Feld nahm sie der Kartenliste Platz.
import { useId } from 'react'

import { Feldhilfe } from '../bausteine/Feldhilfe'
import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { Schnellfilterstand } from '../haken/useSchnellfilter'

interface SchnellfilterEigenschaften {
  filter: Schnellfilterstand

  // gesamt und gezeigt füttern die Trefferzahl — sie kommt aus der Liste, die
  // den Filter anwendet, nicht aus der Antwort: was die Marken zusätzlich
  // wegnehmen, zählt mit.
  gesamt: number
  gezeigt: number
}

export function Schnellfilter({
  filter,
  gesamt,
  gezeigt,
}: SchnellfilterEigenschaften) {
  const worte = useTexte().seiten.karten.schnellfilter
  const kennung = useId()
  const hilfeID = `${kennung}-hilfe`
  const aktiv = filter.anfrage !== ''

  return (
    <div className="mb-2 flex flex-col gap-1">
      <div className="flex items-center gap-1">
        <label htmlFor={kennung} className="sr-only">
          {worte.bezeichnung}
        </label>
        <input
          id={kennung}
          type="search"
          value={filter.anfrage}
          placeholder={worte.platzhalter}
          aria-describedby={hilfeID}
          onChange={(ereignis) => {
            filter.setzeAnfrage(ereignis.target.value)
          }}
          onKeyDown={(ereignis) => {
            if (ereignis.key === 'Escape' && aktiv) {
              ereignis.preventDefault()
              filter.setzeAnfrage('')
            }
          }}
          className="min-w-0 flex-1 rounded-md border border-kante bg-flaeche px-2 py-1 text-sm placeholder:italic placeholder:text-leise focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
        {aktiv && (
          <button
            type="button"
            onClick={() => {
              filter.setzeAnfrage('')
            }}
            className="shrink-0 rounded-sm px-1 text-sm underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {worte.leeren}
          </button>
        )}
        <Feldhilfe zu={worte.bezeichnung} text={worte.hilfe} />
      </div>

      <p id={hilfeID} className="sr-only">
        {worte.hilfe}
      </p>

      {aktiv && filter.fehler === undefined && (
        <p role="status" className="text-xs text-leise">
          {einsetzen(worte.treffer, {
            gezeigt: String(gezeigt),
            gesamt: String(gesamt),
          })}
        </p>
      )}

      {filter.fehler !== undefined && (
        <Hinweiskasten art="fehler" meldung={filter.fehler} />
      )}
    </div>
  )
}
