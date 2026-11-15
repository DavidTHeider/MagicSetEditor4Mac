// **Der Haken hinter dem portablen Umschlag** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Wonach hier gefragt wird
//
// Nach der **Trennung von Ansehen und Übernehmen**, nach dem abgebrochenen
// Dialog, der keine Meldung wert ist, und nach §7 Regel 3: ein Fehler lässt
// die Vorschau stehen, damit es ein zweites Mal geht, ohne noch einmal zu
// wählen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { useUmschlag } from './useUmschlag'

afterEach(() => {
  raeumeApi()
})

function rahmen({ children }: { children: React.ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function zeige() {
  return renderHook(() => useUmschlag(), { wrapper: rahmen })
}

const leer = {
  objekte: [],
  urheber: '',
  fingerabdruck: '',
  pfad: '',
  abgebrochen: false,
  fehler: null,
}

const einUmschlag = {
  ...leer,
  pfad: '/tmp/u.msepackage',
  urheber: 'k51q',
  fingerabdruck: 'AB CD',
  objekte: [
    { art: 'set', bezeichnung: 'set.mse-set', groesse: 12, braucht: [] },
  ],
}

describe('schnüren', () => {
  it('reicht die Setkennung durch und hält das Ergebnis am Set fest', async () => {
    const gerufen = vi
      .fn()
      .mockResolvedValue({ ...leer, objekte: einUmschlag.objekte })
    belegeApi({ UmschlagSchnueren: gerufen })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('7')
    })

    await waitFor(() => {
      expect(result.current.geschnuert?.objekte).toHaveLength(1)
    })
    // **Die Art steht vorn** (T7, #1472) — `set` ist die Vorgabe.
    expect(gerufen).toHaveBeenCalledWith('set', '7')
    expect(result.current.fuer).toBe('7')
  })

  it('sagt nichts, wenn jemand den Dialog wegklickt', async () => {
    const gerufen = vi.fn().mockResolvedValue({ ...leer, abgebrochen: true })
    belegeApi({ UmschlagSchnueren: gerufen })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('7')
    })

    await waitFor(() => {
      expect(gerufen).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(result.current.laeuft('schnueren-7')).toBe(false)
    })

    expect(result.current.geschnuert).toBeUndefined()
    expect(result.current.fehler).toBeUndefined()
  })

  it('meldet den Fehler und schreibt keinen Umschlag fest', async () => {
    belegeApi({
      UmschlagSchnueren: vi.fn().mockResolvedValue({
        ...leer,
        fehler: { code: 'validation', meldung: 'geht nicht', schluessel: '' },
      }),
    })

    const { result } = zeige()

    act(() => {
      result.current.schnueren('7')
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe('geht nicht')
    })
    expect(result.current.geschnuert).toBeUndefined()
  })
})

describe('ansehen und übernehmen', () => {
  it('zeigt erst, was drin ist, und schreibt nichts', async () => {
    const angesehen = vi.fn().mockResolvedValue(einUmschlag)
    const uebernommen = vi.fn()
    belegeApi({ UmschlagWaehlen: angesehen, UmschlagUebernehmen: uebernommen })

    const { result } = zeige()

    act(() => {
      result.current.waehlen()
    })

    await waitFor(() => {
      expect(result.current.angesehen?.pfad).toBe('/tmp/u.msepackage')
    })

    // **Das ist die Zusage aus #1457**: Ansehen schreibt nicht.
    expect(uebernommen).not.toHaveBeenCalled()
  })

  it('nimmt danach herein und räumt die Vorschau weg', async () => {
    belegeApi({
      UmschlagWaehlen: vi.fn().mockResolvedValue(einUmschlag),
      UmschlagUebernehmen: vi.fn().mockResolvedValue({
        setID: '9',
        karten: 3,
        befunde: [],
        angelegt: [],
        abgebrochen: false,
        schonOffen: false,
        unvollstaendig: false,
        fehler: null,
      }),
    })

    const { result } = zeige()

    act(() => {
      result.current.waehlen()
    })
    await waitFor(() => {
      expect(result.current.angesehen).toBeDefined()
    })

    act(() => {
      result.current.uebernehmen('/tmp/u.msepackage')
    })

    await waitFor(() => {
      expect(result.current.uebernommen).toBe('9')
    })
    expect(result.current.angesehen).toBeUndefined()
  })

  it('lässt die Vorschau nach einem gescheiterten Übernehmen stehen', async () => {
    belegeApi({
      UmschlagWaehlen: vi.fn().mockResolvedValue(einUmschlag),
      UmschlagUebernehmen: vi.fn().mockResolvedValue({
        setID: '',
        karten: 0,
        befunde: [],
        angelegt: [],
        abgebrochen: false,
        schonOffen: false,
        unvollstaendig: false,
        fehler: { code: 'validation', meldung: 'kaputt', schluessel: '' },
      }),
    })

    const { result } = zeige()

    act(() => {
      result.current.waehlen()
    })
    await waitFor(() => {
      expect(result.current.angesehen).toBeDefined()
    })

    act(() => {
      result.current.uebernehmen('/tmp/u.msepackage')
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe('kaputt')
    })

    // **§7 Regel 3**: der angesehene Umschlag bleibt liegen — sonst müsste
    // jemand nach einem Fehler noch einmal wählen.
    expect(result.current.angesehen?.pfad).toBe('/tmp/u.msepackage')
    expect(result.current.uebernommen).toBeUndefined()
  })

  it('sagt nichts, wenn die Wahl weggeklickt wird', async () => {
    const gerufen = vi.fn().mockResolvedValue({ ...leer, abgebrochen: true })
    belegeApi({ UmschlagWaehlen: gerufen })

    const { result } = zeige()

    act(() => {
      result.current.waehlen()
    })

    await waitFor(() => {
      expect(gerufen).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(result.current.laeuft('umschlag-waehlen')).toBe(false)
    })

    expect(result.current.angesehen).toBeUndefined()
    expect(result.current.fehler).toBeUndefined()
  })

  it('meldet einen Fehler beim Ansehen', async () => {
    belegeApi({
      UmschlagWaehlen: vi.fn().mockResolvedValue({
        ...leer,
        fehler: { code: 'validation', meldung: 'nicht lesbar', schluessel: '' },
      }),
    })

    const { result } = zeige()

    act(() => {
      result.current.waehlen()
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe('nicht lesbar')
    })
    expect(result.current.angesehen).toBeUndefined()
  })
})

it('räumt alles weg, was dastand', async () => {
  belegeApi({ UmschlagWaehlen: vi.fn().mockResolvedValue(einUmschlag) })

  const { result } = zeige()

  act(() => {
    result.current.waehlen()
  })
  await waitFor(() => {
    expect(result.current.angesehen).toBeDefined()
  })

  act(() => {
    result.current.raeumen()
  })

  expect(result.current.angesehen).toBeUndefined()
  expect(result.current.geschnuert).toBeUndefined()
  expect(result.current.fuer).toBeUndefined()
  expect(result.current.uebernommen).toBeUndefined()
  expect(result.current.fehler).toBeUndefined()
})
