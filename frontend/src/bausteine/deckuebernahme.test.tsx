// **Ein Deck aus einem fremden Profil holen, im Fenster** — Werkstück 12 von
// Phase 4.12
// (#1351).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from '../seiten/Decks'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe } from '../seiten/deckprobe'
import { kataloge } from '../sprache/kataloge'
import { einNutzer } from '../test/anmeldebelegung'
import { probeprofil } from '../test/profil'

import type { Api, Deckuebernahmeantwort } from '../api'
import { kurzform } from '../sprache/kurzform'

const eigene = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

// zeige zeichnet die Deckseite über den gegebenen Belegen.
//
// **Über die ganze Seite und nicht über das Bauteil allein**: es hängt am
// `Deckstand`, und ein selbst gebauter Stand prüfte die eigene
// Zusammensetzung statt des Wegs (§3).
function zeige(teile: Partial<Api> = {}) {
  const decks = vi.fn(() =>
    Promise.resolve({ decks: [deckprobe()], fehler: null }),
  )

  belegeApi({
    Decks: decks,
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    // **Die Wahl fragt nach Menschen, dem eigenen Profil und den Decks von
    // drüben** (UX-08.AK-04, #1680) — Bob mit Profil 2 und dem Deck 7.
    Nutzer: () =>
      Promise.resolve({
        nutzer: [
          einNutzer(),
          {
            ...einNutzer(),
            id: '2',
            name: 'Bob',
            profile: [probeprofil({ id: '2', nutzerID: '2', name: 'Fremd' })],
          },
        ],
        fehler: null,
      }),
    GezeigtesProfil: () =>
      Promise.resolve({ profil: probeprofil(), fehler: null }),
    DeckeZurUebernahme: () =>
      Promise.resolve({
        decks: [deckprobe({ id: '7', name: 'Nachtlauf' })],
        fehler: null,
      }),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Decks wechseln={vi.fn()} />
    </Sprachraum>,
  )

  return decks
}

// aufklappen öffnet den Kasten, in dem die Übernahme seit U10 steht — sie ist
// ein Werkzeug im Navigator und keine Navigation (siehe [Deckbaunavigator]).
//
// **Seit AP-6 (#2363) in zwei Schritten**: die seltenen Wege stehen in einem
// Kasten „Holen und weitergeben“, die Übernahme darin in ihrem eigenen.
async function aufklappen() {
  await userEvent.click(
    await screen.findByRole('button', { name: eigene.wegeTitel }),
  )
  await userEvent.click(
    await screen.findByRole('button', { name: eigene.uebernahme }),
  )
}

// hole wählt Bobs Profil und sein Deck und drückt den Knopf — **ohne eine
// Kennung zu tippen** (UX-08.AK-04, #1680).
async function hole() {
  await aufklappen()

  await userEvent.selectOptions(
    await screen.findByLabelText(eigene.uebernahmeProfilwahl),
    await screen.findByRole('option', { name: 'Bob · Fremd' }),
  )
  await userEvent.selectOptions(
    await screen.findByLabelText(eigene.uebernahmeDeckwahl),
    await screen.findByRole('option', { name: 'Nachtlauf' }),
  )
  await userEvent.click(
    screen.getByRole('button', { name: eigene.uebernahmeHolen }),
  )
}

describe('Die Deckübernahme', () => {
  it('sagt vorher, was nicht mitkommt', async () => {
    zeige()
    await aufklappen()

    expect(
      await screen.findByText(kurzform(eigene.uebernahmeErklaerung)),
    ).toBeInTheDocument()
  })

  it('holt ein Deck und liest die Liste danach neu', async () => {
    const holen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe(), befunde: [], fehler: null }),
    )

    const decks = zeige({ DeckUebernehmen: holen })

    await hole()
    await antwortAbwarten()

    expect(holen).toHaveBeenCalledWith('2', '7')

    // **Regel 3: nach dem geglückten Holen wird nachgeladen.**
    expect(decks).toHaveBeenCalledTimes(2)
  })

  it('zeigt, was nicht mitkam — und lässt es stehen', async () => {
    zeige({
      DeckUebernehmen: vi.fn(() =>
        Promise.resolve({
          deck: deckprobe(),
          befunde: ['Zwölf Karten blieben im anderen Profil.'],
          fehler: null,
        }),
      ),
    })

    await hole()

    expect(
      await screen.findByText('Zwölf Karten blieben im anderen Profil.'),
    ).toBeInTheDocument()
  })

  // **Was geschrieben ist, steht getrennt von dem, was nicht mitkam** —
  // S08.NAT-03 (#1680). Bis hierher stand nur die Liste der Reste da, und ob
  // das Deck überhaupt angekommen ist, sagte erst die Liste darüber.
  it('nennt das geholte Deck getrennt von dem, was nicht mitkam', async () => {
    zeige({
      DeckUebernehmen: vi.fn(() =>
        Promise.resolve({
          deck: deckprobe({ name: 'Leihgabe' }),
          befunde: ['Zwölf Karten blieben im anderen Profil.'],
          fehler: null,
        }),
      ),
    })

    await hole()

    const geschrieben = await screen.findByText(
      /„Leihgabe“ steht jetzt bei deinen Decks/,
    )
    expect(geschrieben).toHaveAttribute('role', 'status')
    expect(geschrieben).not.toHaveTextContent('Zwölf Karten')
    expect(screen.getByText(eigene.uebernahmeNichtMit)).toBeInTheDocument()
  })

  it('meldet, woran das Holen scheitert', async () => {
    zeige({
      // **Der Typ steht dran**, weil `vi.fn` die Rückgabe sonst zu
      // `code: string` verbreitert und `FehlerDTO` eine Aufzählung verlangt.
      DeckUebernehmen: vi.fn((): Promise<Deckuebernahmeantwort> =>
        Promise.resolve({
          deck: null,
          befunde: [],
          fehler: { code: 'erlaubnis', schluessel: 'erlaubnis', meldung: '' },
        }),
      ),
    })

    await hole()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.erlaubnis,
    )
  })

  it('hält den Knopf an, solange ein Feld leer ist', async () => {
    zeige()
    await aufklappen()

    const knopf = await screen.findByRole('button', {
      name: eigene.uebernahmeHolen,
    })

    expect(knopf).toBeDisabled()

    // **Die Kennungen bleiben nennbar** — für Fachleute, unter einem Kasten.
    await userEvent.click(
      screen.getByRole('button', { name: eigene.uebernahmeFachleute }),
    )
    await userEvent.type(screen.getByLabelText(eigene.uebernahmeProfil), '2')

    // **Ein Feld allein reicht nicht** — ohne diesen Schritt bewiese der Lauf
    // nur, dass der Knopf immer angehalten ist (§10).
    expect(knopf).toBeDisabled()

    await userEvent.type(screen.getByLabelText(eigene.uebernahmeDeck), '7')

    expect(knopf).toBeEnabled()
  })
})

// **Profil und Deck zur Wahl** — UX-08.AK-04 (#1680).
describe('Die Wahl der Übernahme', () => {
  it('nennt nur fremde Profile und fragt die Decks erst dann', async () => {
    const DeckeZurUebernahme = vi.fn(() =>
      Promise.resolve({
        decks: [deckprobe({ id: '7', name: 'Nachtlauf' })],
        fehler: null,
      }),
    )
    zeige({ DeckeZurUebernahme })
    await aufklappen()

    const profilwahl = await screen.findByLabelText(eigene.uebernahmeProfilwahl)

    // **Das eigene Profil steht nicht da** — aus ihm wird nicht übernommen.
    expect(
      screen.queryByRole('option', { name: 'Ich · Mein Profil' }),
    ).toBeNull()
    expect(DeckeZurUebernahme).not.toHaveBeenCalled()

    await userEvent.selectOptions(
      profilwahl,
      screen.getByRole('option', { name: 'Bob · Fremd' }),
    )

    expect(
      await screen.findByRole('option', { name: 'Nachtlauf' }),
    ).toBeInTheDocument()
    expect(DeckeZurUebernahme).toHaveBeenCalledWith('2')
  })

  // **Ein Nein steht da, wo sonst die Decks stünden** — und niemand erfährt,
  // welche es drüben gibt.
  it('zeigt das Nein, wenn die Erlaubnis fehlt', async () => {
    zeige({
      DeckeZurUebernahme: () =>
        Promise.resolve({
          decks: [],
          fehler: { code: 'erlaubnis', schluessel: 'erlaubnis', meldung: '' },
        }),
    })
    await aufklappen()

    await userEvent.selectOptions(
      await screen.findByLabelText(eigene.uebernahmeProfilwahl),
      await screen.findByRole('option', { name: 'Bob · Fremd' }),
    )

    expect(
      await screen.findByText(new RegExp(kataloge.de.fehler.erlaubnis)),
    ).toBeInTheDocument()
    expect(screen.queryByLabelText(eigene.uebernahmeDeckwahl)).toBeNull()
  })

  it('sagt, wenn drüben kein Deck liegt', async () => {
    zeige({
      DeckeZurUebernahme: () => Promise.resolve({ decks: [], fehler: null }),
    })
    await aufklappen()

    await userEvent.selectOptions(
      await screen.findByLabelText(eigene.uebernahmeProfilwahl),
      await screen.findByRole('option', { name: 'Bob · Fremd' }),
    )

    expect(
      await screen.findByText(eigene.uebernahmeKeinDeck),
    ).toBeInTheDocument()
  })

  it('sagt, wenn es kein anderes Profil gibt', async () => {
    zeige({
      Nutzer: () => Promise.resolve({ nutzer: [einNutzer()], fehler: null }),
    })
    await aufklappen()

    expect(
      await screen.findByText(eigene.uebernahmeKeinProfil),
    ).toBeInTheDocument()
  })

  it('bietet Wiederholen, wenn die Profile nicht kommen', async () => {
    const Nutzer = vi
      .fn()
      .mockRejectedValueOnce(new Error('Bestand gesperrt'))
      .mockResolvedValue({ nutzer: [einNutzer()], fehler: null })
    zeige({ Nutzer })
    await aufklappen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    )

    expect(
      await screen.findByText(eigene.uebernahmeKeinProfil),
    ).toBeInTheDocument()
  })
})
