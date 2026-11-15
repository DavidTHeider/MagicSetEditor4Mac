// **Die Deckseite am Fenster** — Werkstück 7 von Phase 4.6
// (#1293).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import {
  aufschlagen,
  bauartprobe,
  deckprobe,
  formatprobe,
  platzprobe,
} from './deckprobe'

import type { Api } from '../api'
import { deckhandlung, platzeintrag } from '../test/deckhandlung'

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

describe('die Deckseite', () => {
  it('zeigt die Decks samt ihren Zahlen', async () => {
    zeige()

    expect(await screen.findByText('Aggro')).toBeInTheDocument()
    expect(screen.getByText(/0 im Hauptdeck/)).toBeInTheDocument()
  })

  it('sagt es, wenn noch kein Deck da ist', async () => {
    zeige({ Decks: () => Promise.resolve({ decks: [], fehler: null }) })

    expect(await screen.findByText(texte.leer)).toBeInTheDocument()
  })

  it('meldet einen Ladefehler', async () => {
    zeige({
      Decks: () =>
        Promise.resolve({
          decks: [],
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'kaputt',
            feld: '',
            werte: {},
          },
        }),
    })

    expect(
      await screen.findByText(kataloge.de.fehler.unbekannt),
    ).toBeInTheDocument()
  })

  it('legt ein Deck an und lädt danach neu', async () => {
    const DeckAnlegen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe(), fehler: null }),
    )
    const Decks = vi.fn(() => Promise.resolve({ decks: [], fehler: null }))

    zeige({ DeckAnlegen, Decks })

    await screen.findByText(texte.leer)

    // **In der Anlage und nicht irgendwo** — „Name" steht auch bei den
    // Formaten, und ein `getByLabelText` über die ganze Seite fände beide.
    // **Erst den Kasten öffnen** — die Anlage steht seit U10 in einem.
    await userEvent.click(screen.getByRole('button', { name: texte.neu }))

    const anlage = screen.getByRole('region', { name: texte.neu })

    await userEvent.type(within(anlage).getByLabelText(texte.name), 'Aggro')
    await userEvent.click(
      within(anlage).getByRole('button', { name: texte.anlegen }),
    )

    await waitFor(() => {
      expect(DeckAnlegen).toHaveBeenCalledWith('Aggro', 'magic', '')
    })

    // **Nach einem geglückten Schreiben wird nachgeladen** (§7, Regel 3).
    await waitFor(() => {
      expect(Decks).toHaveBeenCalledTimes(2)
    })
  })

  it('meldet einen Fehler beim Anlegen und lädt nicht nach', async () => {
    const Decks = vi.fn(() => Promise.resolve({ decks: [], fehler: null }))

    zeige({
      Decks,
      DeckAnlegen: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'validierung',
            schluessel: 'pflichtfeld',
            meldung: 'ohne Namen',
            feld: 'name',
            werte: {},
          },
        }),
    })

    await screen.findByText(texte.leer)

    // **Erst den Kasten öffnen** — die Anlage steht seit U10 in einem.
    await userEvent.click(screen.getByRole('button', { name: texte.neu }))

    const anlage = screen.getByRole('region', { name: texte.neu })

    await userEvent.click(
      within(anlage).getByRole('button', { name: texte.anlegen }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.pflichtfeld),
    ).toBeInTheDocument()

    // **Nach einem gescheiterten Schreiben wird gar nicht nachgeladen** —
    // sonst überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
    expect(Decks).toHaveBeenCalledTimes(1)
  })

  it('fragt vor dem Löschen nach und lässt sich abbrechen', async () => {
    const DeckLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({
      DeckLoeschen,
      Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    })

    // **Löschen steht am offenen Deck** (UXR-11, #1919).
    await aufschlagen()
    await deckhandlung('Aggro', texte.loeschen)

    expect(screen.getByText(texte.loeschenFrage)).toBeInTheDocument()
    expect(DeckLoeschen).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenNein }),
    )

    expect(screen.queryByText(texte.loeschenFrage)).not.toBeInTheDocument()

    await deckhandlung('Aggro', texte.loeschen)
    await userEvent.click(
      screen.getByRole('button', { name: texte.loeschenJa }),
    )

    await waitFor(() => {
      expect(DeckLoeschen).toHaveBeenCalledWith('1')
    })
  })

  it('schlägt ein Deck auf und zeigt seine Plätze', async () => {
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({ plaetze: [platzprobe()], hauptdeck: 4 }),
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    // **In der Platzliste** — die Menge steht da, die Handlungen hinter
    // „…“ (AP-6, #2363).
    const plaetze = await screen.findByRole('list', { name: texte.plaetze })

    expect(within(plaetze).getByText('4')).toBeInTheDocument()
    expect(
      within(plaetze).getByRole('button', {
        name: new RegExp(`^${texte.plaetze}: `),
      }),
    ).toHaveTextContent(texte.handlungenKurz)
  })

  // **Der Vermerk für die Übersicht ist Bestenfalls-Arbeit** (#1422): lehnt
  // `DeckBenutzt` ab, bleibt das Deck aufgeschlagen, und keine Meldung steht da.
  it('schlägt ein Deck auf, auch wenn der Vermerk scheitert', async () => {
    const DeckBenutzt = vi.fn().mockRejectedValue(new Error('kein Verlauf'))
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({ plaetze: [platzprobe()], hauptdeck: 4 }),
          fehler: null,
        }),
      DeckBenutzt,
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    const plaetze = await screen.findByRole('list', { name: texte.plaetze })
    expect(within(plaetze).getByText('4')).toBeInTheDocument()
    expect(DeckBenutzt).toHaveBeenCalledWith('1')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('meldet einen Fehler beim Aufschlagen', async () => {
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: null,
          fehler: {
            code: 'nichtGefunden',
            schluessel: 'nichtGefunden',
            meldung: 'weg',
            feld: '',
            werte: {},
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    expect(
      await screen.findByText(kataloge.de.fehler.nichtGefunden),
    ).toBeInTheDocument()
  })

  it('zeigt die leere Platzliste', async () => {
    zeige({
      Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    expect(await screen.findByText(texte.plaetzeLeer)).toBeInTheDocument()
  })

  it('zeigt eine eigene Karte über ihre Kennung', async () => {
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({
            plaetze: [
              platzprobe({ karteID: '9', printing: '', bereich: 'sideboard' }),
            ],
          }),
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    const plaetze = await screen.findByRole('list', { name: texte.plaetze })

    expect(within(plaetze).getByText('9')).toBeInTheDocument()
    expect(within(plaetze).getByText(texte.sideboard)).toBeInTheDocument()
  })

  it('legt eine Karte ins Deck und nimmt sie wieder heraus', async () => {
    const DeckplatzLegen = vi.fn(() =>
      Promise.resolve({ platz: platzprobe(), fehler: null }),
    )
    const DeckplatzNehmen = vi.fn(() => Promise.resolve({ fehler: null }))

    zeige({
      DeckplatzLegen,
      DeckplatzNehmen,
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({ plaetze: [platzprobe()] }),
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    await userEvent.type(await screen.findByLabelText(texte.druck), 'abc')
    await userEvent.click(screen.getByRole('button', { name: texte.legen }))

    await waitFor(() => {
      expect(DeckplatzLegen).toHaveBeenCalledWith(
        '1',
        '',
        'abc',
        4,
        'hauptdeck',
      )
    })

    // **Auf den bedienbaren Eintrag, nicht auf den Ruf** (#1809).
    const nehmen = await platzeintrag(
      platzprobe().bezeichnung || platzprobe().printing,
      texte.nehmen,
    )
    await waitFor(() => {
      expect(nehmen).not.toHaveAttribute('aria-disabled')
    })
    await userEvent.click(nehmen)

    await waitFor(() => {
      expect(DeckplatzNehmen).toHaveBeenCalledWith('1', '3')
    })
  })

  it('stellt ein Deck auf ein Format und wieder herunter', async () => {
    const DeckFormatWaehlen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ formatID: '5' }), fehler: null }),
    )

    zeige({
      DeckFormatWaehlen,
      Deck: () => Promise.resolve({ deck: deckprobe(), fehler: null }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    await userEvent.selectOptions(
      await screen.findByLabelText(texte.formatWaehlen),
      '5',
    )

    await waitFor(() => {
      expect(DeckFormatWaehlen).toHaveBeenCalledWith('1', '5')
    })
  })
})

const marke = kataloge.de.bausteine.druck.vorlaeufig

// **Ein Druck ohne Alias ist vorläufig** — Werkstück A2 von Phase 5.3 (#1441).
describe('die Deckseite und die vorläufige Identität', () => {
  it('markiert den Platz und nennt seine Zeile statt eines Alias', async () => {
    zeige({
      Deck: () =>
        Promise.resolve({
          deck: deckprobe({
            plaetze: [
              platzprobe({ printing: '', druckID: '9', vorlaeufig: true }),
            ],
            hauptdeck: 4,
          }),
          fehler: null,
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: `${texte.waehlen}: Aggro` }),
    )

    const plaetze = await screen.findByRole('list', { name: texte.plaetze })

    expect(within(plaetze).getByText(marke)).toBeInTheDocument()
    expect(within(plaetze).getByText('9')).toBeInTheDocument()
  })
})
