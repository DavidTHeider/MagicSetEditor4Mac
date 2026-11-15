// **Klick und Doppelklick an einer Karte, in der Zeit** — P4
// (#2372):
// ein Doppelklick auf ein Grundland gibt Mana und tappt es nicht vorher; ein
// Klick, auf den kein zweiter folgt, tappt nach der Frist; und nach dem Abbau
// stellt kein wartender Klick mehr einen Befehl.
import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { doppelklickfrist } from '../spiel/kartengriff'
import { useKartengriff } from './useKartengriff'

import type { Objekthandlung } from '../apipartie'

const grundland: Objekthandlung = {
  objekt: 'sitz1/wald/0',
  handlungen: ['objekt.tappen', 'objekt.mana'],
  primaer: 'objekt.tappen',
  zweit: 'objekt.mana',
}

const baer: Objekthandlung = {
  objekt: 'sitz1/baer/0',
  handlungen: ['objekt.tappen'],
  primaer: 'objekt.tappen',
  zweit: '',
}

beforeEach(() => {
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

function zeige() {
  const befehlen = vi.fn()
  const waehle = vi.fn()
  const haken = renderHook(
    ({ stand }: { stand: string }) =>
      useKartengriff({
        handlungen: [grundland, baer],
        befehlen,
        waehle,
        stand,
      }),
    { initialProps: { stand: 'p1/sitz1' } },
  )

  return { befehlen, waehle, haken }
}

describe('der Griff in der Zeit', () => {
  it('gibt bei einem Doppelklick Mana und tappt nicht vorher', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 1, metaKey: false })
      haken.result.current('sitz1/wald/0', { detail: 2, metaKey: false })
      vi.advanceTimersByTime(doppelklickfrist * 2)
    })

    expect(befehlen.mock.calls).toEqual([
      ['objekt.mana', { objekt: 'sitz1/wald/0' }],
    ])
  })

  it('tappt nach der Frist, wenn kein zweiter Klick kommt', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 1, metaKey: false })
    })
    expect(befehlen).not.toHaveBeenCalled()

    act(() => {
      vi.advanceTimersByTime(doppelklickfrist)
    })
    expect(befehlen.mock.calls).toEqual([
      ['objekt.tappen', { objekt: 'sitz1/wald/0' }],
    ])
  })

  it('wirkt ohne Zweithandlung sofort und genau einmal', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/baer/0', { detail: 1, metaKey: false })
      haken.result.current('sitz1/baer/0', { detail: 2, metaKey: false })
    })

    expect(befehlen.mock.calls).toEqual([
      ['objekt.tappen', { objekt: 'sitz1/baer/0' }],
    ])
  })

  it('wählt mit ⌘', () => {
    const { befehlen, waehle, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/baer/0', { detail: 1, metaKey: true })
    })

    expect(waehle).toHaveBeenCalledWith('sitz1/baer/0')
    expect(befehlen).not.toHaveBeenCalled()
  })

  it('stellt nach dem Abbau keinen wartenden Befehl mehr', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 1, metaKey: false })
    })
    haken.unmount()
    act(() => {
      vi.advanceTimersByTime(doppelklickfrist * 2)
    })

    expect(befehlen).not.toHaveBeenCalled()
  })

  // **Ein Klick auf eine andere Karte stellt den wartenden sofort** — sonst
  // ginge er verloren oder träfe nach der Frist auf einen anderen Stand.
  it('stellt den wartenden Klick, sobald eine andere Karte gegriffen wird', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 1, metaKey: false })
      haken.result.current('sitz1/baer/0', { detail: 1, metaKey: false })
    })

    expect(befehlen.mock.calls).toEqual([
      ['objekt.tappen', { objekt: 'sitz1/wald/0' }],
      ['objekt.tappen', { objekt: 'sitz1/baer/0' }],
    ])
  })

  // **Ein zweiter Klick ohne wartenden ersten** — etwa nach einem Griff an
  // einer anderen Karte dazwischen — löst trotzdem die Zweithandlung aus.
  it('löst die Zweithandlung auch ohne wartenden ersten Klick aus', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 2, metaKey: false })
    })

    expect(befehlen.mock.calls).toEqual([
      ['objekt.mana', { objekt: 'sitz1/wald/0' }],
    ])
  })

  // **Und nach einem Wechsel der Sicht ebenso** — Zweitblick über Phase 7,
  // Befund 10. Ein Sitzwechsel oder der Vorhang baut den Tisch nicht ab;
  // der wartende Klick hielt die vorige Sicht und schrieb nach der Frist
  // noch für den vorigen Sitz.
  it('stellt nach einem Wechsel der Sicht keinen wartenden Befehl mehr', () => {
    const { befehlen, haken } = zeige()

    act(() => {
      haken.result.current('sitz1/wald/0', { detail: 1, metaKey: false })
    })
    haken.rerender({ stand: 'p1/sitz2' })
    act(() => {
      vi.advanceTimersByTime(doppelklickfrist * 2)
    })

    expect(befehlen).not.toHaveBeenCalled()
  })
})
