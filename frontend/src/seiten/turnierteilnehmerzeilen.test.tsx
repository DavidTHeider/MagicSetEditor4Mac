// **Abmelden steht an einer Zeile, nicht an sechzehn** — UXR-12
// (#1920).
//
// # Der Befund
//
// Gemessen unter jsdom mit `test/turnierbestand.ts`: die Lasche
// *Teilnehmende* trug sechzehn Knöpfe *Abmelden* untereinander, die Werkbank
// damit 25 Bedienelemente — R2 aus `docs/UIREVIEW.md` erlaubt 20 im Bild,
// und wer mit der Tastatur zum Anmelden darunter wollte, drückte sechzehnmal
// Tabulator.
//
// # Was jetzt gilt
//
// **Dasselbe Muster wie die Plätze eines Decks (UXR-11) und die
// Paketliste (UXR-17)**: die Liste ist eine Tabstelle, die Pfeile wandern,
// und der Knopf steht an der Zeile mit der Tabstelle und an der unter dem
// Zeiger. → führt in die Zeile, ← heraus. Erreichbar bleibt das Abmelden
// jedes Teilnehmenden — mit der Maus über die Zeile, mit der Tastatur über
// Pfeile.
import { fireEvent, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { aufgeschlagen } from './turnierwerkbank'
import { raeumeApi } from '../test/api'
import { zaehle } from '../test/bedienzaehlung'
import { verstoesse } from '../test/axe'
import { abend, rangliste } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import type { Api } from '../api'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

async function teilnehmende(weitere: Partial<Api> = {}) {
  const werkbank = await aufgeschlagen(
    abend,
    {
      TurnierRangliste: () => Promise.resolve({ rangliste, fehler: null }),
      Decks: () => Promise.resolve({ decks: [], fehler: null }),
      ...weitere,
    },
    'teilnehmer',
  )
  const liste = within(werkbank).getByRole('list', { name: texte.teilnehmer })

  return {
    werkbank,
    liste,
    zeilen: within(liste).getAllByRole('listitem'),
  }
}

describe('die Teilnehmenden eines Abends', () => {
  it('tragen den Knopf an einer Zeile und nicht an sechzehn', async () => {
    const { werkbank, liste, zeilen } = await teilnehmende()

    expect(zeilen).toHaveLength(16)
    expect(within(liste).getAllByRole('button')).toHaveLength(1)
    expect(
      within(liste).getByRole('button', { name: `${texte.aussteigen}: Ada` }),
    ).toBeInTheDocument()
    expect(zaehle(werkbank).bedien).toBeLessThanOrEqual(20)
    expect(await verstoesse(werkbank)).toEqual([])
  })

  it('wandert mit den Pfeilen, und der Knopf wandert mit', async () => {
    const { liste, zeilen } = await teilnehmende()

    zeilen[0]?.focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(zeilen[1]).toHaveFocus()
    expect(within(liste).getAllByRole('button')).toHaveLength(1)
    expect(
      within(liste).getByRole('button', { name: `${texte.aussteigen}: Bert` }),
    ).toBeInTheDocument()

    await userEvent.keyboard('{End}')
    expect(zeilen[15]).toHaveFocus()
  })

  it('meldet mit der Tastatur ab — → in die Zeile, ⏎', async () => {
    const aussteigen = vi.fn<Api['TurnierAussteigen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const { zeilen } = await teilnehmende({ TurnierAussteigen: aussteigen })

    zeilen[0]?.focus()
    await userEvent.keyboard('{ArrowDown}{ArrowDown}{ArrowRight}{Enter}')

    await waitFor(() => {
      expect(aussteigen).toHaveBeenCalledWith('7', 't3', true)
    })
  })

  it('zeigt den Knopf unter dem Zeiger', async () => {
    const aussteigen = vi.fn<Api['TurnierAussteigen']>(() =>
      Promise.resolve({ fehler: null }),
    )
    const { liste, zeilen } = await teilnehmende({
      TurnierAussteigen: aussteigen,
    })

    const sechste = zeilen.at(5)
    if (sechste === undefined) {
      throw new Error('keine sechste Zeile')
    }

    // **Der Zeiger über der sechsten Zeile**, die Tabstelle bleibt an der
    // ersten — zwei Knöpfe, mehr nicht. Geht er, geht der Knopf mit.
    fireEvent.pointerEnter(sechste)
    expect(within(liste).getAllByRole('button')).toHaveLength(2)

    fireEvent.pointerLeave(sechste)
    expect(
      screen.queryByRole('button', { name: `${texte.aussteigen}: Fenja` }),
    ).toBeNull()

    fireEvent.pointerEnter(sechste)
    await userEvent.click(
      within(liste).getByRole('button', { name: `${texte.aussteigen}: Fenja` }),
    )
    await waitFor(() => {
      expect(aussteigen).toHaveBeenCalledWith('7', 't6', true)
    })
  })
})
