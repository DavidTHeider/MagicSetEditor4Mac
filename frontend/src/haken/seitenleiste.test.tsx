// **Die Seitenleiste merkt sich, ob sie offen ist** — Werkstück U14 von
// Phase 5.1 (#1490): eine Einstellung auf Geräteebene, optimistisch gesetzt
// und bei einem Fehler zurückgenommen (§7, Regel 3).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ReactNode } from 'react'

import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Sprachraum } from '../sprache/kontext'
import { useEinstellungen } from './useEinstellungen'
import { seitenleistenschluessel, useSeitenleiste } from './useSeitenleiste'

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

function Rahmen({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function aufbauen() {
  // **Mit dem Stand, den `App` hält** — der Haken lädt nichts selbst.
  return renderHook(() => useSeitenleiste(useEinstellungen()), {
    wrapper: Rahmen,
  })
}

afterEach(() => {
  raeumeApi()
})

describe('useSeitenleiste', () => {
  // **Die Lage kommt aus derselben Einstellungsliste** — UX-14.AK-02
  // (#1686): senkrecht ist die Vorgabe, waagerecht die Wahl.
  it('liest die Lage der Leiste', async () => {
    belegeApi({
      Einstellungen: einstellungen({ leistenlage: 'waagerecht' }),
    })

    const { result } = aufbauen()
    expect(result.current.waagerecht).toBe(false)

    await waitFor(() => {
      expect(result.current.waagerecht).toBe(true)
    })
  })

  it('ist offen, solange nichts geladen ist, und liest dann die Einstellung', async () => {
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'verborgen' }),
    })

    const { result } = aufbauen()
    expect(result.current.offen).toBe(true)

    await waitFor(() => {
      expect(result.current.offen).toBe(false)
    })
  })

  it('schaltet um, schreibt die Einstellung und meldet den Lauf', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'sichtbar' }),
      EinstellungSetzen,
    })

    const { result } = aufbauen()
    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.umschalten()
    })
    expect(result.current.offen).toBe(false)

    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenCalledWith(
        seitenleistenschluessel,
        'verborgen',
      )
    })
    await waitFor(() => {
      expect(result.current.laeuft).toBe(false)
    })

    act(() => {
      result.current.umschalten()
    })
    await waitFor(() => {
      expect(EinstellungSetzen).toHaveBeenLastCalledWith(
        seitenleistenschluessel,
        'sichtbar',
      )
    })
    expect(result.current.offen).toBe(true)
  })

  // **Vor dem Laden gilt der Wunsch, geschrieben wird nichts**: wer klickt,
  // bevor die Einstellungen da sind, sieht die Leiste trotzdem gehen.
  it('schaltet vor dem Laden nur örtlich um', () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({
      Einstellungen: () => new Promise(() => undefined),
      EinstellungSetzen,
    })

    const { result } = aufbauen()
    expect(result.current.bereit).toBe(false)

    act(() => {
      result.current.umschalten()
    })

    expect(result.current.offen).toBe(false)
    expect(EinstellungSetzen).not.toHaveBeenCalled()
  })

  // **Geschrieben ist geschrieben — und nicht geschrieben ist zurückgenommen**
  // (§7, Regel 3): der Fehler steht da, die Leiste steht, wie sie stand.
  it('nimmt den Wunsch zurück, wenn das Schreiben scheitert', async () => {
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'sichtbar' }),
      EinstellungSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung' as const,
            schluessel: 'probe',
            meldung: 'Der Bestand ist schreibgeschützt — erst entsperren.',
          },
        }),
    })

    const { result } = aufbauen()
    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.umschalten()
    })
    expect(result.current.offen).toBe(false)

    await waitFor(() => {
      expect(result.current.fehler).toContain('schreibgeschützt')
    })
    expect(result.current.offen).toBe(true)
  })

  // **Ein Vorgang je Datensatz** (§7, Regel 4): ein zweiter Klick während des
  // ersten Laufs tut nichts.
  it('lässt keinen zweiten Lauf zu, solange der erste schreibt', async () => {
    let freigeben: (() => void) | undefined
    const angehalten = new Promise<void>((aufloesen) => {
      freigeben = aufloesen
    })
    const EinstellungSetzen = vi.fn(async () => {
      await angehalten

      return { fehler: null }
    })
    belegeApi({
      Einstellungen: einstellungen({ [seitenleistenschluessel]: 'sichtbar' }),
      EinstellungSetzen,
    })

    const { result } = aufbauen()
    await waitFor(() => {
      expect(result.current.bereit).toBe(true)
    })

    act(() => {
      result.current.umschalten()
    })
    act(() => {
      result.current.umschalten()
    })
    expect(EinstellungSetzen).toHaveBeenCalledTimes(1)
    expect(result.current.offen).toBe(false)

    freigeben?.()
    await waitFor(() => {
      expect(result.current.laeuft).toBe(false)
    })
  })

  // **Mit der Leiste oben gibt es nichts einzuklappen** — Befund G6 des
  // Zweitblicks über Phase 6.3: ⌥⌘S und *Darstellung › Seitenleiste* legten
  // einen unsichtbaren Zustand um, und wer die Leiste zurück nach links
  // stellte, fand sie eingeklappt. Beide Wege enden hier, also sperrt es
  // hier. Dass jemand zuhört, belegt der Lauf »schaltet um« oben — derselbe
  // Ruf ohne `leistenlage`.
  it('schaltet nichts um, solange die Leiste oben steht', async () => {
    const EinstellungSetzen = vi.fn(() => Promise.resolve({ fehler: null }))
    belegeApi({
      Einstellungen: einstellungen({
        [seitenleistenschluessel]: 'sichtbar',
        leistenlage: 'waagerecht',
      }),
      EinstellungSetzen,
    })

    const { result } = aufbauen()
    await waitFor(() => {
      expect(result.current.waagerecht).toBe(true)
    })

    act(() => {
      result.current.umschalten()
    })
    await antwortAbwarten()

    expect(result.current.offen).toBe(true)
    expect(EinstellungSetzen).not.toHaveBeenCalled()
  })
})
