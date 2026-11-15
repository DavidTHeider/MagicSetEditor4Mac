// **Was ⌘C, ⌘X und ⌘V über Karten bedeuten** — Werkstück 1 von Phase 3.7
// (#1168).
import { describe, expect, it } from 'vitest'

import { ablagetaste, type Ablagetat } from './ablagetasten'

describe('ablagetaste', () => {
  it.each<[string, boolean, boolean, Ablagetat | null]>([
    ['c', true, false, 'kopieren'],
    ['x', true, false, 'ausschneiden'],
    ['v', true, false, 'einfuegen'],

    // **Großgeschrieben ist dieselbe Taste** — ⇧ ändert nichts daran, welche
    // gedrückt wurde.
    ['C', true, false, 'kopieren'],

    // **Ohne ⌘ ist es ein Buchstabe.**
    ['c', false, false, null],

    // **Im Textfeld gehört ⌘C dem Textfeld** — dieselbe Regel wie in MSE2,
    // wo die drei am fokussierten Bedienelement hängen.
    ['c', true, true, null],
    ['x', true, true, null],
    ['v', true, true, null],

    // **Und was sonst noch mit ⌘ kommt, geht vorbei.**
    ['a', true, false, null],
    ['z', true, false, null],
  ])('⌘%s (Feld: %s) → %s', (taste, meta, imEingabefeld, erwartet) => {
    expect(ablagetaste({ taste, meta, imEingabefeld })).toBe(erwartet)
  })
})

// **Ein Leser kopiert, mehr nicht** — #2288: ⌘X und ⌘V schreiben ins Set.
describe('ablagetaste für einen Leser', () => {
  it.each<[string, Ablagetat | null]>([
    ['c', 'kopieren'],
    ['x', null],
    ['v', null],
  ])('⌘%s → %s', (taste, erwartet) => {
    expect(
      ablagetaste({ taste, meta: true, imEingabefeld: false, nurLesend: true }),
    ).toBe(erwartet)
  })
})
