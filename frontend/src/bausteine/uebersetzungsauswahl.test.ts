// **Welche Sätze der Editor zeigt** — LOC-02
// (#1688).
import { describe, expect, it } from 'vitest'

import { auswahl, zeilenJeSeite } from './uebersetzungsauswahl'

const vorlage = {
  'a.titel': 'Sets',
  'b.fehler': 'Die Datei ist zu groß.',
  'c.knopf': 'Anlegen',
}

const stand = {
  texte: {
    'a.titel': { text: 'Conjuntos', quelle: 'Sets' },
    'b.fehler': { text: 'Grande', quelle: 'Alt' },
  },
  fehlendePfade: ['c.knopf'],
  veraltetePfade: ['b.fehler'],
}

describe('auswahl', () => {
  it('zeigt fehlend, veraltet, übersetzt oder alles', () => {
    expect(auswahl(vorlage, stand, 'fehlend', '', 1).pfade).toEqual(['c.knopf'])
    expect(auswahl(vorlage, stand, 'veraltet', '', 1).pfade).toEqual([
      'b.fehler',
    ])
    expect(auswahl(vorlage, stand, 'uebersetzt', '', 1).pfade).toEqual([
      'a.titel',
      'b.fehler',
    ])
    expect(auswahl(vorlage, stand, 'alle', '', 1).pfade).toEqual([
      'a.titel',
      'b.fehler',
      'c.knopf',
    ])
  })

  // **Gesucht wird in Pfad, deutschem Satz und Übersetzung** — ohne Rücksicht
  // auf Groß und Klein.
  it('sucht in Pfad, Satz und Übersetzung', () => {
    expect(auswahl(vorlage, stand, 'alle', 'KNOPF', 1).pfade).toEqual([
      'c.knopf',
    ])
    expect(auswahl(vorlage, stand, 'alle', 'datei', 1).pfade).toEqual([
      'b.fehler',
    ])
    expect(auswahl(vorlage, stand, 'alle', 'conjun', 1).pfade).toEqual([
      'a.titel',
    ])
  })

  it('blättert und klemmt die Seite', () => {
    const viele = Object.fromEntries(
      Array.from({ length: zeilenJeSeite + 3 }, (_, i) => [
        `p.${String(i).padStart(3, '0')}`,
        'x',
      ]),
    )
    const leer = { texte: {}, fehlendePfade: [], veraltetePfade: [] }

    const zweite = auswahl(viele, leer, 'alle', '', 2)
    expect(zweite).toMatchObject({ seite: 2, seiten: 2 })
    expect(zweite.pfade).toHaveLength(3)

    expect(auswahl(viele, leer, 'alle', '', 9).seite).toBe(2)
    expect(auswahl(viele, leer, 'alle', '', 0).seite).toBe(1)
    expect(auswahl({}, leer, 'alle', '', 1)).toMatchObject({
      pfade: [],
      seite: 1,
      seiten: 1,
    })
  })
})
