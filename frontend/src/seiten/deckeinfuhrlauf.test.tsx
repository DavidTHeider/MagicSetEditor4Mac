// **Eine Vorschau gehört zu dem Text, aus dem sie stammt** — dieselbe Klasse
// wie beim CSV der Sammlung, angestrichen im Nachlauf über Phase 6.6 (T8,
// #1473).
//
// Wer tippt, während die Vorschau unterwegs ist, bekommt sie nicht mehr: sie
// beschriebe nicht, was im Feld steht, und *Übernehmen* legte ihre Zeilen
// (§7, Regel 2). Ein Fehler zu dem Text von vorhin ebenso nicht.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Deckeinfuhr } from '../bausteine/Deckeinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { DeckDTO } from '../apidecks'

const eigene = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

const deck: DeckDTO = {
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

// unterwegs ist eine Antwort, die erst auf Zuruf ankommt.
async function unterwegs() {
  let ankommen: (antwort: unknown) => void = () => undefined
  const antwort = new Promise((aufgeloest) => {
    ankommen = aufgeloest
  })

  belegeApi({ DecklisteVorschau: () => antwort as never })

  render(
    <Sprachraum sprache="de">
      <Deckeinfuhr deck={deck} gelegt={() => undefined} />
    </Sprachraum>,
  )

  const feld = screen.getByLabelText(eigene.einfuhrFeld)
  await userEvent.type(feld, '4 Lightning Bolt (M11) 149')
  await userEvent.click(
    screen.getByRole('button', { name: eigene.einfuhrVorschau }),
  )
  await userEvent.type(feld, '0')

  return (wert: unknown) => {
    ankommen(wert)
  }
}

it('verwirft eine Vorschau, die zu einem geänderten Text kommt', async () => {
  const ankommen = await unterwegs()

  ankommen({
    vorschau: {
      deckID: 7,
      stand: 'abdruck',
      hauptdeck: 4,
      sideboard: 0,
      befunde: [],
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
      ],
    },
    fehler: null,
  })
  await antwortAbwarten()

  expect(screen.queryByText(/Hauptdeck · 4 · Lightning Bolt/)).toBeNull()
  expect(
    screen.queryByRole('button', { name: eigene.einfuhrUebernehmen }),
  ).toBeNull()
})

it('verwirft einen Fehler, der zu einem geänderten Text kommt', async () => {
  const ankommen = await unterwegs()

  ankommen({
    vorschau: null,
    fehler: {
      code: 'validierung',
      schluessel: 'validierung',
      meldung: 'kaputt',
    },
  })
  await antwortAbwarten()

  expect(screen.queryByRole('alert')).toBeNull()
})
