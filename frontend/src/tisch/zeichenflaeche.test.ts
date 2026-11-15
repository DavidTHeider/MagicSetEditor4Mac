// **Die Zeichenfläche in Punkten, bei jeder Dichte** — gefunden am 2026-09-24
// in der WebKit-Probe
// (#1828).
//
// **Gegen das echte `ViewSystem` von PixiJS** und nicht gegen ein Doppel: die
// Frage ist gerade, in welcher Einheit PixiJS seine Maße führt, und ein Doppel
// gäbe die Antwort, die sein Verfasser erwartet (AGENTS.md §3, „eine
// Vorrichtung, die großzügiger ist als der Betrieb, prüft sich selbst").
import { ViewSystem } from 'pixi.js'

import type { Renderer } from 'pixi.js'
import { describe, expect, it } from 'vitest'

import { zeichenflaeche } from './zeichenflaeche'

// rendererMit baut, was der Renderer von PixiJS nach außen gibt, über seinem
// echten `ViewSystem`.
//
// **Die Getter sind die des Renderers**: `width` ist
// `view.texture.frame.width`, `resolution` ist `view.resolution`, `screen` ist
// `view.screen` (pixi.js 8.20.1,
// `lib/rendering/renderers/shared/system/AbstractRenderer.mjs`). **Auch die,
// die `zeichenflaeche` nicht liest**: wer zu `width / resolution` zurückkehrt,
// soll hier rot werden und nicht an einem fehlenden Feld scheitern. Die Größe
// setzt `resize` — derselbe Weg, den `resizeTo` nimmt, wenn das Fenster sich
// ändert.
//
// **Seit 8.22 braucht `ViewSystem` seinen Renderer**: es meldet die Größe des
// Bildschirms je `uid` an `TexturePool` und `CanvasPool`
// (`_updateScreenSize`, pixi.js 8.22.0). Mehr als die `uid` liest es nicht.
function rendererMit(dichte: number, breite: number, hoehe: number) {
  const sicht = new ViewSystem({ uid: 1 } as unknown as Renderer)
  sicht.init({
    ...ViewSystem.defaultOptions,
    resolution: dichte,
    autoDensity: true,
  })
  sicht.resize(breite, hoehe, dichte)

  return {
    get width() {
      return sicht.texture.frame.width
    },
    get height() {
      return sicht.texture.frame.height
    },
    get resolution() {
      return sicht.resolution
    },
    get screen() {
      return sicht.screen
    },
    get pixel() {
      return { breite: sicht.canvas.width, hoehe: sicht.canvas.height }
    },
  }
}

describe('die Zeichenfläche des Tischs', () => {
  it.each([1, 2, 3])(
    'ist bei %d-facher Dichte so groß wie die Leinwand in Punkten',
    (dichte) => {
      const renderer = rendererMit(dichte, 1095, 602)

      expect(zeichenflaeche(renderer)).toEqual({ breite: 1095, hoehe: 602 })
    },
  )

  // **Die Leinwand selbst hat die Pixel** — dass die Dichte überhaupt
  // ankommt, sonst prüfte der Lauf darüber bei jeder Dichte dasselbe.
  it('lässt die Pixel bei der Leinwand', () => {
    expect(rendererMit(2, 1095, 602).pixel).toEqual({
      breite: 2190,
      hoehe: 1204,
    })
  })
})
