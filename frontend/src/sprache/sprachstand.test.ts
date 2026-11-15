// **Der Satz unter der Sprachwahl und das Register sagen dasselbe** —
// LOC-01.AK-03 (#1687).
//
// Der Hinweis in den Einstellungen sagt: Deutsch ist die Ausgangssprache,
// alle anderen sind maschinell übersetzt. Wird eine Sprache geprüft oder
// abgenommen, ist der Satz falsch — und dieser Test rot, bis er es nicht mehr
// ist.
import { describe, expect, it } from 'vitest'

import { sprachen, sprachstaende } from './kataloge'

describe('Der Stand der Sprachen', () => {
  it('stimmt mit dem Satz unter der Sprachwahl überein', () => {
    expect(sprachstaende.de).toBe('ausgangssprache')

    for (const sprache of sprachen.filter((s) => s !== 'de')) {
      expect(sprachstaende[sprache], sprache).toBe('maschinell')
    }
  })
})
