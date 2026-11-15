// **Ein Deck als Umschlag** — Werkstück T7 von Phase 5.6
// (#1472).
//
// # Was hier gemessen wird
//
// **Dass ohne offenes Deck nichts hinausgeht** — ein Umschlag ohne Inhalt ist
// keiner, und die Meldung sagt, was zu tun ist.
//
// **Dass ein abgebrochener Dialog keine Meldung wert ist** (§6) — weder beim
// Schnüren noch beim Annehmen.
//
// **Und dass ein angenommenes Deck die Liste nachlädt.**
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckumschlag } from '../bausteine/Deckumschlag'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { DeckDTO } from '../apidecks'

const eigene = kataloge.de.seiten.decks

// einDeck ist das offene Deck dieser Läufe.
function einDeck(): DeckDTO {
  return {
    id: '7',
    fassung: '1',
    name: 'Wanderdeck',
    spiel: 'magic',
    notiz: '',
    herkunft: { art: '', quelle: '', quellProfil: '', uebernommen: '' },
    formatID: '',
    plaetze: [],
    hauptdeck: 0,
    sideboard: 0,
  }
}

// zeige stellt den Kasten hin.
function zeige(deck: DeckDTO | null, angenommen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Deckumschlag deck={deck} angenommen={angenommen} />
    </Sprachraum>,
  )

  return { hand: userEvent.setup(), angenommen }
}

describe('Ein Deck als Umschlag', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('geht ohne offenes Deck nicht hinaus', async () => {
    const geschnuert = vi.fn()

    belegeApi({ UmschlagSchnueren: geschnuert })

    const { hand } = zeige(null)

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagSchnueren }),
    )

    expect(await screen.findByText(eigene.umschlagOhneDeck)).toBeInTheDocument()

    expect(geschnuert).not.toHaveBeenCalled()
  })

  it('schnürt das offene Deck', async () => {
    const geschnuert = vi.fn(() =>
      Promise.resolve({
        objekte: [],
        urheber: '',
        fingerabdruck: '',
        pfad: '',
        abgebrochen: false,
        fehler: null,
      }),
    )

    belegeApi({ UmschlagSchnueren: geschnuert })

    const { hand } = zeige(einDeck())

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagSchnueren }),
    )

    expect(
      await screen.findByText(eigene.umschlagGeschnuert),
    ).toBeInTheDocument()

    // **Die Art steht vorn** — ohne sie ginge ein Set hinaus.
    expect(geschnuert).toHaveBeenCalledWith('deck', '7')
  })

  it('sagt nichts, wenn jemand den Dialog wegklickt', async () => {
    belegeApi({
      UmschlagSchnueren: () =>
        Promise.resolve({
          objekte: [],
          urheber: '',
          fingerabdruck: '',
          pfad: '',
          abgebrochen: true,
          fehler: null,
        }),
    })

    const { hand } = zeige(einDeck())

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagSchnueren }),
    )

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('nimmt ein Deck an und lädt die Liste nach', async () => {
    const angenommen = vi.fn()

    belegeApi({
      UmschlagDeck: () => Promise.resolve({ deck: einDeck(), fehler: null }),
    })

    const { hand } = zeige(null, angenommen)

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagAnnehmen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent('Wanderdeck')
    expect(angenommen).toHaveBeenCalled()
  })

  // **Angenommen ist angenommen** (§7, Regel 3; #1970): kam das Nachlesen
  // nicht, nennt Go das Deck, wie es angelegt wurde — und die Seite nimmt es
  // als angenommen und lädt die Liste nach.
  it('nimmt ein Deck mit Nachladefehler als angenommen', async () => {
    const angenommen = vi.fn()

    belegeApi({
      UmschlagDeck: () =>
        Promise.resolve({
          deck: einDeck(),
          nachladefehler: {
            code: 'nichtGefunden' as const,
            schluessel: '',
            meldung: 'weg',
          },
          fehler: null,
        }),
    })

    const { hand } = zeige(null, angenommen)

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagAnnehmen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent('Wanderdeck')
    expect(angenommen).toHaveBeenCalled()
  })

  it('sagt nichts, wenn die Annahme weggeklickt wird', async () => {
    const angenommen = vi.fn()

    belegeApi({
      UmschlagDeck: () => Promise.resolve({ deck: null, fehler: null }),
    })

    const { hand } = zeige(null, angenommen)

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagAnnehmen }),
    )

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
    expect(angenommen).not.toHaveBeenCalled()
  })

  it('meldet einen Fehler beim Schnüren', async () => {
    belegeApi({
      UmschlagSchnueren: () =>
        Promise.resolve({
          objekte: [],
          urheber: '',
          fingerabdruck: '',
          pfad: '',
          abgebrochen: false,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'umschlag',
            meldung: 'kein Platz',
          },
        }),
    })

    const { hand } = zeige(einDeck())

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagSchnueren }),
    )

    expect(await screen.findByText('kein Platz')).toBeInTheDocument()
  })

  it('meldet, was schiefgeht', async () => {
    belegeApi({
      UmschlagDeck: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'umschlag',
            meldung: 'so nicht',
          },
        }),
    })

    const { hand } = zeige(einDeck())

    await hand.click(
      screen.getByRole('button', { name: eigene.umschlagAnnehmen }),
    )

    expect(await screen.findByText('so nicht')).toBeInTheDocument()
  })
})
