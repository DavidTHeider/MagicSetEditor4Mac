// **Die Deckliste zählt nach jedem Schreiben neu** — gefunden von der
// WebKit-Probe für UX-03 (#1675).
//
// # Der Befund
//
// Drei Karten lagen im Deck, und die Liste daneben sagte weiter „40 im
// Hauptdeck". **Jede Handlung am offenen Deck las es in `fuehre` neu** — und
// ein Lesen zählt als neuer Lauf. `fuehre` hielt die Handlung danach für
// überholt und ließ das Nachladen der Liste aus. Die Anlage eines Decks
// kannte die Falle schon und las **danach** statt darin; die Handlungen am
// Deck nicht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type { DeckplatzAntwort } from '../apidecks'
import { platzeintrag } from '../test/deckhandlung'

const texte = kataloge.de.seiten.decks

const platz = platzprobe({ id: '3', bezeichnung: 'Blitz', menge: 4 })

afterEach(() => {
  raeumeApi()
})

// zeige stellt die Deckseite auf; die Liste nennt, was der Bestand gerade
// hält — **nach** dem Schreiben eine Karte mehr.
function zeige(teile: Partial<Api>) {
  let hauptdeck = 4
  const deckliste = vi.fn(() =>
    Promise.resolve({
      decks: [
        deckprobe({ hauptdeck }),
        deckprobe({ id: '2', name: 'Kontrolle' }),
      ],
      fehler: null,
    }),
  )
  const Deck = vi.fn((deckID: string) =>
    Promise.resolve({
      deck:
        deckID === '1'
          ? deckprobe({ plaetze: [platz], hauptdeck })
          : deckprobe({ id: '2', name: 'Kontrolle' }),
      fehler: null,
    }),
  )

  belegeApi({
    Decks: deckliste,
    Deck,
    DeckBenutzt: () => Promise.resolve({ fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  return {
    Deck,
    mehr: () => {
      hauptdeck = 5
    },
  }
}

async function aufschlagen(name: string) {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: ${name}` }),
  )
}

describe('die Deckliste nach dem Schreiben', () => {
  it('zeigt die neue Zahl, wenn eine Karte dazukam', async () => {
    const stand = zeige({
      DeckplatzLegen: vi.fn(() => {
        stand.mehr()

        return Promise.resolve({ platz: { ...platz, menge: 5 }, fehler: null })
      }),
    })
    await aufschlagen('Aggro')
    await screen.findByRole('list', { name: texte.plaetze })
    expect(await screen.findByText(/4 im Hauptdeck/)).toBeInTheDocument()

    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))

    expect(await screen.findByText(/5 im Hauptdeck/)).toBeInTheDocument()
  })

  // **Nachgelesen wird nur, was noch offen ist** (§7, Regel 1 und 2): wer
  // während des Schreibens ein anderes Deck aufschlug, bekommt das vorige
  // nicht zurück.
  it('schlägt das vorige Deck nicht wieder auf, wenn inzwischen ein anderes offen ist', async () => {
    let freigeben: ((antwort: DeckplatzAntwort) => void) | undefined
    const stand = zeige({
      DeckplatzLegen: () =>
        new Promise<DeckplatzAntwort>((los) => {
          freigeben = los
        }),
    })
    await aufschlagen('Aggro')
    await screen.findByRole('list', { name: texte.plaetze })

    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))
    await aufschlagen('Kontrolle')
    await screen.findByRole('heading', { name: 'Kontrolle' })
    const vorher = stand.Deck.mock.calls.length

    freigeben?.({ platz: { ...platz, menge: 5 }, fehler: null })

    await waitFor(() => {
      expect(
        screen.queryByRole('list', { name: texte.plaetze }),
      ).not.toBeInTheDocument()
    })
    expect(
      stand.Deck.mock.calls.slice(vorher).some(([kennung]) => kennung === '1'),
    ).toBe(false)
    expect(screen.getByRole('heading', { name: 'Kontrolle' })).toBeVisible()
  })

  // **Nach einem gescheiterten Schreiben wird nicht nachgelesen** (§7,
  // Regel 3) — sonst überschriebe ein Lesefehler die Meldung, die sagt, was
  // zu tun ist.
  it('liest nach einem gescheiterten Schreiben nicht nach und zeigt den Grund', async () => {
    const stand = zeige({
      DeckplatzLegen: () =>
        Promise.resolve({
          platz: null,
          fehler: {
            code: 'nichtGefunden',
            meldung: 'nicht gefunden',
            feld: '',
            schluessel: 'nichtGefunden',
          },
        }),
    })
    await aufschlagen('Aggro')
    await screen.findByRole('list', { name: texte.plaetze })
    const vorher = stand.Deck.mock.calls.length

    await userEvent.click(await platzeintrag('Blitz', texte.mehrVon))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.nichtGefunden,
    )
    expect(stand.Deck.mock.calls.length).toBe(vorher)
  })
})
