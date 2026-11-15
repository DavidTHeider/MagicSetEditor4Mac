// **Die Kachel „Decks“ nennt die Decks mit fehlenden Karten** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`. Gezählt hat die Aufmerksamkeit (`deckFehlend`, über den
// Abgleich mit der Sammlung, ohne Netz); Kachel und Liste lesen **dieselbe
// Antwort** — einmal gefragt, nicht zweimal gerechnet.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Anschluss } from '../anschluss'
import type { Anschlussstelle } from '../anschlussraum'
import type { HinweisDTO } from '../apiaufmerksamkeit'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Start } from './Start'

const texte = kataloge.de.seiten.start

const uebergeben: Anschluss[] = []
vi.mock('../anschlussraum', async (echt) => ({
  ...(await echt<typeof import('../anschlussraum')>()),
  useAnschluss: (): Anschlussstelle => ({
    offen: undefined,
    uebergib: (anschluss) => {
      uebergeben.push(anschluss)
    },
    verbrauche: () => undefined,
  }),
}))

afterEach(() => {
  raeumeApi()
})

function fehlend(deckID: string, name: string, anzahl: number): HinweisDTO {
  return {
    art: 'deckFehlend',
    objekt: { art: 'deck', kennung: deckID, name, angeheftet: false },
    anzahl,
    grund: 'Moosbär',
    ort: '',
    schluessel: '',
    werte: [],
  }
}

function zeige(hinweise: HinweisDTO[]) {
  const Aufmerksamkeit = vi.fn(() =>
    Promise.resolve({
      hinweise,
      durchgesehen: { sets: 0, decks: 3, pakete: 0 },
      befunde: [],
      fehler: null,
    }),
  )
  belegeApi({
    Sets: () => Promise.resolve({ sets: [], fehler: null }),
    Decks: () =>
      Promise.resolve({
        decks: [],
        fehler: null,
      }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [],
        gesamt: 0,
        insgesamt: 0,
        stueck: 0,
        fehler: null,
      }),
    Setprobleme: () => Promise.resolve({ staende: [], fehler: null }),
    Aufmerksamkeit,
  })
  render(
    <Sprachraum sprache="de">
      <Start wechseln={vi.fn()} waehleSet={vi.fn()} bestandsstand={0} />
    </Sprachraum>,
  )

  return { Aufmerksamkeit }
}

const ueberblick = () =>
  within(screen.getByRole('list', { name: texte.ueberblick.titel }))

describe('die Kachel der Decks', () => {
  it('nennt die Decks mit fehlenden Karten und führt zu genau ihnen', async () => {
    const { Aufmerksamkeit } = zeige([
      fehlend('1', 'Waldschritt', 4),
      fehlend('3', 'Moorwacht', 1),
    ])

    await userEvent.click(
      await ueberblick().findByRole('button', {
        name: /Decks 2 mit fehlenden Karten$/,
      }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'deckauswahl',
      name: '2 Decks mit fehlenden Karten',
      deckIDs: ['1', '3'],
    })
    // **Einmal gefragt** — die Liste darunter liest dieselbe Antwort.
    expect(
      await screen.findByRole('button', { name: /Waldschritt/ }),
    ).toBeInTheDocument()
    expect(Aufmerksamkeit).toHaveBeenCalledTimes(1)
  })

  it('öffnet ein einzelnes Deck mit seiner Prüfung', async () => {
    zeige([fehlend('3', 'Moorwacht', 1)])

    await userEvent.click(
      await ueberblick().findByRole('button', {
        name: /Decks 1 mit fehlenden Karten$/,
      }),
    )
    expect(uebergeben.at(-1)).toEqual({
      art: 'deckansicht',
      name: 'Moorwacht',
      deckID: '3',
      lasche: 'pruefung',
    })
  })
})
