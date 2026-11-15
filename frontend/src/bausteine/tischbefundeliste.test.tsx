// **Was beim Zeichnen auffiel, bleibt eine Liste** — axe in WebKit am
// nativen Fenster (`tischhand09`, P4,
// #2372):
// `role="status"` stand an der `ul`, und eine Liste mit fremder Rolle ist
// keine — „das Set nennt keinen Stil“ stand ohne sie da (Regel `listitem`).
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { verstoesse } from '../test/axe'
import { Tischbefunde } from './Tischbefunde'

it('führt die Befunde des Zeichnens als Liste in einer Statuszeile', async () => {
  const { container } = render(
    <Sprachraum sprache="de">
      <Tischbefunde
        sitzbefunde={[]}
        texturfehler={undefined}
        texturbefunde={['das Set nennt keinen Stil']}
      />
    </Sprachraum>,
  )

  const status = screen.getByRole('status')
  expect(within(status).getByRole('list')).toBeInTheDocument()
  expect(await verstoesse(container)).toEqual([])
})
