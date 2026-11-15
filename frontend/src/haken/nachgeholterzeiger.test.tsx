// **Ein nachgeholter Zeiger gilt ab seinem Eingang** — Z2 aus dem Zweitblick
// über Phase 6.11 (#1672).
//
// Verborgen baut `<Activity>` den Horcher ab, und `nachholen.ts` spielt beim
// Einblenden nach, was kam. Der Zeiger nahm dabei *jetzt* als Eingang: eine
// Meldung von vor einer Minute galt noch einmal fünfzehn Sekunden.
import { act, renderHook } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { laufzeitprobe, type Laufzeitprobe } from '../test/laufzeitprobe'
import {
  Praesenzfrist,
  ereignisNebenbei,
  usePartienebenbei,
} from './usePartienebenbei'

import type { ReactNode } from 'react'
import type { Nebenbei } from '../apinebenbei'

let probe: Laufzeitprobe | undefined

afterEach(() => {
  probe?.raeume()
  probe = undefined
  vi.useRealTimers()
})

function raum({ children }: { children: ReactNode }) {
  return <Sprachraum sprache="de">{children}</Sprachraum>
}

function einZeiger(von: string): Nebenbei {
  return {
    art: 'praesenz',
    von,
    text: '',
    zeigerart: 'zeiger',
    objekt: 'k1',
    x: 1,
    y: 0,
  }
}

function einSatz(von: string, text: string): Nebenbei {
  return { art: 'chat', von, text, zeigerart: '', objekt: '', x: 0, y: 0 }
}

it('lässt einen Zeiger fallen, dessen Frist im Verborgenen ablief', () => {
  vi.useFakeTimers()
  probe = laufzeitprobe()
  const { melde } = probe

  const erster = renderHook(
    () => usePartienebenbei('anna', 'zeigerprobe', 'p1'),
    {
      wrapper: raum,
    },
  )
  erster.unmount()

  act(() => {
    melde(ereignisNebenbei, 'p1', einZeiger('bert'))
    melde(ereignisNebenbei, 'p1', einSatz('bert', 'Hallo'))
  })

  vi.advanceTimersByTime(Praesenzfrist + 1)

  const zweiter = renderHook(
    () => usePartienebenbei('anna', 'zeigerprobe', 'p1'),
    {
      wrapper: raum,
    },
  )

  // **Der Satz kommt nach** — dass überhaupt nachgeholt wurde (§3).
  expect(zweiter.result.current.gespraech.map((zeile) => zeile.text)).toEqual([
    'Hallo',
  ])
  expect(zweiter.result.current.zeiger).toEqual([])
})

it('behält einen nachgeholten Zeiger nur für den Rest seiner Frist', () => {
  vi.useFakeTimers()
  probe = laufzeitprobe()
  const { melde } = probe

  const erster = renderHook(
    () => usePartienebenbei('anna', 'restprobe', 'p1'),
    {
      wrapper: raum,
    },
  )
  erster.unmount()

  act(() => {
    melde(ereignisNebenbei, 'p1', einZeiger('bert'))
  })

  vi.advanceTimersByTime(Praesenzfrist - 2_000)

  const zweiter = renderHook(
    () => usePartienebenbei('anna', 'restprobe', 'p1'),
    {
      wrapper: raum,
    },
  )

  expect(zweiter.result.current.zeiger).toHaveLength(1)

  act(() => {
    vi.advanceTimersByTime(3_000)
  })

  expect(zweiter.result.current.zeiger).toEqual([])
})
