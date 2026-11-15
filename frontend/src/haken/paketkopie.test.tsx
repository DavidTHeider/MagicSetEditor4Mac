// **Die eigene Kopie, am Haken selbst** — FMT-02.AK-04
// (#1691).
//
// Ein Pfad, den die Liste nicht (mehr) kennt, kopiert nichts: sie kann
// zwischen Zeichnen und Klick nachgeladen worden sein.
import { act, renderHook } from '@testing-library/react'
import { type ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { paketDTO } from '../test/vorlagenprobe'
import { usePaketkopie } from './usePaketkopie'

afterEach(raeumeApi)

function sprachraum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

describe('usePaketkopie', () => {
  it('kopiert nichts, wenn die Liste den Pfad nicht kennt', async () => {
    const kopieren = vi.fn(() =>
      Promise.resolve({ paket: null, fehler: null, abgebrochen: false }),
    )
    belegeApi({ PaketKopieren: kopieren })
    const bekannt = paketDTO('probe.mse-style')

    const { result } = renderHook(
      () =>
        usePaketkopie({
          wahl: null,
          eintraege: [bekannt],
          waehlen: vi.fn(),
          paketeNeuLaden: vi.fn(),
        }),
      { wrapper: sprachraum },
    )

    await act(() => result.current.kopieren('/gibt/es/nicht'))
    expect(kopieren).not.toHaveBeenCalled()

    // **Der Beleg, dass der Weg trägt**: derselbe Ruf mit einem bekannten
    // Pfad fragt nach.
    await act(() => result.current.kopieren(bekannt.pfad))
    expect(kopieren).toHaveBeenCalledWith(bekannt.art, bekannt.name, '')
  })
})
