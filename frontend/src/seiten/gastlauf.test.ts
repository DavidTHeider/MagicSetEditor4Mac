// **Was der Gast am fremden Tisch nicht kann, tut nichts** — der Vertrag von
// `gastalsLauf` (S12.NAT-04, #1672).
//
// Anfangen, Ansehen, Wechseln und Räumen gehören dem Wirt: der Gast sitzt an
// seinem Sitz und hält keine fremde Partie an. **Am Tisch steht dafür kein
// Bedienelement mehr** (`partiegastspiel.test.tsx`); dieser Lauf hält fest,
// dass ein Ruf von woanders ins Leere geht, statt etwas zu bewegen.
import { expect, it, vi } from 'vitest'

import { gastalsLauf } from './Partiegast'
import { eineAntwort } from './partieprobe'

import type { Gastlauf } from '../haken/usePartiegast'

it('lässt Anfangen, Ansehen, Wechseln und Räumen ins Leere gehen', () => {
  const wuenschen = vi.fn()
  const gast: Gastlauf = {
    beitreten: vi.fn(),
    aufstehen: vi.fn(),
    ansehen: vi.fn(),
    wuenschen,
    laeuft: false,
    wartend: 2,
    stand: undefined,
    tisch: undefined,
    zuege: [],
    luecke: false,
    fehler: undefined,
  }

  const lauf = gastalsLauf(gast, eineAntwort())

  lauf.anfangen({ spiel: 'magic', fuer: 'ada', sitze: [] } as never)
  lauf.ansehen('p1', 'bo')
  lauf.wechseln('pausieren')
  lauf.raeumen()

  expect(wuenschen).not.toHaveBeenCalled()
  expect(gast.ansehen).not.toHaveBeenCalled()

  // **Was wartet, reicht der Lauf weiter** (P5, #2373).
  expect(lauf.wartend).toBe(2)

  // **Und befehlen heißt wünschen** — dass der Lauf überhaupt durchreicht.
  lauf.befehlen('spieler.muenzwurf', {})
  expect(wuenschen).toHaveBeenCalledWith('spieler.muenzwurf', {})
})
