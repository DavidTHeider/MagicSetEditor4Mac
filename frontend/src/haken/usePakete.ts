// **Die Pakete, einmal gelesen** — AP-9, Teil 2
// (#2366).
//
// Die Leiste über den Paketen und die Liste darunter nennen dieselben
// Pakete; **die Seite liest sie**, wie seit U14 die Abhängigkeitsbefunde
// (#2159) — zwei Rufe für dieselbe Frage gaben zwei Stände.
import { useCallback } from 'react'

import { api } from '../fassade'
import type { PaketeAntwort } from '../pakete'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'

// usePakete liest nach, wenn `runde` steigt — der Stand bleibt dabei stehen.
export function usePakete(runde: number): Ladezustand<PaketeAntwort> {
  return useLaden(
    useCallback(async () => {
      const antwort = await api().Pakete()

      return auspacken(antwort, () => antwort)
    }, []),
    [],
    false,
    runde,
  )
}
