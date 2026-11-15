// **Die aufgeschlagene Karte gilt im ganzen Fenster** —
// #892, und
// seit UX-08 (#1680)
// auch die, die niemand angeklickt hat.
import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useKartenmeldung } from './useKartenmeldung'

describe('useKartenmeldung', () => {
  // **Die Vorwahl beim Öffnen** (#668) kommt ohne Klick — und muss trotzdem
  // ankommen, sonst sagt die Konsole „keine Karte", während sie daneben
  // offen steht.
  it('meldet eine Wahl, die das Fenster noch nicht kennt', () => {
    const melden = vi.fn()

    renderHook(() => {
      useKartenmeldung('11', undefined, melden)
    })

    expect(melden).toHaveBeenCalledWith('11')
  })

  it('meldet einen Wechsel, sobald er geschieht', () => {
    const melden = vi.fn()

    const { rerender } = renderHook(
      ({ gewaehlt }) => {
        useKartenmeldung(gewaehlt, '11', melden)
      },
      { initialProps: { gewaehlt: '11' } },
    )

    expect(melden).not.toHaveBeenCalled()

    rerender({ gewaehlt: '22' })

    expect(melden).toHaveBeenCalledWith('22')
  })

  // **Was das Fenster schon kennt, wird nicht wiederholt** — sonst meldete
  // jede Zeichnung dieselbe Karte, und jede Meldung zeichnete neu.
  it('schweigt, solange Wahl und Fenster übereinstimmen', () => {
    const melden = vi.fn()

    const { rerender } = renderHook(
      ({ vorgabe }) => {
        useKartenmeldung('11', vorgabe, melden)
      },
      { initialProps: { vorgabe: '11' } },
    )

    rerender({ vorgabe: '11' })

    expect(melden).not.toHaveBeenCalled()
  })

  // **Keine Karte ist keine Meldung** — ein leeres Set hat keine erste.
  it('meldet nichts, solange keine Karte gewählt ist', () => {
    const melden = vi.fn()

    renderHook(() => {
      useKartenmeldung(undefined, undefined, melden)
    })

    expect(melden).not.toHaveBeenCalled()
  })
})
