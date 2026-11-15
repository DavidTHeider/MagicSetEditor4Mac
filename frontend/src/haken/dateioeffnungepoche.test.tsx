// **Eine Finder-Meldung aus einem älteren Bestand öffnet nichts** — J05
// (#2269),
// Befund 2 des Nachlaufs: die Epoche kam nur aus einer geglückten Antwort
// an, und eine Meldung mit höherer Epoche lehrte das Fenster nichts.
import { act, renderHook } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import {
  epocheGesehen,
  epocheVergessen,
  kennungGeloescht,
} from '../bestandsepoche'
import { useDateioeffnung } from './useDateioeffnung'
import { useLauf } from './useLauf'

afterEach(() => {
  epocheVergessen()
})

function oeffnung(aufschlagen: (setID: string) => void) {
  return renderHook(() => useDateioeffnung('de', aufschlagen, useLauf()))
}

it('verwirft eine Kennung aus einer älteren Epoche, auch nach einem Fehler', () => {
  const aufschlagen = vi.fn()
  const { result } = oeffnung(aufschlagen)

  // **Ein gescheiterter Tausch meldet seine Epoche** — wie jede Antwort.
  epocheGesehen(2)

  act(() => {
    result.current.dateiGeoeffnet('7', 1)
  })

  expect(aufschlagen).not.toHaveBeenCalled()
})

it('lernt aus einer Meldung mit höherer Epoche', () => {
  const aufschlagen = vi.fn()
  const { result } = oeffnung(aufschlagen)

  act(() => {
    result.current.dateiGeoeffnet('8', 3)
    result.current.dateiGeoeffnet('9', 2)
  })

  expect(aufschlagen.mock.calls).toEqual([['8']])
})

it('öffnet eine Meldung ohne Epoche', () => {
  const aufschlagen = vi.fn()
  const { result } = oeffnung(aufschlagen)

  epocheGesehen(5)
  act(() => {
    result.current.dateiGeoeffnet('10')
  })

  expect(aufschlagen).toHaveBeenCalledWith('10')
})

// **Eine Meldung für ein Set, das inzwischen gelöscht ist, öffnet nichts**
// (#2306) — und nur sie: eine für ein anderes Set kommt an, und nach einem
// Tausch des Bestands gilt die Kennung wieder.
it('verwirft eine Meldung für ein gelöschtes Set und nur für dieses', () => {
  const aufschlagen = vi.fn()
  const { result } = oeffnung(aufschlagen)

  epocheGesehen(1)
  kennungGeloescht('5', 1)

  act(() => {
    result.current.dateiGeoeffnet('5', 1)
    result.current.dateiGeoeffnet('6', 1)
    result.current.dateiGeoeffnet('5', 2)
  })

  expect(aufschlagen.mock.calls).toEqual([['6'], ['5']])
})
