// **Ob eine Medienabfrage zutrifft** — der Haken für mehr Kontrast und dunkel
// aus UX-09 (#1681).
//
// **Dieselbe Gestalt wie `schmalesfenster.test.ts`**: gelesen wird die Liste,
// nicht das Ereignis, und ohne `matchMedia` gilt »nein«.
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { useMedienfrage } from './useMedienfrage'

interface Nachbau {
  stand: boolean
  horcher: Set<() => void>
  abmeldungen: number
}

function belegeMatchMedia(anfangs: boolean): Nachbau {
  const nachbau: Nachbau = {
    stand: anfangs,
    horcher: new Set(),
    abmeldungen: 0,
  }

  window.matchMedia = (() => ({
    get matches() {
      return nachbau.stand
    },
    addEventListener: (_: string, horcher: () => void) =>
      nachbau.horcher.add(horcher),
    removeEventListener: (_: string, horcher: () => void) => {
      nachbau.abmeldungen += 1
      nachbau.horcher.delete(horcher)
    },
  })) as unknown as typeof window.matchMedia

  return nachbau
}

afterEach(() => {
  Reflect.deleteProperty(window, 'matchMedia')
})

describe('useMedienfrage', () => {
  // **Ohne `matchMedia` gibt es keine Systemeinstellung** — und keinen
  // Absturz, sondern »nein«.
  it('sagt nein, wenn es keine Abfrage gibt', () => {
    Reflect.deleteProperty(window, 'matchMedia')

    const { result } = renderHook(() =>
      useMedienfrage('(prefers-contrast: more)'),
    )

    expect(result.current).toBe(false)
  })

  it('folgt einem Wechsel und hört beim Abbau auf', () => {
    const nachbau = belegeMatchMedia(false)
    const { result, unmount } = renderHook(() =>
      useMedienfrage('(prefers-color-scheme: dark)'),
    )
    expect(result.current).toBe(false)

    nachbau.stand = true
    act(() => {
      for (const horcher of nachbau.horcher) horcher()
    })
    expect(result.current).toBe(true)

    unmount()
    expect(nachbau.abmeldungen).toBe(1)
  })
})
