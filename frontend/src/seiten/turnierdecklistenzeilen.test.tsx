// **Die Knöpfe einer Deckliste stehen an einer Zeile** — UXR-12
// (#1920).
//
// # Der Befund
//
// Gemessen unter jsdom mit `test/turnierbestand.ts` — acht der sechzehn
// haben ein Deck eingereicht: die Lasche *Decklisten* trug je Zeile den
// Knopf mit der Kartenzahl und *Deck aufschlagen*, sechzehn Knöpfe, die
// Werkbank damit 24 (R2 in `docs/UIREVIEW.md`). Reichen alle sechzehn ein,
// sind es 32.
//
// # Was jetzt gilt
//
// **Wie in der Lasche *Teilnehmende***: eine Tabstelle, die Pfeile wandern,
// und die Knöpfe stehen an der Zeile mit der Tabstelle und an der unter dem
// Zeiger. An den übrigen steht die Kartenzahl als Text — sie ist eine
// Angabe, und erst an der Zeile, an der man ist, ein Knopf.
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Turnierdecklisten } from '../bausteine/Turnierdecklisten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { abend } from '../test/turnierbestand'
import { kataloge } from '../sprache/kataloge'
import { teilnehmerprobe, turnierprobe } from './turnierprobe'
import type { TurnierDTO } from '../apiturniere'

const texte = kataloge.de.seiten.turniere

afterEach(() => {
  raeumeApi()
})

// zeile gibt eine Zeile der Liste — oder sagt, dass es sie nicht gibt.
function zeile(zeilen: HTMLElement[], stelle: number): HTMLElement {
  const gefunden = zeilen.at(stelle)
  if (gefunden === undefined) {
    throw new Error(`keine Zeile ${String(stelle)}`)
  }

  return gefunden
}

function zeige(turnier: TurnierDTO = abend) {
  belegeApi({})

  const { container } = render(
    <Sprachraum sprache="de">
      <Turnierdecklisten turnier={turnier} />
    </Sprachraum>,
  )
  const liste = screen.getByRole('list', { name: texte.gebietDecklisten })

  return {
    container,
    liste,
    zeilen: within(liste)
      .getAllByRole('listitem')
      .filter((zeile) => zeile.parentElement === liste),
  }
}

describe('die Decklisten eines Abends', () => {
  it('tragen die Knöpfe an einer Zeile und die Zahl an den übrigen', async () => {
    const { container, liste, zeilen } = zeige()

    expect(zeilen).toHaveLength(8)
    expect(within(liste).getAllByRole('button')).toHaveLength(2)
    expect(
      within(liste).getByRole('button', { name: `${texte.deckabzug}: Ada` }),
    ).toBeInTheDocument()
    expect(
      within(liste).getByRole('button', {
        name: `${texte.deckAufschlagen}: Ada`,
      }),
    ).toBeInTheDocument()

    // **Die Kartenzahl bleibt an jeder Zeile stehen** — als Text.
    expect(within(zeile(zeilen, 1)).getByText('4')).toBeInTheDocument()
    expect(await verstoesse(container)).toEqual([])
  })

  it('wandert mit den Pfeilen, und die Knöpfe wandern mit', async () => {
    const { liste, zeilen } = zeige()

    zeile(zeilen, 0).focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(zeile(zeilen, 1)).toHaveFocus()
    expect(within(liste).getAllByRole('button')).toHaveLength(2)
    expect(
      within(liste).getByRole('button', {
        name: `${texte.deckAufschlagen}: Cleo`,
      }),
    ).toBeInTheDocument()
  })

  it('zeigt die Knöpfe unter dem Zeiger', () => {
    const { liste, zeilen } = zeige()
    const vierte = zeile(zeilen, 3)

    fireEvent.pointerEnter(vierte)
    expect(within(liste).getAllByRole('button')).toHaveLength(4)

    fireEvent.pointerLeave(vierte)
    expect(within(liste).getAllByRole('button')).toHaveLength(2)
  })

  // **Ohne Deck kein Aufschlagen** — auch an der Zeile, an der man ist.
  it('bietet an einer Liste ohne Deck nur die Liste an', () => {
    const { liste } = zeige(
      turnierprobe({
        teilnehmer: [
          teilnehmerprobe({
            name: 'Bert',
            deckabzug: [
              {
                id: 'z1',
                bezeichnung: 'Glutfalke',
                printing: '',
                druckID: '',
                menge: 4,
                bereich: 'hauptdeck',
              },
            ],
          }),
        ],
      }),
    )

    expect(within(liste).getAllByRole('button')).toHaveLength(1)
    expect(
      within(liste).getByRole('button', { name: `${texte.deckabzug}: Bert` }),
    ).toBeInTheDocument()
  })
})
