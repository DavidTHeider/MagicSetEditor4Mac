// **Die Größe der Vorschau** — Werkstück U5 von Phase 5.1
// (#1416).
import { describe, expect, it } from 'vitest'

import {
  beschnittProzent,
  cssBreite,
  passendeBreite,
  vergroesserungFuer,
} from './zoom'

describe('vergroesserungFuer', () => {
  it('rastert ab 150 % neu und darunter in Vorlagenpunkten', () => {
    expect(vergroesserungFuer('anpassen')).toBe(1)
    expect(vergroesserungFuer('hundert')).toBe(1)
    expect(vergroesserungFuer(50)).toBe(1)
    expect(vergroesserungFuer('druck')).toBe(1)
    expect(vergroesserungFuer(150)).toBe(1.5)
    expect(vergroesserungFuer(200)).toBe(2)
  })
})

describe('cssBreite', () => {
  it('rechnet jede Stufe auf die Vorlagenpunkte zurück', () => {
    expect(cssBreite('anpassen', { breite: 375, dpi: 150 })).toBeUndefined()
    expect(cssBreite('hundert', { breite: 375, dpi: 150 })).toBe('375px')
    expect(cssBreite(50, { breite: 375, dpi: 150 })).toBe('187.5px')
    // Bei 150 % ist das Bild schon 1,5-fach gerastert: 562,5 Bildpunkte sind
    // 375 Vorlagenpunkte, gezeigt in 562,5 CSS-Punkten.
    expect(cssBreite(150, { breite: 562.5, dpi: 150 })).toBe('562.5px')
    expect(cssBreite(200, { breite: 750, dpi: 150 })).toBe('750px')
  })

  it('gibt die Druckgröße in Zoll — und ohne Auflösung gar nicht', () => {
    expect(cssBreite('druck', { breite: 375, dpi: 150 })).toBe('2.5in')
    expect(cssBreite('druck', { breite: 375, dpi: 0 })).toBeUndefined()
  })
})

// **„Anpassen" heißt: die Karte passt** — UXR-06
// (#1914).
// Gemessen bei 1.280 × 800: die Vorschau nahm zwei Fünftel des Zentrums,
// die Karte war 256 Punkte breit, begann bei y = 627 und stand zu sechzig
// Prozent unter dem Fensterrand. Im Arbeitsplatz füllt sie jetzt die Breite
// des Zentrums — aber nie mehr, als die Höhe unter ihr zulässt.
describe('passendeBreite', () => {
  it('nimmt die Spalte, aber nur so breit, wie die Höhe es zulässt', () => {
    expect(passendeBreite({ breite: 375, hoehe: 523 }, 400)).toBe(
      'min(100%, 286px)',
    )
  })

  it('rechnet mit den Punkten des Bildes, wie es gerastert ist', () => {
    expect(passendeBreite({ breite: 100, hoehe: 140 }, 700)).toBe(
      'min(100%, 500px)',
    )
  })
})

describe('beschnittProzent', () => {
  it('legt drei Millimeter in Prozent der Kanten um', () => {
    const rand = beschnittProzent({ breite: 375, hoehe: 523, dpi: 150 })

    expect(rand?.x).toBeCloseTo(4.724, 2)
    expect(rand?.y).toBeCloseTo(3.388, 2)
  })

  it('kennt ohne Auflösung keinen Beschnitt', () => {
    expect(
      beschnittProzent({ breite: 375, hoehe: 523, dpi: 0 }),
    ).toBeUndefined()
  })
})
