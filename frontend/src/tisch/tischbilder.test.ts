// **Die Tischbilder** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Der Kontrast wird gerechnet, nicht aufgeschrieben
//
// **Sechs Sätze von Hand zu messen veraltet beim siebten.** `docs/KONTRASTE.md`
// führt die Paare des Grundsatzes, weil sie in `stil.css` stehen und ein Auge
// sie prüfen muss; die fünf weiteren Sätze sind Zahlen im Quelltext, und
// Zahlen im Quelltext prüft ein Lauf.
//
// **Gerechnet wird nach WCAG 2.2**: sRGB-Linearisierung mit der Schwelle
// 0,04045, Gewichte 0,2126 / 0,7152 / 0,0722, Kontrast als
// (L₁ + 0,05) / (L₂ + 0,05). Dieselbe Rechnung wie in `scripts/kontrastlauf.mjs`
// — hier ohne Canvas, weil hier keine `oklch()` vorkommen, sondern Zahlen.
import { describe, expect, it } from 'vitest'

import { Tischersatz } from './tischfarben'
import { tischbild, tischbilder } from './tischbilder'

const linear = (kanal: number) =>
  kanal <= 0.04045 ? kanal / 12.92 : ((kanal + 0.055) / 1.055) ** 2.4

const leuchte = (farbe: number) => {
  const [r, g, b] = [16, 8, 0].map((schub) => ((farbe >> schub) & 0xff) / 255)

  return (
    0.2126 * linear(r ?? 0) + 0.7152 * linear(g ?? 0) + 0.0722 * linear(b ?? 0)
  )
}

const kontrast = (eine: number, andere: number) => {
  const [hell, dunkel] = [leuchte(eine), leuchte(andere)].sort((a, b) => b - a)

  return ((hell ?? 0) + 0.05) / ((dunkel ?? 0) + 0.05)
}

describe('die Rechnung selbst', () => {
  // **Erst der Maßstab, dann die Messung** (§10): eine Kontrastrechnung, die
  // nie gegen einen bekannten Wert gehalten wurde, misst vielleicht etwas
  // ganz anderes.
  it('trifft die bekannten Eckwerte', () => {
    expect(kontrast(0xffffff, 0x000000)).toBeCloseTo(21, 6)
    expect(kontrast(0x000000, 0x000000)).toBeCloseTo(1, 6)
    expect(kontrast(0x777777, 0xffffff)).toBeCloseTo(4.48, 2)
  })
})

describe('tischbilder', () => {
  it('sind die sechs, die das Issue vorschlägt', () => {
    expect(tischbilder.map((eines) => eines.name)).toEqual([
      'holz',
      'dunkles-holz',
      'picknick',
      'dungeon',
      'taverne',
      'turniertisch',
    ])
  })

  it('fangen mit dem Satz an, der in stil.css steht', () => {
    const [erstes] = tischbilder

    expect(erstes.hell.grund).toBe(Tischersatz.grund)
    expect(erstes.hell.sitz).toBe(Tischersatz.sitz)
    expect(erstes.hell.eigen).toBe(Tischersatz.eigenerSitz)
  })

  // **WCAG 1.4.11**: ein grafisches Objekt braucht 3:1 gegen die Farben
  // daneben. Die Kante trennt jede Fläche von jeder — sie muss es gegen
  // **jeden** Satz halten, sonst wäre ein Tischbild eine Falle.
  it.each(tischbilder)(
    '$name hält die Kante gegen jede seiner Flächen',
    (bild) => {
      for (const flaeche of [bild.hell.grund, bild.hell.sitz, bild.hell.eigen])
        expect(kontrast(Tischersatz.kante, flaeche)).toBeGreaterThanOrEqual(3)
    },
  )

  // **Die dunkle Kante ist eine andere** — sie steht in `stil.css` unter
  // `prefers-color-scheme: dark` und heißt dort `#9b8768`.
  const dunkleKante = 0x9b8768

  it.each(tischbilder)(
    '$name hält die dunkle Kante gegen jede seiner Flächen',
    (bild) => {
      for (const flaeche of [
        bild.dunkel.grund,
        bild.dunkel.sitz,
        bild.dunkel.eigen,
      ])
        expect(kontrast(dunkleKante, flaeche)).toBeGreaterThanOrEqual(3)
    },
  )

  // **Der Auswahlrahmen muss sich von der Kante unterscheiden**, und das ist
  // die Prüfung, auf die es ankommt: eine Auswahl, die aussieht wie ein
  // gewöhnlicher Kartenrand, sagt nichts.
  //
  // **Der erste Anlauf prüfte 4,5:1 gegen die Flächen** — die Textschwelle,
  // und für ein grafisches Objekt zu streng (WCAG 1.4.11 sagt 3:1). Gegen die
  // **Kante** stand er dabei bei **1,32:1**, und das ist der Wert, der zählt.
  // Eine strengere Zahl an der falschen Stelle ersetzt keine Prüfung an der
  // richtigen (§10).
  const dunkleAuswahl = 0xfbf0dc

  it('hebt sich von der Kante ab, in beiden Erscheinungsbildern', () => {
    expect(
      kontrast(Tischersatz.auswahl, Tischersatz.kante),
    ).toBeGreaterThanOrEqual(3)
    expect(kontrast(dunkleAuswahl, dunkleKante)).toBeGreaterThanOrEqual(3)
  })

  it.each(tischbilder)('$name trägt den Auswahlrahmen', (bild) => {
    for (const flaeche of [bild.hell.grund, bild.hell.sitz, bild.hell.eigen])
      expect(kontrast(Tischersatz.auswahl, flaeche)).toBeGreaterThanOrEqual(3)

    for (const flaeche of [
      bild.dunkel.grund,
      bild.dunkel.sitz,
      bild.dunkel.eigen,
    ])
      expect(kontrast(dunkleAuswahl, flaeche)).toBeGreaterThanOrEqual(3)
  })
})

describe('tischbild', () => {
  it('findet einen Satz an seinem Namen', () => {
    expect(tischbild('dungeon').name).toBe('dungeon')
  })

  it('fällt auf Holz zurück, wo der Name nichts trifft', () => {
    // **Ein gespeicherter Name, den es nicht mehr gibt**, ließe den Tisch
    // sonst ohne Farbe dastehen.
    expect(tischbild('gibtesnicht').name).toBe('holz')
    expect(tischbild('').name).toBe('holz')
  })
})
