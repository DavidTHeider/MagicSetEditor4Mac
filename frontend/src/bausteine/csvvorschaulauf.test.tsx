// **Eine Vorschau gehört zu dem Text, aus dem sie stammt** — Befund 3 des
// Zweitblicks über Phase 6.6 (DOM-03.AK-01,
// #1698).
//
// Wer tippt, während die Vorschau unterwegs ist, bekommt sie nicht mehr zu
// sehen — sonst stünde eine Vorschau über einem CSV, das sie nicht beschreibt,
// und *Einlesen* schriebe das andere (§7, Regel 2). Dasselbe gilt für einen
// Fehler: er gehört zu dem Text von vorhin.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sammlungsausfuhr } from './Sammlungsausfuhr'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

// unterwegs ist eine Vorschau, die erst auf Zuruf ankommt.
function unterwegs() {
  let ankommen: (antwort: unknown) => void = () => undefined
  const antwort = new Promise((aufgeloest) => {
    ankommen = aufgeloest
  })

  return {
    antwort,
    ankommen: (wert: unknown) => {
      ankommen(wert)
    },
  }
}

async function zeige(
  vorschau: Promise<unknown>,
  einlesen: Api['SammlungCSVEinlesen'],
) {
  belegeApi({
    SammlungAusgeben: (form: string) =>
      Promise.resolve({
        form,
        text: 'printing,menge\ndruck-1,4\n',
        fehler: null,
      }),
    SammlungCSVVorschauen: () => vorschau as never,
    SammlungCSVEinlesen: einlesen,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsausfuhr aufEingelesen={vi.fn()} />
    </Sprachraum>,
  )

  await userEvent.click(
    screen.getByRole('button', { name: texte.ausfuhrAlsCSV }),
  )
  await userEvent.click(
    await screen.findByRole('button', { name: texte.ausfuhrZurueck }),
  )
  await userEvent.type(
    screen.getByRole('textbox', { name: texte.ausfuhrErgebnis }),
    'x',
  )
}

const dublette = {
  posten: [
    {
      zeile: 1,
      bezeichnung: 'Lightning Bolt · M11 #149',
      menge: 4,
      zustand: 'nearMint',
      sprache: 'en',
      wunsch: false,
      art: 'zusammen',
      vorher: 2,
      nachher: 6,
    },
  ],
  befunde: [],
}

it('verwirft eine Vorschau, die zu einem geänderten Text kommt', async () => {
  const lauf = unterwegs()
  const einlesen = vi.fn<Api['SammlungCSVEinlesen']>()
  await zeige(lauf.antwort, einlesen)

  lauf.ankommen({ vorschau: dublette, fehler: null })
  await antwortAbwarten()

  expect(screen.queryByRole('region', { name: texte.vorschauTitel })).toBeNull()
  expect(einlesen).not.toHaveBeenCalled()
})

it('verwirft auch eine Antwort ohne Dublette — und liest nichts ein', async () => {
  const lauf = unterwegs()
  const einlesen = vi.fn<Api['SammlungCSVEinlesen']>()
  await zeige(lauf.antwort, einlesen)

  lauf.ankommen({ vorschau: { posten: [], befunde: [] }, fehler: null })
  await antwortAbwarten()

  expect(einlesen).not.toHaveBeenCalled()
})

it('verwirft einen Fehler, der zu einem geänderten Text kommt', async () => {
  const lauf = unterwegs()
  await zeige(lauf.antwort, vi.fn<Api['SammlungCSVEinlesen']>())

  lauf.ankommen({
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
