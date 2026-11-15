// **Ein Paket als `.mse-installer` weitergeben, am Haken** — FMT-02.AK-04
// (#1691).
import { act, renderHook } from '@testing-library/react'
import { type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { abgebrochen } from '../test/laufauftrag'
import { paketDTO } from '../test/vorlagenprobe'
import { usePaketweitergabe } from './usePaketweitergabe'

afterEach(raeumeApi)

function sprachraum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const spiel = { ...paketDTO('probe'), art: 'mse-game' }
const stil = { ...paketDTO('probe-schlicht'), art: 'mse-style' }

describe('usePaketweitergabe', () => {
  it('gibt weiter und meldet die Datei nur am Paket, das es betrifft', async () => {
    const weitergeben = vi.fn(() =>
      Promise.resolve({
        datei: '/Users/ich/probe.mse-installer',
        abgebrochen: false,
        fehler: null,
      }),
    )
    belegeApi({ PaketWeitergeben: weitergeben })

    const { result } = renderHook(() => usePaketweitergabe(), {
      wrapper: sprachraum,
    })

    await act(() => result.current.weitergeben(spiel))

    expect(weitergeben).toHaveBeenCalledWith('probe.mse-game')
    expect(result.current.meldung(spiel)).toEqual({
      text: 'Als Installer gesichert: /Users/ich/probe.mse-installer',
      fehler: false,
    })
    expect(result.current.meldung(stil)).toBeUndefined()
  })

  it('meldet nach einem Abbruch nichts und einen Fehler am Paket', async () => {
    const weitergeben = vi
      .fn()
      .mockResolvedValueOnce({ datei: '', abgebrochen: true, fehler: null })
      .mockRejectedValueOnce(new Error('kaputt'))
    belegeApi({ PaketWeitergeben: weitergeben })

    const { result } = renderHook(() => usePaketweitergabe(), {
      wrapper: sprachraum,
    })

    await act(() => result.current.weitergeben(spiel))
    expect(result.current.meldung(spiel)).toBeUndefined()

    await act(() => result.current.weitergeben(spiel))
    expect(result.current.meldung(spiel)?.fehler).toBe(true)
  })

  // **Ein Vorgang je Paket** (§7, Regel 4) — ein zweiter Klick, während der
  // erste läuft, fragt nicht noch einmal.
  it('nimmt keinen zweiten Vorgang am selben Paket an', async () => {
    let fertig: (wert: unknown) => void = () => undefined
    const weitergeben = vi.fn(
      () =>
        new Promise((aufloesen) => {
          fertig = aufloesen
        }),
    )
    belegeApi({ PaketWeitergeben: weitergeben as never })

    const { result } = renderHook(() => usePaketweitergabe(), {
      wrapper: sprachraum,
    })

    let erster: Promise<void> = Promise.resolve()
    act(() => {
      erster = result.current.weitergeben(spiel)
    })

    expect(result.current.laeuft(spiel)).toBe(true)
    expect(result.current.laeuft(stil)).toBe(false)

    await act(() => result.current.weitergeben(spiel))
    expect(weitergeben).toHaveBeenCalledOnce()

    await act(async () => {
      fertig({ datei: '/x.mse-installer', abgebrochen: false, fehler: null })
      await erster
    })
    expect(result.current.laeuft(spiel)).toBe(false)
  })
})

// **Ein angehaltener Lauf meldet nichts am Paket** (FMT-03, #1692) — der
// Installer entsteht erst am Ende, und der Laufstand sagt, dass keiner liegt.
it('meldet nach einem Anhalten nichts am Paket', async () => {
  belegeApi({
    PaketWeitergeben: () =>
      Promise.resolve({ datei: '', abgebrochen: false, fehler: abgebrochen }),
  })

  const { result } = renderHook(() => usePaketweitergabe(), {
    wrapper: sprachraum,
  })

  await act(() => result.current.weitergeben(spiel))
  expect(result.current.meldung(spiel)).toBeUndefined()
})
