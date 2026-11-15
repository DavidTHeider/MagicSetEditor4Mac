// **Welches Feld einen KI-Vorschlag bekommt** — AI-05.AK-06
// (#2042).
// Name und Flavortext, und zwar **mit einem `key` je Karte und Feld**: ein
// Kartenwechsel räumt den offenen Vorschlag mit ab (§7, Regel 1 und 2).
import type { ReactNode } from 'react'

import { KIFeldvorschlag } from '../bausteine/KIFeldvorschlag'
import { istKleinesFeld } from '../bausteine/kifeldauftrag'

import type { FeldbeschreibungDTO, KarteDTO } from '../api'

export function feldvorschlag(
  karte: KarteDTO,
  feld: string,
  nachName: ReadonlyMap<string, FeldbeschreibungDTO>,
  geaendert: () => void,
): ReactNode | undefined {
  if (!istKleinesFeld(feld)) {
    return undefined
  }

  return (
    <KIFeldvorschlag
      key={`${karte.id}/${feld}`}
      karteID={karte.id}
      feld={feld}
      aufschrift={nachName.get(feld)?.aufschrift ?? feld}
      werte={karte.werte}
      formen={nachName}
      geaendert={geaendert}
    />
  )
}
