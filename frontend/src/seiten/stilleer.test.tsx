// **Ohne installierten Stil führt die Seite zu den Paketen** — G02
// (#2249).
//
// „Kein Stil installiert" stand da, und der Weg zu einem stand nirgends.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Stil } from './Stil'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

describe('Leerzustand der Stilwahl', () => {
  it('bietet ohne Stil den Weg zu den Paketen an', async () => {
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Karten: () => Promise.resolve({ karten: [] }),
    })

    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Stil setID="1" geaendert={vi.fn()} wechseln={wechseln} />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(texte.seiten.karten.stilLeer),
    ).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.allgemein.arbeitsplatz.zuDenPaketen,
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('pakete')
  })
})
