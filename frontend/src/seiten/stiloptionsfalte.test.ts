// **Welche Stiloptionen vorn stehen und welche unter „Weitere“** — UXR-07
// (#1915),
// nach dem Artboard *Eigenschaften*: „4 von 11“, der Rest eingeklappt und
// nicht entfernt.
import { describe, expect, it } from 'vitest'

import { falte } from './stiloptionsfalte'

// o ist eine Option der Art `art` — für die Falte zählt nur, ob sie eine
// Zwischenüberschrift ist.
const o = (name: string, art = 'boolean') => ({ name, art })

const namen = (liste: readonly { name: string }[]) =>
  liste.map((option) => option.name)

describe('falte', () => {
  it('lässt bis zu vier Optionen ungefaltet', () => {
    const optionen = [o('a'), o('b'), o('c'), o('d')]

    expect(falte(optionen)).toEqual({
      vorn: optionen,
      weitere: [],
      gezeigt: 4,
      gesamt: 4,
    })
  })

  it('stellt die ersten vier vorn und den Rest dahinter', () => {
    const gefaltet = falte([o('a'), o('b'), o('c'), o('d'), o('e'), o('f')])

    expect(namen(gefaltet.vorn)).toEqual(['a', 'b', 'c', 'd'])
    expect(namen(gefaltet.weitere)).toEqual(['e', 'f'])
    expect(gefaltet.gezeigt).toBe(4)
    expect(gefaltet.gesamt).toBe(6)
  })

  // **Eine Zwischenüberschrift ist keine Option** (`InfoField`, editable =
  // false): sie zählt nicht mit — und sie gehört zu dem, was ihr folgt. Stünde
  // sie vorn und ihre Optionen dahinter, überschriebe sie nichts.
  it('zählt Zwischenüberschriften nicht und gibt sie dem, was ihnen folgt', () => {
    const gefaltet = falte([
      o('Grundlagen', 'info'),
      o('a'),
      o('b'),
      o('c'),
      o('d'),
      o('Erweitert', 'info'),
      o('Für Fortgeschrittene', 'info'),
      o('e'),
    ])

    expect(namen(gefaltet.vorn)).toEqual(['Grundlagen', 'a', 'b', 'c', 'd'])
    expect(namen(gefaltet.weitere)).toEqual([
      'Erweitert',
      'Für Fortgeschrittene',
      'e',
    ])
    expect(gefaltet.gezeigt).toBe(4)
    expect(gefaltet.gesamt).toBe(5)
  })

  it('zeigt eine Liste nur aus Zwischenüberschriften ganz', () => {
    const gefaltet = falte([o('Hinweis', 'info')])

    expect(namen(gefaltet.vorn)).toEqual(['Hinweis'])
    expect(gefaltet.weitere).toEqual([])
    expect(gefaltet.gesamt).toBe(0)
  })
})
