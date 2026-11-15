// **Der Rückweg von einem Exemplar zu seinen Decks** — Werkstück U9 von
// Phase 5.1 (#1420).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Exemplardecks } from './Exemplardecks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

describe('die Decks eines Exemplars', () => {
  it('nennt Deck und Menge', async () => {
    belegeApi({
      ExemplarDecks: () =>
        Promise.resolve({
          decks: [{ deckID: '1', name: 'Aggro', menge: 2 }],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Exemplardecks exemplarID={7} />
      </Sprachraum>,
    )

    const liste = await screen.findByLabelText(texte.inDecks)
    expect(liste).toHaveTextContent('2 × Aggro')
  })

  it('bleibt aus, wenn kein Deck es benutzt', async () => {
    const gefragt = vi.fn().mockResolvedValue({ decks: [], fehler: null })

    belegeApi({ ExemplarDecks: gefragt })

    render(
      <Sprachraum sprache="de">
        <Exemplardecks exemplarID={7} />
      </Sprachraum>,
    )

    // **Auf die Antwort warten, nicht auf den Ruf** — sonst prüfte die
    // Verneinung einen Baum, den es noch nicht gibt (§3).
    await vi.waitFor(() => {
      expect(gefragt).toHaveBeenCalledWith('7')
    })

    expect(screen.queryByLabelText(texte.inDecks)).not.toBeInTheDocument()
  })

  it('trägt eine späte Antwort für ein anderes Exemplar nicht ein', async () => {
    let spaet: (() => void) | undefined

    belegeApi({
      ExemplarDecks: (exemplarID: string) =>
        exemplarID === '7'
          ? new Promise((aufloesen) => {
              spaet = () => {
                aufloesen({
                  decks: [{ deckID: '1', name: 'Alt', menge: 1 }],
                  fehler: null,
                })
              }
            })
          : Promise.resolve({
              decks: [{ deckID: '2', name: 'Neu', menge: 3 }],
              fehler: null,
            }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Exemplardecks exemplarID={7} />
      </Sprachraum>,
    )

    rerender(
      <Sprachraum sprache="de">
        <Exemplardecks exemplarID={8} />
      </Sprachraum>,
    )

    expect(await screen.findByLabelText(texte.inDecks)).toHaveTextContent(
      '3 × Neu',
    )

    spaet?.()

    await vi.waitFor(() => {
      expect(screen.getByLabelText(texte.inDecks)).toHaveTextContent('3 × Neu')
    })

    expect(screen.queryByText('1 × Alt')).not.toBeInTheDocument()
  })
})
