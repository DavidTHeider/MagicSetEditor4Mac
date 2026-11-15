// **Bildchen erst, wenn das Rollen innehält** — E07
// (#2279).
//
// Das erste Rollen durch Black (3.788 Karten) fragte 87-mal nach Bildchen —
// für jedes Fenster, an dem die Liste vorbeizog —, und Go renderte sie
// gleichzeitig, während WebKit die Bilder des Rollens zeichnen wollte
// (`docs/belege/huellenprobe/e07rollen.md`). Gefragt wird jetzt das Fenster,
// bei dem die Liste zur Ruhe kommt.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { listenbilderruhe, useListenbilder } from './useListenbilder'

afterEach(() => {
  raeumeApi()
})

function Probe({ ids }: { ids: string[] }) {
  const bilder = useListenbilder('7', ids, 0)

  return <p data-testid="bilder">{JSON.stringify(bilder)}</p>
}

it('fragt beim schnellen Rollen nur das Fenster, bei dem die Liste innehält', async () => {
  const geholt = vi.fn().mockResolvedValue({
    bilder: { '9': 'data:image/png;base64,AAA' },
    fehler: undefined,
  })
  belegeApi({ Kartenbilder: geholt })

  const { rerender } = render(<Probe ids={['1', '2']} />)
  rerender(<Probe ids={['4', '5']} />)
  rerender(<Probe ids={['8', '9']} />)

  await waitFor(() => {
    expect(screen.getByTestId('bilder')).toHaveTextContent('data:image/png')
  })

  // **Und es kommt kein zweiter Ruf nach** — gewartet über die Ruhe hinaus.
  await new Promise((weiter) => setTimeout(weiter, listenbilderruhe * 3))

  expect(geholt).toHaveBeenCalledTimes(1)
  expect(geholt).toHaveBeenCalledWith('7', ['8', '9'])
})
