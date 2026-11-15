// **Erst der Plan, dann die Änderung** — DOM-01.AK-02 und AK-03
// (#1696).
//
// Der Haken fragt den Plan, hält ihn fest und führt **genau den geplanten
// Vorgang** aus — mit der Quelle aus dem Plan, nicht mit einer neuen Wahl.
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaketDTO, PaketplanDTO } from '../api'
import { belegeApi, raeumeApi } from '../test/api'
import { usePaketplan } from './usePaketplan'

afterEach(() => {
  cleanup()
  raeumeApi()
})

const paket = { art: 'mse-style', name: 'probe-schlicht' } as PaketDTO

function plan(teil: Partial<PaketplanDTO>): PaketplanDTO {
  return {
    vorgang: 'entfernen',
    paket: 'probe-schlicht.mse-style',
    quelle: '',
    von: '1.0',
    zu: '',
    geaendert: false,
    moeglich: true,
    sets: [],
    konflikte: [],
    ...teil,
  }
}

const fehler = {
  code: 'konflikt' as const,
  schluessel: 'fehler.paketGeaendert',
  meldung: 'Lokal geändert.',
}

describe('usePaketplan', () => {
  it('plant ein Entfernen und führt es mit Art und Namen aus', async () => {
    const PaketPlan = vi.fn(() =>
      Promise.resolve({ plan: plan({}), abgebrochen: false }),
    )
    const PaketEntfernen = vi.fn(() => Promise.resolve({}))
    const geaendert = vi.fn()
    belegeApi({ PaketPlan, PaketEntfernen })

    const { result } = renderHook(() => usePaketplan(geaendert))

    await act(() => result.current.planen('entfernen', paket))
    expect(PaketPlan).toHaveBeenCalledWith(
      'entfernen',
      'mse-style',
      'probe-schlicht',
      '',
    )
    expect(result.current.plan?.vorgang).toBe('entfernen')

    await act(() => result.current.ausfuehren(false))
    expect(PaketEntfernen).toHaveBeenCalledWith('mse-style', 'probe-schlicht')
    expect(result.current.plan).toBeUndefined()
    expect(geaendert).toHaveBeenCalledTimes(1)
  })

  it('bindet aus der Quelle des Plans ein und ersetzt mit Bestätigung', async () => {
    const PaketPlanWaehlen = vi
      .fn()
      .mockResolvedValueOnce({ plan: null, abgebrochen: true })
      .mockResolvedValueOnce({
        plan: plan({ vorgang: 'einbinden', quelle: '/q/neu.mse-style' }),
        abgebrochen: false,
      })
      .mockResolvedValueOnce({
        plan: plan({ vorgang: 'ersetzen', quelle: '/q/alt.mse-style' }),
        abgebrochen: false,
      })
    const PaketEinbindenVon = vi.fn(() =>
      Promise.resolve({ paket: null, abgebrochen: false }),
    )
    const PaketErsetzen = vi.fn(() =>
      Promise.resolve({ paket: null, abgebrochen: false }),
    )
    belegeApi({ PaketPlanWaehlen, PaketEinbindenVon, PaketErsetzen })

    const { result } = renderHook(() => usePaketplan(vi.fn()))

    // **Ein abgebrochener Dialog plant nichts.**
    await act(() => result.current.waehlen())
    expect(result.current.plan).toBeUndefined()

    await act(() => result.current.waehlen())
    await act(() => result.current.ausfuehren(false))
    expect(PaketEinbindenVon).toHaveBeenCalledWith('/q/neu.mse-style')

    await act(() => result.current.waehlen())
    await act(() => result.current.ausfuehren(true))
    expect(PaketErsetzen).toHaveBeenCalledWith('/q/alt.mse-style', true)
  })

  it('setzt zurück und lässt den Plan bei einem Fehler offen', async () => {
    const PaketPlan = vi.fn(() =>
      Promise.resolve({
        plan: plan({ vorgang: 'zurueck' }),
        abgebrochen: false,
      }),
    )
    const PaketZurueck = vi
      .fn()
      .mockResolvedValueOnce({ paket: null, abgebrochen: false, fehler })
      .mockResolvedValueOnce({ paket: null, abgebrochen: false })
    const geaendert = vi.fn()
    belegeApi({ PaketPlan, PaketZurueck })

    const { result } = renderHook(() => usePaketplan(geaendert))

    await act(() => result.current.planen('zurueck', paket))
    await act(() => result.current.ausfuehren(false))

    // **Die Meldung steht beim Plan**, und er bleibt offen.
    expect(result.current.fehler).toBe('Lokal geändert.')
    expect(result.current.plan?.vorgang).toBe('zurueck')

    // **Geladen wird trotzdem neu** — ein gescheiterter Tausch kann die
    // Platte schon angefasst haben.
    expect(geaendert).toHaveBeenCalledTimes(1)

    await act(() => result.current.ausfuehren(false))
    expect(PaketZurueck).toHaveBeenLastCalledWith('mse-style', 'probe-schlicht')
    expect(result.current.plan).toBeUndefined()
    expect(result.current.fehler).toBeUndefined()
  })

  it('meldet einen Fehler beim Planen und verwirft auf Zuruf', async () => {
    belegeApi({
      PaketPlan: vi.fn(() =>
        Promise.resolve({ plan: null, abgebrochen: false, fehler }),
      ),
    })

    const { result } = renderHook(() => usePaketplan(vi.fn()))

    await act(() => result.current.planen('entfernen', paket))
    expect(result.current.fehler).toBe('Lokal geändert.')

    act(() => {
      result.current.verwerfen()
    })
    expect(result.current.fehler).toBeUndefined()

    // **Ohne Plan führt nichts aus.**
    await act(() => result.current.ausfuehren(false))
    expect(result.current.plan).toBeUndefined()
  })

  it('trägt eine Antwort nicht ein, wenn inzwischen verworfen wurde', async () => {
    let antworten: (wert: unknown) => void = () => undefined
    belegeApi({
      PaketPlan: vi.fn(
        () =>
          new Promise((fertig) => {
            antworten = fertig
          }),
      ) as never,
    })

    const { result } = renderHook(() => usePaketplan(vi.fn()))

    let unterwegs: Promise<void> = Promise.resolve()
    act(() => {
      unterwegs = result.current.planen('entfernen', paket)
    })
    await waitFor(() => {
      expect(result.current.laeuft).toBe(true)
    })

    act(() => {
      result.current.verwerfen()
    })
    await act(async () => {
      antworten({ plan: plan({}), abgebrochen: false })
      await unterwegs
    })

    expect(result.current.plan).toBeUndefined()
    expect(result.current.laeuft).toBe(false)

    // **Und ein überholter Fehler ebenso wenig.**
    act(() => {
      unterwegs = result.current.planen('entfernen', paket)
    })
    await waitFor(() => {
      expect(result.current.laeuft).toBe(true)
    })
    act(() => {
      result.current.verwerfen()
    })
    await act(async () => {
      antworten({ plan: null, abgebrochen: false, fehler })
      await unterwegs
    })

    expect(result.current.fehler).toBeUndefined()
  })
})
