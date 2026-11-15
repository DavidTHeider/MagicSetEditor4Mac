// **Ein Entwurf überlebt das Verlassen der Seite** — UX-03.AK-04
// (#1675):
// *„Entwürfe werden sicher behandelt."*
//
// # Der Befund
//
// **Der getippte Stand einer Paketdatei lebte im Zustand des Hakens.** Wer
// die Vorlagenwerkstatt mit ungespeicherten Änderungen verließ — über die
// Leiste, ⌘K oder ein Kürzel —, fand sie beim Zurückkommen nicht mehr, und
// nichts hatte gefragt. Die Rückfrage gab es nur beim Wechsel der **Datei**.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import {
  useVorlagenpuffer,
  vergissVorlagenentwuerfe,
} from './useVorlagenpuffer'

import type { ReactNode } from 'react'

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const gelesen = 'mse version: 2.0.0\n'

function puffer(datei = 'style') {
  return renderHook(() => useVorlagenpuffer('/pakete/probe.mse-style', datei), {
    wrapper: Rahmen,
  })
}

async function mitEntwurf(text: string) {
  const erster = puffer()
  await waitFor(() => {
    expect(erster.result.current.inhalt).toBe(gelesen)
  })

  act(() => {
    erster.result.current.setzeInhalt(text)
  })

  return erster
}

describe('der Entwurf einer Vorlagendatei', () => {
  afterEach(() => {
    raeumeApi()
    vergissVorlagenentwuerfe()
  })

  it('steht nach dem Verlassen und Zurückkommen wieder da, als ungespeichert', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: gelesen, stand: 'alt', fehler: null }),
    })

    const erster = await mitEntwurf('getippt')
    erster.unmount()

    const zweiter = puffer()
    await waitFor(() => {
      expect(zweiter.result.current.laedt).toBe(false)
    })

    expect(zweiter.result.current.inhalt).toBe('getippt')
    expect(zweiter.result.current.basis).toBe(gelesen)
    expect(zweiter.result.current.geaendert).toBe(true)
  })

  it('gilt nur für seine Datei', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: gelesen, stand: 'alt', fehler: null }),
    })

    const erster = await mitEntwurf('getippt')
    erster.unmount()

    const andere = puffer('card')
    await waitFor(() => {
      expect(andere.result.current.inhalt).toBe(gelesen)
    })
    expect(andere.result.current.geaendert).toBe(false)
  })

  it('ist nach dem Speichern fort', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: gelesen, stand: 'alt', fehler: null }),
      PaketdateiSchreiben: () =>
        Promise.resolve({ stand: 'neu', fehler: null }),
    })

    const erster = await mitEntwurf('getippt')
    await act(async () => {
      await erster.result.current.speichern()
    })
    erster.unmount()

    const zweiter = puffer()
    await waitFor(() => {
      expect(zweiter.result.current.inhalt).toBe(gelesen)
    })
    expect(zweiter.result.current.geaendert).toBe(false)
  })

  // **Wer „Verwerfen" sagt, bekommt ihn nicht zurück** — auch nicht, wenn
  // er später zur Datei zurückkehrt.
  it('ist nach dem Verwerfen fort', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: gelesen, stand: 'alt', fehler: null }),
    })

    const erster = await mitEntwurf('getippt')
    act(() => {
      erster.result.current.verwerfen()
    })
    expect(erster.result.current.inhalt).toBe(gelesen)
    erster.unmount()

    const zweiter = puffer()
    await waitFor(() => {
      expect(zweiter.result.current.inhalt).toBe(gelesen)
    })
  })

  // **Ein Profilwechsel nimmt ihn mit** — ein Entwurf aus einem anderen
  // Profil wären fremde Daten (UX-03.AK-04).
  it('ist nach dem Vergessen fort, wie beim Profilwechsel', async () => {
    belegeApi({
      PaketdateiLesen: () =>
        Promise.resolve({ inhalt: gelesen, stand: 'alt', fehler: null }),
    })

    const erster = await mitEntwurf('getippt')
    erster.unmount()
    vergissVorlagenentwuerfe()

    const zweiter = puffer()
    await waitFor(() => {
      expect(zweiter.result.current.inhalt).toBe(gelesen)
    })
    expect(zweiter.result.current.geaendert).toBe(false)
  })
})
