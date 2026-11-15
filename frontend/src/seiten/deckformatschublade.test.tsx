// **Die Formate gehen in einer Schublade auf** — AP-6
// (#2363).
// Sie lagen seit Werkstück I von Phase 4.10 hinter einer Lasche neben den
// Decks; das Laschenpaar zählte zweimal zu den Bedienelementen im Bild (R2),
// und der Entwurf `Bereiche09` hat keines. **Ein Knopf im Kopf und nicht an
// der Formatwahl**: die steht nur im offenen Deck, und ohne Deck kämen
// niemand mehr an die Formate.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe } from './deckprobe'
import { Decks } from './Decks'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

it('öffnet die Formate in einer Schublade, auch ohne Deck, und schließt sie', async () => {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
  })
  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: texte.formateKnopf }),
  )

  const schublade = await screen.findByRole('dialog', {
    name: texte.gebietFormate,
  })
  expect(
    await within(schublade).findByText(texte.formateLeer),
  ).toBeInTheDocument()
  expect(screen.queryByRole('tablist')).toBeNull()

  await userEvent.click(
    within(schublade).getByRole('button', {
      name: kataloge.de.allgemein.arbeitsplatz.schubladen.schliessen,
    }),
  )
  expect(screen.queryByRole('dialog', { name: texte.gebietFormate })).toBeNull()
})
