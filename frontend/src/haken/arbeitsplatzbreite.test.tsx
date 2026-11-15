// **Eine Spaltenbreite wünschen** — AP-12
// (#2369),
// Befund B-25: die Breiten standen nur als Einstellung. Der Trenner zwischen
// den Spalten schreibt in dieselben Schlüssel, und das Gedächtnis hält dabei
// die Regeln aus §7: sofort zu sehen, ein Vorgang je Breite, der letzte
// Wunsch gewinnt, ein Fehlschlag nimmt nur diese Breite zurück.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'

import { belegeApi, raeumeApi } from '../test/api'
import { Sprachraum } from '../sprache/kontext'
import { useArbeitsplatzgedaechtnis } from './useArbeitsplatzgedaechtnis'

function einstellungen(werte: Record<string, string>) {
  return () =>
    Promise.resolve({
      einstellungen: Object.entries(werte).map(([schluessel, wert]) => ({
        schluessel,
        wert,
        wirksam: wert,
        werte: [],
        bereich: 'arbeitsplatz',
      })),
    })
}

const vollstaendig = {
  'arbeitsplatz.navigator': 'sichtbar',
  'arbeitsplatz.inspektor': 'sichtbar',
  'arbeitsplatz.navigatorbreite': '240',
  'arbeitsplatz.inspektorbreite': '320',
}

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

async function aufbauen() {
  const ansicht = renderHook(() => useArbeitsplatzgedaechtnis(), {
    wrapper: Rahmen,
  })

  await waitFor(() => {
    expect(ansicht.result.current.bereit).toBe(true)
  })

  return ansicht
}

afterEach(() => {
  raeumeApi()
})

describe('eine Spaltenbreite wünschen', () => {
  it('zeigt die Breite sofort und schreibt sie unter ihrem Schlüssel', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({})
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })
    const { result } = await aufbauen()

    act(() => {
      result.current.breiteWuenschen('inspector', 400)
    })

    expect(result.current.inspectorbreite).toBe(400)
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledExactlyOnceWith(
        'arbeitsplatz.inspektorbreite',
        '400',
      )
    })
    expect(result.current.navigatorbreite).toBe(240)
  })

  // **Ein Vorgang je Breite** (§7, Regel 4) — und der letzte Wunsch gilt:
  // wer mit den Pfeiltasten drei Schritte geht, während der erste noch
  // geschrieben wird, will den dritten, nicht den ersten.
  it('schreibt nach einem laufenden Schreiben nur den letzten Wunsch', async () => {
    let freigeben: (wert: unknown) => void = () => undefined
    const EinstellungSetzen = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((erfuellen) => {
            freigeben = erfuellen
          }),
      )
      .mockResolvedValue({})
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })
    const { result } = await aufbauen()

    act(() => {
      result.current.breiteWuenschen('navigator', 256)
    })
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    act(() => {
      result.current.breiteWuenschen('navigator', 272)
      result.current.breiteWuenschen('navigator', 288)
    })

    expect(result.current.navigatorbreite).toBe(288)
    expect(EinstellungSetzen).toHaveBeenCalledTimes(1)

    await act(async () => {
      freigeben({})
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(2)
    })
    expect(EinstellungSetzen).toHaveBeenLastCalledWith(
      'arbeitsplatz.navigatorbreite',
      '288',
    )
  })

  // **Ein älterer Fehlschlag nimmt keinen jüngeren Wunsch zurück** —
  // Zweitblick über Phase 6 von 0.9.0, zweiter Teil, Befund 4: scheiterte
  // 256, während 288 schon wartete, war 288 verloren.
  it('schreibt nach einem gescheiterten Schreiben den jüngeren Wunsch', async () => {
    let ablehnen: (wert: unknown) => void = () => undefined
    let zweites: (wert: unknown) => void = () => undefined
    const EinstellungSetzen = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((erfuellen) => {
            ablehnen = erfuellen
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((erfuellen) => {
            zweites = erfuellen
          }),
      )
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })
    const { result } = await aufbauen()

    act(() => {
      result.current.breiteWuenschen('navigator', 256)
    })
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    act(() => {
      result.current.breiteWuenschen('navigator', 288)
    })

    await act(async () => {
      ablehnen({
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'Der Bestand ist gesperrt.',
        },
      })
      await Promise.resolve()
    })

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(2)
    })
    expect(EinstellungSetzen).toHaveBeenLastCalledWith(
      'arbeitsplatz.navigatorbreite',
      '288',
    )

    // **Gewartet wird auf die Antwort des zweiten Schreibens** (Nachlauf,
    // R7) — 288 stand schon vorher als Wunsch da.
    await act(async () => {
      zweites({})
      await Promise.resolve()
    })
    await waitFor(() => {
      expect(result.current.fehler).toBeUndefined()
    })
    expect(result.current.navigatorbreite).toBe(288)
  })

  // **Ein gescheiterter Wunsch fällt auf die bestätigte Breite zurück, nicht
  // auf die geladene** — Nachlauf zum Zweitblick über Phase 6, zweiter Teil,
  // R4: 256 war geschrieben, 288 scheiterte, und angezeigt wurden wieder die
  // 240 vom Start.
  it('nimmt einen gescheiterten Wunsch auf die zuletzt geschriebene Breite zurück', async () => {
    const EinstellungSetzen = vi
      .fn()
      .mockResolvedValueOnce({})
      .mockResolvedValueOnce({
        fehler: {
          code: 'validation',
          meldung: '288 Punkte passen nicht.',
        },
      })
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })
    const { result } = await aufbauen()
    expect(result.current.navigatorbreite).toBe(240)

    act(() => {
      result.current.breiteWuenschen('navigator', 256)
    })
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    })
    act(() => {
      result.current.breiteWuenschen('navigator', 288)
    })

    await waitFor(() => {
      expect(result.current.fehler).toBe('288 Punkte passen nicht.')
    })
    expect(EinstellungSetzen).toHaveBeenCalledTimes(2)
    expect(result.current.navigatorbreite).toBe(256)
  })

  // **Rücknahme nur des betroffenen Datensatzes** (§7).
  it('nimmt bei einem Fehlschlag nur diese Breite zurück und sagt warum', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({
      fehler: {
        code: 'validation',
        meldung: '900 Punkte liegen außerhalb von 120 bis 800.',
      },
    })
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })
    const { result } = await aufbauen()

    act(() => {
      result.current.breiteWuenschen('navigator', 300)
    })
    expect(result.current.navigatorbreite).toBe(300)

    await waitFor(() => {
      expect(result.current.fehler).toBe(
        '900 Punkte liegen außerhalb von 120 bis 800.',
      )
    })
    expect(result.current.navigatorbreite).toBe(240)
    expect(result.current.inspectorbreite).toBe(320)
  })

  it('zeigt ohne die Schlüssel die Breite, schreibt aber nichts', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({})
    belegeApi({
      Einstellungen: einstellungen({ sprache: 'de' }),
      EinstellungSetzen,
    })
    const { result } = renderHook(() => useArbeitsplatzgedaechtnis(), {
      wrapper: Rahmen,
    })
    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    act(() => {
      result.current.breiteWuenschen('navigator', 300)
    })

    expect(result.current.navigatorbreite).toBe(300)
    // **Ein zweites Ereignis durch denselben Weg** (§3): die Sichtbarkeit
    // wird geschrieben, wo es ihren Schlüssel gibt — hier nicht, und die
    // Breite auch nicht.
    await act(async () => {
      await Promise.resolve()
    })
    expect(EinstellungSetzen).not.toHaveBeenCalled()
  })
})
