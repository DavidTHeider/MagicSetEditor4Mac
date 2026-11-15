// **Die Seitenleiste sagt der Menüleiste, wie sie dasteht** — UXR-F14
// (#2029).
//
// *Darstellung* nennt den Eintrag nach dem Stand und graut ihn, wo es nichts
// einzuklappen gibt. Der Haken hält die Wahrheit, die der Menüeintrag auch
// schaltet (`seitenleiste.umschalten`), und meldet sie in `vorhanden` und
// `offen` — hier gegen einen mitschreibenden Melder statt gegen Go.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'

import {
  Darstellungsmeldung,
  fehlt,
  type Melder,
  type Teilmeldung,
} from '../darstellungsstand'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { useEinstellungen } from './useEinstellungen'
import { seitenleistenschluessel, useSeitenleiste } from './useSeitenleiste'

afterEach(() => {
  raeumeApi()
})

function einstellungen(werte: Record<string, string>) {
  return () =>
    Promise.resolve({
      einstellungen: Object.entries(werte).map(([schluessel, wert]) => ({
        schluessel,
        wert,
        wirksam: wert,
        werte: [],
        bereich: 'geraet',
      })),
    })
}

// aufbauen baut den Haken in einem Raum, der jede Meldung festhält.
function aufbauen() {
  const meldungen: Teilmeldung[] = []
  const melder: Melder = (meldung) => {
    meldungen.push(meldung)
  }

  const huelle = ({ children }: { children: ReactNode }) => (
    <Sprachraum sprache="de">
      <Darstellungsmeldung value={melder}>{children}</Darstellungsmeldung>
    </Sprachraum>
  )

  const ergebnis = renderHook(() => useSeitenleiste(useEinstellungen()), {
    wrapper: huelle,
  })

  // letzte gibt, was die Seitenleiste zuletzt gemeldet hat.
  const letzte = () => meldungen.filter((m) => m.teil === 'seitenleiste').at(-1)

  return { ...ergebnis, letzte }
}

describe('useSeitenleiste meldet der Menüleiste', () => {
  // **Offen, solange nichts geladen ist** — wie die Leiste selbst: eine, die
  // beim Start kurz fehlt und dann erscheint, springt.
  it('meldet die Leiste links als vorhanden und offen', async () => {
    belegeApi({ Einstellungen: einstellungen({}) })

    const { letzte } = aufbauen()

    expect(letzte()).toEqual({
      teil: 'seitenleiste',
      wert: { vorhanden: true, offen: true },
    })

    // Die Einstellungen kommen eine Aufgabe später; ihr Eintrag gehört noch
    // in diesen Test und nicht in den nächsten.
    await antwortAbwarten()
  })

  it('meldet sie zu, wenn die Einstellung sie verbirgt', async () => {
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'verborgen' }),
    })

    const { letzte } = aufbauen()

    await waitFor(() => {
      expect(letzte()?.wert).toEqual({ vorhanden: true, offen: false })
    })
  })

  // **Der Eintrag und die Meldung sind derselbe Weg**: was `umschalten` tut,
  // steht danach in der Leiste.
  it('meldet den Stand nach dem Umschalten', async () => {
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'sichtbar' }),
      EinstellungSetzen: vi.fn(() => Promise.resolve({ fehler: null })),
    })

    const { result, letzte } = aufbauen()
    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.umschalten()
    })

    expect(letzte()?.wert).toEqual({ vorhanden: true, offen: false })
  })

  // **Oben gibt es nichts einzuklappen** (Befund G6 des Zweitblicks über Phase
  // 6.3) — der Eintrag ist dort gegraut, und ⌥⌘S täte ohnehin nichts.
  it('meldet die Leiste oben als nicht vorhanden', async () => {
    belegeApi({
      Einstellungen: einstellungen({ leistenlage: 'waagerecht' }),
    })

    const { letzte } = aufbauen()

    await waitFor(() => {
      expect(letzte()?.wert).toEqual({ vorhanden: false, offen: true })
    })
  })

  it('nimmt die Meldung zurück, wenn der Haken geht', async () => {
    belegeApi({ Einstellungen: einstellungen({}) })

    const { unmount, letzte } = aufbauen()
    unmount()

    expect(letzte()?.wert).toEqual(fehlt)
    await antwortAbwarten()
  })
})
