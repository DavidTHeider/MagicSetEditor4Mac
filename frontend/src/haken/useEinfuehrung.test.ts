// **Ob die ersten Schritte stehen** — Werkstück U11 von Phase 5.1 (#1422).
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useEinfuehrung } from './useEinfuehrung'

describe('useEinfuehrung', () => {
  afterEach(() => {
    sessionStorage.clear()
  })

  it('steht anfangs, lässt sich schließen und erneut verlangen', () => {
    const { result } = renderHook(() => useEinfuehrung())
    expect(result.current.zugeklappt).toBe(false)
    expect(result.current.erneut).toBe(false)

    act(() => {
      result.current.schliessen()
    })
    expect(result.current.zugeklappt).toBe(true)

    act(() => {
      result.current.zeigen()
    })
    expect(result.current.zugeklappt).toBe(false)
    expect(result.current.erneut).toBe(true)

    // Schließen nimmt das „erneut" mit — die nächste Sitzung entscheidet wieder
    // die Einstellung.
    act(() => {
      result.current.schliessen()
    })
    expect(result.current.erneut).toBe(false)
  })

  // **Geschlossen bleibt geschlossen, auch nach einem Neuladen** (#2428):
  // endet der Webprozess, lädt Wails die Seite neu, und der Zustand der
  // Seite ist fort — die Sitzung des Fensters nicht.
  it('bleibt nach einem Neuladen der Seite geschlossen, bis sie verlangt wird', () => {
    const erste = renderHook(() => useEinfuehrung())

    act(() => {
      erste.result.current.schliessen()
    })
    erste.unmount()

    const zweite = renderHook(() => useEinfuehrung())
    expect(zweite.result.current.zugeklappt).toBe(true)

    act(() => {
      zweite.result.current.zeigen()
    })
    zweite.unmount()

    expect(renderHook(() => useEinfuehrung()).result.current.zugeklappt).toBe(
      false,
    )
  })

  it('steht, wenn der Speicher der Sitzung nicht zu lesen ist', () => {
    const lesen = vi
      .spyOn(Storage.prototype, 'getItem')
      .mockImplementation(() => {
        throw new Error('gesperrt')
      })
    const schreiben = vi
      .spyOn(Storage.prototype, 'setItem')
      .mockImplementation(() => {
        throw new Error('gesperrt')
      })
    const loeschen = vi
      .spyOn(Storage.prototype, 'removeItem')
      .mockImplementation(() => {
        throw new Error('gesperrt')
      })

    const { result } = renderHook(() => useEinfuehrung())
    expect(result.current.zugeklappt).toBe(false)

    act(() => {
      result.current.schliessen()
    })
    expect(result.current.zugeklappt).toBe(true)

    act(() => {
      result.current.zeigen()
    })
    expect(result.current.zugeklappt).toBe(false)

    lesen.mockRestore()
    schreiben.mockRestore()
    loeschen.mockRestore()
  })
})
