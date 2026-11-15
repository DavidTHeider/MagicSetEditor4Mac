// **Das passende Bedienelement in der Kartenwahl** — UXR-11
// (#1919),
// R7 in `docs/UIREVIEW.md`: Segmente bis vier Werte, eine Auswahl mit Suche
// ab fünfzehn.
//
// Die Karte eines Sets stand in einem Auswahlfeld mit jeder Karte des Sets —
// bei einem gewöhnlichen Set 250 Einträge ohne Suche. Der Bereich war ein
// Auswahlfeld mit zwei Werten. Jetzt grenzt eine Suche die Karten ein, sobald
// es fünfzehn sind, und der Bereich sind zwei Optionsfelder.
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

// karten gibt ein Set mit so vielen Karten, wie es heißt.
function karten(anzahl: number) {
  return Array.from({ length: anzahl }, (_, stelle) => ({
    id: String(100 + stelle),
    nummer: stelle + 1,
    name: `Glutkarte ${String(stelle + 1)}`,
    felder: 4,
  }))
}

async function zeige(anzahl: number, teile: Partial<Api> = {}) {
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
            name: 'Aschenland',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '',
            herkunftFehlt: false,
            karten: anzahl,
          },
        ],
        fehler: null,
      }),
    Kartenzeilen: () =>
      Promise.resolve({ zeilen: karten(anzahl), fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
  )
  await screen.findByRole('option', { name: 'Aschenland' })
  await userEvent.selectOptions(screen.getByLabelText(texte.setWaehlen), '7')
  await screen.findByRole('option', { name: 'Glutkarte 1' })
}

// angeboten nennt die Karten, die das Auswahlfeld gerade anbietet.
function angeboten(): string[] {
  return within(screen.getByLabelText(texte.karteWaehlen))
    .getAllByRole('option')
    .map((option) => option.textContent)
    .slice(1)
}

describe('die Wahl der Karte', () => {
  it('bietet bei 250 Karten eine Suche an, die die Wahl eingrenzt', async () => {
    await zeige(250)

    expect(angeboten()).toHaveLength(250)

    await userEvent.type(screen.getByLabelText(texte.karteSuchen), 'karte 12')

    expect(angeboten()).toEqual([
      'Glutkarte 12',
      'Glutkarte 120',
      'Glutkarte 121',
      'Glutkarte 122',
      'Glutkarte 123',
      'Glutkarte 124',
      'Glutkarte 125',
      'Glutkarte 126',
      'Glutkarte 127',
      'Glutkarte 128',
      'Glutkarte 129',
    ])
  })

  it('behält die gewählte Karte, auch wenn die Suche sie nicht mehr trifft', async () => {
    await zeige(250)

    await userEvent.selectOptions(
      screen.getByLabelText(texte.karteWaehlen),
      '100',
    )
    await userEvent.type(screen.getByLabelText(texte.karteSuchen), '250')

    expect(angeboten()).toEqual(['Glutkarte 1', 'Glutkarte 250'])
    expect(screen.getByLabelText(texte.karteWaehlen)).toHaveValue('100')
  })

  // **Gefiltert leer ist nicht leer** (§6): das Set hat Karten, nur passt
  // keine — und der Satz sagt, wonach gesucht wurde.
  it('sagt es, wenn keine Karte zur Suche passt', async () => {
    await zeige(250)

    await userEvent.type(screen.getByLabelText(texte.karteSuchen), 'Drache')

    expect(angeboten()).toEqual([])
    expect(
      screen.getByText(einsetzen(texte.karteKeine, { suche: 'Drache' })),
    ).toBeInTheDocument()
  })

  it('braucht bei vierzehn Karten keine Suche', async () => {
    await zeige(14)

    expect(screen.queryByLabelText(texte.karteSuchen)).toBeNull()
  })
})

describe('der Bereich', () => {
  // **Zwei Optionsfelder und kein Knopfpaar** — ein gedrückter Knopf trägt
  // hier das Gewicht des Hauptknopfs, und neben „Ins Deck legen" stünden
  // zwei gefüllte Flächen in einer Region (R1). Optionsfelder sind das
  // Segment, das keinem Hauptknopf gleicht.
  it('sind zwei Optionsfelder, von denen eines gewählt ist', async () => {
    await zeige(2)

    const bereich = screen.getByRole('group', { name: texte.bereich })

    expect(
      within(bereich).getByRole('radio', { name: texte.hauptdeck }),
    ).toBeChecked()
    expect(
      within(bereich).getByRole('radio', { name: texte.sideboard }),
    ).not.toBeChecked()
    expect(screen.getByRole('button', { name: texte.legen })).toHaveClass(
      'bg-marke',
    )
    expect(
      within(screen.getByRole('region', { name: 'Aggro' }))
        .getAllByRole('button')
        .filter((knopf) => knopf.classList.contains('bg-marke')),
    ).toHaveLength(1)
  })

  it('legt ins Sideboard, wenn es gewählt ist', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: null, fehler: null }),
    )

    await zeige(2, { DeckplatzLegen })

    await userEvent.selectOptions(
      screen.getByLabelText(texte.karteWaehlen),
      '101',
    )
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
        '101',
        '',
        4,
        'sideboard',
      )
    })
  })
})
