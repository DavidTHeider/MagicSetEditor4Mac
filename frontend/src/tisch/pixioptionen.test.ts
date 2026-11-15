// **Tab bleibt beim Tisch** — L02
// (#2190).
//
// PixiJS 8.22 aktiviert seine Barrierefreiheitsschicht wieder, wenn auf einer
// frischen Seite Tab gedrückt wird (`activateOnTab`, Vorgabe `true`; seit
// 8.14 war das ausgefallen — Freigabenotiz von 8.22.0, nachgelesen am
// 2026-10-08). Dieser Tisch bedient sich über eigene Knöpfe; mit 8.20.1 lag
// über der Leinwand keine Schicht von PixiJS, und so soll es bleiben.
import { describe, expect, it } from 'vitest'

import { tischoptionen } from './pixioptionen'

describe('Die Optionen des gezeichneten Tischs', () => {
  it('lassen Tab die Barrierefreiheitsschicht von PixiJS nicht aktivieren', () => {
    const leinwand = document.createElement('canvas')

    expect(tischoptionen(leinwand, 0x000000).accessibilityOptions).toEqual({
      activateOnTab: false,
    })
  })

  it('nehmen ohne Fenster die Dichte 1 und die Leinwand als Rahmen', () => {
    const ohneFenster = document.implementation.createHTMLDocument('')
    const leinwand = ohneFenster.createElement('canvas')

    expect(tischoptionen(leinwand, 0)).toMatchObject({
      resolution: 1,
      resizeTo: leinwand,
    })
  })

  it('zeichnen auf Zuruf und in der Dichte des Bildschirms', () => {
    const rahmen = document.createElement('div')
    const leinwand = document.createElement('canvas')
    rahmen.append(leinwand)

    const optionen = tischoptionen(leinwand, 0x123456)

    expect(optionen).toMatchObject({
      canvas: leinwand,
      autoStart: false,
      autoDensity: true,
      background: 0x123456,
      resizeTo: rahmen,
    })
  })
})
