// **Der Haken hört zu, liest die Liste und hört wieder auf** —
// #1630.
//
// **Warum das ein eigener Lauf ist**: die Läufe am Rahmen setzen eine feste
// Breite und prüfen, was daraus folgt. Dass ein **Wechsel** der Breite
// ankommt, prüft keiner von ihnen — und genau das ist der Fall im Betrieb:
// jemand zieht das Fenster kleiner, und die Leiste soll nachgeben, ohne dass
// die Anwendung neu gezeichnet wird.
//
// **Und der dritte Lauf hält eine Messung fest.** Am 2026-09-19 kam im
// Vorschaubereich ein `change` mit `matches: false` an, während
// `matchMedia(…).matches` im selben Augenblick schon `true` sagte. Wer dem
// Ereignis glaubt, trägt eine Antwort ein, die nicht zu ihrem Stand gehört
// (§7, Regel 2) — hier steht der Fall als Lauf.
import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { schmalGrenze, useSchmalesFenster } from './useSchmalesFenster'

interface Nachbau {
  frage: string
  // stand ist das, was die Liste **jetzt** sagt — getrennt von dem, was ein
  // Ereignis behauptet. Genau darin lag der gemessene Unterschied.
  stand: boolean
  horcher: Set<() => void>
  abmeldungen: number
}

function belegeMatchMedia(anfangs: boolean): Nachbau {
  const nachbau: Nachbau = {
    frage: '',
    stand: anfangs,
    horcher: new Set(),
    abmeldungen: 0,
  }

  window.matchMedia = ((frage: string) => {
    nachbau.frage = frage

    return {
      get matches() {
        return nachbau.stand
      },
      addEventListener: (_: string, horcher: () => void) =>
        nachbau.horcher.add(horcher),
      removeEventListener: (_: string, horcher: () => void) => {
        nachbau.abmeldungen += 1
        nachbau.horcher.delete(horcher)
      },
    }
  }) as unknown as typeof window.matchMedia

  return nachbau
}

function melde(nachbau: Nachbau): void {
  act(() => {
    for (const horcher of nachbau.horcher) {
      horcher()
    }
  })
}

afterEach(() => {
  Reflect.deleteProperty(window, 'matchMedia')
  vi.restoreAllMocks()
})

describe('useSchmalesFenster', () => {
  it('fragt nach der Grenze, die es zusagt', () => {
    const nachbau = belegeMatchMedia(false)

    const { result } = renderHook(() => useSchmalesFenster())

    expect(nachbau.frage).toBe(`(max-width: ${String(schmalGrenze - 1)}px)`)
    expect(result.current).toBe(false)
  })

  it('merkt, wenn das Fenster unter die Grenze geht', () => {
    const nachbau = belegeMatchMedia(false)

    const { result } = renderHook(() => useSchmalesFenster())
    expect(result.current).toBe(false)

    nachbau.stand = true
    melde(nachbau)

    expect(result.current).toBe(true)
  })

  it('liest dabei die Liste und nicht das Ereignis', () => {
    const nachbau = belegeMatchMedia(false)

    const { result } = renderHook(() => useSchmalesFenster())

    // **Die Liste sagt schon „schmal", das Ereignis noch nicht.** Der
    // Nachbau meldet ohne Argument — ein Haken, der aus `ereignis.matches`
    // läse, bekäme hier `undefined` und bliebe auf `false` stehen.
    nachbau.stand = true
    melde(nachbau)

    expect(result.current).toBe(true)
  })

  it('folgt auch einem blanken resize ohne Medienereignis', () => {
    const nachbau = belegeMatchMedia(false)

    const { result } = renderHook(() => useSchmalesFenster())
    expect(result.current).toBe(false)

    nachbau.stand = true
    act(() => {
      window.dispatchEvent(new Event('resize'))
    })

    expect(result.current).toBe(true)
  })

  it('hört auf zu horchen, wenn es abgebaut wird', () => {
    const nachbau = belegeMatchMedia(true)

    const { result, unmount } = renderHook(() => useSchmalesFenster())
    expect(result.current).toBe(true)
    expect(nachbau.horcher.size).toBe(1)

    unmount()

    expect(nachbau.abmeldungen).toBe(1)
    expect(nachbau.horcher.size).toBe(0)

    // **Und der zweite Weg ist auch wieder zu**: ein `resize` danach bewegt
    // nichts mehr. Ohne diesen Beleg stünde nur die Abmeldung der einen
    // Seite da, und die zweite wäre ungeprüft (§10).
    nachbau.stand = false
    act(() => {
      window.dispatchEvent(new Event('resize'))
    })

    expect(result.current).toBe(true)
  })

  it('gilt ohne matchMedia als breit', () => {
    Reflect.deleteProperty(window, 'matchMedia')

    const { result } = renderHook(() => useSchmalesFenster())

    expect(result.current).toBe(false)
  })
})
