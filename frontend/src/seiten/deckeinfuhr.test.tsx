// **Eine Deckliste wird ein Deck** — Werkstück T8 von Phase 5.6
// (#1473).
//
// # Was hier gemessen wird
//
// **Dass die Vorschau nichts legt** und dass der Stand aus ihr beim
// Übernehmen zurückgeht — ohne ihn wäre die Bindung an das Deck, über das
// entschieden wurde, nur ein Satz in einem Kommentar.
//
// **Dass eine offene Zeile nicht mitgeht.** „Keine geratenen Drucke" steht so
// im Issue, und es ist hier eine Messung.
//
// **Dass ein geänderter Text die Vorschau ungültig macht** — sie gehört zu
// dem, was dastand (§7, Regel 1).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Deckeinfuhr } from '../bausteine/Deckeinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { DeckDTO } from '../apidecks'
import type { Deckvorschau } from '../apideckeinfuhr'

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

// eineVorschau ist eine Antwort mit einer sicheren und einer offenen Zeile.
function eineVorschau(): Deckvorschau {
  return {
    deckID: 7,
    stand: 'abdruck-von-vorhin',
    hauptdeck: 4,
    sideboard: 3,
    befunde: ['Zeile 9 ist nicht zu lesen'],
    zeilen: [
      {
        bereich: 'hauptdeck',
        menge: 4,
        name: 'Lightning Bolt',
        setcode: 'm11',
        nummer: '149',
        printing: 'a1b2c3d4-0000-0000-0000-00000000000a',
        treffer: [],
        drucke: 1,
        befund: '',
        uebernehmbar: true,
      },
      {
        bereich: 'sideboard',
        menge: 3,
        name: 'Doppelschlag',
        setcode: 'm11',
        nummer: '50',
        printing: 'b1b2c3d4-0000-0000-0000-00000000000a',
        treffer: [],
        drucke: 1,
        befund: '',
        uebernehmbar: true,
      },
      {
        bereich: 'hauptdeck',
        menge: 2,
        name: 'Counterspell',
        setcode: '',
        nummer: '',
        printing: '',
        treffer: [],
        drucke: 72,
        befund: '72 Drucke passen — welcher ist gemeint?',
        uebernehmbar: false,
      },
    ],
  }
}

// zeige stellt den Kasten hin.
function zeige(deck: DeckDTO, gelegt = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Deckeinfuhr deck={deck} gelegt={gelegt} />
    </Sprachraum>,
  )

  return { hand: userEvent.setup(), gelegt }
}

// setzeListe tippt eine Liste in das Feld.
async function setzeListe(hand: ReturnType<typeof userEvent.setup>) {
  await hand.type(
    screen.getByLabelText(eigene.einfuhrFeld),
    '4 Lightning Bolt (M11) 149',
  )
}

describe('Eine Deckliste wird ein Deck', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sieht eine leere Liste nicht an', async () => {
    const angesehen = vi.fn()

    belegeApi({ DecklisteVorschau: angesehen })

    const { hand } = zeige(einDeck())

    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )

    expect(await screen.findByText(eigene.einfuhrLeer)).toBeInTheDocument()
    expect(angesehen).not.toHaveBeenCalled()
  })

  it('zeigt die Vorschau und legt dabei nichts', async () => {
    const gelegtRuf = vi.fn()

    belegeApi({
      DecklisteVorschau: vi.fn().mockResolvedValue({
        vorschau: eineVorschau(),
        fehler: null,
      }),
      DecklisteUebernehmen: gelegtRuf,
    })

    const { hand } = zeige(einDeck())

    await setzeListe(hand)
    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )

    expect(
      await screen.findByText(/Hauptdeck · 4 · Lightning Bolt/),
    ).toBeInTheDocument()

    // **Die offene Zeile steht mit ihrem Grund da** — wer sie ausließe,
    // ließe jemanden rätseln, wo seine Karte geblieben ist.
    // **Und der Bereich steht übersetzt da** — `sideboard` gehört in den
    // Bestand, nicht auf den Bildschirm (§9).
    expect(screen.getByText(/Sideboard · 3 · Doppelschlag/)).toBeInTheDocument()

    expect(
      screen.getByText(/Counterspell — 72 Drucke passen/),
    ).toBeInTheDocument()

    expect(screen.getByText('Zeile 9 ist nicht zu lesen')).toBeInTheDocument()
    expect(gelegtRuf).not.toHaveBeenCalled()
  })

  it('übernimmt nur die eindeutige Zeile — mit dem Stand aus der Vorschau', async () => {
    const uebernommen = vi.fn().mockResolvedValue({
      bericht: { gelegt: 1, stueck: 4, stand: 'abdruck-danach' },
      fehler: null,
    })

    belegeApi({
      DecklisteVorschau: vi
        .fn()
        .mockResolvedValue({ vorschau: eineVorschau(), fehler: null }),
      DecklisteUebernehmen: uebernommen,
    })

    const { hand, gelegt } = zeige(einDeck())

    await setzeListe(hand)
    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )
    await screen.findByText(/Hauptdeck · 4 · Lightning Bolt/)

    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrUebernehmen }),
    )

    expect(await screen.findByText(/1 — 4/)).toBeInTheDocument()

    // **Beide Bereiche gehen mit, und jeder mit seinem Namen** — wer nur
    // das Hauptdeck schickte, ließe das Sideboard still fallen.
    expect(uebernommen).toHaveBeenCalledWith(7, 'abdruck-von-vorhin', [
      {
        printing: 'a1b2c3d4-0000-0000-0000-00000000000a',
        menge: 4,
        bereich: 'hauptdeck',
      },
      {
        printing: 'b1b2c3d4-0000-0000-0000-00000000000a',
        menge: 3,
        bereich: 'sideboard',
      },
    ])

    expect(gelegt).toHaveBeenCalled()

    // **Die Vorschau ist danach verbraucht.** Ihr Stand gilt nicht mehr, und
    // der Knopf ist wieder gesperrt.
    expect(
      screen.queryByRole('button', { name: eigene.einfuhrUebernehmen }),
    ).not.toBeInTheDocument()
  })

  it('macht die Vorschau ungültig, sobald der Text sich ändert', async () => {
    belegeApi({
      DecklisteVorschau: vi
        .fn()
        .mockResolvedValue({ vorschau: eineVorschau(), fehler: null }),
    })

    const { hand } = zeige(einDeck())

    await setzeListe(hand)
    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )
    await screen.findByText(/Hauptdeck · 4 · Lightning Bolt/)

    await hand.type(screen.getByLabelText(eigene.einfuhrFeld), '\n2 Island')

    expect(
      screen.queryByText(/Hauptdeck · 4 · Lightning Bolt/),
    ).not.toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: eigene.einfuhrUebernehmen }),
    ).not.toBeInTheDocument()
  })

  it('sagt es, wenn keine Zeile eindeutig ist', async () => {
    const nurOffen = eineVorschau()
    nurOffen.zeilen = nurOffen.zeilen.filter((zeile) => !zeile.uebernehmbar)

    belegeApi({
      DecklisteVorschau: vi
        .fn()
        .mockResolvedValue({ vorschau: nurOffen, fehler: null }),
      DecklisteUebernehmen: vi.fn(),
    })

    const { hand } = zeige(einDeck())

    await setzeListe(hand)
    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )
    await screen.findByText(/Counterspell — 72 Drucke passen/)

    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrUebernehmen }),
    )

    expect(
      await screen.findByText(eigene.einfuhrNichtsOffen),
    ).toBeInTheDocument()
  })

  it('zeigt die Klage beider Wege', async () => {
    belegeApi({
      DecklisteVorschau: vi.fn().mockRejectedValue(new Error('kein Netz')),
    })

    const { hand } = zeige(einDeck())

    await setzeListe(hand)
    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('kein Netz')

    // **Und beim Übernehmen ebenso** — mit einer Vorschau, die glückt, und
    // einer Übernahme, die klagt.
    raeumeApi()
    belegeApi({
      DecklisteVorschau: vi
        .fn()
        .mockResolvedValue({ vorschau: eineVorschau(), fehler: null }),
      DecklisteUebernehmen: vi
        .fn()
        .mockRejectedValue(new Error('das Deck hat sich geändert')),
    })

    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrVorschau }),
    )
    await screen.findByText(/Hauptdeck · 4 · Lightning Bolt/)

    await hand.click(
      screen.getByRole('button', { name: eigene.einfuhrUebernehmen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'das Deck hat sich geändert',
    )
  })
})
