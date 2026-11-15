// **Ein Fach am Markttisch** — P12
// (#2380),
// Entwurf `Markttisch09`: Name, darunter Zustand, Sprache und Menge; eine
// Karte ist ein Knopf, der sich greifen, wählen und ziehen lässt.
//
// **Was es tut, steht im Haken** (`useMarktgriff`), hier nur, wie es
// aussieht. **Leere Fächer sagen den nächsten Schritt** (§6).
//
// # Ein großes Fach baut nicht alles auf
//
// **#2538**, gefunden mit dem Messlauf aus #2532: 25.000 eigene Exemplare als
// Knöpfe hielten die Seite beim Aufschlagen 23 bis 45 Sekunden an, während
// die Go-Seite 0,4 s brauchte. Über [FACHGRENZE] Karten zeigt das Fach die
// ersten, eine Suche nach dem Namen und „Mehr zeigen“. **Keine Fensterliste**
// wie in der Sammlung: dort ist eine Zeile ein Gitterfeld mit eigener
// Tastatur, hier ein Knopf, den man greift und zieht (`useMarktgriff`).
import { useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Fachkarte } from './marktlegen'

export interface Fachgriff {
  auswahl: readonly string[]
  klicken: (karte: Fachkarte, dazu: boolean) => void
  taste: (karte: Fachkarte, name: string) => boolean
  ziehen: (karte: Fachkarte, ereignis: React.PointerEvent) => void
  gegriffen: (karte: Fachkarte) => boolean
}

interface Eigenschaften {
  karten: readonly Fachkarte[] | undefined
  leer: string
  griff: Fachgriff
}

// FACHGRENZE ist, wie viele Karten ein Fach auf einmal zeigt — so viele wie
// ein Fenster der Sammlung (`sammlungssicht.ts`).
export const FACHGRENZE = 200

export function Fachliste({ karten, leer, griff }: Eigenschaften) {
  const texte = useTexte()
  const tisch = texte.seiten.handeln.markttisch
  const zustaende = texte.seiten.sammlung.zustaende
  const [suche, setzeSuche] = useState('')
  const [zahl, setzeZahl] = useState(FACHGRENZE)

  if (karten === undefined) {
    return <p className="text-sm text-leise">{tisch.laedt}</p>
  }

  if (karten.length === 0) {
    return <p className="text-sm text-leise">{leer}</p>
  }

  const gross = karten.length > FACHGRENZE
  const gesucht = suche.trim().toLowerCase()
  const treffer =
    gesucht === ''
      ? karten
      : karten.filter((karte) => karte.name.toLowerCase().includes(gesucht))
  const gezeigt = gross ? treffer.slice(0, zahl) : treffer

  return (
    <div className="flex flex-col gap-2">
      {gross && (
        <input
          type="search"
          aria-label={tisch.fachSuchen}
          placeholder={tisch.fachSuchen}
          value={suche}
          onChange={(ereignis) => {
            setzeSuche(ereignis.target.value)
            setzeZahl(FACHGRENZE)
          }}
          className="rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      )}
      {gross && treffer.length === 0 && (
        <p className="text-sm text-leise">{tisch.fachKeinTreffer}</p>
      )}
      <Kartenreihe karten={gezeigt} griff={griff} zustaende={zustaende} />
      {gross && treffer.length > 0 && (
        <p role="status" className="text-xs text-leise">
          {einsetzen(tisch.fachZaehlung, {
            gezeigt: String(gezeigt.length),
            gesamt: String(treffer.length),
          })}
        </p>
      )}
      {gezeigt.length < treffer.length && (
        <button
          type="button"
          onClick={() => {
            setzeZahl((bisher) => bisher + FACHGRENZE)
          }}
          className="self-start rounded-sm px-2 py-1 text-sm text-marke focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {tisch.fachMehr}
        </button>
      )}
    </div>
  )
}

// Kartenreihe sind die Knöpfe eines Fachs.
function Kartenreihe({
  karten,
  griff,
  zustaende,
}: {
  karten: readonly Fachkarte[]
  griff: Fachgriff
  zustaende: Record<string, string>
}) {
  const tisch = useTexte().seiten.handeln.markttisch

  return (
    <ul className="flex flex-col gap-1">
      {karten.map((karte) => (
        <li key={karte.kennung}>
          <button
            type="button"
            aria-label={karte.name}
            aria-pressed={griff.auswahl.includes(karte.kennung)}
            className={kartenklassen(griff.gegriffen(karte))}
            onClick={(ereignis) => {
              griff.klicken(karte, ereignis.metaKey)
            }}
            onKeyDown={(ereignis) => {
              if (griff.taste(karte, ereignis.key)) {
                ereignis.preventDefault()
              }
            }}
            onPointerDown={(ereignis) => {
              griff.ziehen(karte, ereignis)
            }}
          >
            <span className="block truncate text-sm">{karte.name}</span>
            <span className="block text-xs text-leise">
              {einsetzen(tisch.karte, {
                zustand: zustaende[karte.zustand] ?? karte.zustand,
                sprache: karte.sprache,
                menge: String(karte.menge),
              })}
            </span>
          </button>
        </li>
      ))}
    </ul>
  )
}

// kartenklassen sind die einer Karte — **gegriffen mit dem Fokusring** (gemessen in `KONTRASTE.md`), gewählt über
// `aria-pressed`.
function kartenklassen(gegriffen: boolean): string {
  return (
    'w-full cursor-grab touch-none select-none rounded-sm border border-kante px-2 py-1 text-left ' +
    'aria-pressed:bg-betont focus-visible:outline focus-visible:outline-2 ' +
    'focus-visible:outline-offset-2 focus-visible:outline-ring' +
    (gegriffen ? ' outline outline-2 outline-offset-2 outline-ring' : '')
  )
}
