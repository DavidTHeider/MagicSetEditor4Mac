// Wächter: ein Element mit `role="alert"` trägt keine Kennung (#1989).
//
// # Warum
//
// **Eine Kennung ist dazu da, dass `aria-describedby` auf sie zeigt** — und
// zeigt es auf ein Element mit `role="alert"`, liest VoiceOver in WKWebView
// die Beschreibung **nicht** vor. Gemessen am 2026-10-01 in der Abnahme-VM
// (macOS 26.4, #1871) mit einer eigenen WKWebView-Probe:
//
// | Bauform                                                | angesagt | gelesen |
// | ------------------------------------------------------ | -------- | ------- |
// | das Ziel trägt `role="alert"`                           | ja       | nein    |
// | eine Hülle trägt `role="alert"`, das Ziel darin keine  | ja       | ja      |
// | das Ziel trägt `aria-live`                             | nein     | ja      |
//
// „Angesagt" heißt: beim Verlassen des Felds, wenn der Befund erscheint;
// „gelesen" heißt: beim Wiederbetreten, über `aria-describedby`. Nur die
// Hülle kann beides. Am Setnamen sagte VoiceOver vorher nur „Name of the set,
// invalid data, edit text".
//
// # Was dieser Wächter nicht prüft
//
// - Eine Kennung, die über eine **Ausbreitung** (`{...verbindung}`) kommt —
//   er liest Attribute, keine Werte.
// - `role="status"` und `role="log"`: nicht gemessen, und deshalb hier nicht
//   behauptet.
// - Ob die Hülle wirklich angesagt wird: das hört nur VoiceOver
//   (`docs/HANDPRUEFUNGEN.md`).
import { readFileSync } from 'node:fs'
import { join, relative } from 'node:path'

import { dateienUnter } from './test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

// alarmMitKennung gibt die Stellen, an denen ein Element `role="alert"` und
// `id` zugleich trägt.
function alarmMitKennung(datei: string, text: string): number[] {
  const quelle = ts.createSourceFile(
    datei,
    text,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  )
  const zeilen: number[] = []

  const besuchen = (knoten: ts.Node) => {
    if (ts.isJsxOpeningElement(knoten) || ts.isJsxSelfClosingElement(knoten)) {
      let alarm = false
      let kennung = false

      for (const eigenschaft of knoten.attributes.properties) {
        if (!ts.isJsxAttribute(eigenschaft)) {
          continue
        }

        const name = eigenschaft.name.getText(quelle)
        const wert = eigenschaft.initializer

        if (
          name === 'role' &&
          wert !== undefined &&
          ts.isStringLiteral(wert) &&
          wert.text === 'alert'
        ) {
          alarm = true
        }

        if (name === 'id') {
          kennung = true
        }
      }

      if (alarm && kennung) {
        zeilen.push(
          quelle.getLineAndCharacterOfPosition(knoten.getStart(quelle)).line +
            1,
        )
      }
    }

    ts.forEachChild(knoten, besuchen)
  }

  besuchen(quelle)

  return zeilen
}

describe('ein Alarm trägt keine Kennung', () => {
  it('in keiner Komponente', () => {
    const funde = dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx'],
      absolut: true,
    }).flatMap((datei) =>
      alarmMitKennung(datei, readFileSync(datei, 'utf8')).map(
        (zeile) => `${relative(quellen, datei)}:${String(zeile)}`,
      ),
    )

    expect(funde).toEqual([])
  })

  // **Die Gegenprobe gehört dazu**: ein Wächter, der nichts findet, weil er
  // nichts sieht, ist grün.
  it('findet die Bauform, die er abweist', () => {
    const probe = [
      'export function Probe() {',
      '  return (',
      '    <div>',
      '      <span role="alert">',
      '        <span id="gut">ohne Rolle</span>',
      '      </span>',
      '      <span id="schlecht" role="alert" />',
      '    </div>',
      '  )',
      '}',
    ].join('\n')

    expect(alarmMitKennung('probe.tsx', probe)).toEqual([7])
  })
})
