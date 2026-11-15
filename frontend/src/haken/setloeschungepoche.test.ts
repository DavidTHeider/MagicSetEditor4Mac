// **Ein gelöschtes Set merkt sich das Fenster** (#2306) — damit eine
// Finder-Meldung, die vor dem Löschen eingelesen wurde, es nicht wieder
// aufschlägt. Nur ein **geglücktes** Löschen zählt.
import { act, renderHook } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import {
  epocheGesehen,
  epocheVergessen,
  kennungIstGeloescht,
} from '../bestandsepoche'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { useSetloeschung } from './useSetloeschung'

afterEach(() => {
  raeumeApi()
  epocheVergessen()
})

it('merkt sich das Set nach einem geglückten Löschen', async () => {
  belegeApi({
    SetLoeschen: () => Promise.resolve({ befunde: [], fehler: null }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(kataloge.de, vi.fn(), vi.fn()),
  )

  await act(async () => {
    await result.current.loeschen('7', 'Mein Set')
  })

  expect(kennungIstGeloescht('7')).toBe(true)
})

it('merkt sich nichts, wenn das Löschen scheitert', async () => {
  belegeApi({
    SetLoeschen: () =>
      Promise.resolve({
        befunde: [],
        fehler: {
          code: 'konflikt',
          meldung: 'Das Set heißt anders.',
          schluessel: 'konflikt',
        },
      }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(kataloge.de, vi.fn(), vi.fn()),
  )

  await act(async () => {
    await result.current.loeschen('7', 'Falsch')
  })

  expect(kennungIstGeloescht('7')).toBe(false)
})

// **Eine Löschantwort aus dem vorigen Bestand merkt sich nichts** (Zweitblick
// über die Folgen vor 0.8.0, Befund 5): während das Löschen lief, wurde
// zurückgespielt, und dieselbe Kennung kann im neuen Bestand ein anderes Set
// meinen — eine Finder-Meldung dafür verwürfe das Fenster sonst für immer.
it('merkt sich nichts, wenn der Bestand inzwischen getauscht wurde', async () => {
  epocheGesehen(1)

  let antworte: () => void = () => undefined
  belegeApi({
    SetLoeschen: () =>
      new Promise((ja) => {
        antworte = () => {
          ja({ befunde: [], fehler: null })
        }
      }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(kataloge.de, vi.fn(), vi.fn()),
  )

  let fertig: Promise<void> = Promise.resolve()
  act(() => {
    fertig = result.current.loeschen('7', 'Mein Set')
  })

  epocheGesehen(2)

  await act(async () => {
    antworte()
    await fertig
  })

  expect(kennungIstGeloescht('7')).toBe(false)
})

// **Auch ein abgelehnter Ruf** — die Brücke selbst scheitert, statt eine
// Fehlerantwort zu geben — zeigt nach dem Tausch keinen Fehler mehr an.
it('ein abgelehnter Ruf aus dem vorigen Bestand zeigt keinen Fehler', async () => {
  epocheGesehen(1)

  let lehneAb: () => void = () => undefined
  belegeApi({
    SetLoeschen: () =>
      new Promise((_, nein) => {
        lehneAb = () => {
          nein(new Error('Brücke weg'))
        }
      }),
  })

  const { result } = renderHook(() =>
    useSetloeschung(kataloge.de, vi.fn(), vi.fn()),
  )

  let fertig: Promise<void> = Promise.resolve()
  act(() => {
    fertig = result.current.loeschen('7', 'Mein Set')
  })

  epocheGesehen(2)

  await act(async () => {
    lehneAb()
    await fertig
  })

  expect(result.current.fehler('7')).toBeUndefined()
})

// **Und sie tut auch sonst nichts** (Nachlauf über die Folgen vor 0.8.0,
// Befund 1): weder meldet sie das Set als gelöscht — das schlösse womöglich
// ein Set, das inzwischen offen ist — noch lädt sie die Liste nach oder
// zeigt Reste oder einen Fehler. Sie gehört zu einem Bestand, den das
// Fenster nicht mehr zeigt (§7, Regel 2).
it.each([
  ['geglückt', { befunde: ['Rest'], fehler: null }],
  [
    'gescheitert',
    {
      befunde: [],
      fehler: {
        code: 'konflikt' as const,
        meldung: 'Anders.',
        schluessel: 'konflikt',
      },
    },
  ],
])(
  'eine %s Antwort aus dem vorigen Bestand ändert nichts im Fenster',
  async (_, antwortwert) => {
    epocheGesehen(1)

    let antworte: () => void = () => undefined
    belegeApi({
      SetLoeschen: () =>
        new Promise((ja) => {
          antworte = () => {
            ja(antwortwert)
          }
        }),
    })

    const nachladen = vi.fn()
    const geloescht = vi.fn()
    const { result } = renderHook(() =>
      useSetloeschung(kataloge.de, nachladen, geloescht),
    )

    let fertig: Promise<void> = Promise.resolve()
    act(() => {
      fertig = result.current.loeschen('7', 'Mein Set')
    })

    epocheGesehen(2)

    await act(async () => {
      antworte()
      await fertig
    })

    expect(geloescht).not.toHaveBeenCalled()
    expect(nachladen).not.toHaveBeenCalled()
    expect(result.current.befunde('7')).toEqual([])
    expect(result.current.fehler('7')).toBeUndefined()
  },
)
