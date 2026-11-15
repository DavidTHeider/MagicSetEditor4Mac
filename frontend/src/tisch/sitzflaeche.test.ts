// **Wo ein Sitz am Tisch liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Nachgelesen, nicht entworfen (§11, §15)
//
// **CR 403.1** (Fassung 2026-08-19) sagt zweierlei über die Anordnung: das
// Schlachtfeld ist die Fläche *zwischen* den Spielern, und die Permanenten
// eines Spielers liegen *vor ihm*. Daraus folgt beides, was unten geprüft
// wird — dass jeder Sitz eine eigene Fläche hat und dass die Flächen
// einander gegenüberliegen.
//
// **Was daraus nicht folgt, ist der Bezugssitz.** Wer unten liegt, sagt keine
// Regel; es ist eine Entscheidung dieser Oberfläche, und sie steht in
// `sitzflaeche.ts` mit ihrem Grund.
import { describe, expect, it } from 'vitest'

import { Sitzabstand, Sitzbreite, Sitzhoehe } from './masse'
import { sitzflaechen } from './sitzflaeche'

const flaeche = (flaechen: ReturnType<typeof sitzflaechen>, sitz: string) => {
  const eine = flaechen.find((jede) => jede.sitz === sitz)
  if (!eine) throw new Error(`keine Fläche für ${sitz}`)
  return eine
}

const ueberschneidet = (
  a: ReturnType<typeof sitzflaechen>[number],
  b: ReturnType<typeof sitzflaechen>[number],
) =>
  a.x < b.x + b.breite &&
  b.x < a.x + a.breite &&
  a.y < b.y + b.hoehe &&
  b.y < a.y + a.hoehe

describe('sitzflaechen', () => {
  it('legt einen einzelnen Sitz in den Ursprung', () => {
    const allein = flaeche(sitzflaechen(['ich'], 'ich'), 'ich')

    expect(allein).toMatchObject({ sitz: 'ich', eigen: true, x: 0, y: 0 })
    expect(allein.breite).toBe(Sitzbreite)
    expect(allein.hoehe).toBe(Sitzhoehe)
  })

  it('legt den eigenen Sitz unter den fremden', () => {
    const zwei = sitzflaechen(['du', 'ich'], 'ich')

    expect(flaeche(zwei, 'du').y).toBe(0)
    expect(flaeche(zwei, 'ich').y).toBe(Sitzhoehe + Sitzabstand)
  })

  it('legt den eigenen Sitz auch dann unten hin, wenn er zuerst genannt ist', () => {
    const zwei = sitzflaechen(['ich', 'du'], 'ich')

    expect(flaeche(zwei, 'ich').y).toBeGreaterThan(flaeche(zwei, 'du').y)
  })

  it('reiht die fremden Sitze nebeneinander', () => {
    const vier = sitzflaechen(['a', 'b', 'c', 'ich'], 'ich')

    expect(flaeche(vier, 'a').x).toBe(0)
    expect(flaeche(vier, 'b').x).toBe(Sitzbreite + Sitzabstand)
    expect(flaeche(vier, 'c').x).toBe(2 * (Sitzbreite + Sitzabstand))
  })

  it('setzt den eigenen Sitz mittig unter die fremden', () => {
    const vier = sitzflaechen(['a', 'b', 'c', 'ich'], 'ich')
    const oben = flaeche(vier, 'a').x + flaeche(vier, 'c').x + Sitzbreite
    const eigen = flaeche(vier, 'ich')

    expect(eigen.x + eigen.breite / 2).toBeCloseTo(oben / 2, 6)
  })

  it('lässt keine Fläche links oder oben aus dem Tisch ragen', () => {
    for (const zahl of [1, 2, 3, 4]) {
      const sitze = ['a', 'b', 'c', 'd'].slice(0, zahl)
      for (const eine of sitzflaechen(sitze, sitze[zahl - 1] ?? '')) {
        expect(eine.x).toBeGreaterThanOrEqual(0)
        expect(eine.y).toBeGreaterThanOrEqual(0)
      }
    }
  })

  it('lässt zwei Flächen einander nie überdecken', () => {
    const vier = sitzflaechen(['a', 'b', 'c', 'ich'], 'ich')

    for (const eine of vier)
      for (const andere of vier)
        if (eine !== andere) expect(ueberschneidet(eine, andere)).toBe(false)
  })

  it('dreht die fremden Sitze zur Mitte', () => {
    const zwei = sitzflaechen(['du', 'ich'], 'ich')

    expect(flaeche(zwei, 'du').zurMitte).toBe(1)
    expect(flaeche(zwei, 'ich').zurMitte).toBe(-1)
  })

  it('nimmt den ersten Sitz als Bezug, wenn keiner der eigene ist', () => {
    const zuschauer = sitzflaechen(['a', 'b'], '')

    expect(flaeche(zuschauer, 'a').eigen).toBe(false)
    expect(flaeche(zuschauer, 'a').y).toBeGreaterThan(flaeche(zuschauer, 'b').y)
  })

  it('gibt für eine Partie ohne Sitze nichts zurück', () => {
    expect(sitzflaechen([], 'ich')).toEqual([])
  })
})
