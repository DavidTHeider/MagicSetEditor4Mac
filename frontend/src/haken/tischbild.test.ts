// **Das eigene Tischbild im Fenster** — Werkstück G3 von Phase 5.5
// (#1460).
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einNutzer } from '../test/anmeldebelegung'
import { useTischbild } from './useTischbild'

afterEach(() => {
  raeumeApi()
})

const profil = () => einNutzer().profile[0] ?? null

const belege = (dazu: Parameters<typeof belegeApi>[0] = {}) => {
  belegeApi({
    GezeigtesProfil: () => Promise.resolve({ profil: profil() }),
    Tischbild: () => Promise.resolve({ adresse: '' }),
    ...dazu,
  })
}

describe('useTischbild', () => {
  it('holt die Adresse des abgelegten Bilds', async () => {
    belege({
      Tischbild: () => Promise.resolve({ adresse: '/tischbild/1.png' }),
    })

    const lauf = renderHook(() => useTischbild())

    await waitFor(() => {
      expect(lauf.result.current.adresse).toBe('/tischbild/1.png')
    })
  })

  it('bleibt leer, wo keines abgelegt ist', async () => {
    belege()

    const lauf = renderHook(() => useTischbild())

    await waitFor(() => {
      expect(lauf.result.current.laeuft).toBe(false)
    })
    expect(lauf.result.current.adresse).toBe('')
  })

  it('wählt eines und nimmt die neue Adresse', async () => {
    const waehlen = vi.fn(() =>
      Promise.resolve({ adresse: '/tischbild/1.png' }),
    )
    belege({ TischbildWaehlen: waehlen })

    // **Erst wenn die Nutzerkennung da ist, hat das Wählen ein Ziel** — und
    // sie kommt mit dem Profil, nicht mit dem ersten Render. Ohne dieses
    // Warten prüfte der Lauf einen Haken, der noch niemandem gehört (§3).
    const lauf = renderHook(() => useTischbild())
    const anfangs = lauf.result.current.waehlen

    await waitFor(() => {
      expect(lauf.result.current.waehlen).not.toBe(anfangs)
    })

    act(() => {
      lauf.result.current.waehlen()
    })

    await waitFor(() => {
      expect(lauf.result.current.adresse).toBe('/tischbild/1.png')
    })
    expect(waehlen).toHaveBeenCalledWith(einNutzer().id)
  })

  it('nimmt es wieder weg', async () => {
    belege({
      Tischbild: () => Promise.resolve({ adresse: '/tischbild/1.png' }),
      TischbildEntfernen: () => Promise.resolve({ adresse: '' }),
    })

    const lauf = renderHook(() => useTischbild())
    await waitFor(() => {
      expect(lauf.result.current.adresse).toBe('/tischbild/1.png')
    })

    act(() => {
      lauf.result.current.entfernen()
    })

    await waitFor(() => {
      expect(lauf.result.current.adresse).toBe('')
    })
  })

  // **Ein gescheitertes Schreiben lädt nicht nach** (§7, Regel 3): ein
  // Lesefehler überschriebe sonst die Meldung, die sagt, was zu tun ist.
  it('meldet, was beim Wählen schiefging, und lässt das Bild stehen', async () => {
    belege({
      Tischbild: () => Promise.resolve({ adresse: '/tischbild/1.png' }),
      TischbildWaehlen: () =>
        Promise.resolve({
          adresse: '',
          fehler: {
            code: 'validierung' as const,
            schluessel: 'bild.unlesbar',
            meldung: 'Kein Bild.',
          },
        }),
    })

    const lauf = renderHook(() => useTischbild())
    await waitFor(() => {
      expect(lauf.result.current.adresse).toBe('/tischbild/1.png')
    })

    act(() => {
      lauf.result.current.waehlen()
    })

    await waitFor(() => {
      expect(lauf.result.current.fehler).toBe('Kein Bild.')
    })
    expect(lauf.result.current.adresse).toBe('/tischbild/1.png')
  })

  // **Ohne Nutzer gibt es nichts zu wählen** — ein Zuschauer ohne Profil
  // hätte sonst einen Dialog ohne Ziel.
  it('bleibt still, wo kein Nutzer dahintersteht', async () => {
    const waehlen = vi.fn(() => Promise.resolve({ adresse: '' }))
    belegeApi({
      GezeigtesProfil: () => Promise.resolve({ profil: null }),
      TischbildWaehlen: waehlen,
    })

    const lauf = renderHook(() => useTischbild())
    await waitFor(() => {
      expect(lauf.result.current.laeuft).toBe(false)
    })

    act(() => {
      lauf.result.current.waehlen()
    })

    expect(waehlen).not.toHaveBeenCalled()
  })
})
