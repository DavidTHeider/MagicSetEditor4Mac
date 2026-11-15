// **Dieselben Stellen wie in Go** — LOC-03.AK-02
// (#1689).
//
// **Dass beide Listen gleich sind, hält ein Wächter auf der Go-Seite**
// (`internal/werkzeuge/waehrungsstellen_test.go`). Hier steht, was er nicht
// sieht: dass ein Code so gelesen wird, wie ein Mensch ihn tippt.
import { expect, it } from 'vitest'

import { waehrungsstellen } from './waehrungsstellen'

// **Klein geschrieben und mit Rand** — wie in `domain.Waehrungsstellen`.
it.each([
  ['JPY', 0],
  [' jpy ', 0],
  ['bhd', 3],
  ['EUR', 2],
  ['', 2],
  ['EURO', 2],
])('%j hat %i Stellen', (code, stellen) => {
  expect(waehrungsstellen(code)).toBe(stellen)
})
