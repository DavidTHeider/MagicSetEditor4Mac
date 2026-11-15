// **Am fremden Tisch keine Vorderseiten aus dem eigenen Bestand** — gefunden
// mit zwei Fenstern (Hüllenprobe `nat04`, S12.NAT-04).
//
// Die Kartenkennungen am Tisch eines Wirts sind seine. Der eigene Bestand
// kennt sie nicht — oder kennt unter derselben Kennung eine andere Karte,
// und dann stünde deren Bild am Tisch.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { useSpieltexturen } from './useSpieltexturen'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { Sprachraum } from '../sprache/kontext'

import type { ReactNode } from 'react'
import type { TexturAntwort, Texturwunsch } from '../apitextur'

afterEach(() => {
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const vorderseite: Texturwunsch = {
  art: 'karte',
  karteID: '1',
  seite: 'vorne',
  guete: 'tisch',
}
const ruecken: Texturwunsch = { art: 'ruecken', spiel: 'magic', guete: 'tisch' }

function textur(): Promise<TexturAntwort> {
  return Promise.resolve({
    textur: {
      schluessel: 'abc',
      adresse: '/spielbild/abc.png',
      guete: 'tisch',
      seite: 'vorne',
      breite: 512,
      hoehe: 714,
      befunde: [],
      zweiseitig: false,
    },
    fehler: null,
  })
}

// **Am fremden Tisch fragt das Fenster den Wirt** — P9 (#2377), Vermerk
// 0198: nach Objekt und Bildkennung, nie nach einer Zeile. Den Rücken holt
// es weiter selbst.
it('holt am fremden Tisch die Vorderseite beim Wirt — und den Rücken selbst', async () => {
  const vorn = vi.fn(textur)
  const fremd = vi.fn(textur)
  const hinten = vi.fn(textur)
  belegeApi({
    Spieltextur: vorn,
    Fremdkartenbild: fremd,
    Rueckentextur: hinten,
  })

  const { result } = renderHook(() => useSpieltexturen(true, 'p1'), {
    wrapper: raum,
  })
  const beimWirt: Texturwunsch = {
    ...vorderseite,
    karteID: '',
    bild: 'b1',
    objekt: 'o1',
  }

  act(() => {
    result.current.belege(beimWirt)
    result.current.belege(ruecken)
  })
  await waitFor(() => {
    expect(fremd).toHaveBeenCalledExactlyOnceWith(
      'p1',
      'o1',
      'b1',
      'vorne',
      'tisch',
    )
  })
  await waitFor(() => {
    expect(result.current.adresse(beimWirt)).toBe('/spielbild/abc.png')
  })
  expect(hinten).toHaveBeenCalledOnce()
  expect(vorn).not.toHaveBeenCalled()
})

// **Eine Zeile ohne Bildkennung nimmt der fremde Tisch nicht** — so schickt
// ein Wirt vor 0.9.0 seine eigenen Zeilen, und hier nachgeschlagen fänden sie
// andere Karten.
it('holt am fremden Tisch nichts zu einer Zeile ohne Bildkennung', async () => {
  const vorn = vi.fn(textur)
  const fremd = vi.fn(textur)
  const hinten = vi.fn(textur)
  belegeApi({
    Spieltextur: vorn,
    Fremdkartenbild: fremd,
    Rueckentextur: hinten,
  })

  const { result } = renderHook(() => useSpieltexturen(true, 'p1'), {
    wrapper: raum,
  })

  act(() => {
    result.current.belege(vorderseite)
    result.current.vorladen([vorderseite])
    result.current.gib(vorderseite)
    result.current.belege(ruecken)
  })
  await waitFor(() => {
    expect(hinten).toHaveBeenCalledOnce()
  })
  await effekteNachholen()

  expect(vorn).not.toHaveBeenCalled()
  expect(fremd).not.toHaveBeenCalled()
})

// **Das eigene Deck zeichnet der Gast selbst** — Zeile und Bildkennung
// zusammen heißen: aus diesem Bestand.
it('zeichnet am fremden Tisch das eigene Deck selbst', async () => {
  const vorn = vi.fn(textur)
  const fremd = vi.fn(textur)
  belegeApi({
    Spieltextur: vorn,
    Fremdkartenbild: fremd,
    Rueckentextur: vi.fn(textur),
  })

  const { result } = renderHook(() => useSpieltexturen(true, 'p1'), {
    wrapper: raum,
  })

  act(() => {
    result.current.belege({ ...vorderseite, bild: 'b1', objekt: 'o1' })
  })
  await waitFor(() => {
    expect(vorn).toHaveBeenCalledExactlyOnceWith('1', 'vorne', 'tisch')
  })
  expect(fremd).not.toHaveBeenCalled()
})

it('holt am eigenen Tisch die Vorderseite', async () => {
  const vorn = vi.fn(textur)
  belegeApi({ Spieltextur: vorn, Rueckentextur: vi.fn(textur) })

  const { result } = renderHook(() => useSpieltexturen(false, 'p1'), {
    wrapper: raum,
  })

  act(() => {
    result.current.belege(vorderseite)
  })

  await waitFor(() => {
    expect(vorn).toHaveBeenCalledOnce()
  })
  expect(result.current.befunde).toEqual([])
})

// **Am eigenen Tisch fragt das Fenster den Gast** — P9b (#2377): eine Karte
// aus einem mitgebrachten Deck trägt keine Zeile dieses Bestands, aber eine
// Bildkennung. Ohne beides gibt es nichts zu holen.
it('holt am eigenen Tisch die Karte eines Gastes über die Bildkennung', async () => {
  const vorn = vi.fn(textur)
  const fremd = vi.fn(textur)
  belegeApi({
    Spieltextur: vorn,
    Fremdkartenbild: fremd,
    Rueckentextur: vi.fn(textur),
  })

  const { result } = renderHook(() => useSpieltexturen(false, 'p1'), {
    wrapper: raum,
  })

  act(() => {
    result.current.belege({
      ...vorderseite,
      karteID: '',
      bild: 'b1',
      objekt: 'o1',
    })
    result.current.belege({ ...vorderseite, karteID: '' })
  })
  await waitFor(() => {
    expect(fremd).toHaveBeenCalledExactlyOnceWith(
      'p1',
      'o1',
      'b1',
      'vorne',
      'tisch',
    )
  })
  await effekteNachholen()

  expect(vorn).not.toHaveBeenCalled()
})
