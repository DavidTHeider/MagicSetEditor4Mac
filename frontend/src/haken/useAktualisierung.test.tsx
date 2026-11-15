// **Die Aktualisierung am Fenster** —
// #1100.
//
// Der Weg durch die Anwendung steht in `src/fassungsband.test.tsx`. Hier
// stehen die Zusagen des Hakens: einmal je Fassung, ein Vorgang je Zeit, und
// ein Fortschritt, der wirklich ankommt.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useAktualisierung } from './useAktualisierung'

// horcher hält die Rückrufe, die der Haken bei der Laufzeit anmeldet.
const horcher = new Map<string, (...daten: unknown[]) => void>()

beforeEach(() => {
  horcher.clear()
  localStorage.clear()
  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      horcher.set(name, rueckruf)

      return () => horcher.delete(name)
    },
  }
})

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

const neuere = {
  gefragt: true,
  neuer: true,
  fassung: '9.9.9',
  seite: 'https://beispiel.test/v9',
  holen: true,
  fehler: null,
}

describe('Der Aktualisierungshaken: melden', () => {
  it('meldet eine neuere Fassung', async () => {
    belegeApi({ NachNeuerFassungSehen: () => Promise.resolve(neuere) })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })

    expect(result.current.seite).toBe('https://beispiel.test/v9')
  })

  it('schweigt bei derselben Fassung', async () => {
    const NachNeuerFassungSehen = vi.fn().mockResolvedValue({
      gefragt: true,
      neuer: false,
      fassung: '0.1.0',
      seite: '',
      holen: true,
      fehler: null,
    })

    belegeApi({ NachNeuerFassungSehen })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(NachNeuerFassungSehen).toHaveBeenCalled()
    })

    expect(result.current.fassung).toBe('')
  })

  it('fragt nicht, wenn die Stufe aus ist', async () => {
    const NachNeuerFassungSehen = vi.fn().mockResolvedValue({
      gefragt: false,
      neuer: false,
      fassung: '',
      seite: '',
      holen: true,
      fehler: null,
    })

    belegeApi({ NachNeuerFassungSehen })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(NachNeuerFassungSehen).toHaveBeenCalled()
    })

    expect(result.current.fassung).toBe('')
  })

  // **Einmal je Fassung, nicht bei jedem Start** — Werkstück 3 des Issues.
  it('schweigt über eine Fassung, die schon weggeklickt wurde', async () => {
    belegeApi({ NachNeuerFassungSehen: () => Promise.resolve(neuere) })

    const erster = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(erster.result.current.fassung).toBe('9.9.9')
    })

    act(() => {
      erster.result.current.weglegen()
    })

    expect(erster.result.current.fassung).toBe('')

    // **Ein zweiter Start**, dieselbe Antwort — und diesmal kein Band.
    const zweiter = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(zweiter.result.current.gefragt).toBe(true)
    })

    expect(zweiter.result.current.fassung).toBe('')
  })

  // **Weglegen ohne Band merkt sich nichts.**
  //
  // Der Knopf steht nur da, solange eine Fassung gemeldet ist — dieser Weg ist
  // also eine Wache und kein Bedienweg. Ohne sie stünde nach einem Klick ins
  // Leere eine **leere** Fassung im Speicher, und die nächste echte Meldung
  // verglich sich gegen sie.
  it('merkt sich nichts, wenn es nichts wegzulegen gibt', async () => {
    belegeApi({
      NachNeuerFassungSehen: () =>
        Promise.resolve({ ...neuere, neuer: false, fassung: '' }),
    })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.gefragt).toBe(true)
    })

    act(() => {
      result.current.weglegen()
    })

    expect(localStorage.getItem('mse.fassung.weggelegt')).toBeNull()
  })

  // **Eine andere Fassung meldet sich wieder** — die Gegenprobe dazu.
  it('meldet eine Fassung, die noch niemand weggeklickt hat', async () => {
    localStorage.setItem('mse.fassung.weggelegt', '9.9.8')
    belegeApi({ NachNeuerFassungSehen: () => Promise.resolve(neuere) })

    const { result } = renderHook(() => useAktualisierung())

    await waitFor(() => {
      expect(result.current.fassung).toBe('9.9.9')
    })
  })
})
