// **Ein Sitz ist eine Gruppe** — P01
// (#2165).
//
// Zeichen, Name und Deck standen mit Profil und „Wegnehmen" in einer
// einzigen umbrechenden Reihe; welcher Name zu welchem Deck gehört, sagte
// nur die Lage, und ein Umbruch trennte sie. Jetzt bilden sie je Sitz eine
// benannte Gruppe, und „Wegnehmen" steht außerhalb der Felder an einem
// festen Platz.
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { Sitzliste, type Sitzwunsch } from './Sitzliste'

import type { DeckDTO } from '../apidecks'

const eigene = kataloge.de.seiten.spielen

const decks = [
  { id: 'd1', name: 'Erstes Deck', hauptdeck: 60, sideboard: 15 },
] as unknown as DeckDTO[]

const wuensche: Sitzwunsch[] = [
  { name: 'Anna', deckID: 'd1', profilID: '', leiheID: '' },
  { name: 'Bert', deckID: '', profilID: '', leiheID: '' },
]

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Sitzliste
        wuensche={wuensche}
        setzeWunsch={vi.fn()}
        hinzufuegen={vi.fn()}
        wegnehmen={vi.fn()}
        leihen={[]}
        decks={decks}
        profile={[]}
        auskunft=""
      />
    </Sprachraum>,
  )
}

describe('ein Sitz', () => {
  it('fasst Name und Deck in einer benannten Gruppe', () => {
    zeige()

    for (const nummer of [1, 2]) {
      const gruppe = screen.getByRole('group', {
        name: einsetzen(eigene.sitzGruppe, { nummer: String(nummer) }),
      })

      expect(
        within(gruppe).getByRole('textbox', {
          name: `${eigene.sitzName} ${String(nummer)}`,
        }),
      ).toBeInTheDocument()
      expect(
        within(gruppe).getByRole('combobox', { name: eigene.sitzDeck }),
      ).toBeInTheDocument()
    }
  })

  it('stellt „Wegnehmen" neben die Gruppe und nicht in sie', () => {
    zeige()

    const gruppe = screen.getByRole('group', {
      name: einsetzen(eigene.sitzGruppe, { nummer: '1' }),
    })
    const sitz = gruppe.closest('li')

    expect(sitz).not.toBeNull()
    expect(
      within(gruppe).queryByRole('button', { name: eigene.sitzWeg }),
    ).toBeNull()
    expect(
      within(sitz as HTMLElement).getByRole('button', {
        name: eigene.sitzWeg,
      }),
    ).toBeInTheDocument()
  })
})
