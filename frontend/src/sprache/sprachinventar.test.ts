// **Das Sprachinventar und das Register sagen dasselbe** — H04
// (#2258).
//
// `docs/SPRACHINVENTAR.md` führt die Fenstersprachen mit Stand, Herkunft und
// Redaktion; das Fenster zeigt sie aus `sprachstaende` und `sprachherkunft`.
// Zwei Stellen laufen auseinander, sobald eine Sprache geprüft wird — dieser
// Lauf hält sie zusammen, in beide Richtungen.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { sprachen, sprachherkunft, sprachstaende } from './kataloge'

const inventar = readFileSync(
  join(import.meta.dirname, '../../../docs/SPRACHINVENTAR.md'),
  'utf8',
)

// Die Zeilen der Tafel „Die Fenstersprachen": Kürzel, BCP 47, Stand,
// Herkunft, übersetzt, redaktionell geprüft, Paket.
const zeilen = new Map(
  [
    ...inventar.matchAll(
      /^\| `(\w+)` \| `[\w-]+` \| ([^|]+) \| ([^|]+) \| ([^|]+) \| ([^|]+) \| `[\w-]+` \|$/gm,
    ),
  ].map((treffer) => [
    treffer[1],
    treffer.slice(2).map((teil) => teil.trim().replaceAll('*', '')),
  ]),
)

const standworte = {
  ausgangssprache: 'Ausgangssprache',
  maschinell: 'maschinell übersetzt',
  geprueft: 'redaktionell geprüft',
  abgenommen: 'abgenommen',
} as const

const herkunftsworte = {
  ausgangssprache: 'Ausgangssprache',
  mitDemCode: 'mit dem Code',
  sprachmodell: 'Sprachmodell',
} as const

describe('Das Sprachinventar', () => {
  it('führt genau die Fenstersprachen', () => {
    expect([...zeilen.keys()].sort()).toEqual([...sprachen].sort())
  })

  it.each(sprachen)('sagt für %s dasselbe wie das Register', (sprache) => {
    const [stand, herkunft, uebersetzt, geprueft] = zeilen.get(sprache) ?? []
    const register = sprachherkunft[sprache]

    expect(stand).toBe(standworte[sprachstaende[sprache]])
    expect(herkunft).toBe(herkunftsworte[register.herkunft])
    expect(uebersetzt).toBe(
      register.uebersetzt === '' ? '—' : register.uebersetzt,
    )
    expect(geprueft).toBe(register.geprueft === '' ? '—' : register.geprueft)
  })
})
