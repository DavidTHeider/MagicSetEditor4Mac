import { describe, expect, it } from 'vitest'

import { huenke, zeilendiff } from './diffzeilen'

describe('Der Zeilen-Diff', () => {
  it('erkennt Gleiches, Neues und Entferntes je Zeile', () => {
    const zeilen = zeilendiff('a\nb\nc\n', 'a\nx\nc\nd\n')

    expect(zeilen).toEqual([
      { art: 'gleich', text: 'a', alt: 1, neu: 1 },
      { art: 'weg', text: 'b', alt: 2 },
      { art: 'neu', text: 'x', neu: 2 },
      { art: 'gleich', text: 'c', alt: 3, neu: 3 },
      { art: 'neu', text: 'd', neu: 4 },
    ])
  })

  it('hält eine gemeinsame Teilfolge, nicht das erste Zeichen', () => {
    // Ohne die längste gemeinsame Teilfolge sähe ein Verschieben wie ein
    // Ersetzen jeder Zeile aus.
    const zeilen = zeilendiff('a\nb\nc', 'b\nc\na')

    expect(zeilen.filter((zeile) => zeile.art === 'gleich')).toHaveLength(2)
  })

  it('sieht zwei gleiche Texte als eine Folge gleicher Zeilen', () => {
    expect(
      zeilendiff('a\nb', 'a\nb').every((zeile) => zeile.art === 'gleich'),
    ).toBe(true)
    expect(zeilendiff('', '')).toEqual([
      { art: 'gleich', text: '', alt: 1, neu: 1 },
    ])
  })

  it('zählt Zeilen ab eins in beiden Ständen', () => {
    const zeilen = zeilendiff('x', 'y')

    expect(zeilen).toEqual([
      { art: 'weg', text: 'x', alt: 1 },
      { art: 'neu', text: 'y', neu: 1 },
    ])
  })
})

describe('Die Hünke', () => {
  it('zeigen nur die Umgebung einer Änderung', () => {
    const zeilen = zeilendiff('a\nb\nc\nd\ne\nf\ng', 'a\nb\nc\nd\nE\nf\ng')

    expect(huenke(zeilen, 1)).toEqual([
      [
        { art: 'gleich', text: 'd', alt: 4, neu: 4 },
        { art: 'weg', text: 'e', alt: 5 },
        { art: 'neu', text: 'E', neu: 5 },
        { art: 'gleich', text: 'f', alt: 6, neu: 6 },
      ],
    ])
  })

  it('trennen zwei weit auseinanderliegende Änderungen', () => {
    const zeilen = zeilendiff('a\nb\nc\nd\ne\nf', 'A\nb\nc\nd\ne\nF')

    expect(huenke(zeilen, 1)).toHaveLength(2)
  })

  it('sind leer, wenn nichts anders ist', () => {
    expect(huenke(zeilendiff('a', 'a'), 1)).toEqual([])
  })
})
