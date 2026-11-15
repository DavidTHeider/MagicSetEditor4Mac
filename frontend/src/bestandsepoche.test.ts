// **Die Epoche des Bestands, die das Fenster zuletzt erfuhr** — J05
// (#2269),
// Befund 2 des Nachlaufs: sie kam nur aus einer geglückten Antwort an; ein
// gescheiterter Tausch und ein Ereignis mit höherer Epoche blieben ungehört.
import { afterEach, expect, it } from 'vitest'

import {
  epocheBekannt,
  epocheGesehen,
  epocheVeraltet,
  epocheVergessen,
  kennungGeloescht,
  kennungIstGeloescht,
} from './bestandsepoche'

afterEach(() => {
  epocheVergessen()
})

it('hält eine ältere Epoche für veraltet und eine gleiche nicht', () => {
  epocheGesehen(3)

  expect(epocheVeraltet(2)).toBe(true)
  expect(epocheVeraltet(3)).toBe(false)
  expect(epocheVeraltet(4)).toBe(false)
})

it('vergisst eine höhere nicht, wenn danach eine niedrigere kommt', () => {
  epocheGesehen(5)
  epocheGesehen(2)

  expect(epocheVeraltet(4)).toBe(true)
})

it('hält ohne Angabe nichts für veraltet und merkt sich nichts', () => {
  epocheGesehen(undefined)

  expect(epocheVeraltet(undefined)).toBe(false)
  expect(epocheVeraltet(0)).toBe(false)
})

// **Ein gelöschtes Set kommt nicht wieder** (#2306): `sets.id` trägt
// `AUTOINCREMENT`, eine Kennung kehrt in **einem** Bestand nie zurück — in
// einem neuen schon.
it('merkt sich ein gelöschtes Set bis zum nächsten Bestand', () => {
  epocheGesehen(1)
  kennungGeloescht('5', 1)

  expect(kennungIstGeloescht('5')).toBe(true)
  expect(kennungIstGeloescht('6')).toBe(false)

  epocheGesehen(1)
  expect(kennungIstGeloescht('5')).toBe(true)

  epocheGesehen(2)
  expect(kennungIstGeloescht('5')).toBe(false)
})

it('vergisst gelöschte Sets mit der Epoche', () => {
  kennungGeloescht('5', 0)
  epocheVergessen()

  expect(kennungIstGeloescht('5')).toBe(false)
})

it('nennt die bekannte Epoche', () => {
  expect(epocheBekannt()).toBe(0)

  epocheGesehen(4)
  expect(epocheBekannt()).toBe(4)
})

it('merkt sich kein Set aus einer Epoche, die nicht mehr gilt', () => {
  epocheGesehen(2)
  kennungGeloescht('5', 1)

  expect(kennungIstGeloescht('5')).toBe(false)
})
