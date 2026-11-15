// **Was die Befehlspalette findet** — aus `Rahmen.test.tsx` herausgelöst, weil
// jene Datei mit der Kontextleiste (U4, #1415) an die Zeilengrenze kam (§2).
// Reine Funktion, reine Fälle.
import { describe, expect, it } from 'vitest'

import { seiten } from '../navigation'
import { kataloge } from '../sprache/kataloge'
import { passende } from './Palette'

describe('passende', () => {
  it('gibt ohne Eingabe alle Seiten', () => {
    expect(passende(kataloge.de, '')).toHaveLength(seiten.length)
    expect(passende(kataloge.de, '   ')).toHaveLength(seiten.length)
  })

  // **Ohne Anfangsbindung**: wer „stell" tippt, meint die Einstellungen.
  it('findet auch mitten im Wort', () => {
    expect(passende(kataloge.de, 'stell')).toEqual(['einstellungen'])
  })

  it('achtet nicht auf Groß- und Kleinschreibung', () => {
    expect(passende(kataloge.de, 'KARTEN')).toEqual(['karten'])
  })

  it('gibt nichts, wenn nichts passt', () => {
    expect(passende(kataloge.de, 'gibtsnicht')).toEqual([])
  })

  it('sucht in der eingestellten Sprache', () => {
    expect(passende(kataloge.en, 'Cards')).toEqual(['karten'])
    expect(passende(kataloge.en, 'Karten')).toEqual([])
  })
})
