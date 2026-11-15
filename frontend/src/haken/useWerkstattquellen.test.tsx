import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import {
  usePaketsymbole,
  useProbekarten,
  useSymbolverwendungen,
} from './useWerkstattquellen'

import type { ReactNode } from 'react'

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

describe('Die Quellen der Werkstatt', () => {
  afterEach(raeumeApi)

  it('holen die Testkarten eines Pakets', async () => {
    belegeApi({
      Probekarten: () =>
        Promise.resolve({
          karten: [{ id: '7', name: 'Langtext', nummer: 1 }],
          fehler: null,
        }),
    })

    const { result } = renderHook(() => useProbekarten('/p'), {
      wrapper: Rahmen,
    })

    await waitFor(() => {
      expect(result.current.karten).toEqual([
        { id: '7', name: 'Langtext', nummer: 1 },
      ])
    })
  })

  it('melden einen Fehler der Testkarten', async () => {
    belegeApi({
      Probekarten: () =>
        Promise.resolve({
          karten: [],
          fehler: { code: 'unbekannt', schluessel: 'unbekannt', meldung: 'zu' },
        }),
    })

    const { result } = renderHook(() => useProbekarten('/p'), {
      wrapper: Rahmen,
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe(kataloge.de.fehler.unbekannt)
    })

    expect(result.current.karten).toEqual([])
  })

  it('holen die Symbole neu, wenn der Stand wechselt', async () => {
    const gerufen = vi.fn()
    belegeApi({
      Paketsymbole: (paket: string) => {
        gerufen(paket)

        return Promise.resolve({
          symbole: [
            {
              name: 'title',
              art: 'set field',
              datei: 'game',
              zeile: 2,
              spalte: 2,
            },
          ],
          fehler: null,
        })
      },
    })

    const { result, rerender } = renderHook(
      ({ stand }: { stand: string }) => usePaketsymbole('/p', stand),
      { wrapper: Rahmen, initialProps: { stand: 'a' } },
    )

    await waitFor(() => {
      expect(result.current.symbole).toHaveLength(1)
    })

    rerender({ stand: 'b' })

    await waitFor(() => {
      expect(gerufen).toHaveBeenCalledTimes(2)
    })

    expect(result.current.namen).toEqual(['title'])
  })

  it('holen ohne Paket keine Testkarten', async () => {
    const gerufen = vi.fn()
    belegeApi({
      Probekarten: () => {
        gerufen()

        return Promise.resolve({ karten: [], fehler: null })
      },
    })

    const { result } = renderHook(() => useProbekarten(''), { wrapper: Rahmen })

    await waitFor(() => {
      expect(result.current.karten).toEqual([])
    })

    expect(gerufen).not.toHaveBeenCalled()
  })

  it('holen ohne Paket keine Symbole', async () => {
    const gerufen = vi.fn()
    belegeApi({
      Paketsymbole: () => {
        gerufen()

        return Promise.resolve({ symbole: [], fehler: null })
      },
    })

    const { result } = renderHook(() => usePaketsymbole('', ''), {
      wrapper: Rahmen,
    })

    await waitFor(() => {
      expect(result.current.symbole).toEqual([])
    })

    expect(gerufen).not.toHaveBeenCalled()
  })

  it('suchen die Verwendungen eines Namens — und ohne Namen nicht', async () => {
    const gerufen = vi.fn()
    belegeApi({
      Symbolverwendungen: (paket: string, name: string) => {
        gerufen(paket, name)

        return Promise.resolve({
          verwendungen: [{ datei: 'style', zeile: 3, text: 'title:' }],
          fehler: null,
        })
      },
    })

    const { result, rerender } = renderHook(
      ({ name }: { name: string }) => useSymbolverwendungen('/p', name),
      { wrapper: Rahmen, initialProps: { name: '' } },
    )

    await waitFor(() => {
      expect(result.current.laedt).toBe(false)
    })

    expect(gerufen).not.toHaveBeenCalled()
    expect(result.current.verwendungen).toEqual([])

    rerender({ name: 'title' })

    await waitFor(() => {
      expect(result.current.verwendungen).toHaveLength(1)
    })

    expect(gerufen).toHaveBeenCalledWith('/p', 'title')
  })
})
