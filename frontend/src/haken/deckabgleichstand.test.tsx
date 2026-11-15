// **Der Abgleich liest nach fremden Schreibvorgängen neu** — Zweitblick über
// Phase 6 von 0.9.0, zweiter Teil, Befund 3: er hing nur an Deck und
// Deckfassung; eine Änderung der Sammlung von außen ändert die Fassung des
// Decks nicht, und „Fehlt in der Sammlung“ zeigte einen alten Bestand.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import type { HabenstandAntwort } from '../apidecks'
import { belegeApi, raeumeApi } from '../test/api'
import { deckprobe } from '../seiten/deckprobe'
import { useDeckabgleich } from './useDeckabgleich'

afterEach(() => {
  raeumeApi()
})

it('liest neu, wenn der Bestandsstand wächst', async () => {
  const abgleichen = vi.fn(() =>
    Promise.resolve({ zeilen: [], fehlend: 0, unbekannt: 0, fehler: null }),
  )
  belegeApi({ DeckAbgleichen: abgleichen })
  const deck = deckprobe()

  const { rerender } = renderHook(
    ({ stand }: { stand: number }) => useDeckabgleich(deck, stand),
    { initialProps: { stand: 0 } },
  )
  await waitFor(() => {
    expect(abgleichen).toHaveBeenCalledTimes(1)
  })

  rerender({ stand: 1 })

  await waitFor(() => {
    expect(abgleichen).toHaveBeenCalledTimes(2)
  })
})

// **Eine überholte Antwort übermalt die neue nicht** — Nachlauf zum
// Zweitblick über Phase 6, zweiter Teil, R7: der Test darüber zählte nur
// Aufrufe. Hier antwortet der alte Abgleich zuletzt, einmal mit Bestand und
// einmal mit einem Fehler, und gezeigt bleibt, was zum neuen Stand gehört.
it.each([
  ['einer veralteten Zahl', false],
  ['einem Fehler', true],
])('lässt den überholten Abgleich mit %s fallen', async (_fall, scheitert) => {
  const antwort = (fehlend: number): HabenstandAntwort => ({
    zeilen: [],
    fehlend,
    unbekannt: 0,
    fehler: null,
  })
  let alt: () => void = () => undefined
  const abgleichen = vi
    .fn<(deckID: string) => Promise<HabenstandAntwort>>()
    .mockImplementationOnce(
      () =>
        new Promise((fertig, nein) => {
          alt = () => {
            if (scheitert) {
              nein(new Error('planmäßig: der alte Abgleich scheitert'))
            } else {
              fertig(antwort(4))
            }
          }
        }),
    )
    .mockResolvedValueOnce(antwort(1))
  belegeApi({ DeckAbgleichen: abgleichen })
  const deck = deckprobe()

  const { result, rerender } = renderHook(
    ({ stand }: { stand: number }) => useDeckabgleich(deck, stand),
    { initialProps: { stand: 0 } },
  )
  await waitFor(() => {
    expect(abgleichen).toHaveBeenCalledTimes(1)
  })

  rerender({ stand: 1 })
  await waitFor(() => {
    expect(result.current.daten?.fehlend).toBe(1)
  })

  await act(async () => {
    alt()
    await Promise.resolve()
  })

  expect(result.current.daten?.fehlend).toBe(1)
  expect(result.current.fehler).toBeUndefined()
})
