// **Die Schalter über den Spalten eines Arbeitsplatzes** — herausgelöst aus
// `Arbeitsplatz.tsx` für P03
// (#2174),
// das dort an der Zeilengrenze aus §2 stand.
//
// **Ein Knopf über jeder Spalte, klein, und kein Satz dazwischen** — UXR-02
// (#1910): wie die Spaltenköpfe im Artboard. Bis dahin eine eigene Zeile in
// voller Größe, im Mindestfenster samt Satz in zwei umgebrochen; der Satz
// steht jetzt am Knopf (`Regionsknopf`).
import type { RefObject } from 'react'
import { createPortal } from 'react-dom'

import { Regionsknopf } from './Regionsknopf'

// Schalterstand ist, was ein Schalter von seiner Region weiß.
export interface Schalterstand {
  offen: boolean
  eng: boolean
  griff: RefObject<HTMLButtonElement | null>
  steuert: string
  angehalten: boolean
  klappen: () => void
}

interface SchalterEigenschaften {
  navigator?: Schalterstand | undefined
  inspector?: Schalterstand | undefined

  // ort ist, wo die Schalter stehen — P03: die Kartenseite stellt sie in die
  // Zeile mit „Karte anlegen", als Gruppe „Ansicht". `undefined` heißt die
  // eigene Zeile wie bisher, `null` einen angekündigten Ort, der noch nicht
  // da ist: dann stehen sie nirgends statt für einen Augenblick hier. **Sie
  // bleiben die Schalter dieses Arbeitsplatzes** — Zustand, Tasten und die
  // Rückgabe des Fokus ziehen nicht mit um, nur ihr Platz im Dokument.
  ort?: HTMLElement | null | undefined
}

export function Arbeitsplatzschalter({
  navigator,
  inspector,
  ort,
}: SchalterEigenschaften) {
  if (ort === null) {
    return null
  }

  const knopf = (region: 'navigator' | 'inspector', stand: Schalterstand) => (
    <Regionsknopf
      region={region}
      offen={stand.offen}
      eng={stand.eng}
      griff={stand.griff}
      steuert={stand.steuert}
      angehalten={stand.angehalten}
      aufKlick={stand.klappen}
    />
  )

  // **Am fremden Ort nebeneinander**, wie der Ort sie ordnet — die Gruppe
  // dort trägt Name und Umbruch.
  if (ort !== undefined) {
    return createPortal(
      <>
        {navigator !== undefined && knopf('navigator', navigator)}
        {inspector !== undefined && knopf('inspector', inspector)}
      </>,
      ort,
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      {navigator !== undefined && knopf('navigator', navigator)}

      {inspector !== undefined && (
        <div className="ml-auto">{knopf('inspector', inspector)}</div>
      )}
    </div>
  )
}
