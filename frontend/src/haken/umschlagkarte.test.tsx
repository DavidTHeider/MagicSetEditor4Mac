// **Wem die Meldung des Umschlags gehört** — UXR-04
// (#1912).
//
// Seit dem Umbau steht, was das **Ablegen** meldet, an der Setkarte, aus der
// es kam — `fuer` sagt, an welcher. Ein Wählen oder Übernehmen schreibt in
// denselben `fehler`; ohne dass es `fuer` räumt, stünde sein Fehler an der
// Karte des zuletzt abgelegten Sets, als wäre dort etwas schiefgegangen. Die
// Regel aus dem Kopf von `useUmschlag` gilt: **der jüngste Ruf besitzt die
// Fläche.**
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { useUmschlag } from './useUmschlag'

afterEach(() => {
  raeumeApi()
})

function rahmen({ children }: { children: React.ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

const leer = {
  objekte: [],
  urheber: '',
  fingerabdruck: '',
  pfad: '',
  abgebrochen: false,
  fehler: null,
}

const kaputt = {
  ...leer,
  fehler: {
    code: 'unbekannt' as const,
    schluessel: 'unbekannt',
    meldung: 'kaputt',
    feld: '',
  },
}

async function abgelegt() {
  const haken = renderHook(() => useUmschlag(), { wrapper: rahmen })

  act(() => {
    haken.result.current.schnueren('7')
  })
  await waitFor(() => {
    expect(haken.result.current.geschnuert).toBeDefined()
  })
  expect(haken.result.current.fuer).toBe('7')

  return haken
}

it('das Wählen nimmt der Karte ihre Meldung, und sein Fehler steht nicht dort', async () => {
  belegeApi({
    UmschlagSchnueren: vi.fn().mockResolvedValue(leer),
    UmschlagWaehlen: vi.fn().mockResolvedValue(kaputt),
  })
  const { result } = await abgelegt()

  act(() => {
    result.current.waehlen()
  })
  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })

  expect(result.current.fuer).toBeUndefined()
  expect(result.current.geschnuert).toBeUndefined()
})

it('die Übernahme ebenso', async () => {
  belegeApi({
    UmschlagSchnueren: vi.fn().mockResolvedValue(leer),
    UmschlagUebernehmen: vi.fn().mockResolvedValue({ ...kaputt, setID: '' }),
  })
  const { result } = await abgelegt()

  act(() => {
    result.current.uebernehmen('/tmp/u.msepackage')
  })
  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })

  expect(result.current.fuer).toBeUndefined()
})
