// **Karten kommen über ihren Namen ins Deck, nicht über eine Kennung** —
// Werkstück I von Phase 4.10
// (#1325).
//
// # Der Befund
//
// Die Deckwerkbank hatte zwei Textfelder: „Druck (Scryfall-Kennung)" und
// „Eigene Karte (Kennung)". Wer eine eigene Karte ins Deck legen wollte,
// musste ihre Datenbankkennung **tippen** — eine Zahl, die nirgends in der
// Oberfläche steht. Am 2026-09-03 im Fenster nachgesehen: es gibt keinen Weg,
// sie zu erfahren, außer über die Konsole.
//
// **Das ist §6**: die Karte ist das Hauptobjekt, und man wählt sie an ihrem
// Namen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// setzeAuf zeichnet die Deckseite mit einem Set und zwei Karten darin.
function setzeAuf(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
    Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '7',
            name: 'Vanguard',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            herkunftFehlt: false,
            karten: 2,
          },
        ],
        fehler: null,
      }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [
          { id: '11', nummer: 1, name: 'Lightning Bolt', felder: 4 },
          { id: '12', nummer: 2, name: 'Counterspell', felder: 4 },
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
}

// aufschlagen öffnet das eine Deck.
async function aufschlagen() {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )
}

describe('die Kartenwahl im Deck', () => {
  it('legt die gewählte Karte über ihren Namen ins Deck', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    setzeAuf({ DeckplatzLegen })

    await aufschlagen()

    // **Gewartet wird auf die Wirkung, nicht auf das Feld** (§3). Das
    // Auswahlfeld steht sofort da; seine Einträge kommen mit der Antwort, und
    // `selectOptions` auf einer leeren Liste scheitert mit „Value not found".
    await screen.findByRole('option', { name: 'Vanguard' })

    await userEvent.selectOptions(screen.getByLabelText(texte.setWaehlen), '7')

    await screen.findByRole('option', { name: 'Counterspell' })

    await userEvent.selectOptions(
      screen.getByLabelText(texte.karteWaehlen),
      '12',
    )

    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledWith('1', '12', '', 4, 'hauptdeck')
    })
  })

  // **Ohne gewähltes Set gibt es keine Karten zu wählen** — und der Knopf
  // wartet, statt eine leere Kennung zu schicken.
  it('hält das Legen an, solange keine Karte gewählt ist', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    setzeAuf({ DeckplatzLegen })

    await aufschlagen()

    expect(
      await screen.findByRole('button', { name: texte.legen }),
    ).toBeDisabled()
  })
})

// **Die beiden leeren Fälle** — Werkstück I von Phase 4.10
// (#1325).
// §6: ein leerer Zustand erklärt den nächsten Schritt, statt „Keine Daten" zu
// zeigen — oder gar nichts.
describe('die leeren Fälle der Kartenwahl', () => {
  it('sagt es, wenn kein Set offen ist', async () => {
    setzeAuf({ Sets: () => Promise.resolve({ sets: [], fehler: null }) })

    await aufschlagen()

    expect(await screen.findByText(texte.setLeer)).toBeInTheDocument()
  })

  it('sagt es, wenn das gewählte Set keine Karte hat', async () => {
    setzeAuf({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    await aufschlagen()

    await screen.findByRole('option', { name: 'Vanguard' })

    await userEvent.selectOptions(screen.getByLabelText(texte.setWaehlen), '7')

    expect(await screen.findByText(texte.setOhneKarten)).toBeInTheDocument()
  })
})

// **⏎ und Doppelklick legen die Karte** — Befund 3 des Zweitblicks über Phase
// 5.1 (#1405):
// #1421
// verlangt beides, und produktiv gab es nur den Knopf. Der Kernweg `deck`
// rechnet mit ⏎ — ohne ihn bleibt er bei elf Handlungen statt sechs
// (`docs/UX-ARCHITEKTUR.md`).
describe('die Tastenwege der Kartenwahl', () => {
  // gewaehlt schlägt das Deck auf und wählt Set und Karte.
  async function gewaehlt(DeckplatzLegen: Api['DeckplatzLegen']) {
    setzeAuf({ DeckplatzLegen })

    await aufschlagen()
    await screen.findByRole('option', { name: 'Vanguard' })
    await userEvent.selectOptions(screen.getByLabelText(texte.setWaehlen), '7')
    await screen.findByRole('option', { name: 'Counterspell' })
    await userEvent.selectOptions(
      screen.getByLabelText(texte.karteWaehlen),
      '12',
    )
  }

  it('legt mit ⏎ aus der Kartenwahl heraus', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    await gewaehlt(DeckplatzLegen)

    await userEvent.type(screen.getByLabelText(texte.karteWaehlen), '{Enter}')

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledWith('1', '12', '', 4, 'hauptdeck')
    })
  })

  it('legt mit einem Doppelklick auf die Kartenwahl', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    await gewaehlt(DeckplatzLegen)

    await userEvent.dblClick(screen.getByLabelText(texte.karteWaehlen))

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledWith('1', '12', '', 4, 'hauptdeck')
    })
  })

  it('lässt jede andere Taste durch', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    await gewaehlt(DeckplatzLegen)

    // **Ein Pfeil wählt in der Liste** — er legt nichts.
    await userEvent.type(
      screen.getByLabelText(texte.karteWaehlen),
      '{ArrowDown}',
    )

    expect(DeckplatzLegen).not.toHaveBeenCalled()
  })

  it('legt nichts, solange keine Karte gewählt ist', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    setzeAuf({ DeckplatzLegen })

    await aufschlagen()
    await screen.findByRole('option', { name: 'Vanguard' })

    // **Der Beleg, dass jemand zuhört** (§3): dasselbe Feld nimmt eine
    // Eingabe an, und die Wirkung ist zu sehen.
    const setwahl = screen.getByLabelText(texte.setWaehlen)
    await userEvent.selectOptions(setwahl, '7')
    expect(
      await screen.findByRole('option', { name: 'Counterspell' }),
    ).toBeInTheDocument()

    await userEvent.type(screen.getByLabelText(texte.karteWaehlen), '{Enter}')

    expect(DeckplatzLegen).not.toHaveBeenCalled()
  })
})
