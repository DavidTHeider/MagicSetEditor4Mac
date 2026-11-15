// **Die Wurzelschrift, in der der Arbeitsplatz seine Grenze rechnet** — G20
// aus dem Zweitblick über Phase 6.3 (#1664), verlagert in Phase 6.11
// (#1672). Gelesen wird, was das Wurzelelement trägt; ohne lesbaren Wert
// gilt 16.
import { afterEach, expect, it } from 'vitest'

import { wurzelschrift } from './arbeitsplatzmasse'

afterEach(() => {
  document.documentElement.style.fontSize = ''
})

it('liest die Schriftgröße des Wurzelelements', () => {
  document.documentElement.style.fontSize = '20px'

  expect(wurzelschrift()).toBe(20)
})

it('nimmt 16, wenn nichts zu lesen ist', () => {
  document.documentElement.style.fontSize = ''

  expect(wurzelschrift()).toBe(16)
})

// **Null Punkte sind keine Schrift** — eine Grenze von null rem machte jede
// Breite zu breit.
it('nimmt 16, wenn das Wurzelelement null Punkte trägt', () => {
  document.documentElement.style.fontSize = '0px'

  expect(wurzelschrift()).toBe(16)
})
