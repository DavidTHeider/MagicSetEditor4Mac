// **Die Sätze über eine Sprache** — H04
// (#2258).
//
// Der Zweig „redaktionell geprüft am …" wird heute von keiner Sprache
// betreten — deshalb steht er hier an einer erfundenen Herkunft, statt
// ungeprüft im Bauteil zu liegen.
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { herkunftstext, redaktionstext } from './sprachstand'

const texte = kataloge.de.seiten.sprachstaende

describe('Die Sätze über eine Sprache', () => {
  it('nennt eine Redaktion mit ihrem Tag', () => {
    expect(
      redaktionstext(
        texte,
        {
          herkunft: 'sprachmodell',
          uebersetzt: '2026-09-29',
          geprueft: '2026-10-10',
        },
        'de',
      ),
    ).toBe('redaktionell geprüft am 10. Oktober 2026')
  })

  it('behauptet keine, die nicht stattfand', () => {
    expect(
      redaktionstext(
        texte,
        { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
        'de',
      ),
    ).toBe(texte.ungeprueft)
  })

  it('nennt den Tag einer Übersetzung in der Darstellung', () => {
    expect(
      herkunftstext(
        texte,
        { herkunft: 'sprachmodell', uebersetzt: '2026-09-29', geprueft: '' },
        'en-US',
      ),
    ).toBe('von einem Sprachmodell übersetzt am September 29, 2026')
  })
})
