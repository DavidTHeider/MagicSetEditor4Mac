// **Ein Symbolziel überlebt sein Set nicht — und kehrt nicht zurück.**
//
// Zweitblick über Phase 4.11 (Befund 3) und ihr Nachlauf.
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { useSymbolziel } from './useSymbolziel'

describe('useSymbolziel', () => {
  it('trägt ein Kartensymbol ein', () => {
    const { result } = renderHook(({ setID }) => useSymbolziel(setID), {
      initialProps: { setID: '7' },
    })

    act(() => {
      result.current.oeffne('3', 'symbol')
    })

    expect(result.current.ziel).toEqual({ karteID: '3', feld: 'symbol' })
  })

  it('gibt beim Schließen das Setsymbol zurück', () => {
    const { result } = renderHook(({ setID }) => useSymbolziel(setID), {
      initialProps: { setID: '7' },
    })

    act(() => {
      result.current.oeffne('3', 'symbol')
    })
    act(() => {
      result.current.schliesse()
    })

    expect(result.current.ziel).toBeUndefined()
  })

  // **§7 Regel 1**: ein Wechsel des Bezugsobjekts räumt den offenen Kontext.
  it('verliert das Ziel beim Setwechsel', () => {
    const { rerender, result } = renderHook(
      ({ setID }) => useSymbolziel(setID),
      { initialProps: { setID: '7' } },
    )

    act(() => {
      result.current.oeffne('3', 'symbol')
    })

    rerender({ setID: '9' })

    expect(result.current.ziel).toBeUndefined()
  })

  // **Und es kehrt nicht zurück** — der Befund des Nachlaufs.
  //
  // Verbergen ist nicht Räumen: wer A verließ und wiederkam, fand das alte
  // Ziel vor. Nach einem Zurückspielen des Bestands kann dieselbe Kennung
  // inzwischen zu einer **anderen** Karte gehören.
  it('kehrt nach A → B → A nicht zurück', () => {
    const { rerender, result } = renderHook(
      ({ setID }) => useSymbolziel(setID),
      { initialProps: { setID: '7' } },
    )

    act(() => {
      result.current.oeffne('3', 'symbol')
    })

    rerender({ setID: '9' })
    rerender({ setID: '7' })

    expect(result.current.ziel).toBeUndefined()
  })

  // **Und dasselbe über „Set schließen"** — ⌘W setzt die Kennung auf leer.
  it('kehrt nach dem Schließen und Wiederöffnen nicht zurück', () => {
    const offen: { setID: string | undefined } = { setID: '7' }

    const { rerender, result } = renderHook(
      ({ setID }: { setID: string | undefined }) => useSymbolziel(setID),
      { initialProps: offen },
    )

    act(() => {
      result.current.oeffne('3', 'symbol')
    })

    rerender({ setID: undefined })
    rerender({ setID: '7' })

    expect(result.current.ziel).toBeUndefined()
  })

  // **Ohne offenes Set entsteht keines** — der Zweig, den der Typ verlangt.
  it('trägt ohne offenes Set nichts ein', () => {
    const { result } = renderHook(
      ({ setID }: { setID: string | undefined }) => useSymbolziel(setID),
      { initialProps: { setID: undefined } },
    )

    act(() => {
      result.current.oeffne('3', 'symbol')
    })

    expect(result.current.ziel).toBeUndefined()
  })
})
