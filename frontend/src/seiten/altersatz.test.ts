// **Wie lange es her ist** — so grob, wie ein Stand altert (AP-8, #2365).
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { alterSatz } from './altersatz'

const woerter = kataloge.de.seiten.start.ueberblick.alter

describe('der Satz über ein Alter', () => {
  it.each([
    [0, 'gerade eben'],
    [59, 'gerade eben'],
    [60, 'vor 1 min'],
    [3599, 'vor 59 min'],
    [3600, 'vor 1 h'],
    [86399, 'vor 23 h'],
    // **Ein Tag ist einer** — hier stand „vor 1 Tagen“, und auf Englisch
    // „1 days ago“.
    [86400, 'vor einem Tag'],
    [172799, 'vor einem Tag'],
    [172800, 'vor 2 Tagen'],
  ])('%i Sekunden heißen „%s“', (sekunden, satz) => {
    expect(alterSatz(sekunden, woerter)).toBe(satz)
  })

  it('sagt es in jeder Sprache für einen Tag eigens', () => {
    for (const katalog of Object.values(kataloge)) {
      const eigene = katalog.seiten.start.ueberblick.alter
      expect(alterSatz(86400, eigene)).toBe(eigene.tag)
      expect(eigene.tag).not.toBe(eigene.tage)
    }
  })
})
