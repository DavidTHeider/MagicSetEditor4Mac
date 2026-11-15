// **Was in der Deckwahl eines Sitzes steht** — ein Deck, eine angenommene
// Leihe (P1b, #2370) oder eine Einladung (P9c, #2377). Ein Wert, ein Sitz:
// was gewählt ist, schließt die anderen beiden aus.
import { describe, expect, it } from 'vitest'

import { deckwert, einladewert, gewaehlterSitz, leihwert } from './sitzwahl'

const leer = { name: '', deckID: '', profilID: '', leiheID: '' }
const kontakte = [{ kennung: 'p-bo', name: 'Bo' }]

describe('die Deckwahl eines Sitzes', () => {
  it('liest Deck, Leihe und Einladung zurück', () => {
    expect(deckwert({ ...leer, deckID: '7' })).toBe('7')
    expect(deckwert({ ...leer, deckID: '7', eingeladen: '' })).toBe('7')
    expect(deckwert({ ...leer, leiheID: 'l1' })).toBe(leihwert('l1'))
    expect(deckwert({ ...leer, eingeladen: 'p-bo' })).toBe(einladewert('p-bo'))
  })

  it('setzt mit einer Einladung den Namen und räumt Deck und Leihe', () => {
    expect(
      gewaehlterSitz(
        { ...leer, deckID: '7', leiheID: 'l1' },
        einladewert('p-bo'),
        kontakte,
      ),
    ).toEqual({ ...leer, name: 'Bo', eingeladen: 'p-bo' })
  })

  it('behält den Namen, wenn der Kontakt fehlt', () => {
    expect(
      gewaehlterSitz({ ...leer, name: 'Ada' }, einladewert('p-xy'), kontakte),
    ).toEqual({
      ...leer,
      name: 'Ada',
      eingeladen: 'p-xy',
    })
  })

  it('räumt die Einladung mit einer Leihe oder einem Deck', () => {
    const eingeladen = { ...leer, eingeladen: 'p-bo' }

    expect(gewaehlterSitz(eingeladen, leihwert('l1'), kontakte)).toEqual({
      ...leer,
      leiheID: 'l1',
      eingeladen: '',
    })
    expect(gewaehlterSitz(eingeladen, '7', kontakte)).toEqual({
      ...leer,
      deckID: '7',
      eingeladen: '',
    })
  })
})
