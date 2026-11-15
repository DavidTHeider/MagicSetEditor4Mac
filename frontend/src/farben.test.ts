// **Farben zwischen MSEs Form und der des Browsers**
// (#634).
import { describe, expect, it } from 'vitest'

import { alsBrowserfarbe, alsMsefarbe } from './farben'

describe('alsBrowserfarbe', () => {
  it.each([
    ['rgb(0,0,0)', '#000000'],
    ['rgb(255, 255, 255)', '#ffffff'],
    ['rgba(12,34,56,128)', '#0c2238'],
    ['#abc', '#aabbcc'],
    ['#abcd', '#aabbcc'],
    ['#0C2238', '#0c2238'],
    ['#0c223880', '#0c2238'],
    ['transparent', '#000000'],
    // **Geklemmt, nicht abgewiesen** — MSE gibt es an wxColour weiter.
    ['rgb(300,0,0)', '#ff0000'],
  ])('liest %s als %s', (roh, erwartet) => {
    expect(alsBrowserfarbe(roh)).toBe(erwartet)
  })

  // **Was sich nicht lesen lässt, wird nicht erfunden.** Die Farbnamen der
  // wx-Datenbank kennt niemand hier; MSE liest sie, wir zeigen sie nicht als
  // Kachel.
  it.each(['', 'rot', 'papayawhip', 'rgb(1,2)', '#12345', '#zzz'])(
    'gibt für %s nichts',
    (roh) => {
      expect(alsBrowserfarbe(roh)).toBeUndefined()
    },
  )
})

describe('alsMsefarbe', () => {
  it('schreibt die Form, die MSE selbst erzeugt', () => {
    expect(alsMsefarbe('#0c2238')).toBe('rgb(12,34,56)')
  })

  // **Was keine Browserfarbe ist, bleibt, wie es war.** Sonst machte ein
  // unerwarteter Wert aus einer Rückgabe eine Erfindung.
  it('lässt Fremdes stehen', () => {
    expect(alsMsefarbe('papayawhip')).toBe('papayawhip')
  })
})
