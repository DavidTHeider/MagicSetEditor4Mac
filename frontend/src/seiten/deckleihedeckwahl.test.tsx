// **Das Deck einer Leihe wächst mit dem Bestand** — UXR-F1
// (#2007),
// R7: dieselbe Frage wie das Deck am Sitz. Bis vierzehn Decks bleibt es das
// native Feld (`deckleihe.test.tsx`); ab fünfzehn tippt man.
import { screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { einDeck } from './partieprobe'
import { eigene, leiheApi, zeige } from '../test/leiheprobe'

const decks = (anzahl: number) =>
  Array.from({ length: anzahl }, (_, stelle) =>
    einDeck(
      `d${String(stelle + 1)}`,
      `Deck ${String(stelle + 1).padStart(2, '0')}`,
    ),
  )

it('bleibt unter fünfzehn Decks ein natives Feld', async () => {
  leiheApi({
    Decks: () => Promise.resolve({ decks: decks(3), fehler: null }),
  })

  zeige()

  expect(
    (await screen.findByRole('combobox', { name: eigene.leiheDeck })).tagName,
  ).toBe('SELECT')
})

it('ist ab fünfzehn Decks eine Auswahl mit Suche', async () => {
  leiheApi({
    Decks: () => Promise.resolve({ decks: decks(16), fehler: null }),
  })

  const hand = zeige()

  const wahl = await screen.findByRole('combobox', { name: eigene.leiheDeck })

  expect(wahl.tagName).toBe('INPUT')
  expect(wahl).toHaveValue(eigene.leiheDeckWaehlen)

  await hand.click(wahl)
  await hand.keyboard('deck 16{Enter}')

  expect(wahl).toHaveValue('Deck 16')
})
