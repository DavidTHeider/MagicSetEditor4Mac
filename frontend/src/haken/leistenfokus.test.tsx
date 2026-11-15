// **Der Fokushorcher ohne Leiste** — Werkstück U14 von Phase 5.1 (#1490): der
// Haken läuft auch, wenn die Referenz noch niemanden zeigt, und tut dann
// nichts. Der Weg mit Leiste steht in `rahmen/seitenleiste.test.tsx`.
import { fireEvent, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useLeistenfokus } from './useLeistenfokus'

describe('useLeistenfokus', () => {
  it('kommt ohne angehängte Leiste aus', () => {
    const { result, rerender } = renderHook(
      ({ offen }: { offen: boolean }) => useLeistenfokus(offen),
      { initialProps: { offen: true } },
    )

    expect(result.current.leistenbereich.current).toBeNull()

    fireEvent.focusIn(document.body)
    rerender({ offen: false })

    expect(document.activeElement).toBe(document.body)
  })
})
