// **Die Freigabe sagt dem Tisch, wer Gast ist** — P6
// (#2374):
// die Sicht folgt dem Beginn nie zu einem Gastsitz. Unbekannt, solange der
// Stand fehlt; keiner, wenn der Tisch zu ist; sonst jeder Sitz außer dem des
// Wirts. **Gelesen wird beim Aufbau** — bis dahin erst auf Knopfdruck.
import { render, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Partiefreigabe } from './Partiefreigabe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einePartie, einSitz } from './partieprobe'

afterEach(raeumeApi)

function zeige(laeuft: boolean) {
  const stand = vi.fn(() =>
    Promise.resolve({
      laeuft,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [],
      fehler: null,
    }),
  )
  belegeApi({ PartieFreigabeStand: stand })
  const melde = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Partiefreigabe
        partie={einePartie({
          fuer: 'sitz1',
          sitze: [einSitz({ kennung: 'sitz1' }), einSitz({ kennung: 'sitz2' })],
        })}
        meldeGastsitze={melde}
      />
    </Sprachraum>,
  )

  return { melde, stand }
}

it('meldet erst unbekannt, dann die Gastsitze eines offenen Tisches', async () => {
  const { melde, stand } = zeige(true)

  expect(melde).toHaveBeenNthCalledWith(1, undefined)
  await waitFor(() => {
    expect(melde).toHaveBeenLastCalledWith(['sitz2'])
  })
  expect(stand).toHaveBeenCalledTimes(1)
})

it('meldet keinen Gast an einem Tisch, der zu ist', async () => {
  const { melde } = zeige(false)

  await waitFor(() => {
    expect(melde).toHaveBeenLastCalledWith([])
  })
})
