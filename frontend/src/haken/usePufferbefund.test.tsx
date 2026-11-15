import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { usePufferbefund } from './usePufferbefund'

function pruefend() {
  return (inhalt: string) =>
    Promise.resolve(
      inhalt.includes('(')
        ? {
            stelle: { zeile: 1, spalte: inhalt.indexOf('(') + 1 },
            meldung: 'offen',
            schluessel: 'klammerOffen',
            werte: [],
          }
        : { stelle: null, meldung: '', schluessel: '', werte: [] },
    )
}

describe('Der Befund des Puffers', () => {
  afterEach(raeumeApi)

  it('gibt Befund und den gültigen Stand, wenn nichts zu beanstanden ist', async () => {
    belegeApi({ PaketdateiPruefen: pruefend() })

    const { result } = renderHook(() => usePufferbefund('script: 1'))

    await waitFor(() => {
      expect(result.current.geprueft).toBe(true)
    })

    expect(result.current.befund?.stelle).toBeNull()
    expect(result.current.gueltig).toBe('script: 1')
  })

  it('hält den letzten gültigen Stand, solange der Puffer einen Befund hat', async () => {
    belegeApi({ PaketdateiPruefen: pruefend() })

    const { result, rerender } = renderHook(
      ({ text }) => usePufferbefund(text),
      {
        initialProps: { text: 'script: 1' },
      },
    )

    await waitFor(() => {
      expect(result.current.gueltig).toBe('script: 1')
    })

    rerender({ text: 'script: f(1' })

    await waitFor(() => {
      expect(result.current.befund?.stelle).toEqual({ zeile: 1, spalte: 10 })
    })

    expect(result.current.gueltig).toBe('script: 1')
    expect(result.current.geprueft).toBe(true)

    rerender({ text: 'script: f' })

    await waitFor(() => {
      expect(result.current.gueltig).toBe('script: f')
    })

    expect(result.current.befund?.stelle).toBeNull()
  })

  it('hat ohne je gültigen Stand keinen', async () => {
    belegeApi({ PaketdateiPruefen: pruefend() })

    const { result } = renderHook(() => usePufferbefund('script: ('))

    await waitFor(() => {
      expect(result.current.geprueft).toBe(true)
    })

    expect(result.current.gueltig).toBeUndefined()
  })

  it('ist zwischen Änderung und Antwort ungeprüft', async () => {
    let freigeben: (() => void) | undefined
    belegeApi({
      PaketdateiPruefen: () =>
        new Promise((los) => {
          freigeben = () => {
            los({ stelle: null, meldung: '', schluessel: '', werte: [] })
          }
        }),
    })

    const { result } = renderHook(() => usePufferbefund('script: 1'))

    expect(result.current.geprueft).toBe(false)
    expect(result.current.gueltig).toBeUndefined()

    freigeben?.()

    await waitFor(() => {
      expect(result.current.geprueft).toBe(true)
    })
  })
})

describe('Der Befund je Datei', () => {
  afterEach(raeumeApi)

  it('gilt nur für die Datei, für die er gerechnet wurde', async () => {
    belegeApi({ PaketdateiPruefen: pruefend() })

    const { result, rerender } = renderHook(
      ({ text, datei }: { text: string; datei: string }) =>
        usePufferbefund(text, datei),
      { initialProps: { text: 'script: 1', datei: 'a' } },
    )

    await waitFor(() => {
      expect(result.current.gueltig).toBe('script: 1')
    })

    rerender({ text: 'script: (', datei: 'b' })

    await waitFor(() => {
      expect(result.current.geprueft).toBe(true)
    })

    expect(result.current.gueltig).toBeUndefined()
  })

  it('prüft einen leeren Text nicht', async () => {
    const gerufen = vi.fn()
    belegeApi({
      PaketdateiPruefen: () => {
        gerufen()

        return Promise.resolve({
          stelle: null,
          meldung: '',
          schluessel: '',
          werte: [],
        })
      },
    })

    const { result } = renderHook(() => usePufferbefund(''))

    await waitFor(() => {
      expect(result.current.geprueft).toBe(true)
    })

    expect(gerufen).not.toHaveBeenCalled()
    expect(result.current.gueltig).toBeUndefined()
  })
})
