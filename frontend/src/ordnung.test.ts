// **Zwei Arten zu sortieren** — LOC-03.AK-01
// (#1689).

import { expect, it } from 'vitest'

import { schluesselvergleich } from './ordnung'

// **Ein Schlüsselvergleich kennt keine Sprache** — und das ist die Zusage.
//
// **Geprüft am Umlaut**: `localeCompare` ordnet »ä« im Deutschen neben »a«
// und im Schwedischen ans Ende. Ein technischer Schlüssel darf davon nichts
// wissen, sonst sähe dieselbe Liste auf zwei Rechnern verschieden aus.
it('ordnet Schlüssel zeichenweise und ohne Sprache', () => {
  expect(schluesselvergleich('a', 'b')).toBe(-1)
  expect(schluesselvergleich('b', 'a')).toBe(1)
  expect(schluesselvergleich('a', 'a')).toBe(0)

  // **»Z« vor »a«** — Großbuchstaben haben den kleineren Codepunkt. Eine
  // sprachliche Ordnung machte daraus »a« vor »Z«.
  expect(schluesselvergleich('Z', 'a')).toBe(-1)
  expect('Z'.localeCompare('a')).toBeGreaterThan(0)

  // **Und der Umlaut steht hinter dem »z«** — im Codepunkt, nicht im
  // Duden.
  expect(schluesselvergleich('z', 'ä')).toBe(-1)
})

// **Eine Liste bleibt in jeder Sprache dieselbe.**
it('gibt dieselbe Reihenfolge, gleich welche Sprache gilt', () => {
  const kennungen = ['o-9', 'o-10', 'O-1', 'ö-1', 'o-1']

  expect([...kennungen].sort(schluesselvergleich)).toEqual([
    'O-1',
    'o-1',
    'o-10',
    'o-9',
    'ö-1',
  ])
})
