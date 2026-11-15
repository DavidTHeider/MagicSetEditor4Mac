// **Wo jemand seine Karten hingelegt hat, bleibt gemerkt** — PLAY-04.AK-03
// (#1712).
//
// # Was hier geprüft wird
//
// **Dass es den Browserspeicher überlebt — und dass es ohne ihn weitergeht.**
// In einem privaten Fenster, bei gesperrten Seitendaten und in einer
// Vorschau wirft schon der Zugriff; eine Anwendung, die daran scheiterte,
// verlöre den Tisch wegen einer Gewohnheit.
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { lagenschluessel, useEigeneLagen } from './useEigeneLagen'

afterEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('useEigeneLagen', () => {
  it('fängt ohne abgelegte Lagen an', () => {
    const { result } = renderHook(() => useEigeneLagen('p1', false))

    expect(result.current.lagen).toEqual({})
  })

  it('merkt sich eine Lage und schreibt sie weg', () => {
    const { result } = renderHook(() => useEigeneLagen('p1', false))

    act(() => {
      result.current.lege('k1', { x: 10, y: 20 })
    })

    expect(result.current.lagen).toEqual({ k1: { x: 10, y: 20 } })
    expect(localStorage.getItem(lagenschluessel('p1'))).toBe(
      '{"k1":{"x":10,"y":20}}',
    )
  })

  it('liest wieder, was dasteht', () => {
    localStorage.setItem(lagenschluessel('p1'), '{"k1":{"x":1,"y":2}}')

    const { result } = renderHook(() => useEigeneLagen('p1', false))

    expect(result.current.lagen).toEqual({ k1: { x: 1, y: 2 } })
  })

  // **Je Partie und nicht über alle** — dieselbe Karte liegt am nächsten
  // Tisch woanders.
  it('trennt die Partien', () => {
    localStorage.setItem(lagenschluessel('p1'), '{"k1":{"x":1,"y":2}}')

    const { result } = renderHook(() => useEigeneLagen('p2', false))

    expect(result.current.lagen).toEqual({})
  })

  // **Ein kaputter Eintrag ist kein Absturz** (§8): was nicht zu lesen ist,
  // ist wie nichts — und der Tisch steht weiter.
  it('kommt mit unlesbarem Inhalt zurecht', () => {
    localStorage.setItem(lagenschluessel('p1'), 'kein JSON')

    const { result } = renderHook(() => useEigeneLagen('p1', false))

    expect(result.current.lagen).toEqual({})
  })

  // **Und ein Speicher, der gar nicht will, ebenso.** In einem privaten
  // Fenster wirft schon `getItem`.
  // **Gültiges JSON in falscher Gestalt** — `null`, eine Liste, eine Lage
  // ohne `y`: übergangen, die übrigen gelten. Sonst warf der Tisch beim
  // Zeichnen (Befund 4 des Nachlaufs über UX-12).
  it.each([
    ['null', {}],
    ['[1, 2]', {}],
    ['"text"', {}],
    [
      '{"k1":{"x":1,"y":2},"k2":null,"k3":{"x":1},"k4":{"x":"1","y":2},"k5":7}',
      { k1: { x: 1, y: 2 } },
    ],
  ])('übergeht, was keine Lage ist (%s)', (roh, soll) => {
    localStorage.setItem(lagenschluessel('p1'), roh)

    const { result } = renderHook(() => useEigeneLagen('p1', false))

    expect(result.current.lagen).toEqual(soll)
  })

  it('kommt ohne Browserspeicher zurecht', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('kein Zugriff')
    })

    const { result } = renderHook(() => useEigeneLagen('p1', false))

    expect(result.current.lagen).toEqual({})
  })

  // **Ein Schreibfehler hält den Tisch nicht auf.** Die Lage gilt dann für
  // diese Sitzung — das ist weniger, als jemand wollte, und mehr als nichts.
  it('behält die Lage, auch wenn das Wegschreiben scheitert', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('voll')
    })

    const { result } = renderHook(() => useEigeneLagen('p1', false))

    act(() => {
      result.current.lege('k1', { x: 10, y: 20 })
    })

    expect(result.current.lagen).toEqual({ k1: { x: 10, y: 20 } })
  })

  // **Ordnet der Tisch selbst, gilt keine abgelegte Lage** — sonst spränge
  // eine Karte beim Einschalten der Anordnung nicht zurück.
  it('gibt nichts her, wo der Tisch selbst ordnet', () => {
    localStorage.setItem(lagenschluessel('p1'), '{"k1":{"x":1,"y":2}}')

    const { result } = renderHook(() => useEigeneLagen('p1', true))

    expect(result.current.lagen).toEqual({})
  })

  // **Gemerkt wird trotzdem** — wer die Anordnung wieder ausschaltet,
  // findet seinen Tisch vor, wie er ihn gelegt hat.
  it('merkt sich auch dann, was abgelegt wird', () => {
    const { result } = renderHook(() => useEigeneLagen('p1', true))

    act(() => {
      result.current.lege('k1', { x: 10, y: 20 })
    })

    expect(localStorage.getItem(lagenschluessel('p1'))).toBe(
      '{"k1":{"x":10,"y":20}}',
    )
  })
})
