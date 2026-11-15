// **Pakete in MSE2s Ordnung** — `position hint:`, dann der Name (UX-05,
// #1677).
import { describe, expect, it } from 'vitest'

import { nachRang } from './paketordnung'

describe('nachRang', () => {
  it('ordnet erst nach Rang, dann nach dem Namen', () => {
    const pakete = [
      { name: 'magic-custom', rang: 10 },
      { name: 'magic-DayNight', rang: 34 },
      { name: 'magic-m15-altered', rang: 1 },
      { name: 'magic-Zeta', rang: 10 },
    ]

    // **Der Name nach Codeeinheiten**: das große Z steht vor dem kleinen c.
    expect([...pakete].sort(nachRang).map((eines) => eines.name)).toEqual([
      'magic-m15-altered',
      'magic-Zeta',
      'magic-custom',
      'magic-DayNight',
    ])
  })

  it('sagt null, wo nichts entscheidet', () => {
    expect(
      nachRang({ name: 'probe', rang: 5 }, { name: 'probe', rang: 5 }),
    ).toBe(0)
  })
})
