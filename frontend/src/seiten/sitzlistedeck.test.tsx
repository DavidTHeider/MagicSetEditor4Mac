// **Das Deck am Sitz wächst mit dem Bestand** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl ohne Suche ab 15 Decks"*.
//
// Wer fünfzehn Decks oder mehr hat, wählt nicht aus einem Menü, sondern tippt
// (`Auswahlfeld`); darunter bleibt es das native Feld. **Jeder Sitz hat sein
// eigenes Feld**, und eine Wahl am zweiten Sitz ändert nur ihn.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Sitzliste, type Sitzwunsch } from './Sitzliste'

import type { DeckDTO } from '../apidecks'

const eigene = kataloge.de.seiten.spielen

const decks = (anzahl: number) =>
  Array.from({ length: anzahl }, (_, stelle) => ({
    id: `d${String(stelle + 1)}`,
    name: `Deck ${String(stelle + 1).padStart(2, '0')}`,
    hauptdeck: 60,
    sideboard: 15,
  })) as unknown as DeckDTO[]

const wuensche: Sitzwunsch[] = [
  { name: 'Anna', deckID: 'd1', profilID: '', leiheID: '' },
  { name: 'Bert', deckID: '', profilID: '', leiheID: '' },
]

function zeige(anzahl: number) {
  const setzeWunsch = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Sitzliste
        wuensche={wuensche}
        setzeWunsch={setzeWunsch}
        hinzufuegen={vi.fn()}
        wegnehmen={vi.fn()}
        leihen={[]}
        decks={decks(anzahl)}
        profile={[]}
        auskunft=""
      />
    </Sprachraum>,
  )

  return setzeWunsch
}

const deckwahlen = () =>
  screen.getAllByRole('combobox', { name: eigene.sitzDeck })

// deckwahl gibt das Feld des Sitzes mit dieser Stelle — oder scheitert mit
// Nennung, statt `undefined` weiterzureichen.
function deckwahl(stelle: number): HTMLElement {
  const feld = deckwahlen()[stelle]

  if (feld === undefined) {
    throw new Error(`kein Deckfeld am Sitz ${String(stelle)}`)
  }

  return feld
}

describe('Das Deck am Sitz', () => {
  it('bleibt unter fünfzehn Decks ein natives Feld je Sitz', async () => {
    const setzeWunsch = zeige(3)

    expect(deckwahlen().map((feld) => feld.tagName)).toEqual([
      'SELECT',
      'SELECT',
    ])

    await userEvent.selectOptions(deckwahl(1), 'd2')

    expect(setzeWunsch).toHaveBeenCalledExactlyOnceWith(1, {
      name: 'Bert',
      deckID: 'd2',
      profilID: '',
      leiheID: '',
      eingeladen: '',
    })
  })

  it('ist ab fünfzehn Decks je Sitz eine Auswahl mit Suche', () => {
    zeige(16)

    expect(deckwahlen().map((feld) => feld.tagName)).toEqual(['INPUT', 'INPUT'])
    expect(deckwahlen()[0]).toHaveValue('Deck 01')
    expect(deckwahlen()[1]).toHaveValue(eigene.deckWaehlen)
  })

  it('wählt am zweiten Sitz über die Suche und ändert nur ihn', async () => {
    const setzeWunsch = zeige(16)

    await userEvent.click(deckwahl(1))
    await userEvent.keyboard('deck 16{Enter}')

    expect(setzeWunsch).toHaveBeenCalledExactlyOnceWith(1, {
      name: 'Bert',
      deckID: 'd16',
      profilID: '',
      leiheID: '',
      eingeladen: '',
    })
  })

  it('wählt „Deck wählen“ als leeren Wert zurück', async () => {
    const setzeWunsch = zeige(16)

    await userEvent.click(deckwahl(0))
    await userEvent.click(
      screen.getByRole('option', { name: eigene.deckWaehlen }),
    )

    expect(setzeWunsch).toHaveBeenCalledExactlyOnceWith(0, {
      name: 'Anna',
      deckID: '',
      profilID: '',
      leiheID: '',
      eingeladen: '',
    })
  })
})
