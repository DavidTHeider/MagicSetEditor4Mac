// **Mit welchem Deck sich ein Gast setzt** — P9c
// (#2377),
// Artboard `Netzpartie09`. Vorgewählt ist das eigene Deck, mit dem zuletzt
// gearbeitet wurde; „Mit einem Deck des Wirts“ bringt nichts mit.
import { screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einDeck } from './partieprobe'
import { gastApi, waehleWirt, zeigeGast } from './partiegastprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

function mitDecks(beigetreten: ReturnType<typeof vi.fn>) {
  gastApi({
    Decks: () =>
      Promise.resolve({
        decks: [einDeck('1', 'Waldschritt'), einDeck('2', 'Glutpfad')],
        fehler: null,
      }),
    Deckleihen: () => Promise.resolve({ leihen: [], pfad: '', fehler: null }),
    GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
    Weiterarbeiten: () =>
      Promise.resolve({
        geoeffnet: [{ objekt: { art: 'deck', kennung: '2' } }],
        fehler: null,
      }),
    PartieBeitreten: beigetreten,
  } as never)
}

const sitzt = {
  laeuft: true,
  pausiert: false,
  gegenueber: 'bwirt',
  partie: 'p1',
  fehler: null,
}

it('bringt das zuletzt geöffnete Deck mit', async () => {
  const beigetreten = vi.fn(() => Promise.resolve(sitzt))
  mitDecks(beigetreten)
  const hand = zeigeGast()

  await waehleWirt(hand)
  const wahl = await screen.findByRole('combobox', {
    name: eigene.gastMitbringen,
  })
  await waitFor(() => {
    expect(wahl).toHaveValue('2')
  })
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    expect(beigetreten).toHaveBeenCalledWith('bwirt', '', {
      deckID: '2',
      leiheID: '',
    })
  })
})

it('bringt mit einem Deck des Wirts nichts mit', async () => {
  const beigetreten = vi.fn(() => Promise.resolve(sitzt))
  mitDecks(beigetreten)
  const hand = zeigeGast()

  await waehleWirt(hand)
  const wahl = await screen.findByRole('combobox', {
    name: eigene.gastMitbringen,
  })
  await waitFor(() => {
    expect(wahl).toHaveValue('2')
  })
  await hand.selectOptions(wahl, '')
  await hand.click(screen.getByRole('button', { name: eigene.gastBeitreten }))

  await waitFor(() => {
    expect(beigetreten).toHaveBeenCalledWith('bwirt', '', {
      deckID: '',
      leiheID: '',
    })
  })
})
