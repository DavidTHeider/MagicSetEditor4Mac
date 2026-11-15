// Kein Satz in einem `Error` — er wäre unübersetzbar.
//
// # Die Lücke, die dieser Wächter schließt
//
// #133:
// `keintextimjsx.test.ts` liest **JSX**. Ein `new Error('Die Vorschau kam ohne
// Bild zurück.')` daneben ist ihm unsichtbar — und der Satz erreicht über
// `meldungAus` genauso die Oberfläche wie jeder Text, den es abweist. Zwei
// solche Sätze standen im Baum, und beide waren deutsch.
//
// **Geprüft wird die Form, nicht die Sprache.** Ein Wächter, der nach Umlauten
// suchte, ließe „The preview came back empty." durch — und das ist derselbe
// Fehler mit einer anderen Sprache. Verlangt ist deshalb: ein `Error` bekommt
// **keine** Zeichenkette. Wer eine Meldung braucht, wirft einen `Katalogfehler`
// mit seinem Schlüssel.
//
// **Über den Syntaxbaum und nicht über einen Ausdruck.** Der erste Anlauf las
// den rohen Text und schlug an dieser Datei selbst an — an einem Beispiel in
// einem Kommentar. Ein Wächter, der Kommentare für Code hält, wird umgangen,
// indem jemand ihn abschaltet.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { dateienUnter } from './test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

// ausnahme erlaubt einen Satz — mit Grund, in der Datei, für den Wächter lesbar.
//
// **Wie `.covignore`, und aus demselben Grund** (AGENTS.md §3): eine Ausnahme,
// die nur in Prosa steht, kann der Wächter nicht lesen — sie ist dann keine
// Ausnahme, sondern ein roter Lauf mit Erklärung.
//
// Erlaubt wird nur, was die Oberfläche **nicht erreichen kann**: ein Fehler, der
// den Start abbricht, hat kein Fenster, in dem er erschiene.
const ausnahme = /\/\/ *keinfehlertext: *\S/

// fehlermitsatz sagt, ob ein Knoten ein `new Error("…")` ist.
function fehlermitsatz(knoten: ts.Node): boolean {
  if (!ts.isNewExpression(knoten) || !ts.isIdentifier(knoten.expression)) {
    return false
  }

  if (knoten.expression.text !== 'Error') {
    return false
  }

  const erstes = knoten.arguments?.[0]

  return (
    erstes !== undefined &&
    (ts.isStringLiteral(erstes) ||
      ts.isTemplateExpression(erstes) ||
      ts.isNoSubstitutionTemplateLiteral(erstes))
  )
}

// saetzeIn gibt die Zeilen, in denen ein Satz in einem `Error` steht.
function saetzeIn(datei: string, inhalt: string): string[] {
  const quelle = ts.createSourceFile(
    datei,
    inhalt,
    ts.ScriptTarget.Latest,
    true,
  )
  const befunde: string[] = []

  const gehe = (knoten: ts.Node) => {
    if (fehlermitsatz(knoten)) {
      const { line } = quelle.getLineAndCharacterOfPosition(knoten.getStart())

      befunde.push(`${datei}:${String(line + 1)}`)
    }

    ts.forEachChild(knoten, gehe)
  }

  gehe(quelle)

  return befunde
}

describe('kein Satz in einem Error', () => {
  // **Ohne Tests.** Ein Test darf `new Error('kaputt')` werfen: er stellt eine
  // Ursache her, die niemand zu sehen bekommt. Ohne diese Ausnahme prüfte der
  // Wächter die Vorrichtung statt den Betrieb.
  it('wirft in keiner Quelldatei einen Satz', () => {
    const dateien = dateienUnter(quellen, '**/*.{ts,tsx}', {
      ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**'],
    })

    // Ohne diese Zeile wäre der Wächter grün, sobald der Pfad nicht mehr
    // stimmt — und niemand sähe es (§10).
    expect(dateien.length).toBeGreaterThan(20)

    const befunde = dateien.flatMap((datei) => {
      const inhalt = readFileSync(join(quellen, datei), 'utf8')

      return ausnahme.test(inhalt) ? [] : saetzeIn(datei, inhalt)
    })

    expect(befunde).toEqual([])
  })

  // **Die Gegenprobe, und sie steht hier statt in einer Notiz.** Ein Wächter,
  // dessen Erkennung nicht mehr greift, ist grün — und grün heißt dann nichts.
  it('erkennt jede Schreibweise', () => {
    for (const zeile of [
      `throw new Error('ein Satz')`,
      `throw new Error("ein Satz")`,
      'throw new Error(`ein Satz ${wert}`)',
      'throw new Error(`ohne Platzhalter`)',
      `Promise.reject(new Error( 'mit Abstand' ))`,
    ]) {
      expect(saetzeIn('probe.ts', zeile)).toHaveLength(1)
    }
  })

  it('lässt gelten, was keinen Satz trägt', () => {
    for (const zeile of [
      `throw new Error()`,
      `throw new Error(ursache)`,
      `throw new Katalogfehler('vorschauLeer')`,
      `throw new Bindungsfehler(antwort.fehler)`,
      `// throw new Error('nur ein Beispiel im Kommentar')`,
    ]) {
      expect(saetzeIn('probe.ts', zeile)).toEqual([])
    }
  })
})
