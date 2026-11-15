// **Der Haken über die Wirkungen** — PLAY-04.AK-02
// (#1712).
import { act, renderHook } from '@testing-library/react'
import { expect, it } from 'vitest'

import { einObjekt, einSitz, einePartie } from '../seiten/partieprobe'
import { useTischwirkungen } from './useTischwirkungen'

import type { Partie } from '../apipartie'

const aufDerHand = (fassung: string): Partie =>
  einePartie({
    fassung,
    objekte: [einObjekt({ kennung: 'k1', zone: 'hand' })],
    sitze: [einSitz({ kennung: 's1', leben: 20 })],
  })

const aufDemTisch = (fassung: string): Partie =>
  einePartie({
    fassung,
    objekte: [einObjekt({ kennung: 'k1', zone: 'tisch' })],
    sitze: [einSitz({ kennung: 's1', leben: 17 })],
  })

it('meldet beim ersten Bild nichts', () => {
  const { result } = renderHook(() => useTischwirkungen(aufDemTisch('1')))

  expect(result.current.jeDing.size).toBe(0)
  expect(result.current.amTisch).toEqual([])
})

it('meldet, was sich mit der nächsten Fassung geändert hat', () => {
  const { result, rerender } = renderHook(
    ({ partie }: { partie: Partie }) => useTischwirkungen(partie),
    { initialProps: { partie: aufDerHand('1') } },
  )

  act(() => {
    rerender({ partie: aufDemTisch('2') })
  })

  expect(result.current.jeDing.get('k1')).toBe('eintritt')
  expect(result.current.jeDing.get('s1')).toBe('leben')
})

// **Dieselbe Fassung ist dasselbe Bild.** Ein Sichtwechsel ändert, welche
// Objekte dieser Sitz sieht, **ohne** dass etwas geschehen wäre — wer daraus
// Wirkungen läse, ließe beim Umschalten den halben Tisch aufblitzen.
it('meldet nichts für dieselbe Fassung', () => {
  const { result, rerender } = renderHook(
    ({ partie }: { partie: Partie }) => useTischwirkungen(partie),
    { initialProps: { partie: aufDerHand('1') } },
  )

  act(() => {
    rerender({ partie: aufDemTisch('1') })
  })

  expect(result.current.jeDing.size).toBe(0)
})

// **Zwei Wirkungen an einem Ding, und die erste gewinnt.** Eine Karte, die
// getappt aufs Feld kommt, ist zuerst eingetreten; die Drehung ist die
// kleinere Auskunft, und eine Marke kann nur eine tragen.
it('behält die erste Wirkung eines Dings', () => {
  const { result, rerender } = renderHook(
    ({ partie }: { partie: Partie }) => useTischwirkungen(partie),
    { initialProps: { partie: aufDerHand('1') } },
  )

  act(() => {
    rerender({
      partie: einePartie({
        fassung: '2',
        objekte: [einObjekt({ kennung: 'k1', zone: 'tisch', getappt: true })],
        sitze: [einSitz({ kennung: 's1', leben: 20 })],
      }),
    })
  })

  expect(result.current.jeDing.get('k1')).toBe('eintritt')
})

it('nennt die Wirkungen ohne Ding gesondert', () => {
  const { result, rerender } = renderHook(
    ({ partie }: { partie: Partie }) => useTischwirkungen(partie),
    { initialProps: { partie: einePartie({ fassung: '1' }) } },
  )

  act(() => {
    rerender({
      partie: einePartie({
        fassung: '2',
        ausgang: { sieger: 'sitz1', unentschieden: false, ausgeschieden: [] },
      }),
    })
  })

  expect(result.current.amTisch).toEqual(['spielende'])
})
