// **Ein Set wählen, wahlweise mit einer Karte** — AP-8, Teil 2b (#2365).
//
// Die Kacheln und die Aufmerksamkeit öffnen ein Set mit seiner ersten Karte
// mit Befund. **Die Karte reist mit der Wahl** und nicht hinterher: ein
// Effekt danach verlor gegen die erste Karte der Liste, die die Kartenseite
// beim Aufbau schon gemeldet hatte (gesehen am Fenster, `ap8befunde09`).
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useSetwahl } from './useSetwahl'

afterEach(raeumeApi)

function wahl() {
  belegeApi({ SetBenutzt: () => Promise.resolve({ fehler: null }) })
  const setzeSetID = vi.fn()
  const setzeKarteID = vi.fn()
  const { result } = renderHook(() =>
    useSetwahl(setzeSetID, setzeKarteID, vi.fn(), vi.fn()),
  )

  return { setwahl: result.current, setzeSetID, setzeKarteID }
}

describe('die Wahl eines Sets', () => {
  it('räumt die Karte, wenn keine mitkommt', () => {
    const { setwahl, setzeSetID, setzeKarteID } = wahl()
    act(() => {
      setwahl.waehleSet('7')
    })

    expect(setzeSetID).toHaveBeenLastCalledWith('7')
    expect(setzeKarteID).toHaveBeenLastCalledWith(undefined)
  })

  it('schlägt die mitgegebene Karte auf', () => {
    const { setwahl, setzeSetID, setzeKarteID } = wahl()
    act(() => {
      setwahl.waehleSet('7', '9')
    })

    expect(setzeSetID).toHaveBeenLastCalledWith('7')
    expect(setzeKarteID).toHaveBeenLastCalledWith('9')
  })
})
