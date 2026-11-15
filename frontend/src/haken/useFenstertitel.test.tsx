// **Der Fenstertitel nennt das offene Set** — AP-7
// (#2364),
// Befund B-06: bis dahin stand über jedem Fenster nur der Name der
// Anwendung, und wer zwei Sets nacheinander bearbeitete, sah im Fenster- und
// im Dockmenü nicht, welches offen war.
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { useFenstertitel } from './useFenstertitel'

afterEach(() => {
  delete (globalThis as { runtime?: unknown }).runtime
})

const sprachraum = ({ children }: { children: ReactNode }) => (
  <Sprachraum sprache="de">{children}</Sprachraum>
)

describe('useFenstertitel', () => {
  it('nennt das offene Set vor der Anwendung und ohne Set nur sie', () => {
    const TitelSetzen = vi.fn()
    ;(globalThis as { runtime?: unknown }).runtime = { TitelSetzen }

    const anfang: { name: string | undefined } = { name: 'Probeset' }
    const { rerender } = renderHook(
      ({ name }: { name: string | undefined }) => {
        useFenstertitel(name)
      },
      {
        initialProps: anfang,
        wrapper: sprachraum,
      },
    )
    expect(TitelSetzen).toHaveBeenLastCalledWith(
      'Probeset — MagicSetEditor4Mac',
    )

    rerender({ name: undefined })
    expect(TitelSetzen).toHaveBeenLastCalledWith('MagicSetEditor4Mac')
  })

  it('tut ohne Laufzeit nichts, statt zu werfen', () => {
    expect(() =>
      renderHook(
        () => {
          useFenstertitel('Probeset')
        },
        { wrapper: sprachraum },
      ),
    ).not.toThrow()
  })
})
