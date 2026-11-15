import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { abstand, useBestandsstand } from './useBestandsstand'

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true })
})

afterEach(() => {
  vi.useRealTimers()
  raeumeApi()
})

// **Der Zähler steigt, wenn ein anderer Prozess geschrieben hat** (#156).
//
// Die erste Antwort setzt den Ausgangswert und zählt **nicht** — sonst
// meldete jeder Start eine Änderung, die es nicht gab.
describe('useBestandsstand', () => {
  it('zählt erst ab der zweiten, geänderten Antwort', async () => {
    let stand = 7

    belegeApi({
      Datenstand: vi.fn().mockImplementation(() => Promise.resolve({ stand })),
    })

    const { result } = renderHook(() => useBestandsstand())

    await waitFor(() => {
      expect(result.current).toBe(0)
    })

    // Derselbe Stand — nichts geschieht.
    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand)
    })

    expect(result.current).toBe(0)

    // Ein anderer hat geschrieben.
    stand = 8

    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand * 2)
    })

    await waitFor(() => {
      expect(result.current).toBe(1)
    })
  })

  // **Ein Fehler bleibt still.** Er beträfe eine Abfrage, die niemand
  // angefordert hat; sie zwei Sekunden später zu wiederholen ist die bessere
  // Antwort als eine Meldung, die im Weg steht.
  it('bleibt bei einem Fehler still', async () => {
    belegeApi({
      Datenstand: vi.fn().mockResolvedValue({
        stand: 0,
        fehler: { code: 'intern', meldung: 'kaputt' },
      }),
    })

    const { result } = renderHook(() => useBestandsstand())

    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand * 2)
    })

    expect(result.current).toBe(0)
  })

  // **Und bei einer Fassade, die die Methode nicht führt**, ebenso: das ist
  // ein Test, kein Betrieb.
  it('bleibt bei einer unbelegten Fassade still', async () => {
    belegeApi({})

    const { result } = renderHook(() => useBestandsstand())

    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand)
    })

    expect(result.current).toBe(0)
  })

  // **Ein Fenster im Hintergrund fragt nicht.** Die Abfrage wäre Arbeit ohne
  // Betrachter.
  it('fragt nicht, wenn niemand hinsieht', async () => {
    const Datenstand = vi.fn().mockResolvedValue({ stand: 1 })
    belegeApi({ Datenstand })

    const sichtbarkeit = vi
      .spyOn(document, 'visibilityState', 'get')
      .mockReturnValue('hidden')

    renderHook(() => useBestandsstand())

    await act(async () => {
      await vi.advanceTimersByTimeAsync(abstand * 2)
    })

    expect(Datenstand).not.toHaveBeenCalled()

    sichtbarkeit.mockRestore()
  })
})

// **Eine Abfrage zur Zeit** — der Befund aus dem Zweitblick über Phase 6.
//
// Dauert eine Abfrage länger als der Abstand, stieße das Intervall eine zweite
// an; eine spät eintreffende Antwort setzte den gemerkten Stand zurück und
// löste eine Aktualisierung aus, die es nicht gab.
it('stößt keine zweite Abfrage an, solange die erste läuft', async () => {
  let offen = 0
  let hoechstens = 0

  belegeApi({
    Datenstand: vi.fn().mockImplementation(async () => {
      offen++
      hoechstens = Math.max(hoechstens, offen)

      await new Promise((fertig) => setTimeout(fertig, abstand * 3))

      offen--

      return { stand: 1 }
    }),
  })

  renderHook(() => useBestandsstand())

  await act(async () => {
    await vi.advanceTimersByTimeAsync(abstand * 4)
  })

  expect(hoechstens).toBe(1)
})
