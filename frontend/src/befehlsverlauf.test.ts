// **Der Verlauf je Sitzung** — Werkstück U12 von Phase 5.1 (#1423).
import { renderHook, act } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { hoechstens, leere, merke, useBefehlsverlauf } from './befehlsverlauf'

afterEach(() => {
  leere()
})

describe('der Befehlsverlauf', () => {
  it('führt die jüngsten Einträge vorn und höchstens zwanzig', () => {
    const { result } = renderHook(() => useBefehlsverlauf())

    act(() => {
      for (let i = 0; i < hoechstens + 5; i++) {
        merke({
          name: `b${String(i)}`,
          argumente: {},
          ausgang: 'gelungen',
          dauerMs: i,
          zeitpunkt: '2026-09-09T10:00:00Z',
        })
      }
    })

    expect(result.current).toHaveLength(hoechstens)
    expect(result.current[0]?.name).toBe(`b${String(hoechstens + 4)}`)
    expect(result.current[0]?.nummer).toBeGreaterThan(
      result.current[1]?.nummer ?? 0,
    )

    act(() => {
      leere()
    })
    expect(result.current).toEqual([])
  })
})
