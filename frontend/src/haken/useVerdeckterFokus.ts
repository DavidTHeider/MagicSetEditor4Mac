// **Ein Fokus unter der Spalte schließt sie** — N12 aus dem Nachlauf über
// Phase 6.3 (#1664), verlagert in Phase 6.11 (#1672).
//
// # Der Zielkonflikt, und wie er entschieden ist
//
// Aus der Spalte über dem Zentrum führte Tab in Bedienelemente, die unter
// ihr lagen: der Fokus stand auf etwas, das niemand sah (WCAG 2.4.11), und
// Escape schloss die Spalte nicht mehr, weil er nicht in ihr stand. **Das
// Zentrum `inert` zu machen oder die Spalte bei jedem Verlassen zu
// schließen, nähme M1 zurück**: ein sichtbares Feld daneben gehört dem
// Menschen, der mit offener Spalte darin arbeitet, und sein Escape dem Feld.
//
// **Also schließt sie nur, wenn der Fokus auf etwas geht, das sie verdeckt**
// ([verdecktVon]) — der Fokus bleibt dort, und jetzt sieht man ihn.
import { useEffect, type RefObject } from 'react'

import { ueberZentrum, verdecktVon } from '../bausteine/ueberzentrum'

export function useVerdeckterFokus(
  rahmen: RefObject<HTMLElement | null>,
  offen: boolean,
  schliessen: () => void,
): void {
  useEffect(() => {
    const element = rahmen.current

    if (!offen || element === null) {
      return undefined
    }

    const aufFokus = (ereignis: FocusEvent) => {
      const ziel = ereignis.target
      const spalte = element.querySelector(ueberZentrum)

      if (
        ziel instanceof Element &&
        spalte !== null &&
        !spalte.contains(ziel) &&
        verdecktVon(ziel, spalte)
      ) {
        schliessen()
      }
    }

    element.addEventListener('focusin', aufFokus)

    return () => {
      element.removeEventListener('focusin', aufFokus)
    }
  }, [rahmen, offen, schliessen])
}
