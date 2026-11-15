// **Eine Karte aufschlagen — und wohin das Fenster dabei geht.**
//
// Aus `Anwendung.tsx` gezogen, als die Schubladen dazukamen (UX-08,
// #1680):
// seitdem hängt der Seitenwechsel davon ab, **wo** man gerade steht.
import { act, renderHook } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import type { Seitenschluessel } from '../navigation'
import { useKartenoeffnung } from './useKartenoeffnung'

// fenster hält die drei Zustände, die der Haken setzt, und ein Symbolziel,
// dessen Räumen man zählen kann.
function fenster(anfang: Seitenschluessel) {
  const symbolziel = { oeffne: vi.fn(), schliesse: vi.fn() }

  const { result } = renderHook(() => {
    const [aktuell, setzeAktuell] = useState(anfang)
    const [setID, setzeSetID] = useState<string>()
    const [karteID, setzeKarteID] = useState<string>()

    return {
      aktuell,
      setID,
      karteID,
      ...useKartenoeffnung({
        setzeSetID,
        setzeKarteID,
        symbolziel,
        setzeAktuell,
      }),
    }
  })

  return { result, symbolziel }
}

describe('useKartenoeffnung', () => {
  // **Ein Treffer der Suche führt zu den Karten** — und merkt sich Set und
  // Karte, damit die Konsole sie kennt.
  it('schlägt die Karte auf und wechselt zu den Karten', () => {
    const { result, symbolziel } = fenster('suche')

    act(() => {
      result.current.oeffneKarte('7', '22')
    })

    expect(result.current.setID).toBe('7')
    expect(result.current.karteID).toBe('22')
    expect(result.current.aktuell).toBe('karten')

    // **Eine andere Karte räumt das Symbolziel mit** (§7, Regel 1).
    expect(symbolziel.schliesse).toHaveBeenCalled()
  })

  // **Eine offene Schublade bleibt offen.** Die Karten darunter melden ihre
  // Wahl auch dann, wenn niemand sie angeklickt hat — die erste Karte beim
  // Öffnen (#668). Ein Seitenwechsel schlösse die Schublade, die gerade
  // aufging.
  it('lässt eine offene Schublade stehen', () => {
    const { result } = fenster('stil')

    act(() => {
      result.current.oeffneKarte('7', '11')
    })

    expect(result.current.karteID).toBe('11')
    expect(result.current.aktuell).toBe('stil')
  })

  it('führt ein Kartensymbol in den Symboleditor', () => {
    const { result, symbolziel } = fenster('karten')

    act(() => {
      result.current.oeffneSymbol('22', 'symbol')
    })

    expect(symbolziel.oeffne).toHaveBeenCalledWith('22', 'symbol')
    expect(result.current.aktuell).toBe('symbol')
  })
})
