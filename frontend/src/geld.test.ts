// **Geld als Text** — LOC-03.AK-02
// (#1689).
//
// # Wogegen hier geprüft wird
//
//   - **Die Stellen kommen aus der Währung** — 150 Yen sind »150 ¥« und
//     nicht »1,50 ¥«.
//   - **Drei Sprachen** (AK-01): de-DE, en-US und eine CJK-Locale.
//   - **Ein ungültiger Code nimmt nicht die Seite mit** — `Intl` wirft, und
//     wer das nicht fängt, zeichnet gar nichts mehr.
import { describe, expect, it } from 'vitest'

import { alsGeld, geldAusText } from './geld'

// **Die kleinste Einheit trägt die Zahl** — und die Währung sagt, wie viele
// Stellen das sind.
it('schreibt jeden Betrag in den Stellen seiner Währung', () => {
  // **Der Euro hat zwei** — 1999 Cent sind 19,99 €.
  expect(alsGeld(1999, 'EUR', 'de-DE')).toContain('19,99')

  // **Der Yen hat keine** — 150 Yen sind 150 und nicht 1,50. Vorher stand
  // hier `wert / 100`, und daraus wurde »2 ¥«.
  expect(alsGeld(150, 'JPY', 'de-DE')).toContain('150')
  expect(alsGeld(150, 'JPY', 'de-DE')).not.toContain(',')

  // **Der Dinar hat drei** — ein Fils ist ausdrückbar.
  expect(alsGeld(1234, 'BHD', 'de-DE')).toContain('1,234')
  expect(alsGeld(1, 'BHD', 'de-DE')).toContain('0,001')
})

// **Dieselbe Zahl in drei Sprachen** — AK-01 verlangt de-DE, en-US und eine
// CJK-Locale.
//
// **Geprüft wird das Trennzeichen und nicht der ganze Satz**: wo das Symbol
// steht und ob ein schmales Leerzeichen davorsteht, entscheidet die
// Laufzeit, und ein Lauf, der es festschriebe, bräche beim nächsten
// ICU-Wechsel (§3).
it('folgt der Sprache beim Trennzeichen', () => {
  expect(alsGeld(1999, 'EUR', 'de-DE')).toContain('19,99')
  expect(alsGeld(1999, 'USD', 'en-US')).toContain('19.99')
  expect(alsGeld(1999, 'EUR', 'ja-JP')).toContain('19.99')

  // **Und der Yen bleibt in jeder Sprache ohne Nachkommastellen** — das ist
  // eine Eigenschaft der Währung und keine der Sprache (AK-02: keine
  // automatische Umrechnung aus einer Spracheinstellung).
  for (const sprache of ['de-DE', 'en-US', 'ja-JP']) {
    expect(alsGeld(150, 'JPY', sprache)).not.toContain('1,5')
    expect(alsGeld(150, 'JPY', sprache)).not.toContain('1.5')
  }
})

// **Ohne Währung bleibt die Zahl eine Zahl** — „noch nichts freigegeben"
// ist kein Fehler.
it('gibt ohne Währung die nackte Zahl', () => {
  expect(alsGeld(1999, '', 'de-DE')).toBe('1999')
})

// **Ein ungültiger Code nimmt die Seite nicht mit.**
//
// **`Intl` wirft einen `RangeError`** für alles, was nicht drei Buchstaben
// sind (gemessen am 2026-09-22 an »EURO«). Ein Wurf beim Zeichnen nimmt den
// ganzen Baum mit — und der Code kommt aus einem Feld, in das jemand tippt.
it('überlebt einen ungültigen Währungscode', () => {
  expect(alsGeld(1999, 'EURO', 'de-DE')).toBe('1999')
  expect(alsGeld(1999, '€', 'de-DE')).toBe('1999')
})

// **Ein unbekannter, aber gültiger Code bekommt zwei Stellen** — so
// formatiert ihn `Intl`, und die Vorgabe ist dieselbe wie im Kern.
it('schreibt einen unbekannten Code mit zwei Stellen', () => {
  expect(alsGeld(1999, 'XYZ', 'de-DE')).toContain('19,99')
})

// **Die Skala ist die des Bestands und nicht die von `Intl`** — beim
// Forint und beim irakischen Dinar sagen die beiden etwas anderes.
//
// **Der rote Lauf dazu**: vorher teilte das Fenster durch die Stellen, die
// `Intl` nennt, und das sind bei beiden null. 1.999 Fillér standen als
// »1.999 Ft« da, 1.234 Fils als »1.234 IQD« — hundert- und tausendfach.
it('teilt durch die Stellen, die Go speichert', () => {
  expect(alsGeld(1999, 'HUF', 'de-DE')).not.toContain('1.999')
  expect(alsGeld(1234, 'IQD', 'de-DE')).not.toContain('1.234')
})

// **Und zeigt sie auch** — angestrichen vom Zweitblick über Phase 6.3 (G8):
// `Intl` zeigt den Forint ohne Nachkommastellen, und 19,99 Forint standen
// gerundet als »20 HUF« da; 1,234 Dinar als einer. Geprüft werden die
// Stellen — das Zeichen wählt `Intl` nach der Sprache. Eine Anzeige, die rundet,
// verschweigt einen Betrag, der im Bestand steht.
it('zeigt jeden Betrag mit den Stellen, die Go speichert', () => {
  expect(alsGeld(1999, 'HUF', 'de-DE')).toMatch(/^19,99\s/u)
  expect(alsGeld(1234, 'IQD', 'de-DE')).toMatch(/^1,234\s/u)
  expect(geldAusText('19.99', 'HUF', 'de')).toMatch(/^19,99\s/u)
  expect(alsGeld(150, 'JPY', 'de-DE')).toMatch(/^150\s¥$/u)
})

// **Ein Betrag, den Go als Text schickt** — Markt, Portfolio und Verlauf
// (`internal/app/preisdto.go`). Angestrichen vom Zweitblick über Phase 6.3
// (M10): dort stand »1.234 BHD«, und im Deutschen liest sich das als
// eintausendzweihundertvierunddreißig Dinar.
describe('geldAusText', () => {
  it('schreibt einen kanonischen Betrag in der Sprache des Fensters', () => {
    expect(geldAusText('19.99', 'EUR', 'de')).toMatch(/^19,99\s€$/u)
    expect(geldAusText('1.234', 'BHD', 'de')).toMatch(/^1,234\sBHD$/u)
    expect(geldAusText('19.99', 'EUR', 'en')).toBe('€19.99')
  })

  it('lässt stehen, was es nicht schreiben kann', () => {
    expect(geldAusText('', 'EUR', 'de')).toBe('')
    expect(geldAusText('19.99', '', 'de')).toBe('19.99')
    expect(geldAusText('viel', 'EUR', 'de')).toBe('viel EUR')
  })
})
