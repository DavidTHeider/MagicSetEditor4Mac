// **Das Deck eines Teilnehmers wird gewählt, nicht getippt** — Werkstück I von
// Phase 4.10
// (#1325).
//
// **Dieselbe Klasse wie bei der Deckwerkbank** (§10: den Fall behandeln heißt
// die Klasse behandeln). Dort stand „Eigene Karte (Kennung)", hier „Deck
// (Kennung)" — beide verlangten eine Datenbankkennung, die nirgends in der
// Oberfläche steht.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { deckprobe } from './deckprobe'
import { teilnehmerprobe, turnierprobe } from './turnierprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

function setzeAuf(teile: Partial<Api> = {}) {
  const turnier = turnierprobe()

  belegeApi({
    Turniere: () => Promise.resolve({ turniere: [turnier], fehler: null }),
    Turnier: () => Promise.resolve({ turnier, fehler: null }),
    TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    Decks: () =>
      Promise.resolve({
        decks: [deckprobe({ id: '9', name: 'Weiß-Blau' })],
        fehler: null,
      }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Turniere />
    </Sprachraum>,
  )
}

// anmeldebereich gibt den Kasten, in dem Teilnehmer angemeldet werden.
async function anmeldebereich() {
  return await screen.findByRole('region', { name: texte.teilnehmer })
}

describe('die Deckwahl im Turnier', () => {
  it('meldet mit dem gewählten Deck an', async () => {
    const TurnierAnmelden = vi.fn(() =>
      Promise.resolve({ teilnehmer: teilnehmerprobe(), fehler: null }),
    )

    setzeAuf({ TurnierAnmelden })

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    // **Im Anmeldebereich und nicht auf der Seite**: „Name" steht auch im
    // Kasten, mit dem ein Turnier angelegt wird.
    const bereich = await anmeldebereich()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Anna')

    // **Gewartet wird auf den Eintrag, nicht auf das Feld** (§3).
    await within(bereich).findByRole('option', { name: 'Weiß-Blau' })

    await userEvent.selectOptions(
      within(bereich).getByLabelText(texte.deck),
      '9',
    )

    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anmelden }),
    )

    await waitFor(() => {
      expect(TurnierAnmelden).toHaveBeenCalledWith('7', 'Anna', '9')
    })
  })

  // **Ohne Deck geht es auch** — eine Anmeldung ohne eingereichte Liste ist
  // der Normalfall bei einer Freitagsrunde.
  it('meldet ohne Deck an', async () => {
    const TurnierAnmelden = vi.fn(() =>
      Promise.resolve({ teilnehmer: teilnehmerprobe(), fehler: null }),
    )

    setzeAuf({ TurnierAnmelden })

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    const bereich = await anmeldebereich()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Bea')

    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anmelden }),
    )

    await waitFor(() => {
      expect(TurnierAnmelden).toHaveBeenCalledWith('7', 'Bea', '')
    })
  })
})
