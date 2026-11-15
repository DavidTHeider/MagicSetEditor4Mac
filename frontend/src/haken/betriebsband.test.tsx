// **Das Betriebsband fragt höchstens alle 30 Sekunden** — OPS02 (#2348):
// sofort beim Aufschlagen, dann im Takt, nicht im Hintergrund, und eine
// gescheiterte Abfrage lässt die letzten Hinweise stehen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { betriebstakt, useBetriebsband } from './useBetriebsband'
import type { BetriebsbandAntwort } from '../apibetriebsband'

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
})

afterEach(() => {
  vi.useRealTimers()
  raeumeApi()
})

const alt: BetriebsbandAntwort = {
  hinweise: [
    { art: 'sicherungAlt', ziel: 'einstellungen/daten', werte: { tage: '20' } },
  ],
  ungemessen: [],
  aufwandMikros: 300,
  fehler: null,
}

describe('useBetriebsband', () => {
  it('fragt sofort und dann alle 30 Sekunden', async () => {
    const Betriebsband = vi.fn().mockResolvedValue(alt)
    belegeApi({ Betriebsband })

    const { result } = renderHook(() => useBetriebsband())

    await waitFor(() => {
      expect(result.current).toHaveLength(1)
    })

    expect(betriebstakt).toBe(30_000)
    expect(Betriebsband).toHaveBeenCalledTimes(1)

    // **Die halbe Frist und nicht eine Millisekunde davor**: die Uhr geht
    // mit der Wanduhr mit (`shouldAdvanceTime`), und eine Millisekunde ist
    // schneller vergangen, als der Test sie zählt.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt / 2)
    })

    expect(Betriebsband).toHaveBeenCalledTimes(1)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt / 2)
    })

    await waitFor(() => {
      expect(Betriebsband).toHaveBeenCalledTimes(2)
    })
  })

  it('lässt die letzten Hinweise stehen, wenn eine Abfrage scheitert', async () => {
    const Betriebsband = vi
      .fn()
      .mockResolvedValueOnce(alt)
      .mockRejectedValueOnce(new Error('abgebrochen'))
      .mockResolvedValueOnce({
        ...alt,
        hinweise: [],
        fehler: { code: 'unbekannt', schluessel: '', meldung: 'kaputt' },
      })
      .mockResolvedValue({ ...alt, hinweise: [] })
    belegeApi({ Betriebsband })

    const { result } = renderHook(() => useBetriebsband())

    await waitFor(() => {
      expect(result.current).toHaveLength(1)
    })

    for (const aufruf of [2, 3]) {
      await act(async () => {
        await vi.advanceTimersByTimeAsync(betriebstakt)
      })

      await waitFor(() => {
        expect(Betriebsband).toHaveBeenCalledTimes(aufruf)
      })

      expect(result.current).toHaveLength(1)
    }

    // **Und die nächste gute Antwort räumt** — der Beleg, dass jemand zuhört.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt)
    })

    await waitFor(() => {
      expect(result.current).toHaveLength(0)
    })
  })

  it('fragt nicht, solange niemand hinsieht', async () => {
    const Betriebsband = vi.fn().mockResolvedValue(alt)
    belegeApi({ Betriebsband })

    const sichtbar = vi.spyOn(document, 'visibilityState', 'get')
    sichtbar.mockReturnValue('hidden')

    renderHook(() => useBetriebsband())

    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt * 2)
    })

    expect(Betriebsband).not.toHaveBeenCalled()

    sichtbar.mockReturnValue('visible')

    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt)
    })

    await waitFor(() => {
      expect(Betriebsband).toHaveBeenCalledTimes(1)
    })

    sichtbar.mockRestore()
  })

  it('fragt nur eine zur Zeit und trägt nach dem Abhängen nichts ein', async () => {
    let antworten: (antwort: BetriebsbandAntwort) => void = () => undefined
    const Betriebsband = vi.fn().mockImplementation(
      () =>
        new Promise<BetriebsbandAntwort>((fertig) => {
          antworten = fertig
        }),
    )
    belegeApi({ Betriebsband })

    const { result, unmount } = renderHook(() => useBetriebsband())

    await act(async () => {
      await vi.advanceTimersByTimeAsync(betriebstakt * 2)
    })

    expect(Betriebsband).toHaveBeenCalledTimes(1)

    unmount()

    await act(async () => {
      antworten(alt)
      await Promise.resolve()
    })

    expect(result.current).toHaveLength(0)
  })
})
