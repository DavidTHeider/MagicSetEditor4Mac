// **PixiJS bekommt eine Adresse, die es unter `wails://` richtig auflöst** —
// #2066.
//
// Am nativen Fenster zeichnete der Tisch keine einzige Karte, nicht einmal
// den Rücken: `path.isUrl` kennt nur `http(s):`, und für die Basis
// `wails://localhost/` machte `toAbsolute` aus `/spielbild/x.png` die Adresse
// `wails://spielbild/x.png` — den Pfad als Host. `Assets.load` scheiterte,
// und der Lader ließ still den Platzhalter stehen. Im Browsermodus
// (`http://localhost`) ging es, und dort lief jeder Test.
//
// **Gemessen an Pixis eigener Pfadlogik** und nicht an einer Nachbildung: wer
// Pixi hebt und es dort anders macht, sieht es hier.
import { path } from 'pixi.js'
import { describe, expect, it } from 'vitest'

import { fuerPixi } from './pixiadresse'

const nativ = 'wails://localhost/'

describe('Die Adresse für PixiJS', () => {
  it('kommt unter wails:// dort an, wo sie liegt', () => {
    expect(path.toAbsolute(fuerPixi('/spielbild/a.png', nativ), nativ)).toBe(
      'wails://localhost/spielbild/a.png',
    )
  })

  // **Der Befund selbst** — fällt er eines Tages weg, weil Pixi das Schema
  // kennt, darf die Umrechnung bleiben; dann sagt dieser Fall es.
  it('ging ohne die Umrechnung an den falschen Ort', () => {
    expect(path.toAbsolute('/spielbild/a.png', nativ)).toBe(
      'wails://spielbild/a.png',
    )
  })

  it('lässt den Browsermodus, wie er war', () => {
    const browser = 'http://localhost:34115/'

    expect(
      path.toAbsolute(fuerPixi('/tischbild/holz.png', browser), browser),
    ).toBe('http://localhost:34115/tischbild/holz.png')
  })

  it('lässt eine Adresse mit Schema unberührt', () => {
    expect(fuerPixi('data:image/png;base64,AAAA', nativ)).toBe(
      'data:image/png;base64,AAAA',
    )
  })
})
