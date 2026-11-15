// **Ein Dienst auf mehreren Adressen** — G08
// (#2272).
//
// Der Stand nennt jede Adresse, auf der er lauscht, und jede gewählte, auf
// der keiner entstand — mit Grund.
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { laeuft } from '../test/dienststaende'
import { Dienstposten } from './Dienstposten'

const eigene = kataloge.de.seiten.technik.schnittstellen.lauschwahl

afterEach(raeumeApi)

it('nennt jeden Lauscher und jede fehlende Adresse', async () => {
  belegeApi({
    RESTStand: vi.fn().mockResolvedValue({
      ...laeuft,
      stand: {
        ...laeuft.stand,
        adresse: 'http://127.0.0.1:8123',
        lauscher: ['http://127.0.0.1:8123', 'http://[::1]:8123'],
        fehlende: [{ adresse: '192.0.2.1', grund: 'bind: can’t assign' }],
      },
    }),
  })

  render(
    <Sprachraum sprache="de">
      <Dienstposten art="rest" />
    </Sprachraum>,
  )

  expect(await screen.findByText('http://[::1]:8123')).toBeInTheDocument()
  expect(
    screen.getByText(`${eigene.weitere}:`, { exact: false }),
  ).toBeInTheDocument()
  expect(
    screen.getByText(
      eigene.ohne
        .replace('{adresse}', '192.0.2.1')
        .replace('{grund}', 'bind: can’t assign'),
    ),
  ).toBeInTheDocument()
})
