// **Die Kamera über dem Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import {
  begrenzt,
  eingepasst,
  gedreht,
  gezoomt,
  Groesstmass,
  Kleinstmass,
  nachMillimetern,
  nachPunkten,
  verschoben,
} from './kamera'

import type { Kamera } from './kamera'

const flaeche = { breite: 1000, hoehe: 800 }
const plan = { breite: 2000, hoehe: 1000 }
const ruhig: Kamera = { mitteX: 100, mitteY: 50, massstab: 2, drehung: 0 }

const nah = (eines: { x: number; y: number }, x: number, y: number) => {
  expect(eines.x).toBeCloseTo(x, 6)
  expect(eines.y).toBeCloseTo(y, 6)
}

describe('eingepasst', () => {
  it('nimmt den engeren der beiden Maßstäbe', () => {
    // 1000/2000 = 0,5 gegen 800/1000 = 0,8 — der kleinere gewinnt.
    expect(eingepasst(plan, flaeche).massstab).toBeCloseTo(0.5, 6)
  })

  it('stellt die Mitte des Plans in die Mitte der Fläche', () => {
    const kamera = eingepasst(plan, flaeche)

    nah(nachPunkten(kamera, { x: 1000, y: 500 }, flaeche), 500, 400)
  })

  it('lässt Rand, wenn einer verlangt ist', () => {
    expect(eingepasst(plan, flaeche, 100).massstab).toBeCloseTo(0.4, 6)
  })

  it('kommt mit einem Plan ohne Ausdehnung zurecht', () => {
    expect(eingepasst({ breite: 0, hoehe: 0 }, flaeche).massstab).toBe(
      Groesstmass,
    )
  })
})

describe('begrenzt', () => {
  it('hält den Maßstab zwischen seinen Grenzen', () => {
    expect(begrenzt(0.0001)).toBe(Kleinstmass)
    expect(begrenzt(1000)).toBe(Groesstmass)
    expect(begrenzt(1.5)).toBe(1.5)
  })
})

describe('nachPunkten und nachMillimetern', () => {
  it('bilden die Mitte auf die Mitte ab', () => {
    nah(nachPunkten(ruhig, { x: 100, y: 50 }, flaeche), 500, 400)
  })

  it('rechnen mit dem Maßstab', () => {
    nah(nachPunkten(ruhig, { x: 110, y: 50 }, flaeche), 520, 400)
  })

  it('kehren einander um, auch gedreht', () => {
    for (const drehung of [0, 37, 90, 180, 271]) {
      const kamera = { ...ruhig, drehung }
      const ort = { x: 137, y: -42 }
      const zurueck = nachMillimetern(
        kamera,
        nachPunkten(kamera, ort, flaeche),
        flaeche,
      )

      nah(zurueck, ort.x, ort.y)
    }
  })

  it('dreht den Tisch um seine Mitte', () => {
    const kamera = { ...ruhig, drehung: 90 }

    nah(nachPunkten(kamera, { x: 110, y: 50 }, flaeche), 500, 420)
  })
})

describe('verschoben', () => {
  it('zieht den Tisch in die Richtung des Zeigers', () => {
    const kamera = verschoben(ruhig, 100, 0)

    expect(kamera.mitteX).toBeCloseTo(50, 6)
    nah(nachPunkten(kamera, { x: 100, y: 50 }, flaeche), 600, 400)
  })

  it('folgt dem Bildschirm und nicht dem Tisch, wenn gedreht ist', () => {
    // Bei 90° liegt „rechts auf dem Bildschirm" auf der Tischachse y: der
    // Schwenk trifft deshalb mitteY und lässt mitteX in Ruhe.
    const kamera = verschoben({ ...ruhig, drehung: 90 }, 100, 0)

    expect(kamera.mitteX).toBeCloseTo(100, 6)
    expect(kamera.mitteY).toBeCloseTo(100, 6)
    nah(nachPunkten(kamera, { x: 100, y: 50 }, flaeche), 600, 400)
  })
})

describe('gezoomt', () => {
  it('lässt den Ort unter dem Zeiger stehen', () => {
    const zeiger = { x: 800, y: 200 }
    const vorher = nachMillimetern(ruhig, zeiger, flaeche)
    const kamera = gezoomt(ruhig, 2, zeiger, flaeche)

    nah(nachMillimetern(kamera, zeiger, flaeche), vorher.x, vorher.y)
    expect(kamera.massstab).toBeCloseTo(4, 6)
  })

  it('bleibt an der Grenze stehen, statt sie zu überschreiten', () => {
    const weit = gezoomt(ruhig, 1000, { x: 500, y: 400 }, flaeche)

    expect(weit.massstab).toBe(Groesstmass)
    expect(gezoomt(ruhig, 0.0001, { x: 500, y: 400 }, flaeche).massstab).toBe(
      Kleinstmass,
    )
  })
})

describe('gedreht', () => {
  it('hält den Winkel im Kreis', () => {
    expect(gedreht({ ...ruhig, drehung: 350 }, 20).drehung).toBe(10)
    expect(gedreht({ ...ruhig, drehung: 10 }, -20).drehung).toBe(350)
  })
})
