// **Symbolpalette und Wortlistenwahl sind für einen Leser gesperrt** —
// Befund 5 des Zweitblicks über Phase G, zweiter Teil (#2290,
// #2290).
//
// Das Textfeld eines Kartenfelds war schreibgeschützt, seine zwei
// Geschwister nicht: die Wortwahl rief `WortWaehlen` sofort, die Palette
// übernahm einen neuen Entwurf in die Vorschau.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Schreibrechtraum } from '../haken/schreibrecht'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Symbolpalette } from './Symbolpalette'
import { Wortlistenwahl } from './Wortlistenwahl'

const satz = kataloge.de.bausteine.nurlesen.knopf

function alsLeser(knoten: React.ReactNode) {
  render(
    <Sprachraum sprache="de">
      <Schreibrechtraum nurLesend>{knoten}</Schreibrechtraum>
    </Sprachraum>,
  )
}

it('die Symbolpalette fügt für einen Leser nichts ein', async () => {
  const einfuegen = vi.fn()
  alsLeser(
    <Symbolpalette
      eintraege={[
        {
          art: 'code',
          code: 'T',
          beschriftung: 'Tappen',
          rueckfrage: '',
          bild: '',
          kinder: [],
        },
      ]}
      einfuegen={einfuegen}
      angehalten={false}
    />,
  )

  const kachel = screen.getByRole('button', { name: /Tappen/ })
  expect(kachel).toBeDisabled()
  // **Und sie sagt, warum** (H01, #2255) — wie ein gesperrter Knopf.
  expect(kachel).toHaveAccessibleDescription(satz)
  await userEvent.click(kachel)
  expect(einfuegen).not.toHaveBeenCalled()
})

it('die Wortlistenwahl wählt für einen Leser nichts', () => {
  const waehlen = vi.fn()
  alsLeser(
    <Wortlistenwahl
      bereiche={[
        {
          liste: 'type-en',
          text: 'Legendary Creature',
          kaestchen: true,
          gewaehlt: 1,
          eintraege: [
            {
              name: 'Legendary ',
              istPraefix: true,
              zeileDarunter: false,
              aktiv: true,
              kinder: null,
            },
            {
              name: 'Creature',
              istPraefix: false,
              zeileDarunter: false,
              aktiv: true,
              kinder: null,
            },
            {
              name: 'Instant',
              istPraefix: false,
              zeileDarunter: false,
              aktiv: false,
              kinder: null,
            },
          ],
        },
      ]}
      angehalten={false}
      waehlen={waehlen}
    />,
  )

  expect(screen.getByRole('checkbox', { name: 'Legendary' })).toBeDisabled()
  expect(screen.getByRole('combobox')).toBeDisabled()
  expect(
    screen.getByRole('checkbox', { name: 'Legendary' }),
  ).toHaveAccessibleDescription(satz)
  expect(screen.getByRole('combobox')).toHaveAccessibleDescription(satz)
})
