// **Ein Protokoll bleibt eine Liste** — gefunden mit axe in WebKit am
// nativen Fenster (`tischhand09`, P4,
// #2372):
// `role="log"` stand an der `ul` selbst, und eine Liste mit fremder Rolle ist
// keine Liste mehr — ihre Einträge stehen dann ohne sie da (axe, Regel
// `listitem`). Unter jsdom sah es niemand, weil die geprüften Seiten keine
// Ereignisse trugen. Die Rolle gehört an einen Rahmen um die Liste; was dazu
// kommt, sagt er weiter an (#1992).
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Ereignisliste } from './Ereignisliste'
import { Sprachraum } from '../sprache/kontext'
import { einLauf } from './partieprobe'
import { verstoesse } from '../test/axe'

it('führt die Ereignisse als Liste in einem Protokoll', async () => {
  const { container } = render(
    <Sprachraum sprache="de">
      <Ereignisliste
        lauf={einLauf({
          ereignisse: [
            {
              folge: '1',
              art: 'spieler.leben',
              oeffentlich: {},
              privat: {},
              kette: 'abcd',
            },
          ],
        })}
      />
    </Sprachraum>,
  )

  const protokoll = screen.getByRole('log')
  expect(within(protokoll).getByRole('list')).toBeInTheDocument()
  expect(within(protokoll).getAllByRole('listitem')).toHaveLength(1)
  expect(await verstoesse(container)).toEqual([])
})
