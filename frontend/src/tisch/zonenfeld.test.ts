// **Wo eine Zone innerhalb eines Sitzes liegt** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import {
  Ablagebreite,
  Kartenabstand,
  Kartenbreite,
  Kartenhoehe,
  Randbreite,
  Reihenhoehe,
  Sitzbreite,
  Sitzhoehe,
} from './masse'
import { erstes } from '../test/erstes'
import { sitzflaechen } from './sitzflaeche'
import { zonenfelder } from './zonenfeld'

// **Zwei Sitze reichen für beide Fälle** — bei einem einzigen gäbe es keinen
// fremden, und der ist die Hälfte dessen, was hier geprüft wird.
const flaecheVon = (eigen: boolean) =>
  erstes(
    sitzflaechen(['du', 'ich'], 'ich').filter((eine) => eine.eigen === eigen),
  )

const felderVon = (eigen: boolean) => zonenfelder(flaecheVon(eigen))

const feld = (felder: ReturnType<typeof zonenfelder>, zone: string) => {
  const eines = felder.find((jedes) => jedes.zone === zone)
  if (!eines) throw new Error(`kein Feld für ${zone}`)
  return eines
}

describe('zonenfelder', () => {
  it('gibt jedem Sitz seine sechs Zonen', () => {
    expect(felderVon(true).map((eines) => eines.zone)).toEqual([
      'tisch',
      'hand',
      'bibliothek',
      'friedhof',
      'exil',
      'kommando',
    ])
  })

  it('legt das Schlachtfeld zur Tischmitte hin', () => {
    const eigen = felderVon(true)
    const fremd = felderVon(false)

    expect(feld(eigen, 'tisch').y).toBeLessThan(feld(eigen, 'hand').y)
    expect(feld(fremd, 'tisch').y).toBeGreaterThan(feld(fremd, 'hand').y)
  })

  it('bleibt in jedem Fall innerhalb der Sitzfläche', () => {
    for (const eigen of [true, false]) {
      const flaeche = flaecheVon(eigen)

      for (const eines of zonenfelder(flaeche)) {
        expect(eines.x).toBeGreaterThanOrEqual(flaeche.x + Randbreite)
        expect(eines.y).toBeGreaterThanOrEqual(flaeche.y + Randbreite)
        expect(eines.x + eines.breite).toBeLessThanOrEqual(
          flaeche.x + Sitzbreite - Randbreite,
        )
        expect(eines.y + eines.hoehe).toBeLessThanOrEqual(
          flaeche.y + Sitzhoehe - Randbreite,
        )
      }
    }
  })

  it('gibt dem Schlachtfeld zwei Reihen über die volle Breite', () => {
    const tisch = feld(felderVon(true), 'tisch')

    expect(tisch.hoehe).toBe(Reihenhoehe)
    expect(tisch.breite).toBe(Sitzbreite - 2 * Randbreite)
  })

  it('reiht die vier Ablagen rechts neben die Hand', () => {
    const felder = felderVon(true)
    const hand = feld(felder, 'hand')

    expect(feld(felder, 'bibliothek').x).toBe(
      hand.x + hand.breite + Kartenabstand,
    )
    for (const [stelle, zone] of [
      'bibliothek',
      'friedhof',
      'exil',
      'kommando',
    ].entries()) {
      const eines = feld(felder, zone)
      expect(eines.breite).toBe(Kartenbreite)
      expect(eines.hoehe).toBe(Kartenhoehe)
      expect(eines.y).toBe(hand.y)
      expect(eines.x).toBe(
        hand.x +
          hand.breite +
          Kartenabstand +
          stelle * (Kartenbreite + Kartenabstand),
      )
    }
  })

  it('lässt der Hand genau den Platz neben den Ablagen', () => {
    const hand = feld(felderVon(true), 'hand')

    expect(hand.breite).toBe(
      Sitzbreite - 2 * Randbreite - Ablagebreite - Kartenabstand,
    )
  })

  // **CR 400.2** (Fassung 2026-08-19): Bibliothek und Hand sind verborgen,
  // Friedhof, Schlachtfeld, Stapel, Exil und Kommando sind öffentlich.
  it('kennt die verborgenen Zonen', () => {
    const verborgen = felderVon(true)
      .filter((eines) => eines.verborgen)
      .map((eines) => eines.zone)

    expect(verborgen).toEqual(['hand', 'bibliothek'])
  })

  // **CR 400.5**: die Reihenfolge in Bibliothek, Friedhof und Stapel darf
  // nicht geändert werden — alles andere darf sein Beherrscher anordnen.
  it('kennt die geordneten Zonen', () => {
    const geordnet = felderVon(true)
      .filter((eines) => eines.geordnet)
      .map((eines) => eines.zone)

    expect(geordnet).toEqual(['bibliothek', 'friedhof'])
  })

  it('trägt überall den Sitz, zu dem es gehört', () => {
    for (const eines of felderVon(true)) expect(eines.sitz).toBe('ich')
  })
})
