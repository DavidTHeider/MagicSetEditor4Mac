// **Ein Tischbild zeigt keine halbe Bewegung** — Befund 3 des Zweitblicks
// über Phase 6.9 (PLAY-05.AK-04,
// #1713).
//
// Die Datei steht in `.covignore`, weil sie ohne Renderer nichts zu tun hat;
// `lande` aber entscheidet etwas — wo eine Karte im Abzug steht —, und das
// wird gemessen.
import { describe, expect, it, vi } from 'vitest'

import { neueFluege } from './pixibewegung'

import type { Beweger, Takter } from './pixibewegung'

function einSprite(): Beweger & { ort: { x: number; y: number } } {
  const ort = { x: 0, y: 0 }

  return {
    ort,
    position: {
      set: (x: number, y: number) => {
        ort.x = x
        ort.y = y
      },
    },
    rotation: 0,
  }
}

function einTakter(): Takter & { schritte: (() => void)[] } {
  const schritte: (() => void)[] = []

  return {
    schritte,
    started: false,
    add: (schritt: () => void) => schritte.push(schritt),
    start: vi.fn(),
    stop: vi.fn(),
  }
}

describe('lande', () => {
  it('setzt jede fliegende Karte an ihr Ziel und beendet den Flug', () => {
    const fluege = neueFluege()
    const sprite = einSprite()
    const sprites = new Map([['a', sprite]])
    const takter = einTakter()

    fluege.starte(
      'a',
      { x: 0, y: 0, drehung: 0 },
      { x: 100, y: 40, drehung: Math.PI / 2 },
    )
    fluege.treibe(takter, sprites)

    fluege.lande(sprites)

    expect(sprite.ort).toEqual({ x: 100, y: 40 })
    expect(sprite.rotation).toBe(Math.PI / 2)

    // **Und der Flug ist vorbei**: ein weiterer Schritt setzt nichts zurück
    // auf einen Zwischenstand.
    sprite.position.set(7, 7)
    for (const schritt of takter.schritte) schritt()

    expect(sprite.ort).toEqual({ x: 7, y: 7 })
  })

  it('lässt eine Karte ohne Sprite liegen', () => {
    const fluege = neueFluege()

    fluege.starte('weg', { x: 0, y: 0, drehung: 0 }, { x: 1, y: 1, drehung: 0 })

    expect(() => {
      fluege.lande(new Map())
    }).not.toThrow()
  })
})
