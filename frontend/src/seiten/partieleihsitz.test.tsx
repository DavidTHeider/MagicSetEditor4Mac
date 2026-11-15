// **Eine angenommene Leihe sitzt am Tisch** — P1b
// (#2370),
// Befund C-12: der Kern nimmt seit G6 (#1463) eine Leihe statt eines Decks
// an (`LeiheID` im Sitzwunsch), aber das Formular schickte immer eine leere.
// **Angeboten wird, was sich spielen lässt** — geliehen, nicht widerrufen,
// nicht eingelöst, mit dem Recht `spielen`, an keine andere Art gebunden.
// Ob sie abgelaufen ist oder fremd, prüft der Kern; das Formular deutet
// keine Leihe (keine neue Leihfachlogik).
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import type { Leihe } from '../apileihe'
import type { Partieentwurf } from '../apipartie'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { eineAntwort, element } from './partieprobe'
import { mitDecks, zeige } from './spielenprobe'
import {
  oeffneAnnahme,
  umschlagGewaehlt,
  waehleUmschlag,
} from '../test/leiheprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

const leihe = (teil: Partial<Leihe>): Leihe => ({
  kennung: 'l1',
  verleiher: '12D3KooWBerta',
  entleiher: '12D3KooWIch',
  abdruck: 'ab',
  bindungsart: 'partie',
  woran: '',
  rechte: ['spielen'],
  ausgestellt: '2026-10-17',
  bis: '2026-10-24',
  modus: 'zuende',
  verliehen: false,
  verbraucht: '',
  widerrufen: false,
  karten: 60,
  ...teil,
})

const leihen = [
  leihe({}),
  leihe({ kennung: 'l2', verliehen: true }),
  leihe({ kennung: 'l3', widerrufen: true }),
  leihe({ kennung: 'l4', verbraucht: 'p7' }),
  leihe({ kennung: 'l5', rechte: ['einreichen'] }),
  leihe({ kennung: 'l6', bindungsart: 'turnier' }),
]

function mitLeihen() {
  const angelegt = vi.fn<(entwurf: Partieentwurf) => Promise<unknown>>(() =>
    Promise.resolve(eineAntwort()),
  )
  mitDecks({
    PartieAnlegen: angelegt,
    Deckleihen: () => Promise.resolve({ leihen, fehler: null }),
  })

  return angelegt
}

const geliehen = einsetzen(eigene.sitzLeihe, {
  von: '12D3KooWBerta'.slice(-8),
  karten: '60',
})

it('bietet am Sitz nur die spielbare Leihe an und schickt ihre Kennung', async () => {
  const angelegt = mitLeihen()
  const hand = zeige()

  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Ada')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bo')
  const decks = screen.getAllByLabelText(eigene.sitzDeck)
  await hand.selectOptions(element(decks[0]), '1')

  const zweiter = element(decks[1])
  await waitFor(() => {
    expect(
      [...zweiter.querySelectorAll('option')].filter(
        (option) => option.textContent === geliehen,
      ),
    ).toHaveLength(1)
  })
  await hand.selectOptions(zweiter, 'leihe:l1')
  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt.mock.lastCall?.[0].sitze[1]).toMatchObject({
      leiheID: 'l1',
      deckID: '',
    })
  })
  expect(angelegt.mock.lastCall?.[0].sitze[0]).toMatchObject({
    leiheID: '',
    deckID: '1',
  })
})

it('sagt, dass ein eigenes Deck fehlt, wenn jeder Sitz leiht', async () => {
  mitLeihen()
  const hand = zeige()

  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Ada')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bo')
  const decks = screen.getAllByLabelText(eigene.sitzDeck)
  await waitFor(() => {
    expect(
      element(decks[0]).querySelector('option[value="leihe:l1"]'),
    ).not.toBeNull()
  })
  await hand.selectOptions(element(decks[0]), 'leihe:l1')
  await hand.selectOptions(element(decks[1]), 'leihe:l1')

  expect(await screen.findByText(eigene.fehltEigenesDeck)).toBeInTheDocument()
  expect(screen.getByRole('button', { name: eigene.anfangen })).toBeDisabled()
})

// **Angenommen, und gleich am Sitz** — gefunden am nativen Fenster
// (`p1b09`): die Deckwahl las die Leihen einmal beim Aufbau, und die
// Annahme in der Schublade kam danach. Die Seite zählt jetzt jedes
// geglückte Schreiben an einer Leihe, und die Deckwahl liest neu.
it('bietet eine eben angenommene Leihe sofort am Sitz an', async () => {
  const leihenAntworten = vi
    .fn()
    .mockResolvedValueOnce({ leihen: [], fehler: null })
    .mockResolvedValue({ leihen: [leihe({})], fehler: null })
  mitDecks({
    Deckleihen: leihenAntworten,
    DeckleiheAnnehmen: () =>
      Promise.resolve({ leihen: [leihe({})], pfad: '', fehler: null }),
    UmschlagWaehlen: () =>
      Promise.resolve(umschlagGewaehlt('/tmp/fremd.msepackage')),
  })
  const hand = zeige()

  const decks = await screen.findAllByLabelText(eigene.sitzDeck)
  expect(element(decks[1]).querySelector('option[value="leihe:l1"]')).toBeNull()

  await hand.click(screen.getByRole('button', { name: eigene.leiheOeffnen }))
  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/fremd.msepackage')
  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  await waitFor(() => {
    expect(
      element(screen.getAllByLabelText(eigene.sitzDeck)[1]).querySelector(
        'option[value="leihe:l1"]',
      ),
    ).not.toBeNull()
  })
})
