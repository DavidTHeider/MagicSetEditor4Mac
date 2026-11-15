// **Eine überholte Vorschau trägt sich nicht ein** — DOM-03.AK-01
// (#1698),
// §7 Regel 2: die Antwort gehört zu der Liste, aus der sie stammt.
//
// Wer während der unterwegs befindlichen Vorschau den Text ändert, hat eine
// andere Liste vor sich — die alte Vorschau und ihr Fehler gehören nicht
// darüber.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Listeneinfuhr } from './Listeneinfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'
import type {
  Listendeutungszeile,
  Sammelvorschauantwort,
} from '../antwortenlisteneinfuhr'

const texte = kataloge.de.seiten.sammlung

afterEach(() => {
  raeumeApi()
})

const zeile: Listendeutungszeile = {
  roh: '4 Lightning Bolt (M11) 149',
  zeilennummer: 1,
  menge: 4,
  name: 'Lightning Bolt',
  setcode: 'm11',
  nummer: '149',
  finish: 'nonfoil',
  treffer: [
    {
      printing: 'druck-1',
      name: 'Lightning Bolt',
      setcode: 'm11',
      setname: 'Magic 2011',
      nummer: '149',
      seltenheit: 'common',
      bild: '',
    },
  ],
  drucke: 1,
  eindeutig: true,
  befund: '',
  weitere: false,
  fehler: null,
}

const dublette: Sammelvorschauantwort = {
  vorschau: {
    posten: [
      {
        zeile: 1,
        bezeichnung: 'Lightning Bolt',
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
  },
  fehler: null,
}

// zurueckgehalten gibt eine Vorschau, die erst auf Zuruf antwortet.
async function zurueckgehalten(antwort: Sammelvorschauantwort) {
  let loslassen: () => void = () => undefined
  const halt = new Promise<void>((los) => {
    loslassen = los
  })

  const einlesen = vi.fn<Api['ListeEinlesen']>()

  belegeApi({
    ListeDeuten: () =>
      Promise.resolve({
        deutung: { zeilen: [zeile], befunde: [], alterSekunden: 0 },
        fehler: null,
      }),
    ListeVorschauen: async () => {
      await halt

      return antwort
    },
    ListeEinlesen: einlesen,
  })

  render(
    <Sprachraum sprache="de">
      <Listeneinfuhr aufAufgenommen={vi.fn()} />
    </Sprachraum>,
  )

  const feld = screen.getByRole('textbox', { name: texte.listeFeld })
  await userEvent.type(feld, '4 Lightning Bolt (M11) 149')
  await userEvent.click(screen.getByRole('button', { name: texte.listeDeuten }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.listeEinlesen }),
  )

  // **Während die Vorschau unterwegs ist, ändert sich der Text** — und
  // danach kommt sie an.
  await userEvent.type(feld, ' x')
  loslassen()

  // **Gewartet wird auf die Wirkung** (§3): der Vorgang ist zu Ende, wenn
  // *Zuordnen* wieder geht.
  await waitFor(() => {
    expect(
      screen.getByRole('button', { name: texte.listeDeuten }),
    ).toBeEnabled()
  })

  return einlesen
}

describe('Eine überholte Vorschau', () => {
  it('schlägt sich nicht auf die neue Liste', async () => {
    const einlesen = await zurueckgehalten(dublette)

    expect(
      screen.queryByRole('region', { name: texte.vorschauTitel }),
    ).toBeNull()
    expect(einlesen).not.toHaveBeenCalled()
  })

  it('meldet ihren Fehler nicht über der neuen Liste', async () => {
    const einlesen = await zurueckgehalten({
      vorschau: null,
      fehler: {
        code: 'unbekannt',
        schluessel: 'unbekannt',
        meldung: 'Die Sammlung antwortet nicht.',
      },
    })

    expect(screen.queryByRole('alert')).toBeNull()
    expect(einlesen).not.toHaveBeenCalled()
  })
})
