// **Was ein Markenfilter trifft, sagt Go** — UX-12.AK-03
// (#1684).
//
// Der Haken hält die Wahl einer Sicht, liest die Marken je Objekt für die
// Liste der eigenen Marken und fragt `MarkenFiltern`, sobald eine Marke
// gewählt ist. **Er filtert nicht selbst**: „und" und „oder" haben ihre
// Bedeutung an einer Stelle (`domain.Markenfilter`), und die gilt an jeder
// Fläche.
import { useCallback, useState } from 'react'

import { api } from '../api'
import type { Markenart } from '../apiobjektmarken'
import type { Markenwahl } from '../bausteine/Markenfilter'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

export interface Markenfilterstand {
  filter: Markenwahl
  setzeFilter: (filter: Markenwahl) => void
  je: Readonly<Record<string, readonly string[]>>

  // getroffen sind die Kennungen, die durch den Filter gehen — undefined,
  // solange keine Marke gewählt ist.
  getroffen: ReadonlySet<string> | undefined
  fehler: string | undefined
  nachladen: () => void

  // runde zählt jedes Nachladen — die Pflege liest ihren Katalog dann neu.
  runde: number
}

// **Ohne Art fragt der Haken nichts** — eine Liste ohne Marken ruft ihn
// trotzdem, weil ein Haken nicht bedingt gerufen werden darf.
export function useMarkenfilter(art: Markenart | undefined): Markenfilterstand {
  const [filter, setzeFilter] = useState<Markenwahl>({
    marken: [],
    verknuepfung: 'und',
  })
  const [runde, setzeRunde] = useState(0)

  const je = useLaden(async () => {
    if (art === undefined) {
      return {}
    }

    const antwort = await api().MarkenJeObjekt(art)

    return auspacken(antwort, () => antwort.marken)
  }, [art, runde])

  const getroffen = useLaden(async () => {
    if (filter.marken.length === 0 || art === undefined) {
      return undefined
    }

    const antwort = await api().MarkenFiltern(
      art,
      filter.marken,
      filter.verknuepfung,
    )

    return new Set(auspacken(antwort, () => antwort.kennungen))
  }, [art, runde, filter.marken.join('\u0000'), filter.verknuepfung])

  const nachladen = useCallback(() => {
    setzeRunde((vorher) => vorher + 1)
  }, [])

  return {
    filter,
    setzeFilter,
    je: je.daten ?? {},
    getroffen: filter.marken.length === 0 ? undefined : getroffen.daten,
    fehler: je.fehler ?? getroffen.fehler,
    nachladen,
    runde,
  }
}

// nachMarken lässt stehen, was der Filter trifft — und alles, solange keiner
// gilt. **Eine Kennung, die `undefined` ist, bleibt stehen**: eine
// Gruppenzeile im Paketbaum trägt keine Marke und soll nicht verschwinden.
export function nachMarken<T>(
  eintraege: readonly T[],
  kennung: (eintrag: T) => string | undefined,
  getroffen: ReadonlySet<string> | undefined,
): T[] {
  return eintraege.filter((eintrag) => {
    const welche = kennung(eintrag)

    return (
      getroffen === undefined || welche === undefined || getroffen.has(welche)
    )
  })
}
