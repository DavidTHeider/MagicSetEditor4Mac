// **Eine Geste wirkt nur in dem Stand, aus dem sie kommt** — GESTE (#2358),
// AGENTS.md §7, Regel 1 und 2.
//
// Der Fall aus dem Issue zuerst: eine Antwort, die sich verspätet, während
// das Profil wechselt, und ein Ziel-Sitz, den es nicht mehr gibt. Beides darf
// **keine fremde Wirkung** haben — kein Rückruf, keine Ansage über einen
// Stand, den niemand mehr sieht.
import { act, cleanup, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Gestenabsicht } from '../bausteine/gestenvertrag'
import { useGeste, type Gestenstand } from './useGeste'

afterEach(cleanup)

const ziele = [
  { kennung: 'friedhof', name: 'Friedhof', sitz: 'sitz-1' },
  { kennung: 'fach:gegenueber', name: 'Vom Gegenüber', sitz: 'sitz-2' },
]

function verzoegert() {
  let fertig: () => void = () => undefined
  let gescheitert: (fehler: unknown) => void = () => undefined
  const versprechen = new Promise<void>((ja, nein) => {
    fertig = ja
    gescheitert = nein
  })

  return { versprechen, fertig, gescheitert }
}

function aufbauen(stand: Gestenstand) {
  const antwort = verzoegert()
  const ausfuehren = vi.fn<(absicht: Gestenabsicht) => Promise<void>>(
    () => antwort.versprechen,
  )
  const nachher = vi.fn()
  const fehler = vi.fn()
  const haken = renderHook(
    (props: Gestenstand) =>
      useGeste({ stand: props, ausfuehren, nachher, fehler }),
    { initialProps: stand },
  )

  return { haken, antwort, ausfuehren, nachher, fehler }
}

const ada: Gestenstand = {
  profil: 'ada',
  epoche: 1,
  sitze: ['sitz-1', 'sitz-2'],
}

// tastenzug greift mit der Leertaste, geht n Ziele weiter und legt mit ↩ ab.
async function tastenzug(
  haken: ReturnType<typeof aufbauen>['haken'],
  weiter = 0,
) {
  act(() => {
    haken.result.current.greifen({ objekt: 'karte-7', quelle: 'hand', ziele })
  })
  for (let schritt = 0; schritt < weiter; schritt += 1) {
    act(() => {
      haken.result.current.taste('ArrowRight')
    })
  }
  await act(async () => {
    haken.result.current.taste('Enter')
    await Promise.resolve()
  })
}

describe('useGeste', () => {
  it('führt eine Absicht aus und ruft danach zurück', async () => {
    const { haken, antwort, ausfuehren, nachher } = aufbauen(ada)

    await tastenzug(haken)
    expect(ausfuehren).toHaveBeenCalledWith(
      expect.objectContaining({
        objekt: 'karte-7',
        ziel: 'friedhof',
        zielSitz: 'sitz-1',
        profil: 'ada',
        epoche: 1,
      }),
    )
    await act(async () => {
      antwort.fertig()
      await antwort.versprechen
    })

    expect(nachher).toHaveBeenCalledTimes(1)
    expect(haken.result.current.ansage).toBe('Abgelegt: Friedhof.')
  })

  it('verwirft eine verspätete Antwort, wenn das Profil inzwischen gewechselt hat', async () => {
    const { haken, antwort, nachher, fehler } = aufbauen(ada)

    await tastenzug(haken)
    haken.rerender({ profil: 'bert', epoche: 1, sitze: ['sitz-1', 'sitz-2'] })
    const ansageNachWechsel = haken.result.current.ansage
    await act(async () => {
      antwort.fertig()
      await antwort.versprechen
    })

    expect(nachher).not.toHaveBeenCalled()
    expect(fehler).not.toHaveBeenCalled()
    expect(haken.result.current.ansage).toBe(ansageNachWechsel)
  })

  it('verwirft auch einen verspäteten Fehler aus der vorigen Epoche', async () => {
    const { haken, antwort, fehler } = aufbauen(ada)

    await tastenzug(haken)
    haken.rerender({ ...ada, epoche: 2 })
    await act(async () => {
      antwort.gescheitert(new Error('abgewiesen'))
      await antwort.versprechen.catch(() => undefined)
    })

    expect(fehler).not.toHaveBeenCalled()
  })

  it('führt nichts aus, wenn es den Ziel-Sitz nicht mehr gibt', async () => {
    const { haken, ausfuehren } = aufbauen(ada)

    act(() => {
      haken.result.current.greifen({ objekt: 'karte-7', quelle: 'hand', ziele })
    })
    haken.rerender({ ...ada, sitze: ['sitz-1'] })
    // **Der Wechsel der Sitze räumt den Griff** — §7, Regel 1: eine Geste
    // aus dem vorigen Stand legt nichts in den neuen.
    expect(haken.result.current.zustand).toEqual({ art: 'ruhend' })

    await tastenzug(haken, 1)

    expect(ausfuehren).not.toHaveBeenCalled()
    expect(haken.result.current.ansage).toBe(
      'Das Ziel gibt es nicht mehr — nichts wurde bewegt.',
    )
  })

  it('verwirft die Antwort nach einem Abbruch', async () => {
    const { haken, antwort, nachher } = aufbauen(ada)

    await tastenzug(haken)
    act(() => {
      haken.result.current.abbrechen()
    })
    await act(async () => {
      antwort.fertig()
      await antwort.versprechen
    })

    expect(nachher).not.toHaveBeenCalled()
  })

  it('führt eine Absicht aus einem vorigen Stand gar nicht erst aus', async () => {
    const { haken, ausfuehren } = aufbauen(ada)

    await act(async () => {
      haken.result.current.ablegen({
        objekt: 'karte-7',
        quelle: 'hand',
        ziel: 'friedhof',
        zielSitz: 'sitz-1',
        profil: 'ada',
        epoche: 0,
        weg: 'maus',
      })
      await Promise.resolve()
    })

    expect(ausfuehren).not.toHaveBeenCalled()
    expect(haken.result.current.ansage).toBe(
      'Inzwischen hat sich der Stand geändert — nichts wurde bewegt.',
    )
  })

  it('sagt jeden Schritt der Tastatur an', () => {
    const { haken } = aufbauen(ada)

    act(() => {
      haken.result.current.greifen({ objekt: 'karte-7', quelle: 'hand', ziele })
    })
    expect(haken.result.current.ansage).toBe(
      'Gegriffen. Ziel: Friedhof (1 von 2). Pfeiltasten wählen das Ziel, Eingabe legt ab, Escape bricht ab.',
    )
    act(() => {
      expect(haken.result.current.taste('ArrowDown')).toBe(true)
    })
    expect(haken.result.current.ansage).toBe('Ziel: Vom Gegenüber (2 von 2).')
    act(() => {
      haken.result.current.taste('Escape')
    })
    expect(haken.result.current.ansage).toBe(
      'Abgebrochen. Nichts wurde bewegt.',
    )
    act(() => {
      expect(haken.result.current.taste('Escape')).toBe(false)
    })
    act(() => {
      haken.result.current.greifen({
        objekt: 'karte-7',
        quelle: 'hand',
        ziele: [],
      })
    })
    expect(haken.result.current.ansage).toBe(
      'Für dieses Objekt gibt es gerade kein Ziel.',
    )
  })
})

// **Dieselben Gestenwege am Karten- und am Marktobjekt** — die Abnahme von
// GESTE: dieselben Tasten, dieselbe Absicht, dieselben Wächter. Am
// Markttisch gibt es keine Sitze; das Ziel ist ein Fach.
describe('Karten- und Marktobjekt', () => {
  const faelle = [
    {
      name: 'Karte am Tisch',
      stand: ada,
      griff: { objekt: 'karte-7', quelle: 'hand', ziele },
      erwartet: { ziel: 'fach:gegenueber', zielSitz: 'sitz-2' },
    },
    {
      name: 'Exemplar am Markttisch',
      stand: { profil: 'ada', epoche: 1 },
      griff: {
        objekt: 'exemplar-4',
        quelle: 'sammlung',
        ziele: [
          { kennung: 'fach:eigen', name: 'Von dir' },
          { kennung: 'fach:gegenueber', name: 'Vom Gegenüber' },
        ],
      },
      erwartet: { ziel: 'fach:gegenueber' },
    },
  ] as const

  it.each(faelle)(
    '$name: Leertaste, Pfeil, Eingabe legen ab',
    async ({ stand, griff, erwartet }) => {
      const antwort = verzoegert()
      const ausfuehren = vi.fn<(absicht: Gestenabsicht) => Promise<void>>(
        () => antwort.versprechen,
      )
      const haken = renderHook(() => useGeste({ stand, ausfuehren }))

      act(() => {
        haken.result.current.greifen(griff)
      })
      act(() => {
        haken.result.current.taste('ArrowRight')
      })
      await act(async () => {
        haken.result.current.taste('Enter')
        await Promise.resolve()
      })
      await act(async () => {
        antwort.fertig()
        await antwort.versprechen
      })

      expect(ausfuehren).toHaveBeenCalledWith({
        objekt: griff.objekt,
        quelle: griff.quelle,
        profil: 'ada',
        epoche: 1,
        weg: 'tastatur',
        ...erwartet,
      })
      expect(haken.result.current.ansage).toBe('Abgelegt: Vom Gegenüber.')
    },
  )

  it.each(faelle)(
    '$name: ein Fehler im eigenen Stand kommt an',
    async ({ stand, griff }) => {
      const antwort = verzoegert()
      const fehler = vi.fn()
      const haken = renderHook(() =>
        useGeste({ stand, ausfuehren: () => antwort.versprechen, fehler }),
      )

      act(() => {
        haken.result.current.greifen(griff)
      })
      await act(async () => {
        haken.result.current.taste('Enter')
        await Promise.resolve()
      })
      await act(async () => {
        antwort.gescheitert(new Error('abgewiesen'))
        await antwort.versprechen.catch(() => undefined)
      })

      expect(fehler).toHaveBeenCalledWith(
        new Error('abgewiesen'),
        expect.objectContaining({ objekt: griff.objekt }),
      )
    },
  )
})
