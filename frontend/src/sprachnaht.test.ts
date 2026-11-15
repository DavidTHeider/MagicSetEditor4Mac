// **Keine Stelle ordnet oder faltet ohne Sprache** — F01
// (#2243).
//
// `localeCompare` mit einem Argument und `toLocaleLowerCase()` ohne eines
// nehmen die Sprache des **Systems**. Für Menschentext gilt die des Fensters
// (`sprache/vergleich.ts`), für technische Schlüssel keine (`ordnung.ts`).
// Bis F01 standen 25 solche Stellen in 13 Dateien; „Was noch offen ist" in
// `docs/LOKALISIERUNG.md` nannte sie, bis eine dritte Sprache käme — es kamen
// sechs.
//
// # Warum der Syntaxbaum und keine Regex
//
// **Der erste Anlauf suchte mit einer Regex** — und übersah, was nach Regex
// aussah und keines war (F01 im Review zu J05,
// #2269):
// `toLocaleLowerCase( )` mit Leerraum, eine leere Klammer über mehrere Zeilen,
// `localeCompare(f(g(b)))` — eine Regex zählt Klammern nur bis zu einer festen
// Tiefe —, und umgekehrt fand sie `localeCompare(b)` in einem Kommentar. Jetzt
// wird die Datei zerlegt und **gezählt, wie viele Argumente der Aufruf hat**:
// `localeCompare` braucht zwei, die beiden `toLocale…Case` eines. Ein
// `undefined` an der Stelle der Sprache zählt nicht — es nimmt ebenfalls die
// des Systems.
//
// **Was er nicht hält**: ob eine Stelle die richtige Sprache nennt. Er sieht,
// dass eine dasteht, nicht welche. **Und was er nicht sehen kann**: ein Aufruf
// mit Spread (`a.localeCompare(...rest)`) — wie viele Argumente er trägt, weiß
// erst die Laufzeit, und im Zweifel gilt er als gedeckt —, ein Aufruf über
// `.call` oder `.apply` und eine Methode, die unter anderem Namen weitergereicht
// wird.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import ts from 'typescript'
import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'

const quellen = join(import.meta.dirname)

// zuMelden sind die Gestalten, in denen eine Sprache fehlt — **auch die, die
// die Regex des ersten Anlaufs übersah** (F01, #2269): Leerraum in der
// leeren Klammer, eine Klammer über mehrere Zeilen, eine Verschachtelung
// tiefer als eine Ebene.
const zuMelden: [string, string][] = [
  ['ein Argument', 'a.name.localeCompare(b.name)'],
  ['eine verschachtelte Klammer', 'a.localeCompare(f(b))'],
  ['eine Klammer, zwei Ebenen tief', 'a.localeCompare(f(g(b)))'],
  ['ein Komma nur in der Verschachtelung', 'a.localeCompare(f(b, c))'],
  ['eine Klammer in einer Zeichenkette', "a.localeCompare('(')"],
  ['ein Argument über mehrere Zeilen', 'a.localeCompare(\n  f(g(b)),\n)'],
  ['kleinschreiben ohne Argument', 'text.toLocaleLowerCase()'],
  ['großschreiben ohne Argument', 'text.toLocaleUpperCase()'],
  ['kleinschreiben mit Leerraum in der Klammer', 'text.toLocaleLowerCase( )'],
  [
    'großschreiben mit Zeilenumbruch in der Klammer',
    'text.toLocaleUpperCase(\n)',
  ],
  ['einen Umbruch vor dem Punkt', 'a\n  .localeCompare(b)'],
  ['den verketteten Zugriff', 'a?.localeCompare(b)'],
  ['den verketteten Aufruf', 'a.localeCompare?.(b)'],
  ['den Zugriff über eine Zeichenkette', "a['localeCompare'](b)"],
  ['`undefined` als Sprache beim Ordnen', 'a.localeCompare(b, undefined)'],
  [
    '`undefined` als Sprache beim Kleinschreiben',
    'text.toLocaleLowerCase(undefined)',
  ],
  // **Geklammert** (Befund 8 des Nachlaufs über J05, #2269): die Klammer
  // ändert nichts am Aufruf, aber am Baum.
  ['den geklammerten Aufruf', '(text.toLocaleLowerCase)()'],
  ['die geklammerte Methode beim Ordnen', '(a.localeCompare)(b)'],
  ['geklammertes `undefined` als Sprache', 'a.localeCompare(b, (undefined))'],
]

// durchzulassen sind die Gestalten, die eine Sprache nennen — und die, die
// gar kein Aufruf sind: **ein Kommentar und eine Zeichenkette sprechen über
// `localeCompare`, ohne es zu rufen** (das konnte die Regex nicht trennen).
const durchzulassen: [string, string][] = [
  ['die Sprache als zweites Argument', 'a.localeCompare(b, sprache)'],
  [
    'eine verschachtelte Klammer samt Sprache',
    'a.localeCompare(f(g(b)), sprache)',
  ],
  ['Sprache und Wahl', "a.localeCompare(b, 'de', { numeric: true })"],
  [
    'die Sprache über mehrere Zeilen',
    'a.localeCompare(\n  f(b, c),\n  sprache,\n)',
  ],
  ['kleinschreiben mit Sprache', 'text.toLocaleLowerCase(sprache)'],
  [
    'kleinschreiben mit Leerraum um die Sprache',
    'text.toLocaleLowerCase( sprache )',
  ],
  ['großschreiben mit Sprache', 'text.toLocaleUpperCase(sprache)'],
  [
    'einen Spread, der die Sprache tragen kann',
    'a.localeCompare(...argumente)',
  ],
  ['eine Methode mit anderem Namen', 'a.meinLocaleCompare(b)'],
  ['einen Kommentar', '// a.localeCompare(b) und text.toLocaleLowerCase()'],
  ['einen Blockkommentar', '/* text.toLocaleUpperCase() */ const a = 1'],
  ['eine Zeichenkette', "const s = 'a.localeCompare(b)'"],
  ['eine Schablone ohne Aufruf', 'const s = `a.localeCompare(b)`'],
]

// sprachstelle nennt je Methode, das wievielte Argument die Sprache ist
// (von null an).
const sprachstelle = new Map([
  ['localeCompare', 1],
  ['toLocaleLowerCase', 0],
  ['toLocaleUpperCase', 0],
])

// ohneKlammern nimmt Klammern um einen Ausdruck ab — `(a.name)(…)` ruft
// dasselbe wie `a.name(…)` (Befund 8 des Nachlaufs über J05, #2269).
function ohneKlammern(ausdruck: ts.Expression): ts.Expression {
  let innen = ausdruck
  while (ts.isParenthesizedExpression(innen)) {
    innen = innen.expression
  }

  return innen
}

// methodenname gibt den Namen, unter dem ein Aufruf eine Methode ruft —
// `a.name(…)`, `a?.name(…)` und `a['name'](…)`, auch geklammert.
function methodenname(geklammert: ts.Expression): string | undefined {
  const ausdruck = ohneKlammern(geklammert)

  if (ts.isPropertyAccessExpression(ausdruck)) {
    return ausdruck.name.text
  }

  if (
    ts.isElementAccessExpression(ausdruck) &&
    ts.isStringLiteralLike(ausdruck.argumentExpression)
  ) {
    return ausdruck.argumentExpression.text
  }

  return undefined
}

// ohneSprache sagt, ob an der Stelle der Sprache nichts steht — oder
// `undefined`, das dasselbe heißt.
function ohneSprache(aufruf: ts.CallExpression, stelle: number): boolean {
  // **Im Zweifel gedeckt.** Ein Spread kann beliebig viele Argumente tragen;
  // eine Meldung wäre hier geraten.
  if (aufruf.arguments.some((argument) => ts.isSpreadElement(argument))) {
    return false
  }

  const roh = aufruf.arguments[stelle]
  const genannt = roh === undefined ? undefined : ohneKlammern(roh)

  return (
    genannt === undefined ||
    (ts.isIdentifier(genannt) && genannt.text === 'undefined')
  )
}

// fundeIn gibt die Aufrufe, denen die Sprache fehlt, mit ihrer Zeile.
//
// **`datei` bestimmt die Art des Parsers**: ohne `.tsx` im Namen läse er
// `<Komponente />` als Typargument, und jede Komponente wäre ein Fehler.
function fundeIn(text: string, datei = 'beispiel.ts'): string[] {
  const quelle = ts.createSourceFile(datei, text, ts.ScriptTarget.Latest, true)
  const funde: string[] = []

  const gehe = (knoten: ts.Node) => {
    if (ts.isCallExpression(knoten)) {
      const name = methodenname(knoten.expression)
      const stelle = name === undefined ? undefined : sprachstelle.get(name)

      if (stelle !== undefined && ohneSprache(knoten, stelle)) {
        const { line } = quelle.getLineAndCharacterOfPosition(
          knoten.getStart(quelle),
        )

        funde.push(
          `Zeile ${String(line + 1)}: ${knoten.getText(quelle).replace(/\s+/g, ' ')}`,
        )
      }
    }

    ts.forEachChild(knoten, gehe)
  }

  gehe(quelle)

  return funde
}

describe('die Sprachnaht', () => {
  // **Die Gegenprobe steht im Wächter**: er findet, was er finden soll, und
  // lässt durch, was eine Sprache nennt.
  it.each(zuMelden)('meldet %s', (_name, quelle) => {
    expect(fundeIn(quelle)).toHaveLength(1)
  })

  it.each(durchzulassen)('lässt %s durch', (_name, quelle) => {
    expect(fundeIn(quelle)).toHaveLength(0)
  })

  // **Eine Komponente ist keine Skriptdatei**: ohne den Dateinamen läse der
  // Parser den Apostroph im Text als Beginn einer Zeichenkette, und der
  // Aufruf dahinter ginge unter. Nachgemessen: als `.ts` gelesen findet
  // dieselbe Zeile nichts.
  it('liest einen Aufruf hinter einem Apostroph im Text einer Komponente', () => {
    const komponente = "const a = <p>don't {t.toLocaleLowerCase()}</p>"

    expect(fundeIn(komponente, 'beispiel.tsx')).toHaveLength(1)
    expect(fundeIn(komponente, 'beispiel.ts')).toHaveLength(0)
  })

  // **Der Fund nennt die Zeile**: ein Wächter, der nur „irgendwo in der
  // Datei" sagt, lässt den Suchenden mit der Regex allein.
  it('nennt die Zeile des Aufrufs', () => {
    expect(fundeIn('const a = 1\n\nconst b = x.localeCompare(y)')).toEqual([
      'Zeile 3: x.localeCompare(y)',
    ])
  })

  it('findet im Fenster keinen Aufruf ohne Sprache', () => {
    const dateien = dateienUnter(quellen, '**/*.{ts,tsx}', {
      ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**'],
    })
    expect(dateien.length).toBeGreaterThan(100)

    const funde = dateien.flatMap((datei) =>
      fundeIn(readFileSync(join(quellen, datei), 'utf8'), datei).map(
        (fund) => `${datei}: ${fund}`,
      ),
    )

    expect(funde).toEqual([])
  })
})
