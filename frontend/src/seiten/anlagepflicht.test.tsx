// **Nach dem Anlegen steht kein Pflichtfeldfehler da** — UX-08.NAT-02 und
// S08.NAT-01 (#1680):
// *Vier getestete Anlageformulare zeigen nach Erfolg keinen
// Pflichtfeldfehler.*
//
// # Der Befund
//
// `Pflichtfeld` merkte sich, dass das Feld einmal verlassen war, und urteilte
// danach über jeden leeren Wert. Wer einen Namen tippte und „Anlegen"
// drückte, verließ das Feld dabei; das Formular leerte es nach dem Erfolg —
// und darunter stand rot, der Name brauche einen Wert.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Turniere } from './Turniere'
import { SetNeu } from './SetNeu'
import { Paketanlage } from './Paketanlage'
import { effekteNachholen } from '../test/effekte'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { turnierprobe } from './turnierprobe'
import { assistent } from '../test/turnierassistent'
import { kataloge } from '../sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

// keinPflichtfehler sagt, dass kein Feld behauptet, es fehle ein Wert.
function keinPflichtfehler() {
  expect(screen.queryByText(/braucht einen Wert/)).toBeNull()
}

describe('nach einem geglückten Anlegen', () => {
  it('im Deck', async () => {
    const DeckAnlegen = vi.fn(() =>
      Promise.resolve({ deck: deckprobe({ name: 'Neu' }), fehler: null }),
    )
    belegeApi({
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      DeckAnlegen,
      Deck: () =>
        Promise.resolve({ deck: deckprobe({ name: 'Neu' }), fehler: null }),
    })
    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await screen.findByText(texte.decks.leer)
    await userEvent.click(screen.getByRole('button', { name: texte.decks.neu }))
    const anlage = screen.getByRole('region', { name: texte.decks.neu })
    await userEvent.type(within(anlage).getByLabelText(texte.decks.name), 'Neu')
    await userEvent.click(
      within(anlage).getByRole('button', { name: texte.decks.anlegen }),
    )

    await waitFor(() => {
      expect(DeckAnlegen).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(within(anlage).getByLabelText(texte.decks.name)).toHaveValue('')
    })
    keinPflichtfehler()
  })

  it('im Turnier', async () => {
    const TurnierAnlegen = vi.fn(() =>
      Promise.resolve({ turnier: turnierprobe({ name: 'Neu' }), fehler: null }),
    )
    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
      Turnier: () =>
        Promise.resolve({
          turnier: turnierprobe({ name: 'Neu' }),
          fehler: null,
        }),
      TurnierAnlegen,
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
    })
    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    // **Der Assistent steht in einer Schublade** — seit UXR-12 (#1920).
    await assistent()
    const name = await screen.findByLabelText(texte.turniere.name)
    await userEvent.type(name, 'Neu')
    await userEvent.click(
      screen.getByRole('button', { name: texte.turniere.weiter }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.turniere.anlegen }),
    )

    await waitFor(() => {
      expect(TurnierAnlegen).toHaveBeenCalled()
    })

    // **Danach meldet der Assistent an** (DOM-04.AK-01, #1699): das Feld
    // dort ist leer und klagt nicht.
    await waitFor(() => {
      expect(screen.getByLabelText(texte.turniere.name)).toHaveValue('')
    })
    keinPflichtfehler()
  })

  it('im neuen Set', async () => {
    const SetAnlegen = vi.fn(() => Promise.resolve({ id: '9', fehler: null }))
    belegeApi({
      Pakete: () =>
        Promise.resolve({
          pakete: [
            {
              art: 'mse-game',
              name: 'probe',
              anzeigename: 'probe',
              version: '',
              pfad: '',
              herkunft: '',
              pruefsumme: '',
              braucht: [],
              gruppe: '',
              rang: 0,
              artrang: 0,
              herkunftsart: 'mitgeliefert' as const,
              hauptdatei: 'style',
              geaendert: false,
              gebrauchtVon: [],
              benutzt: 0,
            },
            {
              art: 'mse-style',
              name: 'probe-schlicht',
              anzeigename: 'probe-schlicht',
              version: '',
              pfad: '',
              herkunft: '',
              pruefsumme: '',
              braucht: [],
              gruppe: '',
              rang: 0,
              artrang: 0,
              herkunftsart: 'mitgeliefert' as const,
              hauptdatei: 'style',
              geaendert: false,
              gebrauchtVon: [],
              benutzt: 0,
            },
          ],
          befunde: [],
          fehler: null,
        }),
      SetAnlegen,
    })
    render(
      <Sprachraum sprache="de">
        <SetNeu angelegt={vi.fn()} zuDenPaketen={vi.fn()} />
      </Sprachraum>,
    )

    // **Der Stil steht in einer Auswahl mit Suche** seit UXR-04 (#1912) und
    // UXR-F1 (#2007), nicht mehr als Optionsfeld — und vorgewählt.
    expect(
      await screen.findByRole('combobox', { name: texte.sets.neuStil }),
    ).toHaveValue('probe-schlicht')
    await userEvent.click(
      screen.getByRole('button', { name: texte.sets.neuWeiter }),
    )
    const name = await screen.findByLabelText(texte.sets.neuName)
    await userEvent.type(name, 'Neu')
    await userEvent.click(
      screen.getByRole('button', { name: texte.sets.neuAnlegen }),
    )

    await waitFor(() => {
      expect(SetAnlegen).toHaveBeenCalled()
    })
    // **Der Assistent beginnt von vorn** (#1677) — das Namensfeld steht erst
    // im letzten Schritt wieder, und dann leer und ohne Urteil.
    await waitFor(() => {
      expect(screen.queryByLabelText(texte.sets.neuName)).toBeNull()
    })
    keinPflichtfehler()
  })

  it('im Paket', async () => {
    const spiel = {
      art: 'mse-game',
      name: 'eigenes',
      anzeigename: 'eigenes',
      version: '',
      pfad: '/p/eigenes.mse-game',
      herkunft: '',
      pruefsumme: '',
      gruppe: '',
      rang: 0,
      braucht: [],
      artrang: 30,
      herkunftsart: 'mitgeliefert' as const,
      hauptdatei: 'style',
      geaendert: false,
      gebrauchtVon: [],
      benutzt: 0,
    }
    const PaketAnlegen = vi.fn(() =>
      Promise.resolve({
        paket: { ...spiel, name: 'neues' },
        abgebrochen: false,
        fehler: null,
      }),
    )
    belegeApi({
      PaketArten: () =>
        Promise.resolve([
          { art: 'mse-game', brauchtSpiel: false },
          { art: 'mse-style', brauchtSpiel: true },
        ]),
      PaketAnlegen,
    })
    render(
      <Sprachraum sprache="de">
        <Paketanlage spiele={[spiel]} angelegt={vi.fn()} />
      </Sprachraum>,
    )
    await effekteNachholen()

    const name = screen.getByLabelText(/Kurzname/)
    await userEvent.type(name, 'neues')
    await userEvent.click(screen.getByRole('button', { name: 'Anlegen' }))

    await waitFor(() => {
      expect(PaketAnlegen).toHaveBeenCalled()
    })
    await waitFor(() => {
      expect(name).toHaveValue('')
    })
    keinPflichtfehler()
  })

  // **Und wer das Feld leer verlässt, liest es weiter** — der Satz ist nicht
  // abgeschafft, er gilt nur für ein Feld, das jemand leer gelassen hat.
  it('und doch, wo jemand das Feld leer verlässt', async () => {
    belegeApi({
      Turniere: () => Promise.resolve({ turniere: [], fehler: null }),
      TurnierRangliste: () => Promise.resolve({ rangliste: [], fehler: null }),
    })
    render(
      <Sprachraum sprache="de">
        <Turniere />
      </Sprachraum>,
    )

    // **Der Assistent steht in einer Schublade** — seit UXR-12 (#1920).
    await assistent()
    const name = await screen.findByLabelText(texte.turniere.name)
    await userEvent.click(name)
    await userEvent.tab()

    expect(screen.getByText(/braucht einen Wert/)).toBeInTheDocument()
  })
})
