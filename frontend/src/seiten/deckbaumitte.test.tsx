// **Gebaut wird in der Mitte, gewählt links** — UXR-11
// (#1919).
//
// Die Kartenwahl und das Einlesen einer Deckliste standen im Navigator,
// **unter** der Deckliste: bei acht Decks begann die Kartenwahl unterhalb der
// Falz, und wer eine Karte legen wollte, rollte an allen Decks vorbei. Das
// Artboard *Decks* setzt die Suche in die Mitte, über die Plätze, die sie
// füllt — und an den Kopf des Navigators **eine** führende Handlung: *Deck
// anlegen*.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { deckhandlung } from '../test/deckhandlung'
import { bauartprobe } from './deckprobe'
import {
  commanderdeck,
  deckformate,
  deckuebersicht,
  standarddeck,
} from '../test/deckbestand'
import { kataloge } from '../sprache/kataloge'
import { verstoesse } from '../test/axe'

import type { Api } from '../api'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

function zeige(teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: deckuebersicht, fehler: null }),
    Formate: () => Promise.resolve({ formate: deckformate, fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Deck: () => Promise.resolve({ deck: standarddeck, fehler: null }),
    DeckPruefen: () =>
      Promise.resolve({ urteil: null, befunde: [], fehler: null }),
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    MarkenJeObjekt: () => Promise.resolve({ marken: {}, fehler: null }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// gefuellt sind die Knöpfe mit dem Hauptgewicht, die ein Mensch sieht —
// dieselbe Klasse, die `uirevue` am Fenster zählt.
function gefuellt(): HTMLElement[] {
  return screen
    .getAllByRole('button')
    .filter((knopf) => /(^|\s)bg-marke(\s|$)/.test(knopf.className))
}

async function aufschlagen(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` }),
  )
  await screen.findByRole('list', { name: texte.plaetze })

  return screen.getByRole('region', { name: 'Aschenrat' })
}

describe('der Deckbauplatz mit acht Decks', () => {
  it('setzt die Kartenwahl in die Mitte, über die Plätze', async () => {
    zeige()
    const mitte = await aufschlagen()
    const links = screen.getByRole('complementary', { name: texte.titel })

    expect(
      within(mitte).getByRole('button', { name: texte.legen }),
    ).toBeInTheDocument()
    expect(
      within(links).queryByRole('button', { name: texte.legen }),
    ).toBeNull()

    const wahl = within(mitte).getByRole('heading', { name: texte.suche })
    const plaetze = within(mitte).getByRole('heading', { name: texte.plaetze })
    expect(
      wahl.compareDocumentPosition(plaetze) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  // **Das Einlesen steht beim Deck, das es füllt** — seit UXR-11, und seit
  // AP-6 (#2363) in seinem Menü „…“ statt als Kasten in der Mitte: es ist ein
  // seltener Weg, und der Kasten zählte zu den Bedienelementen im Bild (R2).
  // Es geht in einer Schublade auf, wie das Einlesen in der Sammlung.
  it('trägt das Einlesen einer Deckliste im Menü des offenen Decks', async () => {
    zeige()
    const mitte = await aufschlagen()

    expect(
      within(mitte).queryByRole('button', { name: texte.einfuhrTitel }),
    ).toBeNull()

    await deckhandlung('Aschenrat', texte.einfuhrMenue)

    const schublade = await screen.findByRole('dialog', {
      name: texte.einfuhrTitel,
    })
    expect(within(schublade).getByRole('textbox')).toBeInTheDocument()

    await userEvent.click(
      within(schublade).getByRole('button', {
        name: kataloge.de.allgemein.arbeitsplatz.schubladen.schliessen,
      }),
    )
    expect(
      screen.queryByRole('dialog', { name: texte.einfuhrTitel }),
    ).toBeNull()
  })

  // **Ein anderes Deck räumt die Schublade mit** (§7, Regel 1) — sonst
  // stünde sie wieder da, sobald das erste Deck später aufgeht.
  it('räumt die Schublade, wenn ein anderes Deck aufgeht', async () => {
    // **Jedes Deck liest sich selbst** — sonst gäbe `zeige` für jede
    // Kennung das Standarddeck, und ein Wechsel wäre keiner.
    zeige({
      Deck: (deckID: string) =>
        Promise.resolve({
          deck: deckID === commanderdeck.id ? commanderdeck : standarddeck,
          fehler: null,
        }),
    })
    await aufschlagen()
    await deckhandlung('Aschenrat', texte.einfuhrMenue)
    await screen.findByRole('dialog', { name: texte.einfuhrTitel })

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Bernsteinflut` }),
    )
    await screen.findByRole('region', { name: 'Bernsteinflut' })
    expect(
      screen.queryByRole('dialog', { name: texte.einfuhrTitel }),
    ).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: `${texte.waehlen}: Aschenrat` }),
    )
    await screen.findByRole('region', { name: 'Aschenrat' })
    expect(
      screen.queryByRole('dialog', { name: texte.einfuhrTitel }),
    ).toBeNull()
  })

  it('lässt „Deck anlegen" den Navigator führen, solange es zu ist', async () => {
    zeige()
    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` })
    const links = screen.getByRole('complementary', { name: texte.titel })
    const anlegen = within(links).getByRole('button', { name: texte.neu })

    expect(anlegen).toHaveClass('bg-marke')
    expect(anlegen).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(anlegen)

    expect(anlegen).toHaveAttribute('aria-expanded', 'true')
    expect(anlegen).not.toHaveClass('bg-marke')
    expect(
      within(links).getByRole('button', { name: texte.anlegen }),
    ).toHaveClass('bg-marke')
  })

  // **Je Blick ein gefüllter Knopf** — AP-3
  // (#2360),
  // Entwurf `Bereiche09`: im offenen Deck führt „Ins Deck“, und „Mit diesem
  // Deck spielen“ steht daneben. „Partie anfangen“ im Kopf führt auf eine
  // andere Seite und gibt sein Gewicht ab; „Deck anlegen“ führt nur, solange
  // kein Deck offen ist.
  it('führt ohne offenes Deck mit „Deck anlegen“ allein', async () => {
    zeige()
    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` })
    const links = screen.getByRole('complementary', { name: texte.titel })

    expect(gefuellt()).toEqual([
      within(links).getByRole('button', { name: texte.neu }),
    ])
  })

  it('führt im offenen Deck mit „Ins Deck legen“ allein', async () => {
    zeige()
    const mitte = await aufschlagen()

    expect(gefuellt()).toEqual([
      within(mitte).getByRole('button', { name: texte.legen }),
    ])
  })

  it('behält, was in der Anlage stand, wenn sie zu- und wieder aufgeht', async () => {
    zeige()
    await screen.findByRole('button', { name: `${texte.waehlen}: Aschenrat` })
    const links = screen.getByRole('complementary', { name: texte.titel })
    const anlegen = within(links).getByRole('button', { name: texte.neu })

    await userEvent.click(anlegen)
    await userEvent.type(
      within(links).getByLabelText(texte.name),
      'Glutsturm II',
    )
    await userEvent.click(anlegen)

    expect(within(links).getByLabelText(texte.name)).not.toBeVisible()

    await userEvent.click(anlegen)

    expect(within(links).getByLabelText(texte.name)).toHaveValue('Glutsturm II')
  })

  // **axe gegen WCAG 2.2 AA mit offenem Deck** — die Seitenprüfung in
  // `barrierefreiheit.test.tsx` sieht die Deckseite ohne Deck; die
  // Fokuszeilen der Plätze, die Optionsfelder des Bereichs und die
  // Handlungen am offenen Deck stehen erst hier.
  it('besteht axe mit einem offenen Deck aus 23 Plätzen', async () => {
    zeige()
    await aufschlagen()

    expect(
      await verstoesse(screen.getByRole('region', { name: texte.deckplatz })),
    ).toEqual([])
  })
})
