// **Was die Kommandozentrale tut — ohne DOM** — Werkstück U12 von Phase 5.1 (#1423).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { useKommandozentrale } from './useKommandozentrale'
import { leere } from '../befehlsverlauf'
import { Bindungsfehler } from '../bindungsfehler'
import { belegeApi } from '../test/api'
import type { BefehlDTO } from '../apibefehle'

const befehl: BefehlDTO = {
  name: 'lesen.probe',
  faehigkeit: '',
  cli: '',
  rest: '',
  erklaerung: '',
  schreibend: false,
  idempotent: true,
  gefaehrlich: false,
  schema: {
    type: 'object',
    properties: {
      set: { type: 'string' },
      zahl: { type: 'integer' },
      haken: { type: 'boolean' },
      zeilen: { type: 'array' },
    },
    required: ['set'],
  },
}

afterEach(() => {
  leere()
})

describe('useKommandozentrale', () => {
  it('übernimmt Werte aus dem Verlauf in ihrer Gestalt', () => {
    belegeApi({})
    const { result } = renderHook(() => useKommandozentrale(befehl))

    act(() => {
      result.current.uebernehmen({
        set: '7',
        zahl: 3,
        haken: true,
        zeilen: [1],
      })
    })

    expect(result.current.werte).toEqual({
      set: '7',
      zahl: '3',
      haken: true,
      zeilen: '[1]',
    })
    expect(result.current.bereit).toBe(true)
  })

  it('führt nicht aus, was nicht bereit ist — und ein Abbruch darf scheitern', async () => {
    const BefehlAusfuehren = vi.fn()
    const BefehlAbbrechen = vi.fn().mockRejectedValue(new Error('weg'))
    belegeApi({ BefehlAusfuehren, BefehlAbbrechen })
    const { result } = renderHook(() => useKommandozentrale(befehl))

    act(() => {
      result.current.ausfuehren()
    })
    expect(BefehlAusfuehren).not.toHaveBeenCalled()

    act(() => {
      result.current.abbrechen()
    })
    await waitFor(() => {
      expect(BefehlAbbrechen).toHaveBeenCalledTimes(1)
    })
  })
})

// **Eine abgewiesene Zusage ist ein Ausgang** — Befund 8 des Zweitblicks über
// Phase 5.1 (2026-09-10). `then` und `finally` allein ließen die Ablehnung
// ohne Empfänger: `laeuft` ging zurück, und sonst geschah nichts.
describe('wenn die Naht den Befehl gar nicht annimmt', () => {
  it('zeigt den Ausgang, die Meldung und schreibt in den Verlauf', async () => {
    const BefehlAusfuehren = vi
      .fn()
      .mockRejectedValue(
        new Error('Die Methode BefehlAusfuehren ist nicht belegt.'),
      )
    belegeApi({ BefehlAusfuehren })

    const { result } = renderHook(() => useKommandozentrale(befehl))

    act(() => {
      result.current.setzeWert('set', '7')
    })
    act(() => {
      result.current.ausfuehren()
    })

    await waitFor(() => {
      expect(result.current.ergebnis?.ausgang).toBe('gescheitert')
    })

    expect(result.current.ergebnis?.fehler?.meldung).toBe(
      'Die Methode BefehlAusfuehren ist nicht belegt.',
    )

    // **Und der Lauf steht nicht mehr** — sonst bliebe der Knopf für immer
    // angehalten.
    expect(result.current.laeuft).toBe(false)
  })

  it('übernimmt Code und Schlüssel, wenn die Naht sie mitschickt', async () => {
    const BefehlAusfuehren = vi.fn().mockRejectedValue(
      new Bindungsfehler({
        code: 'erlaubnis',
        schluessel: 'erlaubnis',
        meldung: 'Dieses Profil darf das nicht.',
      }),
    )
    belegeApi({ BefehlAusfuehren })

    const { result } = renderHook(() => useKommandozentrale(befehl))

    act(() => {
      result.current.setzeWert('set', '7')
    })
    act(() => {
      result.current.ausfuehren()
    })

    await waitFor(() => {
      expect(result.current.ergebnis?.fehler?.code).toBe('erlaubnis')
    })

    expect(result.current.ergebnis?.fehler?.schluessel).toBe('erlaubnis')
  })

  it('nimmt auch etwas an, das gar kein Fehler ist', async () => {
    // **Eine Zusage darf mit allem abgewiesen werden** — auch mit einer Zahl.
    // Ohne diesen Zweig stünde dort `undefined`.
    const BefehlAusfuehren = vi.fn().mockRejectedValue('weg')
    belegeApi({ BefehlAusfuehren })

    const { result } = renderHook(() => useKommandozentrale(befehl))

    act(() => {
      result.current.setzeWert('set', '7')
    })
    act(() => {
      result.current.ausfuehren()
    })

    await waitFor(() => {
      expect(result.current.ergebnis?.fehler?.meldung).toBe('weg')
    })
  })
})
