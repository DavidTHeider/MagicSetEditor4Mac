// **Die Felder der Deckseite, die der große Ablauf nicht anfasst** —
// abgeteilt von `decks.test.tsx`, damit beide unter der Zeilengrenze aus §2
// bleiben. Der Schnitt folgt der Frage: dort steht, was die Seite **zeigt**,
// hier, was ihre **Formulare** mitnehmen.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { bauartprobe, deckprobe, formatprobe, platzprobe } from './deckprobe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zeige zeichnet die Seite über den gegebenen Belegen.
function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [deckprobe()], fehler: null }),
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
}

describe('die Formulare der Deckseite', () => {
  it('trägt Spiel und Notiz beim Anlegen mit', async () => {
    const DeckAnlegen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe(), fehler: null }),
    )

    // **Das Spiel wird gewählt** — die Liste kommt aus den Sets des Profils
    // (Befund 6 des Zweitblicks über Phase 5.1). Hier stand ein Textfeld, und
    // der Lauf musste es markieren und neu tippen: drei Handlungen für eine
    // Angabe, die im Bestand schon steht.
    zeige({
      DeckAnlegen,
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Sets: () =>
        Promise.resolve({
          sets: [
            {
              id: '7',
              name: 'Vanguard',
              spiel: 'vs',
              stil: 'm15',
              herkunft: '',
              herkunftFehlt: false,
              karten: 0,
            },
          ],
          fehler: null,
        }),
    })

    await screen.findByText(texte.leer)

    // **Erst den Kasten öffnen** — die Anlage steht seit U10 in einem.
    await userEvent.click(screen.getByRole('button', { name: texte.neu }))

    const anlage = screen.getByRole('region', { name: texte.neu })

    await userEvent.type(within(anlage).getByLabelText(texte.name), 'Küche')

    await screen.findByRole('option', { name: 'vs' })
    await userEvent.selectOptions(
      within(anlage).getByLabelText(texte.spiel),
      'vs',
    )

    await userEvent.type(within(anlage).getByLabelText(texte.notiz), 'Probe')
    await userEvent.click(
      within(anlage).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(DeckAnlegen).toHaveBeenCalledWith('Küche', 'vs', 'Probe')
    })
  })

  // **Ein Druck wird weiter über seine Scryfall-Kennung gelegt** — Werkstück
  // I von Phase 4.10
  // (#1325).
  // Für **eigene** Karten stand hier bis dahin dasselbe Feld, und niemand
  // konnte die Kennung erfahren; die Wahl über den Namen prüft
  // `deckkartenwahl.test.tsx`.
  it('legt einen Druck mit Menge und Bereich', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: platzprobe(), fehler: null }),
    )

    zeige({
      DeckplatzLegen,
      Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
      Sets: () => Promise.resolve({ sets: [], fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    await userEvent.type(await screen.findByLabelText(texte.druck), 'abc')
    await userEvent.clear(screen.getByLabelText(texte.menge))
    await userEvent.type(screen.getByLabelText(texte.menge), '2')
    // **Der Bereich sind zwei Optionsfelder** — seit UXR-11 (#1919), R7:
    // zwei Werte sind kein Auswahlfeld.
    await userEvent.click(
      within(screen.getByRole('group', { name: texte.bereich })).getByRole(
        'radio',
        { name: texte.sideboard },
      ),
    )
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledWith(
        '1',
        '',
        'abc',
        2,
        'sideboard',
      )
    })
  })
})
