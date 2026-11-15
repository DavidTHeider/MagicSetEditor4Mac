// **Hin und zurück in drei Sprachen** — LOC-03.AK-01
// (#1689):
// „de-DE, en-US sowie eine CJK-Locale besitzen Roundtriptests."
//
// # Was ein Rundlauf hier heißt
//
// **Was das Fenster zeigt, liest sich als derselbe Wert zurück.** Wer einen
// Betrag vom Bildschirm abschreibt, bekommt denselben Betrag; wer einen
// Zeitpunkt liest, liest dieselbe Minute. Geprüft wird über `formatToParts`
// und nicht über den ganzen Satz: wo ein Symbol steht und welches
// Leerzeichen davor, entscheidet die Laufzeit (§3).
//
// # Warum Japanisch
//
// **Weil es bei Trennzeichen nichts Neues bringt und woanders alles.**
// Gemessen am 2026-09-22: `ja-JP` schreibt Zahlen wie `en-US`. Anders sind
// die **Reihenfolge** eines Datums (Jahr zuerst), das **breite** Yenzeichen
// (U+FFE5 statt U+00A5) und die **Sekunde** (»12 秒«). Die Ordnung von
// Kanji prüft Go (`internal/domain/sammlungsordnung_test.go`), das CSV
// ebenfalls (`internal/service/sammlungssprachen_test.go`).
import { describe, expect, it } from 'vitest'

import { alsGeld } from './geld'
import { alsProzent, lesarten } from './lokal'
import { waehrungsstellen } from './waehrungsstellen'
import { zeitpunktformat } from './zeitpunkt'

const sprachen = ['de-DE', 'en-US', 'ja-JP']

describe.each(sprachen)('%s', (sprache) => {
  // **Dezimal- und Tausenderzeichen**: was `Intl` schreibt, liest sich
  // zurück — mit Gruppen, mit Nachkommastellen, mit Minus.
  it.each(['0.5', '19.99', '1234', '1234.56', '1234567.89', '-5.25'])(
    'liest die Zahl %s zurück',
    (kanonisch) => {
      const stellen = kanonisch.split('.')[1]?.length ?? 0
      const geschrieben = new Intl.NumberFormat(sprache, {
        minimumFractionDigits: stellen,
        maximumFractionDigits: stellen,
      }).format(Number(kanonisch))

      expect(lesarten(geschrieben, sprache, stellen)).toEqual({
        art: 'eindeutig',
        kanonisch,
      })
    },
  )

  // **Geld**: der Betrag, den das Fenster zeigt, ohne sein Zeichen
  // abgeschrieben, ist derselbe — in den Stellen seiner Währung.
  it.each([
    [1999, 'EUR', '19.99'],
    [1999, 'USD', '19.99'],
    [150, 'JPY', '150'],
    [123456, 'JPY', '123456'],
    [1234, 'BHD', '1.234'],
  ])('liest %i %s als %s zurück', (wert, waehrung, kanonisch) => {
    const gezeigt = alsGeld(wert, waehrung, sprache)
    const stellen = waehrungsstellen(waehrung)

    expect(lesarten(ohneZeichen(gezeigt), sprache, stellen)).toEqual({
      art: 'eindeutig',
      kanonisch,
    })
  })

  // **Datum und Uhrzeit**: aus dem, was das Fenster zeigt, kommen Jahr,
  // Monat, Tag, Stunde und Minute zurück — auch dort, wo der Monat ein Wort
  // ist und die Stunde ein »PM« trägt.
  it('liest einen Zeitpunkt auf die Minute zurück', () => {
    const wann = new Date(2026, 8, 22, 14, 5)

    expect(zurueckgelesen(sprache, wann)).toEqual({
      jahr: 2026,
      monat: 8,
      tag: 22,
      stunde: 14,
      minute: 5,
    })
  })

  // **Prozent**: die ganze Zahl kommt zurück, das Zeichen steht, wo die
  // Sprache es hinsetzt.
  it.each([0, 0.07, 0.42, 1, 1.5])('liest %s als Prozent zurück', (anteil) => {
    const teile = new Intl.NumberFormat(sprache, {
      style: 'percent',
      maximumFractionDigits: 0,
    }).formatToParts(anteil)

    expect(alsProzent(anteil, sprache)).toBe(
      teile.map((teil) => teil.value).join(''),
    )
    expect(Number(teil(teile, 'integer')) / 100).toBe(anteil)
  })

  // **Einheiten**: Millisekunden und Megabyte schreibt jede der drei
  // Sprachen gleich — die Kataloge sagen »{ms} ms« in beiden, und das
  // stimmt.
  it.each([
    ['millisecond', '250 ms'],
    ['megabyte', '250 MB'],
  ])('schreibt %s als %s', (einheit, erwartet) => {
    const geschrieben = new Intl.NumberFormat(sprache, {
      style: 'unit',
      unit: einheit,
    }).format(250)

    expect(geschrieben.replace(/\s/gu, ' ')).toBe(erwartet)
  })
})

// **Die Sekunde ist keine Ausnahme von der Regel, sondern eine Wahl.**
// `Intl` schreibt »12 Sek.«, »12 sec« und »12 秒«; die Kataloge schreiben in
// beiden Sprachen das SI-Zeichen »s« — und das ist in jeder Sprache
// richtig. Dieser Lauf hält fest, dass es eine Wahl ist und kein Versehen.
it('schreibt die Sekunde in jeder Sprache anders als das SI-Zeichen', () => {
  const geschrieben = sprachen.map((sprache) =>
    new Intl.NumberFormat(sprache, { style: 'unit', unit: 'second' })
      .format(12)
      .replace(/\s/gu, ' '),
  )

  expect(geschrieben).toEqual(['12 Sek.', '12 sec', '12 秒'])
})

// ohneZeichen nimmt aus einem gezeigten Betrag alles, was keine Zahl ist.
function ohneZeichen(gezeigt: string): string {
  return gezeigt.replace(/[^\d.,'\-\s]/gu, '').trim()
}

// teil gibt den Text eines Teils.
function teil(teile: Intl.NumberFormatPart[], art: string): string {
  return teile
    .filter((stueck) => stueck.type === art)
    .map((stueck) => stueck.value)
    .join('')
}

// zurueckgelesen liest die Teile eines gezeigten Zeitpunkts in Zahlen
// zurück — über Tabellen, die **dieselbe** Sprache geschrieben hat. So
// liest sich »Sep« und »PM« ohne eine Liste von Monatsnamen daneben.
function zurueckgelesen(sprache: string, wann: Date) {
  const format = zeitpunktformat(sprache)
  const text = (datum: Date, art: Intl.DateTimeFormatPartTypes) =>
    format
      .formatToParts(datum)
      .filter((stueck) => stueck.type === art)
      .map((stueck) => stueck.value)
      .join('')

  const monat = text(wann, 'month')
  const stunde = text(wann, 'hour') + text(wann, 'dayPeriod')

  return {
    jahr: Number(text(wann, 'year')),
    monat: [...Array(12).keys()].find(
      (m) => text(new Date(2026, m, 15, 12), 'month') === monat,
    ),
    tag: Number(text(wann, 'day')),
    stunde: [...Array(24).keys()].find((h) => {
      const probe = new Date(2026, 0, 15, h)

      return text(probe, 'hour') + text(probe, 'dayPeriod') === stunde
    }),
    minute: Number(text(wann, 'minute')),
  }
}
