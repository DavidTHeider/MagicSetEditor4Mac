// **Eine leere Ordnung sagt, was leer ist** — G02
// (#2249).
//
// Lagerorte, Dubletten und Wunschliste antworteten bis dahin alle drei mit
// „Nichts zu zeigen." — und wer die Wunschliste leer fand, erfuhr nicht, wie
// etwas daraufkommt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Sammlungsordnung } from './Sammlungsordnung'

const eigene = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

function zeige() {
  belegeApi({
    SammlungLagerorte: () => Promise.resolve({ lagerorte: [], fehler: null }),
    SammlungDubletten: () => Promise.resolve({ dubletten: [], fehler: null }),
    SammlungWuensche: () => Promise.resolve({ exemplare: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsordnung />
    </Sprachraum>,
  )
}

describe('Leerzustände der Ordnung', () => {
  it('sagt bei der Wunschliste, wie etwas daraufkommt', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.ordnungWuensche }),
    )

    const satz = await screen.findByText(eigene.ordnungLeerWuensche)

    // **Mit beiden Wegen**, wie sie dastehen: am Exemplar und am Deck.
    expect(satz.textContent).toContain(eigene.wunsch)
    expect(satz.textContent).toContain(kataloge.de.seiten.decks.aufWunschliste)
  })

  it('nennt bei den Dubletten die Zahl, die behalten wird', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.ordnungDubletten }),
    )

    expect(
      await screen.findByText(
        eigene.ordnungLeerDubletten.replace('{behalten}', '1'),
      ),
    ).toBeInTheDocument()
  })

  it('sagt bei den Lagerorten, dass es kein Exemplar gibt', async () => {
    zeige()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.ordnungLagerorte }),
    )

    expect(
      await screen.findByText(eigene.ordnungLeerLagerorte),
    ).toBeInTheDocument()
  })

  // **Der Satz gehört zur Antwort und nicht zum Feld** (§7, Regel 2): wer
  // danach eine andere Zahl tippt, hat noch nicht neu gefragt.
  it('bleibt bei der Zahl, mit der gefragt wurde', async () => {
    zeige()

    const nutzer = userEvent.setup()
    const feld = screen.getByRole('spinbutton', {
      name: eigene.ordnungBehalten,
    })

    await nutzer.clear(feld)
    await nutzer.type(feld, '2')
    await nutzer.click(
      screen.getByRole('button', { name: eigene.ordnungDubletten }),
    )

    const gefragt = eigene.ordnungLeerDubletten.replace('{behalten}', '2')

    expect(await screen.findByText(gefragt)).toBeInTheDocument()

    await nutzer.clear(feld)
    await nutzer.type(feld, '3')

    expect(feld).toHaveValue(3)
    expect(screen.getByText(gefragt)).toBeInTheDocument()
  })
})
