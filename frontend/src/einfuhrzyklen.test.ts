// Wächter: kein Einfuhrzyklus in der Oberfläche — ENG-03
// (#1728).
//
// # Warum
//
// **Ein Zyklus macht die Reihenfolge zufällig, in der Module ausgewertet
// werden.** Liest `a.ts` beim Laden eine Konstante aus `b.ts`, und `b.ts`
// lädt `a.ts`, dann hängt es davon ab, wer zuerst eingeführt wird, ob die
// Konstante schon dasteht — und ein Test, der eine Datei allein lädt, sieht
// eine andere Reihenfolge als das Bündel. Das bricht nicht sofort; es bricht
// beim nächsten Umbau an einer Stelle, die nichts damit zu tun hat.
//
// **Gezählt am 2026-09-24 waren es drei**, alle aus Teilungen an der
// Zeilengrenze: `tischgeometrie.ts` las `Stapelbreite` aus `tischplan.ts`,
// das `tischgeometrie.ts` liest; drei Seiten lasen `bildwunsch` über die
// Wiederausfuhr in `Partietisch.tsx`, das sie einbindet; `quelltextfarben.ts`
// las einen Typ aus `Quelltextfeld.tsx` mit `import { type … }` — und das ist
// unter `verbatimModuleSyntax` eine Einfuhr, die bleibt.
//
// # Was als Kante zählt
//
// **Was zur Laufzeit bleibt.** `import type` und `export type` fallen weg;
// `import { type X }` bleibt als `import {} from …` stehen (TypeScript 5.0,
// `verbatimModuleSyntax`) und lädt das Modul. Ein dynamisches `import()` zählt
// nicht: es lädt, wenn es gerufen wird, nicht beim Auswerten.
//
// # Was dieser Wächter nicht prüft
//
// **Die Tests und ihre Vorrichtungen** (`*.test.ts[x]`, `test/`) — sie laufen
// nicht im Bündel. Und **Pakete aus `node_modules`**: sie sind keine Kanten
// dieses Graphen.
import { existsSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'

import { dateienUnter } from './test/dateien'
import ts from 'typescript'
import { describe, expect, it } from 'vitest'

const quellen = import.meta.dirname

// laufzeitkanten gibt die relativen Einfuhren eines Quelltextes, die zur
// Laufzeit bleiben.
function laufzeitkanten(text: string): string[] {
  const quelle = ts.createSourceFile(
    'datei.tsx',
    text,
    ts.ScriptTarget.Latest,
    false,
    ts.ScriptKind.TSX,
  )
  const ziele: string[] = []

  for (const knoten of quelle.statements) {
    if (
      ts.isImportDeclaration(knoten) &&
      knoten.importClause?.phaseModifier !== ts.SyntaxKind.TypeKeyword
    ) {
      ziele.push((knoten.moduleSpecifier as ts.StringLiteral).text)
    }

    if (
      ts.isExportDeclaration(knoten) &&
      knoten.moduleSpecifier &&
      !knoten.isTypeOnly
    ) {
      ziele.push((knoten.moduleSpecifier as ts.StringLiteral).text)
    }
  }

  return ziele.filter((ziel) => ziel.startsWith('.'))
}

// aufloesen findet die Datei zu einer relativen Einfuhr — wie Vite, mit
// `.ts`, `.tsx` und `index`.
function aufloesen(von: string, ziel: string): string | undefined {
  const basis = resolve(dirname(von), ziel)

  return [
    basis,
    `${basis}.ts`,
    `${basis}.tsx`,
    join(basis, 'index.ts'),
    join(basis, 'index.tsx'),
  ].find((kandidat) => existsSync(kandidat) && statSync(kandidat).isFile())
}

// zyklen gibt die stark zusammenhängenden Teile eines Graphen mit mehr als
// einem Knoten — und jeden Knoten, der sich selbst einführt (Tarjan).
function zyklen(kanten: Map<string, string[]>): string[][] {
  const index = new Map<string, number>()
  const tiefste = new Map<string, number>()
  const stapel: string[] = []
  const aufDemStapel = new Set<string>()
  const gefunden: string[][] = []
  let zaehler = 0

  const besuche = (knoten: string): void => {
    index.set(knoten, zaehler)
    tiefste.set(knoten, zaehler)
    zaehler++
    stapel.push(knoten)
    aufDemStapel.add(knoten)

    for (const nachbar of kanten.get(knoten) ?? []) {
      if (!index.has(nachbar)) {
        besuche(nachbar)
        tiefste.set(
          knoten,
          Math.min(tiefste.get(knoten) ?? 0, tiefste.get(nachbar) ?? 0),
        )
      } else if (aufDemStapel.has(nachbar)) {
        tiefste.set(
          knoten,
          Math.min(tiefste.get(knoten) ?? 0, index.get(nachbar) ?? 0),
        )
      }
    }

    if (tiefste.get(knoten) !== index.get(knoten)) {
      return
    }

    const teil: string[] = []
    let oben: string | undefined

    do {
      oben = stapel.pop()
      if (oben !== undefined) {
        aufDemStapel.delete(oben)
        teil.push(oben)
      }
    } while (oben !== undefined && oben !== knoten)

    if (teil.length > 1 || (kanten.get(knoten) ?? []).includes(knoten)) {
      gefunden.push(teil.sort())
    }
  }

  for (const knoten of kanten.keys()) {
    if (!index.has(knoten)) {
      besuche(knoten)
    }
  }

  return gefunden
}

describe('kein Einfuhrzyklus in der Oberfläche', () => {
  it('führt keine Datei über einen Umweg auf sich selbst zurück', () => {
    const dateien = dateienUnter(quellen, '**/*.{ts,tsx}', {
      ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**', '**/*.d.ts'],
      absolut: true,
    }).sort()

    expect(dateien.length).toBeGreaterThan(100)

    const bekannt = new Set(dateien)
    const kanten = new Map(
      dateien.map((datei) => [
        datei,
        laufzeitkanten(readFileSync(datei, 'utf8'))
          .map((ziel) => aufloesen(datei, ziel))
          .filter(
            (ziel): ziel is string => ziel !== undefined && bekannt.has(ziel),
          ),
      ]),
    )

    expect(
      zyklen(kanten).map((teil) =>
        teil.map((datei) => relative(quellen, datei)).join(', '),
      ),
    ).toEqual([])
  })

  // **Die Gegenprobe steht im Wächter** (§10): ein Zyklus, den er nicht
  // fände, machte jeden grünen Lauf wertlos.
  it('findet einen Zyklus über zwei und über drei Dateien — und einen auf sich selbst', () => {
    expect(
      zyklen(
        new Map([
          ['a', ['b']],
          ['b', ['a']],
          ['c', ['d']],
          ['d', ['e']],
          ['e', ['c']],
          ['f', ['f']],
          ['g', ['a']],
        ]),
      ),
    ).toEqual([['a', 'b'], ['c', 'd', 'e'], ['f']])
  })

  it('zählt nur, was zur Laufzeit bleibt', () => {
    expect(laufzeitkanten("import type { A } from './a'")).toEqual([])
    expect(laufzeitkanten("export type { A } from './a'")).toEqual([])
    expect(laufzeitkanten("import { type A } from './a'")).toEqual(['./a'])
    expect(laufzeitkanten("import { b, type A } from './b'")).toEqual(['./b'])
    expect(laufzeitkanten("export { c } from '../c'")).toEqual(['../c'])
    expect(laufzeitkanten("import './stil.css'")).toEqual(['./stil.css'])
    expect(laufzeitkanten("import { useState } from 'react'")).toEqual([])
    expect(laufzeitkanten("const x = await import('./spaeter')")).toEqual([])
  })
})
