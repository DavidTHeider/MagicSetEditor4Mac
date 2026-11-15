// **Die Bézierhandgriffe eines Punktes** —
// #890.
//
// **Ein eigenes Bauteil**, weil `Symbolflaeche.tsx` über die Zeilengrenze
// wuchs (§2). Der Schnitt folgt der Frage: dort **die Fläche und ihre
// Punkte**, hier **die Griffe**, die aus einer Ecke eine Rundung machen.
//
// **Nur am gewählten Teil**, so wie MSE es tut (`SymbolPointEditor::draw`,
// `src/gui/symbol/point_editor.cpp`, Fassung f9d9356 vom 2021-07-06): alle
// Griffe aller Teile zugleich wären ein Netz, in dem niemand den einen findet,
// den er meint.

import type { KeyboardEvent, PointerEvent } from 'react'

import type { SymbolpunktDTO } from '../antworten'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface SymbolgriffeEigenschaften {
  punkte: SymbolpunktDTO[]
  teil: number
  groesse: number
  schritt: number
  grosserSchritt: number

  aufsetzen: (
    ereignis: PointerEvent<SVGCircleElement>,
    teil: number,
    punkt: number,
    griff?: 'davor' | 'danach' | null,
  ) => void

  // ziehen setzt die **Griffspitze** auf eine Stelle im Einheitsquadrat.
  //
  // **Auch die Tastatur rechnet in Stellen**, obwohl ein Tastenschritt eine
  // Strecke ist: die Umrechnung geschieht dort, wo der Punkt vorliegt, und
  // damit hat die ganze Naht nur eine Sprache (#890).
  ziehen: (
    teil: number,
    punkt: number,
    davor: boolean,
    x: number,
    y: number,
  ) => void

  waehlen: (teil: number) => void
}

export function Symbolgriffe({
  punkte,
  teil,
  groesse,
  schritt,
  grosserSchritt,
  aufsetzen,
  ziehen,
  waehlen,
}: SymbolgriffeEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  // kurveDavorVon sagt, ob die Strecke **vor** einem Punkt eine Kurve ist.
  //
  // **Sie steht nicht am Punkt, sondern am vorigen**: eine Strecke hat eine
  // Art, und die schreibt der Punkt davor (`enforceConstraints`,
  // `src/data/symbol.cpp`, Zeilen 205–212, Fassung f9d9356 vom 2021-07-06).
  // Der Zug ist geschlossen, also ist der Vorgänger des ersten der letzte.
  //
  // **Ohne Rückfallwert**: die Modulorechnung gibt immer eine gültige Stelle,
  // und ein `?? false` daneben wäre ein Zweig, den kein Aufruf betritt (§3).
  const kurveDavorVon = (punkte: SymbolpunktDTO[], stelle: number) =>
    punkte.some(
      (kandidat, wo) =>
        wo === (stelle + punkte.length - 1) % punkte.length &&
        kandidat.kurveDanach,
    )

  // griffe zeichnet die beiden Handgriffe eines Punktes samt ihrer Linie.
  //
  // **Nur am gewählten Teil**, so wie MSE es tut (`SymbolPointEditor::draw`,
  // `src/gui/symbol/point_editor.cpp`, Fassung f9d9356 vom 2021-07-06): alle
  // Griffe aller Teile zugleich wären ein Netz, in dem niemand den einen
  // findet, den er meint.
  //
  // **Und nur, wo eine Kurve ist.** Ein Griff an einer Geraden zeigt auf eine
  // Rundung, die es nicht gibt.
  const griffePaar = (
    punkt: SymbolpunktDTO,
    punktstelle: number,
    kurveDavor: boolean,
  ) => (
    <g key={`griffe${String(punktstelle)}`}>
      {(
        [
          ['davor', kurveDavor, punkt.griffDavorX, punkt.griffDavorY],
          ['danach', punkt.kurveDanach, punkt.griffDanachX, punkt.griffDanachY],
        ] as const
      ).map(([seite, zeigen, dx, dy]) =>
        zeigen ? (
          <g key={seite}>
            <line
              x1={punkt.x * groesse}
              y1={punkt.y * groesse}
              x2={(punkt.x + dx) * groesse}
              y2={(punkt.y + dy) * groesse}
              className="stroke-leise"
              strokeWidth={1}
            />
            <circle
              cx={(punkt.x + dx) * groesse}
              cy={(punkt.y + dy) * groesse}
              r={4}
              tabIndex={0}
              role="button"
              aria-label={einsetzen(
                seite === 'davor' ? eigene.griffDavor : eigene.griffDanach,
                { nummer: String(punktstelle + 1) },
              )}
              onPointerDown={(ereignis) => {
                aufsetzen(ereignis, teil, punktstelle, seite)
              }}
              onKeyDown={(ereignis) => {
                griffTaste(ereignis, punktstelle, punkt, seite === 'davor')
              }}
              className="cursor-pointer fill-auswahl stroke-auswahlkante focus:outline-2 focus:outline-offset-2 focus:outline-ring"
              strokeWidth={2}
            />
          </g>
        ) : null,
      )}
    </g>
  )

  // griffTaste bewegt einen Griff mit den Pfeiltasten.
  //
  // **Dieselben Schritte wie am Punkt** — wer den einen bedienen kann, kann
  // den anderen.
  const griffTaste = (
    ereignis: KeyboardEvent<SVGCircleElement>,
    punktstelle: number,
    punkt: SymbolpunktDTO,
    davor: boolean,
  ) => {
    const weite = ereignis.shiftKey ? grosserSchritt : schritt

    const richtungen: Record<string, [number, number]> = {
      ArrowLeft: [-weite, 0],
      ArrowRight: [weite, 0],
      ArrowUp: [0, -weite],
      ArrowDown: [0, weite],
    }

    const richtung = richtungen[ereignis.key]
    if (richtung === undefined) {
      return
    }

    ereignis.preventDefault()
    waehlen(teil)

    const heute = davor
      ? { x: punkt.griffDavorX, y: punkt.griffDavorY }
      : { x: punkt.griffDanachX, y: punkt.griffDanachY }

    ziehen(
      teil,
      punktstelle,
      davor,
      punkt.x + heute.x + richtung[0],
      punkt.y + heute.y + richtung[1],
    )
  }

  return (
    <>
      {punkte.map((punkt, punktstelle) =>
        griffePaar(punkt, punktstelle, kurveDavorVon(punkte, punktstelle)),
      )}
    </>
  )
}
