// **Aus einem Set ein Deck** — FMT-04
// (#1693).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

const werkstatt = {
  id: '4',
  name: 'Werkstatt',
  spiel: 'magic',
  stil: 'm15',
  herkunft: '',
  karten: 12,
  herkunftFehlt: false,
}

async function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Sets: () =>
      Promise.resolve({
        sets: [
          werkstatt,
          { ...werkstatt, id: '5', name: 'Zweites', karten: 1 },
          { ...werkstatt, id: '6', name: 'Leer', karten: 0 },
        ],
        fehler: null,
      }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(await screen.findByRole('button', { name: texte.neu }))

  return screen.findByRole('region', { name: texte.ausSetTitel })
}

describe('Ein Deck aus einem Set', () => {
  it('legt es mit der gewählten Menge an und schlägt es auf', async () => {
    const DeckAusSet = vi.fn(() =>
      Promise.resolve({
        deck: deckprobe({ id: '9', name: 'Werkstatt' }),
        fehler: null,
      }),
    )
    const Deck = vi.fn((deckID: string) =>
      Promise.resolve({ deck: deckprobe({ id: deckID }), fehler: null }),
    )
    const bereich = await zeige({ DeckAusSet, Deck })

    // **Das Set nennt seine Kartenzahl** — wer eines wählt, sieht, was kommt.
    expect(
      await within(bereich).findByRole('option', {
        name: 'Werkstatt · 12 Karten',
      }),
    ).toBeInTheDocument()

    // **Ein leeres Set steht nicht zur Wahl** — aus ihm entstünde nichts.
    expect(
      within(bereich).queryByRole('option', { name: /^Leer/ }),
    ).not.toBeInTheDocument()

    await userEvent.selectOptions(
      within(bereich).getByLabelText(texte.ausSetSet),
      '5',
    )
    const menge = within(bereich).getByLabelText(texte.ausSetMenge)
    await userEvent.clear(menge)
    await userEvent.type(menge, '2')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.ausSetAnlegen }),
    )

    await waitFor(() => {
      expect(Deck).toHaveBeenCalledWith('9')
    })
    expect(DeckAusSet).toHaveBeenCalledWith('5', '', 2)
  })

  it('sagt es an der Menge, wenn der Dienst sie abweist', async () => {
    const bereich = await zeige({
      DeckAusSet: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'validierung',
            schluessel: '',
            meldung: 'Die Menge ist 0 — mindestens eines muss es sein.',
            feld: 'menge',
          },
        }),
    })

    await within(bereich).findByRole('option', {
      name: 'Werkstatt · 12 Karten',
    })
    const menge = within(bereich).getByLabelText(texte.ausSetMenge)
    await userEvent.clear(menge)
    await userEvent.type(menge, '0')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.ausSetAnlegen }),
    )

    expect(
      await within(bereich).findByText(/mindestens eines/),
    ).toBeInTheDocument()
    expect(menge).toHaveAttribute('aria-invalid', 'true')
  })

  // **F03** (#2245):
  // „2.5" ging als Kommazahl an ein `int` in Go, und die Absage kam als
  // Satz über JSON und nicht an der Menge; ein leeres Feld wurde null.
  it.each(['2.5', ''])(
    'weist %j an der Menge ab, ohne zu fragen',
    async (roh) => {
      const DeckAusSet = vi.fn()
      const bereich = await zeige({ DeckAusSet })

      await within(bereich).findByRole('option', {
        name: 'Werkstatt · 12 Karten',
      })
      const menge = within(bereich).getByLabelText(texte.ausSetMenge)
      await userEvent.clear(menge)
      if (roh !== '') {
        await userEvent.type(menge, roh)
      }
      await userEvent.click(
        within(bereich).getByRole('button', { name: texte.ausSetAnlegen }),
      )

      expect(menge).toHaveAttribute('aria-invalid', 'true')
      expect(menge).toHaveAccessibleDescription(
        einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '1' }),
      )
      expect(DeckAusSet).not.toHaveBeenCalled()
    },
  )

  it('erklärt, dass es ohne Set nichts gibt', async () => {
    const bereich = await zeige({
      Sets: () => Promise.resolve({ sets: [], fehler: null }),
    })

    expect(
      await within(bereich).findByText(texte.ausSetKeinSet),
    ).toBeInTheDocument()
    expect(
      within(bereich).queryByRole('button', { name: texte.ausSetAnlegen }),
    ).not.toBeInTheDocument()
  })
})
