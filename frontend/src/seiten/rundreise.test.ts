// **Die Rundreise über viele Formen und viele Zeichen** —
// #718.
//
// # Wofür
//
// `mitauszeichnung.test.ts` prüft benannte Fälle: einen Aromatext, einen
// Trenner, eine Schachtelung. Der Zweitblick durch Codex hat am 2026-08-21
// etwas anderes getan — **jede Auszeichnungsform gegen jedes Sonderzeichen**,
// 230 Rundreisen. 46 davon scheiterten, alle an denselben zwei Zeichen.
//
// Ein Wächter, der nur die Fälle kennt, an die jemand gedacht hat, findet so
// etwas nicht. Dieser hier fährt das Kreuzprodukt.
//
// # Was er zusagt
//
// Was jemand tippt, kommt zurück — **ohne die beiden Zeichen, die MSE für sich
// behält**. `ESCAPED_LANGLE` und `CONNECTION_SPACE` sind in einem Kartentext
// nichts, und kodieren ließen sie sich nicht, ohne eine Darstellung zu
// erfinden, die kein anderes Programm liest (`src/util/tagged_string.cpp`,
// Zeilen 14–22, Fassung f9d9356 vom 2021-07-06).
import { describe, expect, it } from 'vitest'

import { GESCHUETZT, VERBINDUNGSLEERZEICHEN } from './marken'
import { mitAuszeichnung } from './mitauszeichnung'
import { ohneAuszeichnung } from './ohneauszeichnung'

// formen sind die Auszeichnungen, die in einem Kartenwert vorkommen.
//
// **Aus MSE2s Textdarstellung** (`src/render/text/element.cpp`, Zeilen 75–209)
// und aus dem, was `magic.mse-game` schreibt.
const formen = [
  'INHALT',
  '<i>INHALT</i>',
  '<b>INHALT</b>',
  '<sym>INHALT</sym>',
  '<i-flavor>INHALT</i-flavor>',
  '<i-auto>INHALT</i-auto>',
  '<word-list-type-en>INHALT</word-list-type-en>',
  '<kw-a>INHALT</kw-a>',
  '<atom-param>INHALT</atom-param>',
  '<soft>INHALT</soft>',
  '<line>INHALT</line>',
  'davor<sep> — </sep>INHALT',
  'davor<sep-soft> — </sep-soft>INHALT',
  '<b><i>INHALT</i></b>',
  '<i>vor</i>INHALT<i>nach</i>',
  '<i-flavor>INHALT</i-flavor><sep><line>\n</line></sep>dahinter',
]

// zeichen sind die Sonderfälle, die in einem getippten Text stehen können.
const zeichen = [
  'x',
  '<',
  '>',
  '&',
  '"',
  '\n',
  '\t',
  'é',
  '🜂',
  GESCHUETZT,
  VERBINDUNGSLEERZEICHEN,
]

// ohneReservierte nimmt heraus, was MSE für sich behält.
function ohneReservierte(text: string): string {
  return text.replaceAll(GESCHUETZT, '').replaceAll(VERBINDUNGSLEERZEICHEN, '')
}

describe('Die Rundreise über das Kreuzprodukt', () => {
  for (const form of formen) {
    for (const [stelle, sonder] of zeichen.entries()) {
      it(`${form.slice(0, 24)} mit Zeichen ${String(stelle)}`, () => {
        const roh = form.replace('INHALT', 'alt')
        const getippt = ohneAuszeichnung(roh).replace('alt', 'a' + sonder + 'b')

        const zurueck = ohneAuszeichnung(mitAuszeichnung(roh, getippt))

        expect(zurueck).toBe(ohneReservierte(getippt))
      })
    }
  }
})
