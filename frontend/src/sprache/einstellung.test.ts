import { describe, expect, it } from 'vitest'

import { gespeicherteSprache, istSprache, spracheMerken } from './einstellung'

function speicher(inhalt: Record<string, string> = {}) {
  return {
    getItem: (name: string) => inhalt[name] ?? null,
    setItem: (name: string, wert: string) => {
      inhalt[name] = wert
    },
    inhalt,
  }
}

describe('die eingestellte Sprache', () => {
  it('nimmt die gewählte', () => {
    expect(gespeicherteSprache(speicher({ 'mse.sprache': 'en' }), 'de')).toBe(
      'en',
    )
  })

  // **Was im Speicher steht, hat jemand hineingeschrieben.** Ohne Prüfung
  // stünde ein `kl` im Kontext, und jeder Katalogzugriff gäbe `undefined`.
  it('übergeht einen unbrauchbaren Eintrag', () => {
    expect(gespeicherteSprache(speicher({ 'mse.sprache': 'kl' }), 'en')).toBe(
      'en',
    )
  })

  it('nimmt sonst die Sprache des Systems, ohne die Gegend', () => {
    expect(gespeicherteSprache(speicher(), 'en-GB')).toBe('en')
    expect(gespeicherteSprache(speicher(), 'de-AT')).toBe('de')
  })

  // Deutsch ist die Ausgangssprache und damit der Rückfall (§9).
  it('fällt auf Deutsch zurück', () => {
    expect(gespeicherteSprache(speicher(), 'pt-BR')).toBe('de')
    expect(gespeicherteSprache(speicher(), '')).toBe('de')
  })

  it('merkt sich die Wahl', () => {
    const s = speicher()

    spracheMerken(s, 'en')

    expect(s.inhalt['mse.sprache']).toBe('en')
  })

  it('sagt, was eine Sprache ist', () => {
    expect(istSprache('de')).toBe(true)
    expect(istSprache('kl')).toBe(false)
    expect(istSprache(42)).toBe(false)
  })
})
