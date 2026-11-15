// **Das Gedächtnis des Arbeitsplatzes** — Werkstück U4 von Phase 5.1
// (#1415).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'

import { belegeApi, raeumeApi } from '../test/api'
import { Sprachraum } from '../sprache/kontext'
import {
  breiteNavigatorVorgabe,
  breiteInspectorVorgabe,
  useArbeitsplatzgedaechtnis,
} from './useArbeitsplatzgedaechtnis'

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
  'arbeitsplatz.navigator': 'verborgen',
  'arbeitsplatz.inspektor': 'sichtbar',
  'arbeitsplatz.navigatorbreite': '300',
  'arbeitsplatz.inspektorbreite': '400',
}

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function aufbauen() {
  return renderHook(() => useArbeitsplatzgedaechtnis(), { wrapper: Rahmen })
}

afterEach(() => {
  raeumeApi()
})

describe('useArbeitsplatzgedaechtnis', () => {
  it('liest die vier Schlüssel aus den Einstellungen', async () => {
    belegeApi({ Einstellungen: einstellungen(vollstaendig) })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    expect(result.current.navigator).toBe(false)
    expect(result.current.inspector).toBe(true)
    expect(result.current.navigatorbreite).toBe(300)
    expect(result.current.inspectorbreite).toBe(400)
  })

  // **Ein Bestand ohne die Schlüssel** — ein älterer Go-Teil oder eine
  // Antwort, die sie nicht trägt — bekommt die Zahlen aus Vermerk 0076, und
  // `bereit` bleibt falsch: wer schriebe, schriebe einen Schlüssel, den es
  // nicht gibt.
  it('fällt ohne die Schlüssel auf die Zahlen aus 0076 zurück', async () => {
    belegeApi({ Einstellungen: einstellungen({ sprache: 'de' }) })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    expect(result.current.bereit).toBe(false)
    expect(result.current.navigator).toBe(true)
    expect(result.current.inspector).toBe(true)
    expect(result.current.navigatorbreite).toBe(breiteNavigatorVorgabe)
    expect(result.current.inspectorbreite).toBe(breiteInspectorVorgabe)
  })

  it('klappt ohne die Schlüssel nur im Baum und schreibt nichts', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({})
    belegeApi({
      Einstellungen: einstellungen({ sprache: 'de' }),
      EinstellungSetzen,
    })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.geladen).toBe(true)
    })

    act(() => {
      result.current.merke('navigator', false)
    })

    expect(result.current.navigator).toBe(false)
    expect(EinstellungSetzen).not.toHaveBeenCalled()
  })

  it('nimmt eine Breite, die keine Zahl ist, als Vorgabe', async () => {
    belegeApi({
      Einstellungen: einstellungen({
        ...vollstaendig,
        'arbeitsplatz.navigatorbreite': 'breit',
      }),
    })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    expect(result.current.navigatorbreite).toBe(breiteNavigatorVorgabe)
  })

  // **Optimistisch, und die Antwort bestätigt** (§6).
  it('merkt sich einen Wechsel sofort und schreibt ihn in die Einstellungen', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({})
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.merke('navigator', true)
    })

    expect(result.current.navigator).toBe(true)
    expect(result.current.laeuft('navigator')).toBe(true)

    await waitFor(() => {
      expect(result.current.laeuft('navigator')).toBe(false)
    })

    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'arbeitsplatz.navigator',
      'sichtbar',
    )
    expect(result.current.navigator).toBe(true)
    expect(result.current.fehler).toBeUndefined()
  })

  it('schreibt den Inspector unter seinem eigenen Schlüssel', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({})
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.merke('inspector', false)
    })

    await waitFor(() => {
      expect(result.current.laeuft('inspector')).toBe(false)
    })

    expect(EinstellungSetzen).toHaveBeenCalledWith(
      'arbeitsplatz.inspektor',
      'verborgen',
    )
    expect(result.current.inspector).toBe(false)
  })

  // **Rücknahme nur des betroffenen Datensatzes** (§7).
  it('nimmt den Wechsel zurück und nennt den Grund, wenn das Schreiben scheitert', async () => {
    const EinstellungSetzen = vi.fn().mockResolvedValue({
      fehler: { code: 'validation', meldung: 'Diese Breite gibt es nicht.' },
    })
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.merke('navigator', true)
    })

    expect(result.current.navigator).toBe(true)

    await waitFor(() => {
      expect(result.current.fehler).toBe('Diese Breite gibt es nicht.')
    })

    expect(result.current.navigator).toBe(false)
    expect(result.current.inspector).toBe(true)
  })

  // **Ein Vorgang je Datensatz** (§7, Regel 4): der zweite Wechsel derselben
  // Region wartet nicht, er unterbleibt — und der Knopf ist solange angehalten.
  it('lässt je Region nur einen Schreibvorgang zu', async () => {
    let freigeben: (wert: object) => void = () => undefined
    const EinstellungSetzen = vi.fn().mockReturnValue(
      new Promise((aufloesen) => {
        freigeben = aufloesen
      }),
    )
    belegeApi({ Einstellungen: einstellungen(vollstaendig), EinstellungSetzen })

    const { result } = aufbauen()

    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.merke('navigator', true)
    })
    act(() => {
      result.current.merke('navigator', false)
    })

    expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    expect(result.current.navigator).toBe(true)

    act(() => {
      freigeben({})
    })

    await waitFor(() => {
      expect(result.current.laeuft('navigator')).toBe(false)
    })

    // **Der Beleg, dass jemand zuhört**: nach dem Ende geht ein neuer Wechsel.
    act(() => {
      result.current.merke('navigator', false)
    })

    expect(EinstellungSetzen).toHaveBeenCalledTimes(2)
  })
})
