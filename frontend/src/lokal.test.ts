// **Eine Eingabe in der Sprache des Fensters** — LOC-03.AK-01 und AK-03
// (#1689).
//
// # Wogegen hier geprüft wird
//
//   - **Was die Sprache selbst so schriebe, wird übersetzt** — »1.234,56«
//     im Deutschen, »1,234.56« im Englischen, »1'234.50« in der Schweiz.
//   - **Was zwei Werte haben kann, wird nicht geraten** (AK-03).
//   - **Was keine Lesart hat, geht unverändert weiter** — Go erklärt es.
//
// Die Rundläufe über drei Sprachen stehen in `lokalrundlauf.test.ts`.
import { describe, expect, it } from 'vitest'

import { alsProzent, alsZahl, lesarten, trennzeichen } from './lokal'

describe('lesarten', () => {
  // **Die Schreibweise der Sprache wird übersetzt** — mit und ohne Gruppen.
  it.each([
    ['1.234,56', 'de-DE', '1234.56'],
    ['1234,56', 'de-DE', '1234.56'],
    ['19,99', 'de-DE', '19.99'],
    ['-5,25', 'de-DE', '-5.25'],
    ['1,234.56', 'en-US', '1234.56'],
    ['1,234', 'ja-JP', '1234'],

    // **Die Schweiz trennt mit einem Apostroph**, Frankreich mit einem
    // schmalen geschützten Leerzeichen — und wer ein gewöhnliches tippt,
    // meint dasselbe (gemessen am 2026-09-22).
    ["1'234.50", 'de-CH', '1234.50'],
    ['1 234,56', 'fr-FR', '1234.56'],

    // **Hindi gruppiert zu zweien** — eine Regel über Dreiergruppen läse
    // das nicht.
    ['12,34,567.89', 'hi-IN', '1234567.89'],
  ])('%s in %s ist %s', (text, sprache, kanonisch) => {
    expect(lesarten(text, sprache)).toEqual({ art: 'eindeutig', kanonisch })
  })

  // **Die kanonische Form gilt in jeder Sprache** — »19.99« im Deutschen
  // ist kein deutsch geschriebener Betrag, aber ein kanonischer.
  it('nimmt die kanonische Form, wo die Sprache sie nicht liest', () => {
    expect(lesarten('19.99', 'de-DE', 2)).toEqual({
      art: 'eindeutig',
      kanonisch: '19.99',
    })
  })

  // **Wo beide Lesarten derselbe Text sind, ist es eine.**
  it('nimmt eine Lesart, wenn beide gleich lauten', () => {
    expect(lesarten('1234.56', 'en-US')).toEqual({
      art: 'eindeutig',
      kanonisch: '1234.56',
    })
  })

  // **Die Stellen der Währung entscheiden, ob die kanonische Form eine
  // Lesart ist.** Beim Euro ist »1.234« keine — drei Stellen hat er nicht —,
  // und dann bleibt die deutsche; beim Yen ebenso.
  it.each([
    ['EUR', 2],
    ['JPY', 0],
  ])('liest »1.234« beim %s deutsch', (_, stellen) => {
    expect(lesarten('1.234', 'de-DE', stellen)).toEqual({
      art: 'eindeutig',
      kanonisch: '1234',
    })
  })

  // **Beim Dinar hat »1.234« zwei Werte** — tausendzweihundertvierunddreißig
  // oder eins Komma zwei drei vier. Das ist der Fall, für den es AK-03
  // gibt: erklären, nicht raten.
  it('erklärt, was zwei Werte haben kann', () => {
    expect(lesarten('1.234', 'de-DE', 3)).toEqual({
      art: 'mehrdeutig',
      lesarten: ['1234', '1.234'],
    })
  })

  // **Was keine Lesart hat, geht weiter** — und Go sagt, warum.
  it.each([
    // **Im Englischen schriebe niemand »19,99«** für neunzehn Komma
    // neunundneunzig, und für neunzehnhundertneunundneunzig auch nicht.
    ['19,99', 'en-US'],
    ['19,99 €', 'de-DE'],
    ['1,2,3', 'de-DE'],

    // **Westliche Gruppen sind in Hindi keine** — dort stünde »12,34,567«.
    ['1,234,567.89', 'hi-IN'],
  ])('lässt %s in %s unverändert', (text, sprache) => {
    expect(lesarten(text, sprache)).toEqual({ art: 'fremd' })
  })

  // **Mehr als zwanzig Stellen trägt `Intl` nicht** — dann bleibt die
  // kanonische Lesart, und Go entscheidet.
  it('nimmt bei zu vielen Stellen die kanonische Form', () => {
    const text = '0.000000000000000000001'

    expect(lesarten(text, 'en-US')).toEqual({
      art: 'eindeutig',
      kanonisch: text,
    })
  })

  // **Kein Text ist kein Betrag** — und keine Frage.
  it.each(['', '   '])('gibt für %j nichts', (text) => {
    expect(lesarten(text, 'de-DE', 2)).toEqual({
      art: 'eindeutig',
      kanonisch: '',
    })
  })
})

describe('alsZahl', () => {
  // **Beide Lesarten in der Schreibweise dessen, der getippt hat.**
  it('schreibt die Stellen, die verlangt sind', () => {
    expect(alsZahl('1234', 'de-DE', 3)).toBe('1.234,000')
    expect(alsZahl('1.234', 'de-DE', 3)).toBe('1,234')
  })
})

describe('trennzeichen', () => {
  // **Gemessen und nicht aufgezählt** — und ein schmales geschütztes
  // Leerzeichen wird ein gewöhnliches.
  it.each([
    ['de-DE', '.', ','],
    ['en-US', ',', '.'],
    ['ja-JP', ',', '.'],
    ['fr-FR', ' ', ','],
  ])('%s trennt mit »%s« und »%s«', (sprache, gruppe, dezimal) => {
    expect(trennzeichen(sprache)).toEqual({ gruppe, dezimal })
  })
})

describe('alsProzent', () => {
  // **Das Englische schreibt kein Leerzeichen**, das Deutsche ein
  // geschütztes — und hier stand in beiden Katalogen »{prozent} %«.
  it('folgt der Sprache beim Abstand', () => {
    expect(alsProzent(0.42, 'de-DE')).toBe('42\u00a0%')
    expect(alsProzent(0.42, 'en-US')).toBe('42%')
    expect(alsProzent(0.42, 'ja-JP')).toBe('42%')
  })

  // **Auf ganze Zahlen** — und gerundet wie `Math.round`.
  it('rundet auf ganze Prozent', () => {
    expect(alsProzent(0.374, 'en-US')).toBe('37%')
    expect(alsProzent(0.375, 'en-US')).toBe('38%')
  })
})
