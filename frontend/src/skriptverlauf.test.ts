// **Der Verlauf der Skriptkonsole** — Werkstück U13 von Phase 5.1 (#1424).
//
// Je Sitzung, begrenzt, löschbar, und ohne Geheimnisse: ein Ausdruck, der
// nach einem Token oder Kennwort klingt, steht als »•••« darin.
import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it } from 'vitest'

import {
  hoechstens,
  leere,
  merke,
  redigiert,
  useSkriptverlauf,
} from './skriptverlauf'

beforeEach(() => {
  leere()
})

describe('Der Skriptverlauf', () => {
  it('hängt den jüngsten Lauf vorn an und zählt durch', () => {
    merke({
      ausdruck: '1 + 1',
      ausgang: 'gelungen',
      dauerMs: 3,
      zeitpunkt: 'a',
    })
    merke({
      ausdruck: 'card.name',
      ausgang: 'gescheitert',
      dauerMs: 0,
      zeitpunkt: 'b',
    })

    const { result } = renderHook(() => useSkriptverlauf())

    expect(result.current.map((e) => e.ausdruck)).toEqual([
      'card.name',
      '1 + 1',
    ])
    expect(result.current[0]?.nummer).toBeGreaterThan(
      result.current[1]?.nummer ?? 0,
    )
  })

  it('behält nur die jüngsten Einträge', () => {
    for (let i = 0; i < hoechstens + 5; i += 1) {
      merke({
        ausdruck: String(i),
        ausgang: 'gelungen',
        dauerMs: 0,
        zeitpunkt: 'z',
      })
    }

    const { result } = renderHook(() => useSkriptverlauf())

    expect(result.current).toHaveLength(hoechstens)
    expect(result.current[0]?.ausdruck).toBe(String(hoechstens + 4))
  })

  it('meldet jede Änderung an den Haken und lässt sich leeren', () => {
    const { result } = renderHook(() => useSkriptverlauf())
    expect(result.current).toEqual([])

    act(() => {
      merke({
        ausdruck: 'x',
        ausgang: 'abgebrochen',
        dauerMs: 9,
        zeitpunkt: 'z',
      })
    })
    expect(result.current).toHaveLength(1)

    act(() => {
      leere()
    })
    expect(result.current).toEqual([])
  })

  // **Geheimnisse stehen nicht im Verlauf.** Ein Ausdruck, der nach einem
  // Token, Kennwort oder Geheimnis klingt, wird als Ganzes redigiert — die
  // Konsole weiß nicht, welcher Teil das Geheimnis ist.
  it.each([
    ['set.token', '•••'],
    ['"mein Kennwort"', '•••'],
    ['password_of(x)', '•••'],
    ['geheimnis', '•••'],
    ['api_secret', '•••'],
    ['card.name', 'card.name'],
    ['1 + 1', '1 + 1'],
  ])('redigiert %s zu %s', (ausdruck, erwartet) => {
    expect(redigiert(ausdruck)).toBe(erwartet)
  })
})
