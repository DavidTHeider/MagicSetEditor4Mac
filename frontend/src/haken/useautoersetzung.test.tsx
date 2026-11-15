import { renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useAutoersetzung } from './useAutoersetzung'

// Der Haken über der Autoersetzung —
// #897.

describe('useAutoersetzung', () => {
  afterEach(raeumeApi)

  it('reicht Set, Text und Stelle an die Fassade weiter', async () => {
    const AutoErsetzen = vi.fn(() =>
      Promise.resolve({ text: '©', stelle: 1, fehler: null }),
    )
    belegeApi({ AutoErsetzen })

    const { result } = renderHook(() => useAutoersetzung('7'))

    await expect(result.current('(C)', 3)).resolves.toEqual({
      text: '©',
      stelle: 1,
    })

    expect(AutoErsetzen).toHaveBeenCalledWith('7', '(C)', 3)
  })

  // **Ein strukturierter Fehler lässt Text und Marke stehen** —
  // #932.
  //
  // Bis zum 2026-08-23 belegte der einzige Fall dieser Datei `fehler: null`.
  // Damit stand die Zusage „Text und Marke bleiben stehen" ungeprüft da — für
  // genau den Fall, für den sie gemacht ist.
  it('gibt Text und Marke zurück, auch wenn Go einen Fehler meldet', async () => {
    const AutoErsetzen = vi.fn(() =>
      Promise.resolve({
        text: '(C)',
        stelle: 3,
        fehler: {
          code: 'nichtGefunden' as const,
          schluessel: 'nichtGefunden',
          meldung: 'das Set führt der Bestand nicht',
        },
      }),
    )
    belegeApi({ AutoErsetzen })

    const { result } = renderHook(() => useAutoersetzung('7'))

    await expect(result.current('(C)', 3)).resolves.toEqual({
      text: '(C)',
      stelle: 3,
    })
  })
})
