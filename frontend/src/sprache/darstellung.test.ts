// **Die volle Locale bleibt erhalten** — LOC-03.AK-03
// (#1689).
//
// # Wogegen hier geprüft wird
//
// **Hier stand `navigator.language.split('-')[0]`**: aus `de-CH` wurde `de`,
// und eine Schweizer Zahl sah aus wie eine deutsche — »1.234,50« statt
// »1'234.50« (gemessen am 2026-09-22). Die Sprache der Texte braucht nur den
// Sprachanteil; die Darstellung braucht die Gegend.
import { describe, expect, it } from 'vitest'

import { darstellungAus, zeitzoneAus } from './darstellung'

describe('darstellungAus', () => {
  // **Die Wahl, sonst das System — ganz —, sonst die Sprache.**
  it.each([
    ['de-CH', 'en-US', 'en', 'de-CH'],
    ['', 'de-CH', 'de', 'de-CH'],
    ['', 'en-GB', 'en', 'en-GB'],
    ['', '', 'de', 'de'],
  ])(
    '%j bei System %j und Sprache %s ist %s',
    (gewaehlt, system, sprache, erwartet) => {
      expect(darstellungAus(gewaehlt, system, sprache)).toBe(erwartet)
    },
  )

  // **Was `Intl` nicht versteht, fällt zurück** — und nimmt nicht die Seite
  // mit. `Intl` wirft bei »de_CH« einen `RangeError`; Go weist es zwar ab,
  // aber ein System kann es trotzdem melden.
  it('fällt zurück, wo Intl eine Locale abweist', () => {
    expect(darstellungAus('de_CH', 'en-US', 'en')).toBe('en-US')
    expect(darstellungAus('', 'de_CH', 'de')).toBe('de')
  })
})

describe('zeitzoneAus', () => {
  // **Leer heißt die Zone des Rechners** — und die kennt der Browser.
  it('gibt eine bekannte Zone zurück und sonst keine', () => {
    expect(zeitzoneAus('Asia/Tokyo')).toBe('Asia/Tokyo')
    expect(zeitzoneAus('')).toBeUndefined()

    // **Eine Zone, die `Intl` nicht kennt, ist die des Systems** — ein
    // `RangeError` beim Zeichnen nähme sonst jede Seite mit einem Datum mit.
    expect(zeitzoneAus('Mond/Krater')).toBeUndefined()
  })
})
