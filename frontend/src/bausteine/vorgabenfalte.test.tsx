// **Die mehrwertigen Vorgaben falten** — #2022, Folge aus UXR-20
// (#1928).
//
// Präferenzdecks und Lagerorte standen als Kästchen je Deck oder Ort da und
// wuchsen mit dem Bestand (R6). Seitdem stehen acht und jedes gewählte, der
// Rest auf Wunsch.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Profilvorgaben } from './Profilvorgaben'

afterEach(() => {
  raeumeApi()
})

const zwoelf = Array.from({ length: 12 }, (_, stelle) => ({
  wert: String(stelle + 1),
  beschriftung: `Deck ${String(stelle + 1)}`,
}))

it('zeigt acht Decks und jedes gewählte, den Rest auf Wunsch', async () => {
  belegeApi({
    Profilvorgaben: () =>
      Promise.resolve({
        vorgaben: [
          {
            art: 'deck',
            mehrwertig: true,
            werte: ['11'],
            moeglich: zwoelf,
            verwaist: [],
          },
        ],
      }),
  })
  render(
    <Sprachraum sprache="de">
      <Profilvorgaben profilID="10" />
    </Sprachraum>,
  )

  const gruppe = await screen.findByRole('group')
  const kaestchen = () => within(gruppe).getAllByRole('checkbox')

  expect(kaestchen()).toHaveLength(9)
  expect(within(gruppe).getByLabelText('Deck 11')).toBeChecked()

  await userEvent.click(screen.getByRole('button', { name: /Weitere 3/ }))

  expect(kaestchen()).toHaveLength(12)
})
