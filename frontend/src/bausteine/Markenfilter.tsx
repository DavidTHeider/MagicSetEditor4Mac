// **Eigene Marken und gespeicherte Filter im Navigator** — UX-12.AK-03
// (#1684),
// Artboard *Sammlung*.
//
// # Was hier nicht entschieden wird
//
// **Was „und" und „oder" heißt** — das beantwortet Go (`MarkenFiltern`, im
// Haken `useMarkenfilter`). Dieser Baustein zeigt die Wahl und nimmt sie
// entgegen; welche Objekte durchgehen, bekommt er fertig herein.
//
// # Gespeichert wird als Einstellung
//
// `filter.gespeichert`, eine Liste über alle Sichten (Vermerk 0135). Damit
// überlebt ein Filter den Neustart und reist mit der Ausfuhr der
// Einstellungen. **Geschrieben wird die ganze Liste** — die Filter der
// anderen Sichten gehen unverändert mit.
import { useId, useState } from 'react'

import type { Markenart, Verknuepfung } from '../apiobjektmarken'
import { einsetzen } from '../sprache/einsetzen'
import { Aktionsknopf } from './Aktionsknopf'
import { useTexte } from '../sprache/kontext'

// Markenwahl ist die Frage, die gerade gestellt ist.
export interface Markenwahl {
  marken: string[]
  verknuepfung: Verknuepfung
}

interface GespeicherterFilter {
  name: string
  sicht: string
  marken: string[]
  verknuepfung: Verknuepfung
}

interface MarkenfilterEigenschaften {
  art: Markenart

  // je sind die Marken jedes Objekts dieser Art — aus ihnen entsteht die
  // Liste der eigenen Marken.
  je: Readonly<Record<string, readonly string[]>>

  filter: Markenwahl
  setzeFilter: (filter: Markenwahl) => void

  // gespeichert ist der Wert der Einstellung, setzeGespeichert schreibt ihn.
  gespeichert: string
  setzeGespeichert: (wert: string) => void

  // getroffen sind die Objekte, die durch den Filter gehen — undefined,
  // solange keiner gilt.
  getroffen: ReadonlySet<string> | undefined
}

// gespeicherteAus liest die Einstellung — **was sich nicht lesen lässt, gilt
// als keine**, und ein Eintrag, der nicht die Gestalt eines Filters hat, wird
// übergangen: Go hat die Liste beim Schreiben geprüft, und eine, die von Hand
// kaputtgeschrieben wurde, soll die Sicht nicht anhalten. Beim Speichern
// bleibt ein solcher Eintrag weg — mit ihm wiese Go die ganze Liste ab.
function gespeicherteAus(wert: string): GespeicherterFilter[] {
  try {
    const gelesen: unknown = JSON.parse(wert)

    return Array.isArray(gelesen) ? gelesen.filter(istFilter) : []
  } catch {
    return []
  }
}

function istFilter(eintrag: unknown): eintrag is GespeicherterFilter {
  if (typeof eintrag !== 'object' || eintrag === null) {
    return false
  }

  const { name, sicht, marken, verknuepfung } = eintrag as Record<
    string,
    unknown
  >

  return (
    typeof name === 'string' &&
    typeof sicht === 'string' &&
    Array.isArray(marken) &&
    marken.every((marke) => typeof marke === 'string') &&
    (verknuepfung === 'und' || verknuepfung === 'oder')
  )
}

// sichtbareMarken ist, wie viele Marken ungefaltet dastehen (#2017) — R2
// lässt zwanzig Bedienelemente im Bild, und der Navigator trägt mehr als
// diesen Filter.
export const sichtbareMarken = 8

export function Markenfilter({
  art,
  je,
  filter,
  setzeFilter,
  gespeichert,
  setzeGespeichert,
  getroffen,
}: MarkenfilterEigenschaften) {
  const eigene = useTexte().bausteine.objektmarken
  const feld = useId()
  const [name, setzeName] = useState('')
  const [alleZeigen, setzeAlleZeigen] = useState(false)

  const eigeneMarken = [...new Set(Object.values(je).flat())].sort()
  // **Gefaltet wird nur, was nicht gewählt ist** — eine gewählte Marke steht
  // immer da, sonst ließe sich ein Filter nicht zurücknehmen.
  const gefaltet = !alleZeigen && eigeneMarken.length > sichtbareMarken
  const gezeigt = gefaltet
    ? eigeneMarken.filter(
        (marke, stelle) =>
          stelle < sichtbareMarken || filter.marken.includes(marke),
      )
    : eigeneMarken
  const alle = gespeicherteAus(gespeichert)
  const hier = alle.filter((eintrag) => eintrag.sicht === art)

  const umschalten = (marke: string) => {
    setzeFilter({
      ...filter,
      marken: filter.marken.includes(marke)
        ? filter.marken.filter((gewaehlt) => gewaehlt !== marke)
        : [...filter.marken, marke],
    })
  }

  const speichern = () => {
    if (name.trim() === '') {
      return
    }

    setzeGespeichert(
      JSON.stringify([
        ...alle,
        {
          name: name.trim(),
          sicht: art,
          marken: filter.marken,
          verknuepfung: filter.verknuepfung,
        },
      ]),
    )
    setzeName('')
  }

  return (
    <div className="flex flex-col gap-3 text-sm">
      <section
        aria-label={eigene.filterUeberschrift}
        className="flex flex-col gap-1"
      >
        <h3 className="font-medium">{eigene.filterUeberschrift}</h3>
        {eigeneMarken.length === 0 ? (
          <p className="text-leise">{eigene.filterKeine}</p>
        ) : (
          <ul className="flex flex-wrap gap-1">
            {gezeigt.map((marke) => (
              <li key={marke}>
                <button
                  type="button"
                  aria-pressed={filter.marken.includes(marke)}
                  aria-label={einsetzen(eigene.filterWaehlen, { marke })}
                  onClick={() => {
                    umschalten(marke)
                  }}
                  className="rounded-full border border-kante px-2 py-0.5 aria-pressed:border-marke aria-pressed:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                >
                  {marke}
                </button>
              </li>
            ))}
          </ul>
        )}

        {eigeneMarken.length > sichtbareMarken && (
          <div>
            <button
              type="button"
              aria-expanded={alleZeigen}
              onClick={() => {
                setzeAlleZeigen(!alleZeigen)
              }}
              className="text-leise underline-offset-2 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {alleZeigen
                ? eigene.wenigerMarken
                : einsetzen(eigene.weitereMarken, {
                    zahl: String(eigeneMarken.length - sichtbareMarken),
                  })}
            </button>
          </div>
        )}

        {filter.marken.length > 0 && (
          <>
            <fieldset className="flex gap-3">
              <legend className="sr-only">{eigene.verknuepfung}</legend>
              {(['und', 'oder'] as const).map((verknuepfung) => (
                <label key={verknuepfung} className="flex items-center gap-1">
                  <input
                    type="radio"
                    name={`${feld}-verknuepfung`}
                    checked={filter.verknuepfung === verknuepfung}
                    onChange={() => {
                      setzeFilter({ ...filter, verknuepfung })
                    }}
                  />
                  {eigene[verknuepfung]}
                </label>
              ))}
            </fieldset>
            {getroffen !== undefined && (
              <p className="text-leise">
                {einsetzen(eigene.gefiltert, {
                  anzahl: String(getroffen.size),
                })}
              </p>
            )}
            <div>
              <button
                type="button"
                onClick={() => {
                  setzeFilter({ ...filter, marken: [] })
                }}
                className="rounded-sm border border-kante px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                {eigene.filterLeeren}
              </button>
            </div>
          </>
        )}
      </section>

      <section
        aria-label={eigene.gespeicherteUeberschrift}
        className="flex flex-col gap-1"
      >
        <h3 className="font-medium">{eigene.gespeicherteUeberschrift}</h3>
        {hier.length === 0 ? (
          <p className="max-w-prose text-leise">{eigene.gespeicherteKeine}</p>
        ) : (
          <ul className="flex flex-col gap-1">
            {hier.map((eintrag) => (
              <li
                key={eintrag.name}
                className="flex items-center justify-between gap-2"
              >
                <button
                  type="button"
                  aria-label={einsetzen(eigene.anwenden, {
                    name: eintrag.name,
                  })}
                  onClick={() => {
                    setzeFilter({
                      marken: eintrag.marken,
                      verknuepfung: eintrag.verknuepfung,
                    })
                  }}
                  className="text-left underline-offset-2 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                >
                  {eintrag.name}
                </button>
                <button
                  type="button"
                  aria-label={einsetzen(eigene.loeschen, {
                    name: eintrag.name,
                  })}
                  onClick={() => {
                    setzeGespeichert(
                      JSON.stringify(
                        alle.filter((anderer) => anderer !== eintrag),
                      ),
                    )
                  }}
                  className="rounded-full px-1 leading-none focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        )}

        {filter.marken.length > 0 && (
          <form
            className="flex gap-2"
            onSubmit={(ereignis) => {
              ereignis.preventDefault()
              speichern()
            }}
          >
            <label className="sr-only" htmlFor={feld}>
              {eigene.speichernName}
            </label>
            <input
              id={feld}
              value={name}
              onChange={(ereignis) => {
                setzeName(ereignis.target.value)
              }}
              className="min-w-0 flex-1 rounded-sm border border-kante bg-transparent px-2 py-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
            <Aktionsknopf
              aktion="sichern"
              klein
              angehalten={name.trim() === ''}
              aufKlick={speichern}
            >
              {eigene.speichern}
            </Aktionsknopf>
          </form>
        )}
      </section>
    </div>
  )
}
