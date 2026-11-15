// **Die Freigabe folgt der Besetzung** — gefunden mit zwei Fenstern
// (Hüllenprobe `nat04`, S12.NAT-04): Bo saß am Tisch und hatte schon eine
// Münze geworfen, und die Freigabe des Wirts sagte „Noch sitzt niemand". Go
// meldet jetzt `partie:freigabe`, und der Haken holt den Stand neu.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { ereignisPartiefreigabe, usePartiefreigabe } from './usePartiefreigabe'

import type { ReactNode } from 'react'
import type { PartiefreigabeAntwort } from '../api'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

it('holt den Stand neu, wenn sich die Besetzung des eigenen Tisches ändert', async () => {
  probe = laufzeitprobe()
  const { melde } = probe
  const stand = vi.fn((): Promise<PartiefreigabeAntwort> =>
    Promise.resolve({
      laeuft: true,
      partie: 'p1',
      wirt: 'ada',
      vergeben: [],
      fehler: null,
    }),
  )
  belegeApi({ PartieFreigabeStand: stand })

  renderHook(() => usePartiefreigabe('p1', 'ada'), { wrapper: raum })

  // **Eine andere Partie geht diesen Tisch nichts an** …
  act(() => {
    melde(ereignisPartiefreigabe, 'p2')
  })
  await effekteNachholen()
  expect(stand).not.toHaveBeenCalled()

  // … **die eigene schon** — dass jemand zuhört (§3).
  act(() => {
    melde(ereignisPartiefreigabe, 'p1')
  })
  await waitFor(() => {
    expect(stand).toHaveBeenCalledOnce()
  })
})
