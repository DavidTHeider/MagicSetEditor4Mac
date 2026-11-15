// **Der offene Kontext der Deckseite, am Haken gemessen** — seit UXR-11
// (#1919).
//
// Seit Umbenennen, Duplizieren und Löschen am **offenen** Deck stehen, löscht
// die Seite nur noch das Deck, das gerade offen ist. Die Zusagen des Hakens
// gelten aber weiter für jedes: wer ein anderes vergisst, schließt das offene
// nicht (§7, Regel 1, nur der betroffene Datensatz) — und ein Neuladen nach
// einem Konflikt ohne offenes Deck schlägt keines auf. Die Seite betritt beide
// Wege nicht mehr; hier stehen sie, damit sie nicht unbemerkt brechen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { deckformate, deckuebersicht, standarddeck } from '../test/deckbestand'
import { useDecks } from './useDecks'

afterEach(() => {
  raeumeApi()
})

function belege() {
  let deckrufe = 0
  belegeApi({
    Decks: () => Promise.resolve({ decks: deckuebersicht, fehler: null }),
    Formate: () => Promise.resolve({ formate: deckformate, fehler: null }),
    Deck: () => {
      deckrufe++

      return Promise.resolve({ deck: standarddeck, fehler: null })
    },
  })

  return () => deckrufe
}

it('schließt das offene Deck nicht, wenn ein anderes vergessen wird', async () => {
  belege()
  const { result } = renderHook(() => useDecks())
  await waitFor(() => {
    expect(result.current.decks.length).toBeGreaterThan(0)
  })

  await act(async () => {
    await result.current.waehle(standarddeck.id)
  })
  expect(result.current.offen?.id).toBe(standarddeck.id)

  act(() => {
    result.current.vergiss('ein-anderes-deck')
  })

  expect(result.current.offen?.id).toBe(standarddeck.id)
})

it('schlägt nach einem Konflikt kein Deck auf, wenn keines offen war', async () => {
  const deckrufe = belege()
  const { result } = renderHook(() => useDecks())
  await waitFor(() => {
    expect(result.current.decks.length).toBeGreaterThan(0)
  })

  act(() => {
    result.current.nachKonflikt()
  })
  await waitFor(() => {
    expect(result.current.decks.length).toBeGreaterThan(0)
  })

  expect(deckrufe()).toBe(0)
  expect(result.current.offen).toBeNull()
})
