// **Wo ein Objekt innerhalb seiner Zone liegt** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import { Kartenabstand, Kartenbreite, Kartenhoehe } from './masse'
import { an, erstes } from '../test/erstes'
import { sitzflaechen } from './sitzflaeche'
import { zonenfelder } from './zonenfeld'
import {
  Haufenversatz,
  Stapelsicht,
  Tappdrehung,
  zonenlagen,
} from './zonenlage'

import type { Gelegtes } from './zonenlage'

const feld = (zone: string) => {
  const flaeche = erstes(sitzflaechen(['ich'], 'ich'))
  const eines = zonenfelder(flaeche).find((jedes) => jedes.zone === zone)
  if (!eines) throw new Error(`kein Feld für ${zone}`)
  return eines
}

const viele = (zahl: number, dazu: Partial<Gelegtes> = {}): Gelegtes[] =>
  Array.from({ length: zahl }, (_, stelle) => ({
    kennung: `k${String(stelle)}`,
    getappt: false,
    verdeckt: false,
    haufen: '',
    ...dazu,
  }))

describe('zonenlagen', () => {
  it('legt wenige Permanente nebeneinander mit vollem Abstand', () => {
    const { lagen } = zonenlagen(feld('tisch'), viele(3))

    expect(an(lagen, 1).x - an(lagen, 0).x).toBe(Kartenbreite + Kartenabstand)
    expect(an(lagen, 0).y).toBe(an(lagen, 2).y)
  })

  it('bricht das Schlachtfeld in die zweite Reihe um', () => {
    const { lagen } = zonenlagen(feld('tisch'), viele(20))

    expect(an(lagen, 0).y).toBeLessThan(an(lagen, 19).y)
    expect(an(lagen, 19).y - an(lagen, 0).y).toBe(Kartenhoehe + Kartenabstand)
  })

  it('setzt zu viele Permanente enger statt über den Rand', () => {
    const feldchen = feld('tisch')
    const { lagen } = zonenlagen(feldchen, viele(60))

    for (const eine of lagen)
      expect(eine.x + eine.breite).toBeLessThanOrEqual(
        feldchen.x + feldchen.breite + 0.000001,
      )
  })

  it('dreht eine getappte Permanente seitwärts', () => {
    // **CR 701.26a** (Fassung 2026-08-19): tappen heißt, sie aus der
    // aufrechten Lage seitwärts zu drehen.
    const { lagen } = zonenlagen(feld('tisch'), viele(1, { getappt: true }))

    expect(an(lagen, 0).drehung).toBe(Tappdrehung)
    expect(an(zonenlagen(feld('tisch'), viele(1)).lagen, 0).drehung).toBe(0)
  })

  it('fächert die Hand in einer einzigen Reihe auf', () => {
    const { lagen } = zonenlagen(feld('hand'), viele(7))

    for (const eine of lagen) expect(eine.y).toBe(an(lagen, 0).y)
  })

  it('legt einen Stapel übereinander und die oberste zuoberst', () => {
    const feldchen = feld('bibliothek')
    const { lagen } = zonenlagen(feldchen, viele(4, { verdeckt: true }))
    const oberste = an(lagen, lagen.length - 1)

    expect(oberste.x).toBe(feldchen.x)
    expect(oberste.y).toBe(feldchen.y)
    expect(oberste.hoehenrang).toBeGreaterThan(an(lagen, 0).hoehenrang)
    expect(an(lagen, 0).x).toBeGreaterThan(oberste.x)
  })

  it('gibt einem hohen Stapel nur die obersten Lagen und zählt den Rest', () => {
    const { lagen, zahl } = zonenlagen(feld('bibliothek'), viele(60))

    expect(zahl).toBe(60)
    expect(lagen).toHaveLength(Stapelsicht)
    expect(an(lagen, lagen.length - 1).kennung).toBe('k59')
  })

  it('trägt weiter, dass etwas verdeckt liegt', () => {
    const { lagen } = zonenlagen(feld('hand'), viele(2, { verdeckt: true }))

    expect(lagen.every((eine) => eine.verdeckt)).toBe(true)
  })

  // **CR 110.1** (Fassung 2026-08-19): nur was auf dem Schlachtfeld liegt,
  // ist eine Permanente — und nur eine Permanente kann getappt sein
  // (CR 701.26a). Ein Stapel dreht deshalb nichts, auch wenn die Eingabe es
  // behauptet.
  it('dreht in einem Stapel nichts, auch nicht auf Zuruf', () => {
    const { lagen } = zonenlagen(feld('friedhof'), viele(3, { getappt: true }))

    expect(lagen.every((eine) => eine.drehung === 0)).toBe(true)
  })

  it('kommt mit einer leeren Zone zurecht', () => {
    for (const zone of ['tisch', 'hand', 'bibliothek'])
      expect(zonenlagen(feld(zone), [])).toEqual({ lagen: [], zahl: 0 })
  })

  it('legt eine einzelne Karte an den Anfang ihres Feldes', () => {
    const feldchen = feld('hand')
    const { lagen } = zonenlagen(feldchen, viele(1))

    expect(an(lagen, 0)).toMatchObject({
      x: feldchen.x,
      y: feldchen.y,
      breite: Kartenbreite,
      hoehe: Kartenhoehe,
      hoehenrang: 0,
    })
  })
})

// **Ein Haufen belegt einen Platz** — CR 400.5, PLAY-03.AK-03
// (#1756).
//
// # Was er einlösen muss
//
// **AK-03**: „Visuelle Zusammenfassung identischer Karten erhält individuelle
// Objekt-IDs, Zustände und Auflösbarkeit." Jede Karte behält also ihre Lage
// und ihre Kennung — sie liegen nur übereinander statt nebeneinander.
describe('ein Haufen auf dem Schlachtfeld', () => {
  // siebenLaender sind sieben Karten, von denen fünf auf einem Haufen liegen.
  const siebenLaender = (): Gelegtes[] => [
    ...viele(5, { haufen: 'k0' }).map((eines, stelle) => ({
      ...eines,
      kennung: `h${String(stelle)}`,
    })),
    ...viele(2).map((eines, stelle) => ({
      ...eines,
      kennung: `frei${String(stelle)}`,
    })),
  ]

  it('gibt jeder Karte des Haufens eine eigene Lage', () => {
    const { lagen, zahl } = zonenlagen(feld('tisch'), siebenLaender())

    expect(lagen).toHaveLength(7)
    expect(zahl).toBe(7)
    expect(new Set(lagen.map((lage) => lage.kennung)).size).toBe(7)
  })

  // **Ein Haufen belegt einen Platz und nicht fünf.** Ohne das wäre er
  // keine Zusammenfassung, sondern fünf Karten nebeneinander mit einem
  // gemeinsamen Namen.
  //
  // **Und der Platz ist so breit, wie der Haufen zu sehen ist** — die
  // Karte und viermal der Versatz. Hier stand bis zum 2026-09-24 genau ein
  // Schritt, und die oberste Karte ragte 10 mm in den nächsten Platz;
  // dieselbe Gestalt wie bei den getappten Karten (`querlage.test.ts`).
  it('rückt die zwei freien Karten auf die Plätze zwei und drei', () => {
    const { lagen } = zonenlagen(feld('tisch'), siebenLaender())

    const haufenbreite = Kartenbreite + 4 * Haufenversatz
    const haufen = lagen.filter((lage) => lage.kennung.startsWith('h'))
    const frei = lagen.filter((lage) => lage.kennung.startsWith('frei'))

    expect(an(frei, 0).x - an(haufen, 0).x).toBe(haufenbreite + Kartenabstand)
    expect(an(frei, 1).x - an(frei, 0).x).toBe(Kartenbreite + Kartenabstand)
  })

  // **Innerhalb des Haufens schaut jede Karte hervor** — sonst wäre die
  // unterste nicht auflösbar, und AK-03 verlangt genau das.
  it('versetzt die Karten des Haufens gegeneinander', () => {
    const { lagen } = zonenlagen(feld('tisch'), siebenLaender())

    const haufen = lagen.filter((lage) => lage.kennung.startsWith('h'))

    expect(an(haufen, 1).x - an(haufen, 0).x).toBeGreaterThan(0)
    expect(an(haufen, 1).x - an(haufen, 0).x).toBeLessThan(Kartenbreite)
    expect(an(haufen, 1).y - an(haufen, 0).y).toBeGreaterThan(0)
  })

  // **Und die Reihenfolge des Haufens steht fest**: wer zuerst kam, liegt
  // unten. Ein Zeichner, der nach Zufall entschiede, ließe den Haufen bei
  // jedem Bild anders aussehen.
  it('legt die späteren Karten des Haufens obenauf', () => {
    const { lagen } = zonenlagen(feld('tisch'), siebenLaender())

    const haufen = lagen.filter((lage) => lage.kennung.startsWith('h'))

    expect(an(haufen, 1).hoehenrang).toBeGreaterThan(an(haufen, 0).hoehenrang)
  })

  // **Zwei Haufen sind zwei Plätze.**
  it('hält zwei Haufen auseinander', () => {
    const gelegte: Gelegtes[] = [
      ...viele(2, { haufen: 'a' }).map((eines, stelle) => ({
        ...eines,
        kennung: `a${String(stelle)}`,
      })),
      ...viele(2, { haufen: 'b' }).map((eines, stelle) => ({
        ...eines,
        kennung: `b${String(stelle)}`,
      })),
    ]

    const { lagen } = zonenlagen(feld('tisch'), gelegte)
    const ersteA = lagen.find((lage) => lage.kennung === 'a0')
    const ersteB = lagen.find((lage) => lage.kennung === 'b0')

    expect(ersteB?.x).toBe(
      (ersteA?.x ?? 0) + Kartenbreite + Haufenversatz + Kartenabstand,
    )
  })
})
