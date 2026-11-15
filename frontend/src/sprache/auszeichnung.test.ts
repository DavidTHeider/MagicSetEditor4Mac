// Wächter: eine Auszeichnung im Katalog kommt auch an
// (Nachlauf zum Zweitblick über die Phasen 4.0 bis 4.13, 2026-09-07).
//
// # Der Befund
//
// Auf der Seite „Nutzer und Profile" stand wörtlich:
//
//     Adresse, Rufnummer und Kennwort stehen unter *Konto*.
//
// **Mit Sternchen.** `Satz` setzt `**fett**` und `` `code` `` und sonst
// nichts — „ein einzelnes Sternchen ist ein Sternchen", so entschieden in
// #699. Wer
// im Katalog `*Konto*` schreibt, meint eine Hervorhebung und bekommt zwei
// Zeichen, die niemand erklärt.
//
// **Es waren 17 Stellen in beiden Sprachen**, nicht eine — und keine davon
// hätte ein Test gefunden: die Kataloge werden gegen**einander** gehalten
// (`sprache.test.ts`), und ein Sternchen, das in beiden steht, fällt dabei
// nicht auf.
//
// # Was er hält und was nicht
//
//   - **Er hält**: kein Katalogtext trägt ein einzelnes `*`-Paar.
//   - **Er hält nicht**, ob die Hervorhebung sinnvoll ist. `**alles fett**`
//     geht durch; das sieht nur ein Mensch (`AGENTS.md` §10).
//   - **Und er sieht Zeichenketten, keine Kommentare.** In einem Kommentar ist
//     ein Sternchen ein Sternchen und stört niemanden.
import { readFileSync } from 'node:fs'

import { dateienUnter } from '../test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

// einzelnesPaar trifft `*x*`, aber nicht `**x**` und nicht `**x** … **y**`.
//
// **Die Randbedingungen sind der ganze Ausdruck**: ohne sie fände er in
// `**Anrede** ist ein freies Feld` das Stück zwischen den inneren Sternchen.
const einzelnesPaar = /(?<!\*)\*(?!\*)([^*`\n]{1,60})\*(?!\*)/

// katalogtexte sammelt jede Zeichenkette eines Katalogs mit ihrer Zeile.
function katalogtexte(datei: string): { zeile: number; text: string }[] {
  const quelle = readFileSync(datei, 'utf8')

  const baum = ts.createSourceFile(
    datei,
    quelle,
    ts.ScriptTarget.Latest,
    /* setParentNodes */ true,
  )

  const gefunden: { zeile: number; text: string }[] = []

  const gehe = (knoten: ts.Node): void => {
    if (
      ts.isStringLiteral(knoten) ||
      ts.isNoSubstitutionTemplateLiteral(knoten)
    ) {
      const { line } = baum.getLineAndCharacterOfPosition(knoten.getStart(baum))

      gefunden.push({ zeile: line + 1, text: knoten.text })
    }

    ts.forEachChild(knoten, gehe)
  }

  gehe(baum)

  return gefunden
}

describe('die Auszeichnung der Kataloge', () => {
  it('kennt kein einzelnes Sternchenpaar', () => {
    const dateien = dateienUnter(import.meta.dirname, '**/*.ts', {
      ohne: ['**/*.test.ts'],
      absolut: true,
    })

    // **Die Gegenprobe zur Messung**: findet sie keine Datei, ist sie grün
    // und sagt nichts (§10).
    expect(dateien.length).toBeGreaterThan(10)

    const befunde: string[] = []

    for (const datei of dateien) {
      for (const { zeile, text } of katalogtexte(datei)) {
        const treffer = einzelnesPaar.exec(text)
        if (treffer) {
          const kurz = datei.slice(datei.indexOf('sprache'))

          befunde.push(`${kurz}:${String(zeile)} — *${treffer[1] ?? ''}*`)
        }
      }
    }

    expect(
      befunde,
      'diese Texte tragen ein einzelnes Sternchenpaar; `Satz` setzt nur ' +
        '`**fett**`, und ein einzelnes Sternchen erscheint wörtlich:\n' +
        befunde.join('\n'),
    ).toEqual([])
  })
})
