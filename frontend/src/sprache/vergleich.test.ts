// **Vergleichen und Kleinschreiben in der Sprache des Fensters** — F01
// (#2243).
//
// `localeCompare` und `toLocaleLowerCase` ohne Sprache nehmen die des
// Systems, nicht die des Fensters. Für Deutsch und Englisch ist das dasselbe;
// seit es acht Oberflächensprachen gibt, ist es das nicht mehr. Die Vektoren
// sind in Node 24.15 (ICU) gemessen, am 2026-10-10.
import { describe, expect, it } from 'vitest'

import { falte, vergleicher } from './vergleich'

const geordnet = (sprache: string, werte: string[]) =>
  [...werte].sort(vergleicher(sprache)).join(' ')

describe('vergleicher', () => {
  it('ordnet Japanisch wie Japanisch — Katakana vor Hiragana desselben Lauts', () => {
    expect(geordnet('ja', ['か', 'ア', 'あ', 'カ'])).toBe('ア あ か カ')
    expect(geordnet('en', ['か', 'ア', 'あ', 'カ'])).toBe('あ ア か カ')
  })

  it('ordnet Chinesisch nach Pinyin und nicht nach Codepunkt', () => {
    expect(geordnet('zh', ['中', '安', '北'])).toBe('安 北 中')
    expect(geordnet('en', ['中', '安', '北'])).toBe('中 北 安')
  })

  it('ordnet im Russischen Kyrillisch vor Lateinisch', () => {
    expect(geordnet('ru', ['b', 'б', 'a', 'а'])).toBe('а б a b')
    expect(geordnet('en', ['b', 'б', 'a', 'а'])).toBe('a b а б')
  })

  it('legt im Deutschen und Englischen »ä« neben »a«', () => {
    for (const sprache of ['de', 'en']) {
      expect(geordnet(sprache, ['Zebra', 'Äpfel', 'apfel'])).toBe(
        'apfel Äpfel Zebra',
      )
    }
  })

  it('gibt je Sprache denselben Vergleicher zurück', () => {
    expect(vergleicher('ja')).toBe(vergleicher('ja'))
  })
})

describe('falte', () => {
  // **Die Sprache des Fensters, nicht die des Systems**: im Türkischen wird
  // aus »I« ein »ı«, und eine Suche nach »INFO« fände »info« nicht. Eine
  // türkische Oberfläche gibt es nicht — ein türkisches System schon.
  it('schreibt klein in der genannten Sprache', () => {
    expect(falte('INFO', 'de')).toBe('info')
    expect(falte('I', 'tr')).toBe('ı')
    expect(falte('I', 'de')).toBe('i')
  })
})
