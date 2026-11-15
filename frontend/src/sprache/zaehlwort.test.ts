// **Eine Zahl und ihr Wort** —
// #1218.
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'
import { zaehlform, zaehlwort } from './zaehlwort'

it('nimmt die Einzahl bei genau einem', () => {
  expect(zaehlwort(1, kataloge.de.allgemein.zaehlwoerter.karten)).toBe(
    '1 Karte',
  )
  expect(zaehlwort(1, kataloge.en.allgemein.zaehlwoerter.karten)).toBe('1 card')
})

it('nimmt die Mehrzahl bei allem anderen', () => {
  for (const anzahl of [0, 2, 107]) {
    expect(zaehlwort(anzahl, kataloge.de.allgemein.zaehlwoerter.karten)).toBe(
      `${String(anzahl)} Karten`,
    )
  }
})

// **Auch bei minus eins** — sie kommt in keiner Kartenzahl vor, und genau
// deshalb steht sie hier: die Bedingung lautet „genau eins" und nicht „Betrag
// eins". Wer sie später auf `Math.abs` umstellt, sieht es an diesem Fall.
it('nimmt die Mehrzahl bei minus eins', () => {
  expect(zaehlwort(-1, kataloge.de.allgemein.zaehlwoerter.karten)).toBe(
    '-1 Karten',
  )
})

// **Das Russische zählt dreifach** — LOC-01 (#1687): eins, zwei bis vier,
// fünf bis zwanzig; und 21 zählt wie 1. Die dritte Form heißt `wenige`, und
// nur Kataloge, die sie führen, zählen danach (`Intl.PluralRules`).
describe('Die russische Zählform', () => {
  const karte = { eine: 'карта', wenige: 'карты', viele: 'карт' }

  it.each([
    [1, '1 карта'],
    [2, '2 карты'],
    [4, '4 карты'],
    [5, '5 карт'],
    [11, '11 карт'],
    [21, '21 карта'],
    [22, '22 карты'],
    [0, '0 карт'],
  ])('%i', (anzahl, erwartet) => {
    expect(zaehlwort(anzahl, karte)).toBe(erwartet)
  })
})

// **Die dritte Zählform folgt der Sprache, die sie trägt** — Befund Z4 des
// Zweitblicks über Phase 6.5 (LOC-02, #1688): eine eigene Übersetzung ins
// Polnische zählt 21 anders als das Russische. **Die Regel reist mit der
// Form** (Befund N5 des Nachlaufs), nicht in einem Zustand des Moduls.
describe('Die Zählregel einer eigenen Übersetzung', () => {
  const karte = { eine: 'karta', wenige: 'karty', viele: 'kart', regel: 'pl' }

  it.each([
    [1, '1 karta'],
    [2, '2 karty'],
    [5, '5 kart'],
    [21, '21 kart'],
    [22, '22 karty'],
  ])('zählt %i polnisch', (anzahl, erwartet) => {
    expect(zaehlwort(anzahl, karte)).toBe(erwartet)
  })

  it('fällt bei einer unbekannten Sprache auf die russische Regel zurück', () => {
    expect(zaehlwort(21, { ...karte, regel: 'xx-x-nichts' })).toBe('21 karta')
  })
})

// **Das Wort allein** — UXR-03 (#1911):
// die Kennzahlen der Startseite setzen die Zahl groß und das Wort klein
// darunter. Gezählt wird dabei nach derselben Regel wie mit der Zahl davor —
// sonst stünde unter einer 1 „Karten".
describe('Das Wort ohne Zahl', () => {
  it('folgt derselben Regel wie mit der Zahl', () => {
    const karten = kataloge.de.allgemein.zaehlwoerter.karten

    expect(zaehlform(1, karten)).toBe('Karte')
    expect(zaehlform(357, karten)).toBe('Karten')
    expect(
      zaehlform(3, { eine: 'карта', wenige: 'карты', viele: 'карт' }),
    ).toBe('карты')
  })
})
