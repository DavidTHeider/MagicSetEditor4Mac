// Die Prüfungen der drei Haken aus AGENTS.md §7.
//
// **Jeder Fall ist der, der ohne den Haken schiefgeht** — nicht der, in dem er
// nichts tut. Ein Test, der nur den geraden Weg geht, hält einen Haken für
// überflüssig und wird beim nächsten Aufräumen gelöscht.
import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useLaden } from './useLaden'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'

describe('useLauf', () => {
  it('erklärt nur den zuletzt begonnenen Lauf für aktuell', () => {
    const { result } = renderHook(() => useLauf())

    const erster = result.current.naechster()
    const zweiter = result.current.naechster()

    expect(result.current.istAktuell(zweiter)).toBe(true)
    expect(result.current.istAktuell(erster)).toBe(false)
  })

  it('beginnt bei eins', () => {
    const { result } = renderHook(() => useLauf())

    expect(result.current.naechster()).toBe(1)
  })
})

describe('useVorgang', () => {
  // **Der Fall, um den es geht:** zweimal schnell auf dieselbe Schaltfläche.
  it('lässt einen zweiten Aufruf für denselben Datensatz nicht zu', async () => {
    const { result } = renderHook(() => useVorgang())

    let freigeben: (() => void) | undefined
    const langsam = () =>
      new Promise<string>((aufloesen) => {
        freigeben = () => {
          aufloesen('fertig')
        }
      })

    let ersterLauf: Promise<string | undefined> | undefined
    act(() => {
      ersterLauf = result.current.starte(7, langsam)
    })

    expect(result.current.laeuft(7)).toBe(true)

    // Der zweite Aufruf beginnt gar nicht — und sagt das, statt so zu tun.
    const zweiter = await result.current.starte(7, langsam)
    expect(zweiter).toBeUndefined()

    await act(async () => {
      freigeben?.()
      await ersterLauf
    })

    expect(result.current.laeuft(7)).toBe(false)
  })

  it('lässt einen anderen Datensatz gleichzeitig zu', async () => {
    const { result } = renderHook(() => useVorgang())

    const eins = result.current.starte(1, () => Promise.resolve('a'))
    const zwei = result.current.starte(2, () => Promise.resolve('b'))

    await act(async () => {
      expect(await eins).toBe('a')
      expect(await zwei).toBe('b')
    })
  })

  // **Ein gescheiterter Vorgang gibt den Datensatz wieder frei.** Sonst sperrt
  // er ihn für den Rest der Sitzung, und niemand sieht, warum die Schaltfläche
  // nicht mehr reagiert.
  it('gibt den Datensatz auch nach einem Fehler frei', async () => {
    const { result } = renderHook(() => useVorgang())

    await act(async () => {
      await expect(
        result.current.starte(3, () => Promise.reject(new Error('kaputt'))),
      ).rejects.toThrow('kaputt')
    })

    expect(result.current.laeuft(3)).toBe(false)

    await act(async () => {
      expect(
        await result.current.starte(3, () => Promise.resolve('geht wieder')),
      ).toBe('geht wieder')
    })
  })
})

describe('useLaden', () => {
  it('trägt die Antwort ein', async () => {
    const { result } = renderHook(() =>
      useLaden(() => Promise.resolve('da'), []),
    )

    await waitFor(() => {
      expect(result.current.laedt).toBe(false)
    })

    expect(result.current.daten).toBe('da')
    expect(result.current.fehler).toBeUndefined()
  })

  it('meldet einen Fehler mit seiner Ursache', async () => {
    const { result } = renderHook(() =>
      useLaden(() => Promise.reject(new Error('nicht lesbar')), []),
    )

    await waitFor(() => {
      expect(result.current.laedt).toBe(false)
    })

    expect(result.current.fehler).toBe('nicht lesbar')
  })

  it('macht auch aus einer Nicht-Fehler-Ursache eine Meldung', async () => {
    const { result } = renderHook(() =>
      // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
      useLaden(() => Promise.reject('etwas anderes'), []),
    )

    await waitFor(() => {
      expect(result.current.fehler).toBe('etwas anderes')
    })
  })

  // **Der Fall, für den der Haken da ist.**
  //
  // Der erste Ladelauf ist noch unterwegs, als das Bezugsobjekt wechselt. Ohne
  // den Wächter trägt seine Antwort sich in den neuen Stand ein — und dort
  // steht dann die Kartenliste des vorigen Sets, ohne Fehlermeldung und ohne
  // erkennbaren Anlass.
  it('trägt die Antwort eines überholten Laufs nicht ein', async () => {
    let ersteAufloesung: ((wert: string) => void) | undefined

    const holen = vi
      .fn<() => Promise<string>>()
      .mockImplementationOnce(
        () =>
          new Promise<string>((aufloesen) => {
            ersteAufloesung = aufloesen
          }),
      )
      .mockImplementationOnce(() => Promise.resolve('zweites Set'))

    const { result, rerender } = renderHook(
      ({ setID }) => useLaden(holen, [setID]),
      {
        initialProps: { setID: 1 },
      },
    )

    // Wechsel, während der erste Lauf hängt.
    rerender({ setID: 2 })

    await waitFor(() => {
      expect(result.current.daten).toBe('zweites Set')
    })

    // Jetzt kommt der erste nach.
    await act(async () => {
      ersteAufloesung?.('erstes Set')
      await Promise.resolve()
    })

    expect(result.current.daten).toBe('zweites Set')
  })

  // Und ebenso für den Fehlerweg — die unauffälligere Hälfte.
  //
  // Eine Meldung „Set nicht lesbar" zu einem Set, das niemand mehr offen hat,
  // sieht aus wie ein Fehler des aktuellen.
  it('meldet den Fehler eines überholten Laufs nicht', async () => {
    let ersteAblehnung: ((grund: Error) => void) | undefined

    const holen = vi
      .fn<() => Promise<string>>()
      .mockImplementationOnce(
        () =>
          new Promise<string>((_, ablehnen) => {
            ersteAblehnung = ablehnen
          }),
      )
      .mockImplementationOnce(() => Promise.resolve('zweites Set'))

    const { result, rerender } = renderHook(
      ({ setID }) => useLaden(holen, [setID]),
      {
        initialProps: { setID: 1 },
      },
    )

    rerender({ setID: 2 })

    await waitFor(() => {
      expect(result.current.daten).toBe('zweites Set')
    })

    await act(async () => {
      ersteAblehnung?.(new Error('erstes Set nicht lesbar'))
      await Promise.resolve()
    })

    expect(result.current.fehler).toBeUndefined()
  })

  it('lädt auf Wunsch nach', async () => {
    const holen = vi
      .fn<() => Promise<string>>()
      .mockResolvedValueOnce('erst')
      .mockResolvedValueOnce('dann')

    const { result } = renderHook(() => useLaden(holen, []))

    await waitFor(() => {
      expect(result.current.daten).toBe('erst')
    })

    act(() => {
      result.current.nachladen()
    })

    await waitFor(() => {
      expect(result.current.daten).toBe('dann')
    })
  })
})

// **`useLauf` gibt bei jedem Zeichnen denselben Verbund.**
//
// Ohne das läuft jeder Effekt, der ihn in den Abhängigkeiten führt, bei jedem
// Zeichnen erneut — und ein Aufruf nach draußen geht doppelt hinaus. Genau so
// meldete die Anwendung ihre Sprache beim Start zweimal an die Menüleiste.
describe('useLauf bleibt derselbe Verbund', () => {
  it('gibt bei jedem Zeichnen dasselbe Objekt', () => {
    const { result, rerender } = renderHook(() => useLauf())
    const erster = result.current

    rerender()

    expect(result.current).toBe(erster)
  })
})
