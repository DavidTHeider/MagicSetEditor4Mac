// **Was sich seit dem letzten Bild geändert hat** — Werkstück G3 von Phase
// 5.5 (#1460).
//
// # Warum das hier steht und nicht in der Naht
//
// **Weil es eine Rechnung ist.** #1460 verlangt „langlebige BoardScene statt
// Neuanlage bei jedem React-Render" — was langlebig bleibt, muss abgeglichen
// werden, und ein Abgleich, der in der Grafiknaht steckte, wäre der eine
// Teil des Tisches, den kein Test je betritt.
import { describe, expect, it } from 'vitest'

import { abgeglichen } from './szenenabgleich'

const eines = (kennung: string) => ({ kennung })

describe('abgeglichen', () => {
  it('nennt alles neu, was noch nie da war', () => {
    const lauf = abgeglichen(new Set<string>(), [eines('a'), eines('b')])

    expect(lauf.neu.map((jedes) => jedes.kennung)).toEqual(['a', 'b'])
    expect(lauf.bleibt).toEqual([])
    expect(lauf.weg).toEqual([])
  })

  it('trennt das Bleibende vom Neuen', () => {
    const lauf = abgeglichen(new Set(['a']), [eines('a'), eines('b')])

    expect(lauf.bleibt.map((jedes) => jedes.kennung)).toEqual(['a'])
    expect(lauf.neu.map((jedes) => jedes.kennung)).toEqual(['b'])
  })

  it('nennt, was verschwunden ist', () => {
    const lauf = abgeglichen(new Set(['a', 'b']), [eines('b')])

    expect(lauf.weg).toEqual(['a'])
  })

  it('hält die Reihenfolge des neuen Standes', () => {
    const lauf = abgeglichen(new Set(['b']), [
      eines('c'),
      eines('b'),
      eines('a'),
    ])

    expect(
      [...lauf.neu, ...lauf.bleibt].map((jedes) => jedes.kennung).sort(),
    ).toEqual(['a', 'b', 'c'])
    expect(lauf.neu.map((jedes) => jedes.kennung)).toEqual(['c', 'a'])
  })

  it('räumt einen ganzen Stand ab', () => {
    const lauf = abgeglichen(new Set(['a', 'b']), [])

    expect(lauf.weg.sort()).toEqual(['a', 'b'])
    expect(lauf.neu).toEqual([])
  })

  it('bleibt ruhig, wenn sich nichts geändert hat', () => {
    const lauf = abgeglichen(new Set(['a']), [eines('a')])

    expect(lauf).toEqual({ neu: [], bleibt: [eines('a')], weg: [] })
  })
})
