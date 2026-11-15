// **Auch das Aufgeschlagene lädt nach, wenn ein anderer schreibt** — Befund 2
// des Nachlaufs zum Gesamtreview über 0.6.0 (#1660).
//
// Seit #2062 laden die Listen von Decks und Turnieren am fremden Stand nach.
// Das offene Deck und das offene Turnier las nur `waehle` — wer an der
// Kommandozeile umbenannte oder ein Ergebnis setzte, sah in der Liste den
// neuen Namen und daneben in der Werkbank den alten.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Decks } from './Decks'
import { Turniere } from './Turniere'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { aufschlagen, bauartprobe, deckprobe, formatprobe } from './deckprobe'
import { rangprobe, teilnehmerprobe, turnierprobe } from './turnierprobe'
import { zumTurniergebiet } from './turnierwerkbank'

afterEach(() => {
  raeumeApi()
})

// **Stabile Rückrufe** — ein neues `vi.fn()` je Zeichnen löste über einen
// Effekt selbst ein Nachladen aus (`fremdstandseiten.test.tsx`).
const wechseln = vi.fn()

describe('Ein fremder Stand', () => {
  it('liest das offene Deck nach, ohne es als benutzt zu vermerken', async () => {
    let name = 'Aggro'
    const DeckBenutzt = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      Decks: () =>
        Promise.resolve({ decks: [deckprobe({ name })], fehler: null }),
      Formate: () =>
        Promise.resolve({ formate: [formatprobe()], fehler: null }),
      Bauarten: () =>
        Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
      Deck: () => Promise.resolve({ deck: deckprobe({ name }), fehler: null }),
      DeckBenutzt,
    })

    const ansicht = (stand: number) => (
      <Sprachraum sprache="de">
        <Decks wechseln={wechseln} bestandsstand={stand} />
      </Sprachraum>
    )
    const { rerender } = render(ansicht(0))

    await aufschlagen()

    name = 'Kontrolle'
    rerender(ansicht(1))

    // **An der Werkbank gemessen und nicht an der Liste.** Der erste Anlauf
    // fragte nach „Umbenennen: Kontrolle" — und war grün, ohne dass das
    // offene Deck nachlas: der Knopf steht in der Deckliste und nimmt den
    // Namen aus ihr. Die Werkbank trägt den Namen des **offenen** Decks.
    await screen.findByRole('region', { name: 'Kontrolle' })

    // **Ein Nachlesen ist kein Benutzen** — die Übersicht auf dem Start
    // sortiert danach (#1422).
    expect(DeckBenutzt).toHaveBeenCalledTimes(1)
  })

  it('liest das offene Turnier nach', async () => {
    let teilnehmer = [teilnehmerprobe()]

    belegeApi({
      Turniere: () =>
        Promise.resolve({
          turniere: [turnierprobe({ teilnehmer })],
          fehler: null,
        }),
      Turnier: () =>
        Promise.resolve({
          turnier: turnierprobe({ teilnehmer }),
          fehler: null,
        }),
      TurnierRangliste: () =>
        Promise.resolve({ rangliste: [rangprobe()], fehler: null }),
    })

    const ansicht = (stand: number) => (
      <Sprachraum sprache="de">
        <Turniere bestandsstand={stand} />
      </Sprachraum>
    )
    const { rerender } = render(ansicht(0))

    const texte = kataloge.de.seiten.turniere
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.waehlen}: Küchentisch`,
      }),
    )
    await zumTurniergebiet('teilnehmer')
    const werkbank = await screen.findByRole('region', { name: 'Küchentisch' })
    within(werkbank).getByText('Ada')

    teilnehmer = [
      teilnehmerprobe(),
      teilnehmerprobe({ id: '4', name: 'Bert', stelle: 2 }),
    ]
    rerender(ansicht(1))

    expect(await within(werkbank).findByText('Bert')).toBeInTheDocument()
  })
})
