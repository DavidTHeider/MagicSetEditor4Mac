// **Gespräch und Zeiger gehören zu einer Partie** — Zweitblick über Phase
// 6.11 (#1672, Z1 betrifft alle drei Partieereignisse).
//
// Wirt und Gast hören dasselbe `partie:nebenbei`, und dieselbe Anwendung kann
// zugleich Wirt eines Tisches und Gast an einem anderen sein. Ohne Filter
// stand der Satz eines Tisches im Gespräch des anderen.
import { act, renderHook, waitFor } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import { belegeApi, raeumeApi } from '../test/api'
import { ereignisNebenbei, usePartienebenbei } from './usePartienebenbei'

import type { ReactNode } from 'react'
import type { Nebenbei } from '../apinebenbei'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  raeumeApi()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function einSatz(text: string): Nebenbei {
  return {
    art: 'chat',
    von: 'bert',
    text,
    zeigerart: '',
    objekt: '',
    x: 0,
    y: 0,
  }
}

it('nimmt nur, was zu seiner Partie gehört', () => {
  probe = laufzeitprobe()
  const { melde } = probe

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'filterprobe', 'p1'),
    { wrapper: raum },
  )

  act(() => {
    melde(ereignisNebenbei, 'p2', einSatz('fremd'))
    melde(ereignisNebenbei, 'p1', einSatz('eigen'))
  })

  expect(result.current.gespraech.map((zeile) => zeile.text)).toEqual(['eigen'])
})

it('nimmt nichts, solange die Partie nicht feststeht', () => {
  probe = laufzeitprobe()
  const { melde } = probe

  const { result } = renderHook(
    () => usePartienebenbei('anna', 'leerprobe', ''),
    { wrapper: raum },
  )

  act(() => {
    melde(ereignisNebenbei, 'p1', einSatz('zu früh'))
  })

  expect(result.current.gespraech).toEqual([])
})

// **Ein Wechsel der Partie räumt, was zur vorigen gehörte** — N4 aus dem
// Nachlauf über Phase 6.11: der Haken blieb eingehängt, und die Sätze von A
// standen im Gespräch von B.
it('räumt Gespräch und Zeiger, wenn die Partie wechselt', () => {
  probe = laufzeitprobe()
  const { melde } = probe

  const { result, rerender } = renderHook(
    ({ partie }: { partie: string }) =>
      usePartienebenbei('anna', 'wechselprobe', partie),
    { wrapper: raum, initialProps: { partie: 'p1' } },
  )

  act(() => {
    melde(ereignisNebenbei, 'p1', einSatz('von A'))
    melde(ereignisNebenbei, 'p1', {
      art: 'praesenz',
      von: 'bert',
      text: '',
      zeigerart: 'zeiger',
      objekt: 'k1',
      x: 1,
      y: 1,
    })
  })
  expect(result.current.gespraech).toHaveLength(1)
  expect(result.current.zeiger).toHaveLength(1)

  rerender({ partie: 'p2' })

  expect(result.current.gespraech).toEqual([])
  expect(result.current.zeiger).toEqual([])
})

it('trägt einen Satz nicht ein, dessen Antwort nach dem Wechsel kommt', async () => {
  probe = laufzeitprobe()
  const offen: ((wert: { fehler: null }) => void)[] = []
  belegeApi({
    PartieSagen: () =>
      new Promise((los) => {
        offen.push(los)
      }),
  })

  const { result, rerender } = renderHook(
    ({ partie }: { partie: string }) =>
      usePartienebenbei('anna', 'antwortprobe', partie),
    { wrapper: raum, initialProps: { partie: 'p1' } },
  )

  act(() => {
    result.current.sagen('für A')
  })
  rerender({ partie: 'p2' })

  // **Erst die alte Antwort, dann eine neue, deren Wirkung man sieht** (§3):
  // steht „für B“ da, ist auch die Antwort davor verarbeitet.
  act(() => {
    offen[0]?.({ fehler: null })
    result.current.sagen('für B')
  })
  await waitFor(() => {
    expect(offen).toHaveLength(2)
  })
  act(() => {
    offen[1]?.({ fehler: null })
  })
  await waitFor(() => {
    expect(result.current.gespraech.map((zeile) => zeile.text)).toContain(
      'für B',
    )
  })

  expect(result.current.gespraech.map((zeile) => zeile.text)).toEqual(['für B'])
})

it('meldet keinen Fehler eines Satzes, der der vorigen Partie galt', async () => {
  probe = laufzeitprobe()
  const offen: ((grund: Error) => void)[] = []
  belegeApi({
    PartieSagen: () =>
      new Promise((_, nein) => {
        offen.push(nein)
      }),
  })

  const { result, rerender } = renderHook(
    ({ partie }: { partie: string }) =>
      usePartienebenbei('anna', 'fehlerprobe', partie),
    { wrapper: raum, initialProps: { partie: 'p1' } },
  )

  act(() => {
    result.current.sagen('für A')
  })
  rerender({ partie: 'p2' })

  act(() => {
    offen[0]?.(new Error('weg'))
    result.current.sagen('für B')
  })
  await waitFor(() => {
    expect(offen).toHaveLength(2)
  })
  expect(result.current.fehler).toBeUndefined()

  // **Und der Fehler der eigenen Partie steht da** — dass jemand zuhört.
  act(() => {
    offen[1]?.(new Error('weg'))
  })
  await waitFor(() => {
    expect(result.current.fehler).toBeDefined()
  })
})
