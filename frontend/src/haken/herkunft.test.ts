// **Von welcher Seite man in die Hilfe kam** — UXR-22
// (#1930).
import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import type { Seitenschluessel } from '../navigation'
import { useHerkunft } from './useHerkunft'

function verlauf(anfang: Seitenschluessel) {
  return renderHook(
    ({ seite }: { seite: Seitenschluessel }) => useHerkunft(seite),
    {
      initialProps: { seite: anfang },
    },
  )
}

describe('useHerkunft', () => {
  it('kennt keine Herkunft, wenn die Hilfe die erste Seite ist', () => {
    const { result } = verlauf('hilfe')

    expect(result.current).toBeUndefined()
  })

  it('merkt die Seite vor der Hilfe', () => {
    const { result, rerender } = verlauf('decks')

    rerender({ seite: 'hilfe' })

    expect(result.current).toBe('decks')
  })

  it('behält sie beim Weggehen und nimmt beim nächsten Besuch die neue', () => {
    const { result, rerender } = verlauf('decks')
    rerender({ seite: 'hilfe' })

    rerender({ seite: 'start' })
    expect(result.current).toBe('decks')

    rerender({ seite: 'statistik' })
    rerender({ seite: 'hilfe' })
    expect(result.current).toBe('statistik')
  })
})
