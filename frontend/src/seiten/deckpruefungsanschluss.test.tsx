// **Ein Deck geht mit seiner Prüfung auf** — AP-8
// (#2365),
// Befund B-14: die Aufmerksamkeit führt zu einem Deck, weil Karten fehlen,
// und das steht in der Lasche „Prüfung“ — nicht in der Auswertung, die
// sonst zuerst aufgeht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Anschlussbuehne, uebergabeknopf } from '../test/anschlussbuehne'
import { Decks } from './Decks'
import {
  bauartprobe,
  deckprobe,
  formatprobe,
  platzprobe,
  urteilprobe,
} from './deckprobe'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

function zeige(anschluss: Anschluss) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () =>
      Promise.resolve({
        deck: deckprobe({ plaetze: [platzprobe()], hauptdeck: 4 }),
        fehler: null,
      }),
    DeckPruefen: () =>
      Promise.resolve({ urteil: urteilprobe(), befunde: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussbuehne anschluss={anschluss}>
        <Decks wechseln={vi.fn()} />
      </Anschlussbuehne>
    </Sprachraum>,
  )
}

describe('ein Deck aus der Aufmerksamkeit', () => {
  it('geht mit der Lasche „Prüfung“ auf', async () => {
    zeige({
      art: 'deckansicht',
      name: 'Aggro',
      deckID: '1',
      lasche: 'pruefung',
    })
    await userEvent.click(
      await screen.findByRole('button', { name: uebergabeknopf }),
    )

    expect(
      await screen.findByRole('tab', { name: texte.gebietPruefung }),
    ).toHaveAttribute('aria-selected', 'true')
  })

  // **Ein Beleg, dass jemand zuhört** (§3): ohne Lasche geht dasselbe Deck
  // mit der Auswertung auf.
  it('geht ohne Lasche mit der Auswertung auf', async () => {
    zeige({ art: 'deckansicht', name: 'Aggro', deckID: '1' })
    await userEvent.click(
      await screen.findByRole('button', { name: uebergabeknopf }),
    )

    expect(
      await screen.findByRole('tab', { name: texte.gebietAuswertung }),
    ).toHaveAttribute('aria-selected', 'true')
  })
})
