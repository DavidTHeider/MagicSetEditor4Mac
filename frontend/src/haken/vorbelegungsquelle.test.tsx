// **Die Quellen der Vorbelegung** — P6
// (#2374):
// der Name des gezeigten Profils und die zuletzt geöffneten **Decks** (keine
// Sets). Ein Lesefehler kostet nur die Vorbelegung.
import { renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useVorbelegung } from './useVorbelegung'

afterEach(raeumeApi)

const schritt = (art: 'set' | 'deck', kennung: string) => ({
  objekt: { art, kennung, name: kennung, angeheftet: false },
  vorgang: 'geoeffnet' as const,
  zeitpunkt: '',
})

it('liest Profil und zuletzt geöffnete Decks', async () => {
  belegeApi({
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: { id: '1', nutzerID: 'n', name: 'Ada' },
        fehler: null,
      } as never),
    Weiterarbeiten: () =>
      Promise.resolve({
        zuletzt: null,
        geoeffnet: [
          schritt('set', 's1'),
          schritt('deck', 'd2'),
          schritt('deck', 'd1'),
        ],
        eingefuehrt: [],
        ausgegeben: [],
        angeheftet: [],
        fehler: null,
      }),
  })

  const { result } = renderHook(() => useVorbelegung())

  await waitFor(() => {
    expect(result.current).toEqual({
      profilname: 'Ada',
      zuletzt: ['d2', 'd1'],
    })
  })
})

it('gibt ohne Antwort eine leere Vorbelegung', async () => {
  belegeApi({
    GezeigtesProfil: () => Promise.reject(new Error('zu')),
    Weiterarbeiten: () => Promise.reject(new Error('zu')),
  })

  const { result } = renderHook(() => useVorbelegung())

  await waitFor(() => {
    expect(result.current).toEqual({
      profilname: '',
      zuletzt: [],
    })
  })
})

// **Ohne gezeigtes Profil kein Name** — der erste Sitz bleibt dann leer, und
// das Formular sagt, dass er fehlt.
it('lässt den Namen leer, wenn kein Profil gezeigt wird', async () => {
  belegeApi({
    GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
    Weiterarbeiten: () =>
      Promise.resolve({
        zuletzt: null,
        geoeffnet: [schritt('deck', 'd1')],
        eingefuehrt: [],
        ausgegeben: [],
        angeheftet: [],
        fehler: null,
      }),
  })

  const { result } = renderHook(() => useVorbelegung())

  await waitFor(() => {
    expect(result.current).toEqual({ profilname: '', zuletzt: ['d1'] })
  })
})
