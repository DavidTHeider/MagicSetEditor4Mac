// **Der Schnellfilter der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ListenspalteDTO } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { useSchnellfilter } from './useSchnellfilter'

afterEach(() => {
  raeumeApi()
})

describe('useSchnellfilter', () => {
  it('fragt ohne Anfrage nicht und lässt alles stehen', async () => {
    const KartenFiltern = vi.fn().mockResolvedValue({ kartenIDs: [] })
    belegeApi({ KartenFiltern })

    const { result } = renderHook(() => useSchnellfilter('3', 0, []))

    // **Der Beleg, dass der Effekt lief**: eine Anfrage wird gestellt …
    act(() => {
      result.current.setzeAnfrage('x')
    })
    await waitFor(() => {
      expect(result.current.treffer).toBeDefined()
    })

    // … und die leere danach nicht.
    act(() => {
      result.current.setzeAnfrage('   ')
    })
    await waitFor(() => {
      expect(result.current.treffer).toBeUndefined()
    })

    expect(KartenFiltern).toHaveBeenCalledTimes(1)
    expect(result.current.laedt).toBe(false)
  })

  it('gibt die Kennungen der Antwort als Menge', async () => {
    const KartenFiltern = vi.fn().mockResolvedValue({ kartenIDs: ['5', '9'] })
    belegeApi({ KartenFiltern })

    const { result } = renderHook(() => useSchnellfilter('3', 0, []))

    act(() => {
      result.current.setzeAnfrage('type:creature')
    })

    await waitFor(() => {
      expect(result.current.treffer?.has('5')).toBe(true)
    })

    expect(KartenFiltern).toHaveBeenCalledWith('3', 'type:creature')
    expect(result.current.treffer?.has('9')).toBe(true)
    expect(result.current.treffer?.has('7')).toBe(false)
  })

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
  it('lässt eine späte Antwort auf eine ältere Anfrage fallen', async () => {
    let ersteAntwort: (wert: object) => void = () => undefined
    const KartenFiltern = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((aufloesen) => {
            ersteAntwort = aufloesen
          }),
      )
      .mockResolvedValueOnce({ kartenIDs: ['2'] })
    belegeApi({ KartenFiltern })

    const { result } = renderHook(() => useSchnellfilter('3', 0, []))

    act(() => {
      result.current.setzeAnfrage('a')
    })
    act(() => {
      result.current.setzeAnfrage('ab')
    })

    await waitFor(() => {
      expect(result.current.treffer?.has('2')).toBe(true)
    })

    act(() => {
      ersteAntwort({ kartenIDs: ['1'] })
    })

    // **Ausdrücklich abgewartet**: nach der späten Antwort steht noch der Stand
    // der letzten Anfrage.
    await act(async () => {
      await Promise.resolve()
    })

    expect(result.current.treffer?.has('1')).toBe(false)
    expect(result.current.treffer?.has('2')).toBe(true)
  })

  it('nennt den Fehler, wenn das Filtern scheitert', async () => {
    belegeApi({
      KartenFiltern: vi.fn().mockResolvedValue({
        fehler: { code: 'validierung', meldung: 'Das Set gibt es nicht.' },
      }),
    })

    const { result } = renderHook(() => useSchnellfilter('3', 0, []))

    act(() => {
      result.current.setzeAnfrage('x')
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe('Das Set gibt es nicht.')
    })
  })
})

// **Ein neuer Bezug fragt neu** — Befund 1 des Zweitblicks über Phase G: die
// Treffer hängen an den sichtbaren Spalten und den Werten, nicht nur an der
// Anfrage. Ohne Anfrage fragt auch ein neuer Bezug nicht.
describe('useSchnellfilter mit Bezug', () => {
  it('fragt bei neuem Bezug neu, ohne Anfrage nicht', async () => {
    const KartenFiltern = vi.fn().mockResolvedValue({ kartenIDs: ['1'] })
    belegeApi({ KartenFiltern })

    const spalte = (sichtbar: boolean): ListenspalteDTO => ({
      feld: 'type',
      aufschrift: 'Typ',
      sichtbar,
      ausrichtung: 'links',
    })
    const { result, rerender } = renderHook(
      ({ stand, spalten }: { stand: number; spalten: ListenspalteDTO[] }) =>
        useSchnellfilter('3', stand, spalten),
      { initialProps: { stand: 0, spalten: [spalte(true)] } },
    )

    rerender({ stand: 1, spalten: [spalte(true)] })

    act(() => {
      result.current.setzeAnfrage('x')
    })
    await waitFor(() => {
      expect(result.current.treffer?.has('1')).toBe(true)
    })
    expect(KartenFiltern).toHaveBeenCalledTimes(1)

    // **Derselbe Aufbau, neu geladen, fragt nicht** …
    rerender({ stand: 1, spalten: [spalte(true)] })
    // … **eine andere Wahl schon**, und ein neuer Stand ebenso.
    rerender({ stand: 1, spalten: [spalte(false)] })

    await waitFor(() => {
      expect(KartenFiltern).toHaveBeenCalledTimes(2)
    })

    rerender({ stand: 2, spalten: [spalte(false)] })

    await waitFor(() => {
      expect(KartenFiltern).toHaveBeenCalledTimes(3)
    })
  })
})
