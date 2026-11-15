// **Vorlage, Format und Deck eines Turniers wachsen mit dem Bestand** —
// UXR-F1 (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl ohne Suche, wächst mit dem Bestand
// über 15"*.
//
// Wer fünfzehn Turniere, Formate oder Decks hat, wählt nicht mehr aus einem
// Menü, sondern tippt (`Auswahlfeld`); darunter bleiben es native Felder
// (`turnierassistent.test.tsx`, `turnierdeckwahl.test.tsx`).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { assistent, buehne, pauper, schritt } from '../test/turnierassistent'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { deckprobe } from './deckprobe'
import { Turniere } from './Turniere'
import { teilnehmerprobe, turnierprobe } from './turnierprobe'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

const nummer = (stelle: number) => String(stelle + 1).padStart(2, '0')

describe('das Format eines neuen Turniers', () => {
  it('ist ab fünfzehn Formaten eine Auswahl mit Suche und schreibt seine Kennung', async () => {
    const formate = Array.from({ length: 16 }, (_, stelle) => ({
      ...pauper,
      id: `f${nummer(stelle)}`,
      name: `Format ${nummer(stelle)}`,
    }))
    const doppel = buehne([], {
      Formate: () => Promise.resolve({ formate, fehler: null }),
    })
    const bereich = await assistent()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Freitag')
    await waitFor(() => {
      expect(
        within(bereich).getByRole('combobox', { name: texte.format }).tagName,
      ).toBe('INPUT')
    })

    await userEvent.click(
      within(bereich).getByRole('combobox', { name: texte.format }),
    )
    await userEvent.keyboard('format 16{Enter}')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )

    expect(schritt()).toContain(texte.schrittRegeln)
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(doppel.anlegen).toHaveBeenCalledWith(
        'Freitag',
        'magic',
        '',
        false,
        'f16',
      )
    })
  })
})

describe('die Vorlage eines neuen Turniers', () => {
  it('ist ab fünfzehn Turnieren eine Auswahl mit Suche', async () => {
    const turniere = Array.from({ length: 16 }, (_, stelle) =>
      turnierprobe({
        id: String(stelle + 1),
        name: `Turnier ${nummer(stelle)}`,
        formatID: '4',
      }),
    )
    const doppel = buehne(turniere)
    const bereich = await assistent()

    await waitFor(() => {
      expect(
        within(bereich).getByRole('combobox', { name: texte.vorlage }).tagName,
      ).toBe('INPUT')
    })
    await userEvent.click(
      within(bereich).getByRole('combobox', { name: texte.vorlage }),
    )
    await userEvent.keyboard('turnier 16{Enter}')

    // **Spiel und Format kommen aus der Vorlage** — und ruhen.
    expect(within(bereich).getByLabelText(texte.spiel)).toBeDisabled()
    expect(
      within(bereich).getByRole('combobox', { name: texte.format }),
    ).toBeDisabled()

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Neu')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.weiter }),
    )
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(doppel.kopieren).toHaveBeenCalledWith('16', 'Neu', true)
    })
  })
})

describe('das Deck eines Teilnehmers', () => {
  it('ist ab fünfzehn Decks eine Auswahl mit Suche', async () => {
    const turnier = turnierprobe()
    const TurnierAnmelden = vi.fn(() =>
      Promise.resolve({ teilnehmer: teilnehmerprobe(), fehler: null }),
    )

    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [turnier], fehler: null }),
      Turnier: () => Promise.resolve({ turnier, fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      Decks: () =>
        Promise.resolve({
          decks: Array.from({ length: 16 }, (_, stelle) =>
            deckprobe({
              id: `d${nummer(stelle)}`,
              name: `Deck ${nummer(stelle)}`,
            }),
          ),
          fehler: null,
        }),
      TurnierAnmelden,
    })

    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )

    const bereich = await screen.findByRole('region', {
      name: texte.teilnehmer,
    })

    await userEvent.type(within(bereich).getByLabelText(texte.name), 'Anna')
    await waitFor(() => {
      expect(
        within(bereich).getByRole('combobox', { name: texte.deck }).tagName,
      ).toBe('INPUT')
    })

    await userEvent.click(
      within(bereich).getByRole('combobox', { name: texte.deck }),
    )
    await userEvent.keyboard('deck 16{Enter}')
    await userEvent.click(
      within(bereich).getByRole('button', { name: texte.anmelden }),
    )

    await waitFor(() => {
      expect(TurnierAnmelden).toHaveBeenCalledWith('7', 'Anna', 'd16')
    })
  })
})
