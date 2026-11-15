// **Eine Tabstelle, und sie zeigt nie ins Leere** —
// #1290.

import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useRovierendeZeile } from './useRovierendeZeile'

describe('useRovierendeZeile', () => {
  it('beginnt bei der ersten Zeile', () => {
    const { result } = renderHook(() => useRovierendeZeile(5))

    expect(result.current.annehmen(0)).toBe(true)
    expect(result.current.annehmen(1)).toBe(false)
  })

  it('rückt auf die genannte Zeile', () => {
    const { result } = renderHook(() => useRovierendeZeile(5))

    act(() => {
      result.current.rueckenAuf(3)
    })

    expect(result.current.annehmen(3)).toBe(true)
    expect(result.current.annehmen(0)).toBe(false)
  })

  it('klemmt, wenn die Liste kürzer wird', () => {
    // **Sonst nähme keine Zeile den Tabulator an**, und das Gitter wäre gar
    // nicht mehr erreichbar — ein Filter, der die Liste kürzt, sperrte es aus.
    const { result, rerender } = renderHook(
      ({ anzahl }) => useRovierendeZeile(anzahl),
      { initialProps: { anzahl: 10 } },
    )

    act(() => {
      result.current.rueckenAuf(9)
    })

    rerender({ anzahl: 3 })

    expect(result.current.aktiv).toBe(2)
    expect(result.current.annehmen(2)).toBe(true)
  })

  it('bleibt bei null, wenn die Liste leer ist', () => {
    const { result } = renderHook(() => useRovierendeZeile(0))

    expect(result.current.aktiv).toBe(0)
  })
})
