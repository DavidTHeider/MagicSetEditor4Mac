// **Die Dauer einer Überlagerung kommt aus dem Stilblatt** — AP-4
// (#2361),
// Befund B-11, Vermerk 0065 Festlegung 10.
//
// Schublade, Dialog und Popover standen auf `duration: 0`. Mantine will eine
// Zahl; die Grenze steht aber als Variable in `stil.css` (`--dauer-mittel`,
// 160 ms), und wenig Bewegung setzt sie auf 0 ms — in der Systemeinstellung
// und in der eigenen (`data-bewegung`). Der Haken liest die Variable und
// keine zweite Zahl daneben.
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { inMillisekunden, useBewegungsdauer } from './useBewegungsdauer'

afterEach(() => {
  document.documentElement.style.removeProperty('--dauer-kurz')
  document.documentElement.style.removeProperty('--dauer-mittel')
  document.documentElement.removeAttribute('data-bewegung')
})

describe('useBewegungsdauer', () => {
  it('liest die Dauer aus der Variablen des Stilblatts', () => {
    document.documentElement.style.setProperty('--dauer-mittel', '160ms')
    document.documentElement.style.setProperty('--dauer-kurz', '120ms')

    expect(renderHook(() => useBewegungsdauer('mittel')).result.current).toBe(
      160,
    )
    expect(renderHook(() => useBewegungsdauer('kurz')).result.current).toBe(120)
  })

  it('folgt der eigenen Einstellung, wenn sie die Bewegung nimmt', async () => {
    document.documentElement.style.setProperty('--dauer-mittel', '160ms')
    const { result } = renderHook(() => useBewegungsdauer('mittel'))
    expect(result.current).toBe(160)

    // **Wie `barrierefreiheit.css`**: die Einstellung setzt das Attribut,
    // und das Stilblatt setzt die Variable auf 0 ms.
    await act(async () => {
      document.documentElement.style.setProperty('--dauer-mittel', '0ms')
      document.documentElement.setAttribute('data-bewegung', 'reduziert')
      await Promise.resolve()
    })

    expect(result.current).toBe(0)
  })

  it('gibt 0, wo die Variable fehlt — lieber keine Bewegung als eine geratene', () => {
    expect(renderHook(() => useBewegungsdauer('mittel')).result.current).toBe(0)
  })
})

describe('inMillisekunden', () => {
  it('liest Millisekunden und Sekunden, und sonst nichts', () => {
    expect(inMillisekunden('160ms')).toBe(160)
    expect(inMillisekunden(' 0.16s ')).toBe(160)
    expect(inMillisekunden('lang')).toBe(0)
  })
})

describe('ohne matchMedia', () => {
  it('liest trotzdem, statt zu werfen', () => {
    vi.stubGlobal('matchMedia', undefined)
    document.documentElement.style.setProperty('--dauer-kurz', '120ms')
    try {
      const { result, unmount } = renderHook(() => useBewegungsdauer('kurz'))
      expect(result.current).toBe(120)
      unmount()
    } finally {
      vi.unstubAllGlobals()
    }
  })
})
