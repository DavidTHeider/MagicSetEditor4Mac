// **Die Liste der Zeitzonen, die das Fenster anbietet** — UXR-F1
// (#2007).
//
// # Was hier steht — und was in Go
//
// `zeitzonen.json` ist die Liste der Auswahl. **Dass Go jede Zeile annimmt**,
// hält `internal/werkzeuge/zeitzonenliste_test.go`; hier steht, dass **das
// Fenster** sie versteht. Beide Seiten zählen: eine Zone, die Go annähme und
// `Intl` nicht kennte, schriebe sich und zeigte die Uhr dann in der Zone des
// Rechners — still, denn `zeitzoneAus` fällt dafür zurück.
import { describe, expect, it } from 'vitest'

import { zeitzoneAus } from './darstellung'
import zonen from './zeitzonen.json'

describe('Die Zeitzonen der Auswahl', () => {
  it('versteht das Fenster in jeder Zeile', () => {
    const unbekannt = zonen.filter((zone) => zeitzoneAus(zone) !== zone)

    expect(unbekannt).toEqual([])
  })

  it('sind streng geordnet und ohne Doppel', () => {
    const geordnet = [...zonen].sort((a, b) => (a < b ? -1 : a > b ? 1 : 0))

    expect(zonen).toEqual(geordnet)
    expect(new Set(zonen).size).toBe(zonen.length)
  })

  it('führen keinen leeren Namen und nicht „Local"', () => {
    expect(zonen).not.toContain('')
    expect(zonen).not.toContain('Local')
  })

  // **Genug, dass eine Suche dazugehört** (R7: ab fünfzehn Werten).
  it('sind mehr als fünfzehn', () => {
    expect(zonen.length).toBeGreaterThan(15)
  })

  it('führen die Zonen, nach denen man zuerst sucht', () => {
    expect(zonen).toEqual(
      expect.arrayContaining([
        'Europe/Berlin',
        'America/New_York',
        'Asia/Tokyo',
        'UTC',
      ]),
    )
  })
})
