// **Die Anfasser eines Teils auf der Zeichenfläche** — Punkte, Griffe und die
// Symmetrieanfasser.
//
// # Warum sie einen eigenen Durchgang haben
//
// **Weil MSE sie zuletzt zeichnet.** `SymbolViewer::draw` malt erst alle
// Teile und ruft danach `drawEditingHints(dc)` (`src/render/symbol/viewer.cpp`,
// Zeilen 109–116, Fassung `f9d9356` vom 2021-07-06).
//
// Das ist keine Feinheit: seit die Fläche **rückwärts** malt (`Teilumriss.tsx`),
// liegt das erste Teil der Liste oben — und ein Anfasser, der im selben `<g>`
// wie sein Umriss stünde, verschwände unter der Füllung des Teils, das danach
// gemalt wird. Werkstück 5 von Phase 3.7
// (#1168).

import type { KeyboardEvent, PointerEvent } from 'react'

import type { SymbolteilDTO, SymbolzugDTO } from '../antworten'
import { Symbolgriffe } from './Symbolgriffe'
import { Symmetriesicht } from './Symmetriesicht'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface SymbolanfasserEigenschaften {
  teil: SymbolteilDTO
  teilstelle: number

  // gewaehlt sagt, ob dieses Teil das ist, an dem gearbeitet wird.
  gewaehlt: boolean

  groesse: number
  schritt: number
  grosserSchritt: number

  kopien: SymbolzugDTO[][]
  mitteAuf: (x: number, y: number) => void
  symmetriegriffAuf: (x: number, y: number) => void
  griffAuf: (
    teil: number,
    punkt: number,
    davor: boolean,
    x: number,
    y: number,
  ) => void

  // **Die Ereignistypen sind die des Kreises**, nicht die von `Element`: die
  // Behandlungen in `Symbolflaeche` sind darauf geschrieben, und ein weiterer
  // Typ hier wäre eine zweite Antwort auf dieselbe Frage.
  zeigerRunter: (
    ereignis: PointerEvent<SVGCircleElement>,
    teilstelle: number,
    punktstelle: number,
    griff?: 'davor' | 'danach' | null,
  ) => void
  taste: (
    ereignis: KeyboardEvent<SVGCircleElement>,
    teilstelle: number,
    punktstelle: number,
  ) => void
  waehlen: (stelle: number) => void
}

export function Symbolanfasser({
  teil,
  teilstelle,
  gewaehlt,
  groesse,
  schritt,
  grosserSchritt,
  kopien,
  mitteAuf,
  symmetriegriffAuf,
  griffAuf,
  zeigerRunter,
  taste,
  waehlen,
}: SymbolanfasserEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  return (
    <g>
      {gewaehlt && teil.art === 'Symmetrie' && (
        <Symmetriesicht
          teil={teil}
          kopien={kopien}
          groesse={groesse}
          schritt={schritt}
          grosserSchritt={grosserSchritt}
          mitteAuf={mitteAuf}
          griffAuf={symmetriegriffAuf}
        />
      )}

      {gewaehlt && (
        <Symbolgriffe
          punkte={teil.punkte}
          teil={teilstelle}
          groesse={groesse}
          schritt={schritt}
          grosserSchritt={grosserSchritt}
          aufsetzen={zeigerRunter}
          ziehen={griffAuf}
          waehlen={waehlen}
        />
      )}

      {teil.punkte.map((punkt, punktstelle) => (
        <circle
          key={punktstelle}
          cx={punkt.x * groesse}
          cy={punkt.y * groesse}
          r={6}
          tabIndex={0}
          role="button"
          aria-label={einsetzen(eigene.punkt, {
            nummer: String(punktstelle + 1),
            name: teil.name,
          })}
          onKeyDown={(ereignis) => {
            taste(ereignis, teilstelle, punktstelle)
          }}
          onPointerDown={(ereignis) => {
            zeigerRunter(ereignis, teilstelle, punktstelle)
          }}
          onFocus={() => {
            waehlen(teilstelle)
          }}
          className="cursor-pointer fill-flaeche stroke-text focus:outline-2 focus:outline-offset-2 focus:outline-ring"
          strokeWidth={2}
        />
      ))}
    </g>
  )
}
