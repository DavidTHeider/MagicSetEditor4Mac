import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useVerzoegert } from './useVerzoegert'

describe('useVerzoegert', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('rückt einen neuen Wert erst nach der Frist nach — und nur den letzten', () => {
    const { result, rerender } = renderHook(
      ({ wert }) => useVerzoegert(wert, 300),
      {
        initialProps: { wert: 'a' },
      },
    )

    expect(result.current).toBe('a')

    rerender({ wert: 'b' })
    rerender({ wert: 'c' })

    act(() => {
      vi.advanceTimersByTime(299)
    })

    expect(result.current).toBe('a')

    act(() => {
      vi.advanceTimersByTime(1)
    })

    expect(result.current).toBe('c')
  })
})
