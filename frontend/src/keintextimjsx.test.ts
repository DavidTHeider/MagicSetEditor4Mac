// Wächter: kein Oberflächentext steht in einer Komponente (AGENTS.md §9).
//
// Die Texte gehören in die Kataloge unter `src/sprache/de/` und `en/`. Ein Text
// im JSX ist ein Befund — nicht, weil er hässlich wäre, sondern weil er die
// zweite Sprache unmöglich macht, ohne dass es jemand merkt: er wird einfach
// nie übersetzt, und niemand sieht eine Lücke.
//
// Gelesen wird über die TypeScript-API, nicht über reguläre Ausdrücke. Ein
// Ausdruck über Quelltext findet `<p>Hallo</p>` und übersieht
// `<p>{"Hallo"}</p>`, `title="Hallo"` und `` {`${n} Karten`} `` — und genau die
// drei sind die Wege, auf denen ein Text zurückkommt, nachdem man den ersten
// geschlossen hat.
//
// # Was dieser Wächter nicht prüft
//
// Er sieht den **Quelltext**, nicht die Laufzeit. Ein Text, der in einer
// TypeScript-Datei ohne JSX steht und von dort in eine Komponente gereicht
// wird, kommt hier nicht vor. Dagegen hilft `sprache.test.ts`: was im deutschen
// Katalog steht, muss im englischen stehen — ein Text außerhalb der Kataloge
// hat dort keine Entsprechung und fällt beim Sprachwechsel auf.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { dateienUnter } from './test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

// beschriftendeEigenschaften sind die Attribute, deren Wert ein Mensch liest.
//
// `className` und `role` stehen bewusst nicht dabei: sie sind für die Maschine.
// `alt` und `placeholder` schon — sie erscheinen auf dem Bildschirm oder im
// Screenreader, und ein unübersetztes `alt` ist eine unübersetzte Seite.
const beschriftendeEigenschaften = new Set([
  'alt',
  'aria-description',
  'aria-label',
  'aria-placeholder',
  'aria-roledescription',
  'aria-valuetext',
  'placeholder',
  'title',
])

// erlaubteZeichenketten sind die benannten Ausnahmen.
//
// Tastennamen sind keine Prosa, sondern Werte der DOM-API. Wer `'Escape'`
// übersetzte, bräche die Tastaturbedienung — und zwar lautlos.
const erlaubteZeichenketten = new Set([
  'Alt',
  'ArrowDown',
  'ArrowLeft',
  'ArrowRight',
  'ArrowUp',
  'Control',
  'Enter',
  'Escape',
  'Meta',
  'Shift',
  'Tab',
  ' ',
])

interface Befund {
  datei: string
  zeile: number
  text: string
  art: string
}

describe('kein Oberflächentext im JSX', () => {
  it('findet in keiner Komponente einen Text', () => {
    const dateien = dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx', 'test/**', 'sprache/de/**', 'sprache/en/**'],
    })

    expect(dateien.length).toBeGreaterThan(0)

    const befunde = dateien.flatMap((datei) => pruefe(datei))

    expect(befunde.map(beschreiben)).toEqual([])
  })
})

function beschreiben(befund: Befund): string {
  return `${befund.datei}:${String(befund.zeile)} — ${befund.art}: ${JSON.stringify(befund.text)}`
}

function pruefe(datei: string): Befund[] {
  const quelle = ts.createSourceFile(
    datei,
    readFileSync(join(quellen, datei), 'utf8'),
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  )

  const befunde: Befund[] = []

  const melde = (knoten: ts.Node, text: string, art: string): void => {
    if (!istProsa(text)) {
      return
    }

    befunde.push({
      datei,
      zeile:
        quelle.getLineAndCharacterOfPosition(knoten.getStart(quelle)).line + 1,
      text,
      art,
    })
  }

  const gehe = (knoten: ts.Node): void => {
    if (ts.isJsxText(knoten)) {
      melde(knoten, knoten.text.trim(), 'Textknoten')
    } else if (ts.isJsxAttribute(knoten)) {
      pruefeEigenschaft(knoten, melde)
    } else if (ts.isJsxExpression(knoten) && knoten.expression !== undefined) {
      pruefeAusdruck(knoten.expression, melde)
    }

    ts.forEachChild(knoten, gehe)
  }

  gehe(quelle)

  return befunde
}

function pruefeEigenschaft(
  eigenschaft: ts.JsxAttribute,
  melde: (knoten: ts.Node, text: string, art: string) => void,
): void {
  const name = eigenschaft.name.getText()
  if (!beschriftendeEigenschaften.has(name)) {
    return
  }

  const wert = eigenschaft.initializer
  if (wert === undefined) {
    return
  }

  if (ts.isStringLiteral(wert)) {
    melde(wert, wert.text, `beschriftende Eigenschaft ${name}`)

    return
  }

  if (ts.isJsxExpression(wert) && wert.expression !== undefined) {
    pruefeAusdruck(wert.expression, melde, `beschriftende Eigenschaft ${name}`)
  }
}

function pruefeAusdruck(
  ausdruck: ts.Node,
  melde: (knoten: ts.Node, text: string, art: string) => void,
  art = 'Zeichenkette im Ausdruck',
): void {
  if (
    ts.isStringLiteral(ausdruck) ||
    ts.isNoSubstitutionTemplateLiteral(ausdruck)
  ) {
    melde(ausdruck, ausdruck.text, art)

    return
  }

  if (ts.isTemplateExpression(ausdruck)) {
    const teile = [
      ausdruck.head.text,
      ...ausdruck.templateSpans.map((span) => span.literal.text),
    ]
    melde(ausdruck, teile.join('…').trim(), `${art} mit Platzhalter`)
  }
}

// istProsa unterscheidet einen Satz von allem anderen.
//
// Drei Merkmale, und alle drei sind nötig gewesen:
//
//   1. Was in `erlaubteZeichenketten` steht, ist eine benannte Ausnahme.
//   2. Was kein Buchstabe ist — Zahlen, Zeichensetzung, Leerraum —, ist kein Text.
//   3. Ein einzelnes Wort ohne Leerzeichen und ganz klein ist ein Bezeichner
//      (`primary`, `sm`), kein Satz. Ein großgeschriebenes Wort dagegen schon:
//      „Karten" ist ein Text, `karten` ist ein Schlüssel.
//
// Die Regel ist absichtlich streng in Richtung Befund: ein falscher Befund
// kostet eine benannte Ausnahme, ein übersehener kostet eine unübersetzte Seite.
function istProsa(text: string): boolean {
  const gestutzt = text.trim()

  if (gestutzt === '' || erlaubteZeichenketten.has(gestutzt)) {
    return false
  }
  if (!/\p{L}/u.test(gestutzt)) {
    return false
  }

  return gestutzt.includes(' ') || /^\p{Lu}/u.test(gestutzt)
}
