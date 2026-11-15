// **Was die Kachel „Sammlung“ unter ihre Zahl schreibt** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`: „486 € · ältester Preis vor 3 Tagen“.
import { describe, expect, it } from 'vitest'

import { geldAusText } from '../geld'
import { kataloge } from '../sprache/kataloge'
import { sammlungsunterzeile } from './sammlungskachel'

const texte = kataloge.de

function zeile(wert: string, waehrung: string, alterStunden: number) {
  return {
    waehrung,
    wert,
    exemplare: 3,
    aeltesterPreis: '2026-10-14T08:00:00Z',
    alterStunden,
    quelle: 'Scryfall',
  }
}

describe('die Unterzeile der Sammlung', () => {
  it('schweigt, solange nichts da ist', () => {
    expect(sammlungsunterzeile(undefined, texte, 'de')).toBeUndefined()
    expect(
      sammlungsunterzeile({ zeilen: [], ohnePreis: 0 }, texte, 'de'),
    ).toBeUndefined()
  })

  it('nennt den Wert und das Alter des ältesten Preises', () => {
    expect(
      sammlungsunterzeile(
        { zeilen: [zeile('486.00', 'EUR', 72)], ohnePreis: 0 },
        texte,
        'de',
      ),
    ).toBe(`${geldAusText('486.00', 'EUR', 'de')} · ältester Preis vor 3 Tagen`)
  })

  it('rechnet nicht um und nimmt das älteste Alter aller Währungen', () => {
    expect(
      sammlungsunterzeile(
        {
          zeilen: [zeile('486.00', 'EUR', 5), zeile('12.50', 'USD', 30)],
          ohnePreis: 0,
        },
        texte,
        'de',
      ),
    ).toBe(
      `${geldAusText('486.00', 'EUR', 'de')} · ` +
        `${geldAusText('12.50', 'USD', 'de')} · ältester Preis vor einem Tag`,
    )
  })

  it('sagt, wie viele Exemplare ohne Preis blieben', () => {
    expect(
      sammlungsunterzeile(
        { zeilen: [zeile('486.00', 'EUR', 0)], ohnePreis: 12 },
        texte,
        'de',
      ),
    ).toBe(
      `${geldAusText('486.00', 'EUR', 'de')} · ältester Preis gerade eben · 12 ohne Preis`,
    )
    expect(sammlungsunterzeile({ zeilen: [], ohnePreis: 1 }, texte, 'de')).toBe(
      '1 ohne Preis',
    )
  })
})
