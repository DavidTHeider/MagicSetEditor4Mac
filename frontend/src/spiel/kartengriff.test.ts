// **Was ein Griff an einer Karte tut** — P4
// (#2372):
// Klick, Doppelklick, ⌘-Klick und Tastatur, und nie die Primärhandlung
// zweimal.
import { describe, expect, test } from 'vitest'

import { griffFuer, handlungZu, ohneFeld } from './kartengriff'

import type { Objekthandlung } from '../apipartie'

const tappen: Objekthandlung = {
  objekt: 'sitz1/baer/0',
  handlungen: ['objekt.tappen', 'objekt.opfern'],
  primaer: 'objekt.tappen',
  zweit: '',
}

const grundland: Objekthandlung = {
  objekt: 'sitz1/wald/0',
  handlungen: ['objekt.tappen', 'objekt.mana'],
  primaer: 'objekt.tappen',
  zweit: 'objekt.mana',
}

const ohnePrimaer: Objekthandlung = {
  objekt: 'sitz1/friedhof/0',
  handlungen: ['objekt.schieben'],
  primaer: '',
  zweit: '',
}

describe('der Griff an einer Karte', () => {
  test.each([
    [
      'ein Klick tappt',
      tappen,
      1,
      false,
      { art: 'befehl', befehl: 'objekt.tappen' },
    ],
    [
      'der zweite Klick eines Doppelklicks tappt nicht noch einmal',
      tappen,
      2,
      false,
      { art: 'nichts' },
    ],
    ['⌘-Klick wählt', tappen, 1, true, { art: 'waehlen' }],
    [
      'ohne Primärhandlung wählt ein Klick',
      ohnePrimaer,
      1,
      false,
      { art: 'waehlen' },
    ],
    [
      'mit Zweithandlung wartet der erste Klick',
      grundland,
      1,
      false,
      { art: 'warten', befehl: 'objekt.tappen' },
    ],
    [
      'und der zweite löst die Zweithandlung aus',
      grundland,
      2,
      false,
      { art: 'zweit', befehl: 'objekt.mana' },
    ],
    [
      'ein dritter Klick tut nichts mehr',
      grundland,
      3,
      false,
      { art: 'nichts' },
    ],
    [
      'die Tastatur wartet nicht',
      grundland,
      0,
      false,
      { art: 'befehl', befehl: 'objekt.tappen' },
    ],
    ['ohne Auskunft wählt ein Klick', undefined, 1, false, { art: 'waehlen' }],
    [
      'und ein Doppelklick ohne Auskunft wählt nicht zweimal',
      undefined,
      2,
      false,
      { art: 'nichts' },
    ],
  ] as const)('%s', (_name, handlung, detail, metaKey, griff) => {
    expect(griffFuer(handlung, { detail, metaKey })).toEqual(griff)
  })

  test('findet die Auskunft eines Objekts', () => {
    expect(handlungZu([tappen, grundland], 'sitz1/wald/0')).toBe(grundland)
    expect(handlungZu([tappen], 'sitz1/wald/0')).toBeUndefined()
  })

  test('kennt die Befehle ohne weiteres Feld', () => {
    expect(ohneFeld('objekt.tappen')).toBe(true)
    expect(ohneFeld('objekt.schieben')).toBe(false)
    expect(ohneFeld('gibt.esnicht')).toBe(false)
  })
})
