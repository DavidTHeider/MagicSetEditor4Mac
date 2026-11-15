// **Die Übersicht über die Sammlung, einmal gefragt** — AP-9
// (#2366).
//
// Die Leiste über der Sammlung und die Übersicht im Navigator nennen
// dieselben Summen; **sie lesen dieselbe Antwort**, damit zwei Zahlen über
// denselben Wert nicht auseinanderlaufen. Gelesen wird, was im Bestand liegt
// — kein Preis wird dafür geholt (§16, Regel 5).
import { useCallback } from 'react'

import type { Portfolioantwort } from '../api'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'

// usePortfolio fragt neu, wenn `runde` steigt — nach einer Änderung an der
// Sammlung.
export function usePortfolio(runde: number): Ladezustand<Portfolioantwort> {
  return useLaden(
    useCallback(async () => {
      const antwort = await api().Portfolio()

      return auspacken(antwort, () => antwort)
    }, []),
    [runde],
  )
}
