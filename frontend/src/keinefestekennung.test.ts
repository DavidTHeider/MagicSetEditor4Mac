// Wächter: keine feste DOM-Kennung in einer Komponente
// (#1553).
//
// # Warum
//
// `id="kibatch-ueberschrift"` trägt, solange das Bauteil **einmal** im
// Dokument steht. Steht es zweimal, sind zwei Knoten mit derselben Kennung
// da — `duplicate-id` ist eine axe-Regel, und `aria-labelledby` zeigt dann auf
// den ersten, gleich welcher gemeint war. Ein Bauteil weiß nicht, wie oft es
// aufgestellt wird; deshalb erzeugt React die Kennung (`useId`).
//
// Gefunden hat das kein Wächter, sondern das Bauen eines zweiten Kastens
// (A13.1b, #1452) — also der Zufall, dass jemand dieselbe Vorlage ein zweites
// Mal benutzte.
//
// # Die zwei Ausnahmen, und warum sie welche sind
//
// Eine Kennung, die **von außen** angesprungen wird, ist eine Zusage und darf
// nicht bei jedem Zeichnen anders lauten. Beide gehören zum Rahmen, und der
// steht genau einmal im Dokument:
//
// | Kennung | Wer sie braucht |
// | --- | --- |
// | `inhalt` | das Sprungziel der Sprungmarke — `href="#inhalt"` steht im Markup und wandert in die Adresszeile |
// | `seitenleiste` | `aria-controls` aus `rahmen/Seitenleistenknopf.tsx`, also **über** die Bauteilgrenze |
//
// **Die zweite steht seit dem 2026-09-19 in `rahmen/Seitenleiste.tsx`**
// und nicht mehr in `Rahmen.tsx`: die Leiste ist ein eigenes Bauteil
// geworden, weil die Datei an der Zeilengrenze stand (§2). Der Ort ändert
// sich, die Begründung nicht — und der Ort gehört zur Ausnahme, sonst wäre
// sie eine Erlaubnis für jede Datei.
//
// **Die erste steht seit dem 2026-10-11 in `rahmen/Inhalt.tsx`** — aus
// demselben Grund: `Rahmen.tsx` wuchs mit dem Namen von `<main>` (H01) über
// die Zeilengrenze.
//
// # Was dieser Wächter nicht prüft
//
// **Er sieht nur die Zeichenkette.** Eine Kennung, die aus einem Wert gebaut
// wird — `` id={`marke-${karteID}`} ``, `id={fehlerkennung(feld)}` —, ist nur
// so eindeutig wie ihre Eingabe, und dazu sagt er nichts. Dasselbe für eine
// Kennung, die über `setAttribute` oder aus einem gestreuten Objekt kommt.
//
// **Und er sagt nichts über den Bezug.** Dass eine erzeugte Kennung auch dort
// ankommt, wo `aria-labelledby` sie sucht, prüft er nicht — das prüft
// `barrierefreiheit.test.tsx` mit axe, und zwar nur für die Seiten, die es
// aufstellt.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { dateienUnter } from './test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

// benannteAusnahmen sind die Kennungen, die fest bleiben — mit der Datei, in
// der sie stehen. Eine Ausnahme ohne Ort wäre eine Erlaubnis für jede Datei.
const benannteAusnahmen = new Set([
  'rahmen/Inhalt.tsx: inhalt',
  'rahmen/Seitenleiste.tsx: seitenleiste',
])

describe('keine feste DOM-Kennung im JSX', () => {
  it('erzeugt jede Kennung, die nicht von außen angesprungen wird', () => {
    const dateien = dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx', 'test/**'],
    })

    expect(dateien.length).toBeGreaterThan(0)

    const befunde = dateien
      .flatMap((datei) => feste(datei))
      .filter((befund) => !benannteAusnahmen.has(`${befund.ort}: ${befund.id}`))

    expect(
      befunde.map(
        (befund) => `${befund.ort}:${String(befund.zeile)} — ${befund.id}`,
      ),
    ).toEqual([])
  })

  // **Die Gegenprobe steht im Wächter** (§10): sonst belegt ein grüner Lauf
  // nur, dass niemand hingesehen hat.
  it('findet eine feste Kennung in einem gepflanzten Quelltext', () => {
    expect(
      ausQuelltext(
        'probe.tsx',
        'export const P = () => <section aria-labelledby="k"><h3 id="k">x</h3></section>',
      ).map((befund) => befund.id),
    ).toEqual(['k'])
  })

  // **Und er geht an einer erzeugten Kennung vorbei** — sonst wäre die
  // Abhilfe selbst ein Befund.
  it('geht an einer erzeugten Kennung vorbei', () => {
    expect(
      ausQuelltext('probe.tsx', 'export const P = () => <h3 id={ueber}>x</h3>'),
    ).toEqual([])
  })
})

// ------------------------------------------------------------------ Hilfsmittel

interface Befund {
  ort: string
  zeile: number
  id: string
}

function feste(datei: string): Befund[] {
  return ausQuelltext(datei, readFileSync(join(quellen, datei), 'utf8'))
}

// ausQuelltext liest die festen `id`-Angaben eines TSX-Quelltextes.
//
// **Über die TypeScript-API und nicht über einen Ausdruck**: `id="x"` und
// `id={'x'}` sind dasselbe, und ein regulärer Ausdruck fände nur das erste
// — dieselbe Überlegung wie in `keintextimjsx.test.ts`.
function ausQuelltext(datei: string, text: string): Befund[] {
  const quelle = ts.createSourceFile(
    datei,
    text,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TSX,
  )

  const befunde: Befund[] = []

  const gehe = (knoten: ts.Node): void => {
    if (ts.isJsxAttribute(knoten) && knoten.name.getText(quelle) === 'id') {
      const wert = festerWert(knoten.initializer)

      if (wert !== undefined) {
        befunde.push({
          ort: datei,
          zeile:
            quelle.getLineAndCharacterOfPosition(knoten.getStart(quelle)).line +
            1,
          id: wert,
        })
      }
    }

    ts.forEachChild(knoten, gehe)
  }

  gehe(quelle)

  return befunde
}

// festerWert gibt die Zeichenkette, wo eine dasteht — sonst nichts.
function festerWert(initialisierung: ts.Node | undefined): string | undefined {
  if (initialisierung === undefined) {
    return undefined
  }

  if (ts.isStringLiteral(initialisierung)) {
    return initialisierung.text
  }

  if (
    ts.isJsxExpression(initialisierung) &&
    initialisierung.expression !== undefined &&
    (ts.isStringLiteral(initialisierung.expression) ||
      ts.isNoSubstitutionTemplateLiteral(initialisierung.expression))
  ) {
    return initialisierung.expression.text
  }

  return undefined
}
