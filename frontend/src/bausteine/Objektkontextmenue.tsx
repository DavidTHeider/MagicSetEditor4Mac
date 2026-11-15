// **Das Kontextmenü des gewählten Objekts, an einer anderen Stelle** — GESTE
// (#2358).
//
// Die Handlungen einer Karte entstehen bei ihrer Zeile (`Kartenhandlungen`),
// gemeint ist im Rechtsklick aber ihre Vorschau — ein anderer Zweig des
// Baums. Statt die Liste nach oben zu reichen (ein Zustand, der bei jedem
// Zeichnen neu käme), liest dieses Bauteil sie aus dem `Objekthandlungsraum`:
// **dieselbe Liste**, wie sie zuletzt gezeichnet war, mit Sperre und Grund.
//
// **Ausgelöst wird über den Raum** — mit den heutigen Rückrufen und denselben
// Sperren wie ein Klick in der Menüleiste. Und nur, wenn das gewählte Objekt
// das hier gezeigte ist: liegt die Vorschau einer anderen Karte darunter,
// bleibt der Rechtsklick bei macOS.
//
// **Der Rahmen steht immer**, auch ohne passende Liste — AP-5 (#2362). Sonst
// wechselte der Baum darunter mit jeder Wahl zwischen zwei Gestalten, und die
// Kartenliste verlöre bei jedem Klick Rollstand und Fokus.
import { use, type ReactNode } from 'react'

import { GewaehltesObjekt } from '../objekthandlungsraum'
import { Kontextmenue, type Anstoss } from './Kontextmenue'
import type { Menueeintrag } from './Menue'

interface Objektkontexteigenschaften {
  art: string
  kennung: string

  // klasse, anstoss, anstossVerbraucht und nurAngestossen gehen an
  // `Kontextmenue` durch.
  klasse?: string | undefined
  anstoss?: Anstoss | undefined
  anstossVerbraucht?: (() => void) | undefined
  nurAngestossen?: boolean

  children: ReactNode
}

export function Objektkontextmenue({
  art,
  kennung,
  klasse,
  anstoss,
  anstossVerbraucht,
  nurAngestossen,
  children,
}: Objektkontexteigenschaften) {
  const gewaehlt = use(GewaehltesObjekt)
  const passt =
    gewaehlt?.meldung.art === art && gewaehlt.meldung.kennung === kennung

  const eintraege = (passt ? gewaehlt.eintraege : []).map(
    (eintrag): Menueeintrag =>
      'aufWahl' in eintrag
        ? {
            ...eintrag,
            aufWahl: () => {
              gewaehlt?.ausloesen(eintrag.schluessel)
            },
          }
        : eintrag,
  )

  return (
    <Kontextmenue
      bezeichnung={passt ? gewaehlt.meldung.titel : ''}
      eintraege={eintraege}
      objekt={passt ? `${art}/${kennung}` : undefined}
      klasse={klasse}
      anstoss={anstoss}
      anstossVerbraucht={anstossVerbraucht}
      nurAngestossen={nurAngestossen}
    >
      {children}
    </Kontextmenue>
  )
}
