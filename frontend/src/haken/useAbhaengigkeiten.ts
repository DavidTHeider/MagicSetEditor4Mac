// **Die Abhängigkeiten der Paketseite, einmal gelesen** — U14
// (#2159).
//
// Die Zahl an jeder Paketzeile (`Paketliste`) und der Abschnitt
// „Abhängigkeiten" fragten je für sich; offen standen damit zwei Rufe für
// dieselbe Frage und zwei Stände — die Liste räumte bei jedem Nachladen, der
// Abschnitt behielt. **Die Seite liest einmal** und reicht beiden dasselbe;
// ein neuer Stand liest still nach (`useLaden` mit Stand, §7 Regel 2).
import { useCallback } from 'react'

import { api, auspacken, type AbhaengigkeitDTO } from '../api'
import { useLaden } from './useLaden'

export function useAbhaengigkeiten(stand: number) {
  const holen = useCallback(async () => {
    const antwort = await api().Abhaengigkeiten()

    return auspacken(antwort, () => antwort.befunde)
  }, [])

  return useLaden<AbhaengigkeitDTO[]>(holen, [], false, stand)
}
