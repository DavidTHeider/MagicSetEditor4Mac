import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { useVorlagenpuffer } from './useVorlagenpuffer'

import type { ReactNode } from 'react'
import type { Feldapi } from '../apifeld'

const texte = kataloge.de.seiten.vorlagen

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function lesend(inhalt = 'mse version: 2.0.0\n', stand = 'alt') {
  return () => Promise.resolve({ inhalt, stand, fehler: null })
}

function puffer(paket = '/pakete/probe.mse-style', datei = 'style') {
  return renderHook(
    ({ p, d }: { p: string; d: string }) => useVorlagenpuffer(p, d),
    { wrapper: Rahmen, initialProps: { p: paket, d: datei } },
  )
}

describe('Der Vorlagenpuffer', () => {
  afterEach(raeumeApi)

  it('liest die Datei mit ihrem Stand und ist danach unverändert', async () => {
    belegeApi({ PaketdateiLesen: lesend() })

    const { result } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('mse version: 2.0.0\n')
    })

    expect(result.current.stand).toBe('alt')
    expect(result.current.basis).toBe('mse version: 2.0.0\n')
    expect(result.current.geaendert).toBe(false)
    expect(result.current.laedt).toBe(false)
  })

  it('liest nichts, solange keine Datei gewählt ist', async () => {
    const gelesen = vi.fn<Feldapi['PaketdateiLesen']>(lesend())
    belegeApi({ PaketdateiLesen: gelesen })

    const { result } = puffer('/pakete/probe.mse-style', '')

    await waitFor(() => {
      expect(result.current.laedt).toBe(false)
    })

    expect(gelesen).not.toHaveBeenCalled()
    expect(result.current.inhalt).toBe('')
  })

  it('schreibt mit dem gelesenen Stand und danach mit dem neuen', async () => {
    const geschrieben = vi.fn<Feldapi['PaketdateiSchreiben']>(() =>
      Promise.resolve({ stand: 'neu', fehler: null }),
    )
    belegeApi({ PaketdateiLesen: lesend(), PaketdateiSchreiben: geschrieben })

    const { result } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('mse version: 2.0.0\n')
    })

    act(() => {
      result.current.setzeInhalt('mse version: 2.0.0\nx')
    })

    expect(result.current.geaendert).toBe(true)
    expect(result.current.gespeichert).toBe(false)

    await act(() => result.current.speichern())

    expect(geschrieben).toHaveBeenCalledWith(
      '/pakete/probe.mse-style',
      'style',
      'mse version: 2.0.0\nx',
      'alt',
    )
    expect(result.current.gespeichert).toBe(true)
    expect(result.current.geaendert).toBe(false)
    expect(result.current.basis).toBe('mse version: 2.0.0\nx')

    act(() => {
      result.current.setzeInhalt('mse version: 2.0.0\nxy')
    })

    await act(() => result.current.speichern())

    expect(geschrieben.mock.calls[1]?.[3]).toBe('neu')
  })

  it('lässt bei einem Konflikt den Text stehen und lädt auf Wunsch neu', async () => {
    let gelesenMit = 'alt'
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({
          inhalt: 'von: ' + gelesenMit,
          stand: gelesenMit,
          fehler: null,
        }),
      PaketdateiSchreiben: () =>
        Promise.resolve({
          stand: '',
          fehler: {
            code: 'konflikt',
            schluessel: 'konflikt',
            meldung: 'geändert',
          },
        }),
    })

    const { result } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('von: alt')
    })

    act(() => {
      result.current.setzeInhalt('von: alt x')
    })

    await act(() => result.current.speichern())

    expect(result.current.fehler).toBe(texte.konflikt)
    expect(result.current.inhalt).toBe('von: alt x')
    expect(result.current.geaendert).toBe(true)

    gelesenMit = 'fremd'

    act(() => {
      result.current.neuLaden()
    })

    await waitFor(() => {
      expect(result.current.inhalt).toBe('von: fremd')
    })

    expect(result.current.stand).toBe('fremd')
    expect(result.current.fehler).toBeUndefined()
    expect(result.current.geaendert).toBe(false)
  })

  it('meldet einen Fehler, der keiner der Bindung ist', async () => {
    belegeApi({
      PaketdateiLesen: lesend(),
      PaketdateiSchreiben: () =>
        Promise.reject(new Error('die Leitung ist weg')),
    })

    const { result } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('mse version: 2.0.0\n')
    })

    act(() => {
      result.current.setzeInhalt('x')
    })

    await act(() => result.current.speichern())

    expect(result.current.fehler).toBe('die Leitung ist weg')
    expect(result.current.inhalt).toBe('x')
  })

  it('meldet einen Fehler beim Lesen und hält kein halbes Feld', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({
          inhalt: '',
          stand: '',
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'weg',
          },
        }),
    })

    const { result } = puffer()

    // **Der Satz kommt aus dem Katalog, nicht aus der Antwort** — `meldungAus`
    // setzt über den Schlüssel den eigenen ein (§9).
    await waitFor(() => {
      expect(result.current.lesefehler).toBe(kataloge.de.fehler.nichtGefunden)
    })

    expect(result.current.fehler).toBeUndefined()

    expect(result.current.laedt).toBe(false)
    expect(result.current.inhalt).toBe('')
  })

  it('räumt den Entwurf beim Dateiwechsel und trägt eine späte Antwort nicht ein', async () => {
    let freigeben:
      ((antwort: { stand: string; fehler: null }) => void) | undefined
    belegeApi({
      PaketdateiLesen: (_paket: string, datei: string) =>
        Promise.resolve({
          inhalt: 'inhalt ' + datei,
          stand: 'stand ' + datei,
          fehler: null,
        }),
      PaketdateiSchreiben: () =>
        new Promise((los) => {
          freigeben = los
        }),
    })

    const { result, rerender } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('inhalt style')
    })

    act(() => {
      result.current.setzeInhalt('inhalt style geändert')
    })

    let laufend: Promise<void> | undefined

    act(() => {
      laufend = result.current.speichern()
    })

    rerender({ p: '/pakete/probe.mse-style', d: 'game' })

    await waitFor(() => {
      expect(result.current.inhalt).toBe('inhalt game')
    })

    expect(result.current.geaendert).toBe(false)

    await act(async () => {
      freigeben?.({ stand: 'spät', fehler: null })
      await laufend
    })

    expect(result.current.gespeichert).toBe(false)
    expect(result.current.stand).toBe('stand game')
  })

  it('schreibt nicht zweimal zugleich', async () => {
    let freigeben:
      ((antwort: { stand: string; fehler: null }) => void) | undefined
    const geschrieben = vi.fn<Feldapi['PaketdateiSchreiben']>(
      () =>
        new Promise((los) => {
          freigeben = los
        }),
    )
    belegeApi({ PaketdateiLesen: lesend(), PaketdateiSchreiben: geschrieben })

    const { result } = puffer()

    await waitFor(() => {
      expect(result.current.inhalt).toBe('mse version: 2.0.0\n')
    })

    act(() => {
      result.current.setzeInhalt('x')
    })

    let erster: Promise<void> | undefined

    act(() => {
      erster = result.current.speichern()
    })

    await act(() => result.current.speichern())

    expect(geschrieben).toHaveBeenCalledTimes(1)
    expect(result.current.schreibt).toBe(true)

    await act(async () => {
      freigeben?.({ stand: 'neu', fehler: null })
      await erster
    })

    expect(result.current.schreibt).toBe(false)
    expect(result.current.gespeichert).toBe(true)
  })
})
