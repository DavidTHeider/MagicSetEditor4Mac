// **Die Schritte des Verlaufs laden** — Werkstück U6 von Phase 5.1
// (#1417).
//
// `verlaufsstand` liegt in den Abhängigkeiten: er steigt nach jeder Bewegung
// und nach jeder eigenen Änderung, und die Liste lädt dann neu — **was sich
// geändert hat, steht im Bestand** (§7, Regel 2 über `useLaden`).
import { useCallback } from 'react'

import { api, auspacken, type VerlaufsschrittDTO } from '../api'
import { useLaden, type Ladezustand } from './useLaden'

export function useVerlaufsliste(
  setID: string,
  verlaufsstand: number,
): Ladezustand<VerlaufsschrittDTO[]> {
  const holen = useCallback(async () => {
    const antwort = await api().Verlauf(setID)

    return auspacken(antwort, () => antwort.schritte)
  }, [setID])

  return useLaden<VerlaufsschrittDTO[]>(holen, [setID], false, verlaufsstand)
}
