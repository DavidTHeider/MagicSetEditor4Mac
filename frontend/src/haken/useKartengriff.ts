// **Klick und Doppelklick an einer Karte, in der Zeit** — P4
// (#2372).
// Was ein Griff bedeutet, steht in `spiel/kartengriff.ts`; hier, **wann** er
// wirkt.
//
// # Ein wartender Klick je Tisch
//
// **Hat ein Objekt eine Zweithandlung, wartet sein erster Klick die Frist
// ab.** Kommt ein zweiter, gilt die Zweithandlung, und der wartende fällt
// weg; greift jemand inzwischen eine andere Karte, wirkt der wartende sofort
// — er ginge sonst verloren. **Nach dem Abbau stellt keiner mehr etwas**: der
// Tisch, an dem geklickt wurde, ist nicht mehr da (§7, Regel 1). **Und nach
// einem Wechsel von `stand` auch nicht** (Zweitblick über Phase 7, Befund
// 10): Sicht und Vorhang bauen den Tisch nicht ab, und der wartende Klick
// schriebe sonst für den vorigen Sitz.
import { useCallback, useEffect, useRef } from 'react'

import type { Objekthandlung } from '../apipartie'
import {
  doppelklickfrist,
  griffFuer,
  handlungZu,
  type Klick,
} from '../spiel/kartengriff'

interface Wartend {
  kennung: string
  befehl: string
  uhr: ReturnType<typeof setTimeout>
}

export function useKartengriff({
  handlungen,
  befehlen,
  waehle,
  stand,
}: {
  handlungen: readonly Objekthandlung[]
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  waehle: (kennung: string) => void

  // stand ist, wofür geklickt wird — Partie, Sicht und ob der Tisch bedienbar
  // ist; wechselt er, verfällt ein wartender Klick.
  stand: string
}): (kennung: string, klick: Klick) => void {
  const wartend = useRef<Wartend>(undefined)

  useEffect(
    () => () => {
      if (wartend.current !== undefined) {
        clearTimeout(wartend.current.uhr)
        wartend.current = undefined
      }
    },
    [stand],
  )

  return useCallback(
    (kennung: string, klick: Klick) => {
      const vorher = wartend.current

      // **Ein Klick auf eine andere Karte stellt den wartenden sofort.**
      if (vorher !== undefined && vorher.kennung !== kennung) {
        clearTimeout(vorher.uhr)
        wartend.current = undefined
        befehlen(vorher.befehl, { objekt: vorher.kennung })
      }

      const griff = griffFuer(handlungZu(handlungen, kennung), klick)

      switch (griff.art) {
        case 'waehlen':
          waehle(kennung)

          return

        case 'befehl':
          befehlen(griff.befehl, { objekt: kennung })

          return

        case 'warten':
          wartend.current = {
            kennung,
            befehl: griff.befehl,
            uhr: setTimeout(() => {
              wartend.current = undefined
              befehlen(griff.befehl, { objekt: kennung })
            }, doppelklickfrist),
          }

          return

        case 'zweit':
          if (wartend.current !== undefined) {
            clearTimeout(wartend.current.uhr)
            wartend.current = undefined
          }
          befehlen(griff.befehl, { objekt: kennung })

          return

        case 'nichts':
      }
    },
    [handlungen, befehlen, waehle],
  )
}
