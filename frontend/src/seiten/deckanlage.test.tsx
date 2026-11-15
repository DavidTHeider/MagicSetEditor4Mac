// **Was nach dem Anlegen geschieht** — Befund 3 des Zweitblicks über Phase
// 5.1 (#1405).
//
// **Eine eigene Datei**, damit `decks.test.tsx` unter der Zeilengrenze aus §2
// bleibt.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// **Ein neues Deck geht auf** — Befund 3 des Zweitblicks über Phase 5.1
// (#1405).
// Wer eines anlegt, will darin bauen; der Kernweg `deck` rechnet ohne den
// Umweg über „Öffnen".
describe('nach dem Anlegen', () => {
  it('schlägt das neue Deck auf', async () => {
    const angelegt = deckprobe({ name: 'Neu', plaetze: [platzprobe()] })

    const Deck = vi.fn(() => Promise.resolve({ deck: angelegt, fehler: null }))

    zeige({
      DeckAnlegen: () => Promise.resolve({ deck: angelegt, fehler: null }),
      Deck,
    })

    await screen.findByText(texte.leer)

    // **Erst den Kasten öffnen** — die Anlage steht seit U10 in einem.
    await userEvent.click(screen.getByRole('button', { name: texte.neu }))

    const anlage = screen.getByRole('region', { name: texte.neu })
    await userEvent.type(within(anlage).getByLabelText(texte.name), 'Neu')
    await userEvent.click(
      within(anlage).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledWith('1')
    })

    // **Und es steht wirklich offen** — die Plätze sind da, nicht nur der Ruf.
    expect(
      await screen.findByRole('list', { name: texte.plaetze }),
    ).toBeInTheDocument()
  })
})

// **Die Spiele der Anlage kommen aus den Sets** — Befund 6 des Zweitblicks
// über Phase 5.1: hier stand ein Textfeld mit `magic` darin, und wer ein Deck
// für ein anderes Spiel anlegte, tippte den Namen von Hand.
describe('die Spiele in der Deckanlage', () => {
  // mitSets zeichnet die Seite mit diesen Spielen in den Sets.
  function mitSets(spiele: readonly string[]) {
    zeige({
      Sets: () =>
        Promise.resolve({
          sets: spiele.map((spiel, stelle) => ({
            id: String(stelle + 1),
            name: `Satz ${spiel}`,
            spiel,
            stil: 'm15',
            herkunft: '',
            herkunftFehlt: false,
            karten: 0,
          })),
          fehler: null,
        }),
    })
  }

  it('nennt jedes Spiel einmal und stellt magic dazu', async () => {
    mitSets(['vs', 'vs', 'probe'])

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neu }),
    )

    const wahl = await screen.findByLabelText(texte.spiel)

    await waitFor(() => {
      expect(
        within(wahl)
          .getAllByRole('option')
          .map((eines) => eines.textContent),
      ).toEqual(['vs', 'probe', 'magic'])
    })

    // **Vorbelegt ist das erste Set** und nicht magic.
    expect((wahl as HTMLSelectElement).value).toBe('vs')
  })

  it('stellt magic nicht zweimal dazu', async () => {
    mitSets(['magic', 'probe'])

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neu }),
    )

    const wahl = await screen.findByLabelText(texte.spiel)

    await waitFor(() => {
      expect(
        within(wahl)
          .getAllByRole('option')
          .map((eines) => eines.textContent),
      ).toEqual(['magic', 'probe'])
    })
  })
})
