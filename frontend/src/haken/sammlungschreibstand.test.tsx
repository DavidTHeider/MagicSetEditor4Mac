// **Ein eigener Schreibvorgang der Sammlung hebt den gemeinsamen Stand** —
// Nachlauf zum Zweitblick über Phase 6 von 0.9.0, zweiter Teil, R3
// (#2366).
//
// Aufnahme, Änderung, Löschen, Einfuhr und Massenvorgänge hoben nur die
// eigene Runde der Sammlung. Die Deckseite bleibt unter `Activity` erhalten,
// und ihr Abgleich „Fehlt in der Sammlung“ las beim Zurückwechseln nicht neu.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { FehlerDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { exemplar } from '../test/sammlungsprobe'
import { useSammlung } from './useSammlung'

afterEach(() => {
  raeumeApi()
})

function belege(
  aendern: () => Promise<FehlerDTO | null> = () => Promise.resolve(null),
) {
  belegeApi({
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [exemplar(0, 'Lightning Bolt')],
        gesamt: 1,
        insgesamt: 1,
        stueck: 1,
        fehler: null,
      }),
    ExemplarAnlegen: () =>
      Promise.resolve({ id: 2, vorhanden: null, fehler: null }),
    ExemplarAendern: vi.fn(aendern),
  })
}

describe('der Stand nach eigenem Schreiben', () => {
  it('meldet ein geglücktes Aufnehmen und einen Massenvorgang', async () => {
    belege()
    const geschrieben = vi.fn()
    const { result } = renderHook(() => useSammlung(undefined, 0, geschrieben))
    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    await act(async () => {
      await result.current.anlegen(exemplar(1, 'Shock'))
    })
    expect(geschrieben).toHaveBeenCalledTimes(1)

    act(() => {
      result.current.nachSchreiben()
    })
    expect(geschrieben).toHaveBeenCalledTimes(2)

    act(() => {
      result.current.nachMasse()
    })
    expect(geschrieben).toHaveBeenCalledTimes(3)
  })

  // **Ein Lesen und ein Fehlschlag sind kein Schreiben** — sonst lüde jede
  // Seite nach einem bloßen „Erneut versuchen“ nach.
  it('meldet weder ein bloßes Nachladen noch einen Fehlschlag', async () => {
    belege(() =>
      Promise.resolve({
        code: 'validierung',
        schluessel: 'validierung',
        meldung: 'Die Menge fehlt.',
      }),
    )
    const geschrieben = vi.fn()
    const { result } = renderHook(() => useSammlung(undefined, 0, geschrieben))
    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    act(() => {
      result.current.nachladen()
    })
    await act(async () => {
      await result.current.aendern(exemplar(0, 'Lightning Bolt'))
    })

    // **Ein Beleg, dass der Fehlschlag angekommen ist.**
    expect(result.current.fehler).toBe(kataloge.de.fehler.validierung)
    expect(geschrieben).not.toHaveBeenCalled()
  })
})
