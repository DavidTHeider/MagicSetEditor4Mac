// **Ein offener Tisch liest nach, was eine andere Fläche an seiner Partie
// schreibt** — #2388.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): die Probe legte über
// `mse partie-befehl` zwei Kreaturen aufs Spielfeld, das Fenster zeigte nach
// 15 s noch die alte Fassung, und der nächste Klick prallte ab. Der fremde
// Stand stieg — nur hing der Tisch nicht daran.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { usePartie } from './usePartie'

import { eineAntwort } from '../seiten/partieprobe'

import type { ReactNode } from 'react'

afterEach(() => {
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

it('liest die offene Partie nach, wenn der fremde Stand steigt', async () => {
  let fassung = '17'
  const gelesen = vi.fn(() =>
    Promise.resolve(eineAntwort({ id: 'p1', fuer: 'sitz1', fassung })),
  )

  belegeApi({ Partie: gelesen })

  const { result, rerender } = renderHook(({ stand }) => usePartie(stand), {
    wrapper: raum,
    initialProps: { stand: 0 },
  })

  act(() => {
    result.current.ansehen('p1', 'sitz1')
  })

  await waitFor(() => {
    expect(result.current.partie?.fassung).toBe('17')
  })

  fassung = '19'
  rerender({ stand: 1 })

  await waitFor(() => {
    expect(result.current.partie?.fassung).toBe('19')
  })

  expect(gelesen).toHaveBeenLastCalledWith('p1', 'sitz1')
})

it('liest ohne offene Partie nichts', async () => {
  const gelesen = vi.fn(() => Promise.resolve(eineAntwort({ id: 'p1' })))

  belegeApi({ Partie: gelesen })

  const { rerender } = renderHook(({ stand }) => usePartie(stand), {
    wrapper: raum,
    initialProps: { stand: 0 },
  })

  rerender({ stand: 1 })
  await effekteNachholen()

  expect(gelesen).not.toHaveBeenCalled()
})
