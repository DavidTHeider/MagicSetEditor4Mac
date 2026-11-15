// **Die erklärten Auszählungsdimensionen eines Sets** —
// #1338.
//
// **Ein eigener Haken**, weil er zwei Ladewege hält: das Angebot hängt am Set,
// die Auszählung am Set **und** an der gewählten Dimension.
//
// **Die Wahl hält er seit UXR-08 nicht mehr selbst**
// (#1916):
// sie ist ein Eintrag der einen Wahl der Statistik, und die räumt sie beim
// Wechsel des Sets (§7, Regel 1). Zwei Stellen, die dieselbe Wahl räumen,
// liefen an dem Tag auseinander, an dem eine es anders macht.
import { useCallback } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'
import type {
  AuszaehldimensionDTO,
  AuszaehlgruppeDTO,
} from '../antwortenauszaehlung'

export interface Auszaehlgriff {
  dimensionen: readonly AuszaehldimensionDTO[]
  gruppen: readonly AuszaehlgruppeDTO[]
  befunde: readonly string[]
  laedt: boolean
  fehler: string | undefined
}

// useAuszaehlung lädt das Angebot eines Sets und zählt nach `gewaehlt` aus —
// leer heißt: keine Dimension gewählt, und dann wird nicht gefragt.
export function useAuszaehlung(setID: string, gewaehlt: string): Auszaehlgriff {
  const holenAngebot = useCallback(async () => {
    const antwort = await api().Auszaehldimensionen(setID)

    return auspacken(antwort, () => antwort.dimensionen)
  }, [setID])

  const holenGruppen = useCallback(async () => {
    // **Ohne Wahl wird nicht gefragt.** Eine Auszählung rechnet über jede
    // Karte des Sets; sie ungefragt zu fahren kostete bei 3.761 Karten
    // 25 Sekunden.
    if (gewaehlt === '') {
      return { gruppen: [], befunde: [] }
    }

    const antwort = await api().Auszaehlen(setID, gewaehlt)

    return auspacken(antwort, () => ({
      gruppen: antwort.gruppen,
      befunde: antwort.befunde,
    }))
  }, [setID, gewaehlt])

  const angebot = useLaden(holenAngebot, [setID])
  const auszaehlung = useLaden(holenGruppen, [setID, gewaehlt])

  return {
    dimensionen: angebot.daten ?? [],
    gruppen: auszaehlung.daten?.gruppen ?? [],
    befunde: auszaehlung.daten?.befunde ?? [],
    laedt: auszaehlung.laedt,
    fehler: auszaehlung.fehler,
  }
}
