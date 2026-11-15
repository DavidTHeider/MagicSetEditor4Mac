// **Die drei Quellen der Vorlagenwerkstatt** — Werkstück U7 von Phase 5.1
// (#1418):
// die Testkarten eines Pakets, seine Symbole und die Verwendungen eines
// Namens. Alle drei sind Lesewege über `useLaden` — ein Paketwechsel räumt
// sie (§7, Regel 1), und ein überholter Lauf trägt sich nicht ein (Regel 2).
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { PaketsymbolDTO, ProbekarteDTO, SymbolverwendungDTO } from '../api'

export interface Probekarten {
  karten: ProbekarteDTO[]
  fehler: string | undefined
}

export function useProbekarten(paket: string): Probekarten {
  const stand = useLaden(async () => {
    if (paket === '') {
      return []
    }

    const antwort = await api().Probekarten(paket)

    return auspacken(antwort, () => antwort.karten)
  }, [paket])

  return { karten: stand.daten ?? [], fehler: stand.fehler }
}

export interface Paketsymbole {
  symbole: PaketsymbolDTO[]

  // namen sind die Symbolnamen — für die Vervollständigung des Felds.
  namen: string[]
}

// usePaketsymbole liest die Symbole neu, wenn `stand` wechselt — nach jedem
// Speichern, denn dann können Definitionen dazugekommen oder verschwunden sein.
export function usePaketsymbole(paket: string, stand: string): Paketsymbole {
  const geladen = useLaden(
    async () => {
      if (paket === '') {
        return []
      }

      const antwort = await api().Paketsymbole(paket)

      return auspacken(antwort, () => antwort.symbole)
    },
    [paket, stand],
    true,
  )

  const symbole = geladen.daten ?? []

  return { symbole, namen: symbole.map((symbol) => symbol.name) }
}

export interface Symbolverwendungen {
  verwendungen: SymbolverwendungDTO[]
  laedt: boolean
}

export function useSymbolverwendungen(
  paket: string,
  name: string,
): Symbolverwendungen {
  const stand = useLaden(async () => {
    if (name === '') {
      return []
    }

    const antwort = await api().Symbolverwendungen(paket, name)

    return auspacken(antwort, () => antwort.verwendungen)
  }, [paket, name])

  return { verwendungen: stand.daten ?? [], laedt: stand.laedt }
}
