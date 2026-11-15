// **⌘Z am Tisch meint den Tisch** — P8
// (#2376),
// Vermerk 0197.
//
// # Was er tut
//
// **Er meldet den Tisch als Ziel von ⌘Z**, solange der Tisch dasteht
// (`useRuecknahmeweg`). Bis dahin traf ⌘Z am Tisch die
// Rücknahme des **Sets** — eine Kartenänderung, die niemand sah.
//
// # Was er nicht tut
//
// **Entscheiden, ob der Zug zurückgeht.** Er stellt `spieler.zuruecknehmen`,
// wenn der Kern sie anbietet; ob sie wirkt, sagt der Kern, und eine
// Abweisung steht wie jede am Tisch. Ohne Angebot tut ⌘Z hier nichts — und
// greift auch nicht auf das Set durch. **Wiederherstellen gibt es nicht.**
import { useMemo } from 'react'

import { useRuecknahmeweg } from '../ruecknahme'

import type { Partielauf } from './usePartie'

export const Zuruecknehmen = 'spieler.zuruecknehmen'

// useTischruecknahme meldet ⌘Z an und sagt, ob der Kern sie anbietet.
//
// **Angemeldet, solange der Tisch dasteht, und nicht nur die Spalte**
// (Zweitblick über Phase 7, Befund 9): beim Sichtwechsel und hinter dem
// Vorhang geht die Spalte mit, und ⌘Z fiel dann auf das Set zurück. Jetzt
// meldet sich der Tisch an; ist er nicht `aktiv`, tut ⌘Z nichts.
export function useTischruecknahme(lauf: Partielauf, aktiv: boolean): boolean {
  const kann = aktiv && lauf.erlaubt.includes(Zuruecknehmen)
  const { befehlen } = lauf

  const weg = useMemo(
    () => ({
      zuruecknehmen: () => {
        if (kann) {
          befehlen(Zuruecknehmen, {})
        }
      },
      wiederherstellen: () => undefined,
    }),
    [kann, befehlen],
  )

  useRuecknahmeweg(weg)

  return kann
}
