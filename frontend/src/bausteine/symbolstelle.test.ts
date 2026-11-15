import { describe, expect, it } from 'vitest'

import { symbolAnStelle } from './symbolstelle'

import type { PaketsymbolDTO } from '../api'

const symbole: PaketsymbolDTO[] = [
  {
    name: 'casting cost',
    art: 'card field',
    datei: 'game',
    zeile: 4,
    spalte: 2,
  },
  { name: 'cost', art: 'script', datei: 'game', zeile: 9, spalte: 2 },
  { name: 'title', art: 'set field', datei: 'game', zeile: 7, spalte: 2 },
]

describe('Das Symbol unter dem Zeiger', () => {
  it('findet einen Feldnamen mit Leerzeichen in der Schreibweise der Datei', () => {
    const text = 'card style:\n\tcasting cost:\n\t\tleft: 1'
    const stelle = text.indexOf('cost') + 1

    expect(symbolAnStelle(text, stelle, symbole)).toEqual({
      name: 'casting cost',
      symbol: symbole[0],
    })
  })

  it('findet ihn auch in der Schreibweise des Skripts', () => {
    const text = 'script: card.casting_cost + 1'

    expect(
      symbolAnStelle(text, text.indexOf('casting') + 3, symbole)?.name,
    ).toBe('casting cost')
  })

  it('nimmt das längste Symbol, das den Zeiger deckt', () => {
    // `cost` ist auch ein Symbol — aber innerhalb von `casting cost` meint
    // niemand das kürzere.
    const text = 'x: casting cost'

    expect(symbolAnStelle(text, text.length, symbole)?.name).toBe(
      'casting cost',
    )
  })

  it('gibt ein Wort ohne Definition ohne Symbol zurück', () => {
    expect(symbolAnStelle('a := to_upper(x)', 8, symbole)).toEqual({
      name: 'to_upper',
      symbol: undefined,
    })
  })

  it('gibt nichts, wenn der Zeiger auf keinem Wort steht', () => {
    expect(symbolAnStelle('a := 1', 3, symbole)).toBeNull()
    expect(symbolAnStelle('a := 1', -1, symbole)).toBeNull()
    expect(symbolAnStelle('a := 1', 99, symbole)).toBeNull()
  })

  it('bleibt in der Zeile des Zeigers', () => {
    const text = 'title\ncasting cost'

    expect(symbolAnStelle(text, 2, symbole)?.name).toBe('title')
    expect(symbolAnStelle(text, text.length, symbole)?.name).toBe(
      'casting cost',
    )
  })

  it('verlangt ganze Wörter', () => {
    expect(symbolAnStelle('subtitle', 4, symbole)).toEqual({
      name: 'subtitle',
      symbol: undefined,
    })
  })
})
