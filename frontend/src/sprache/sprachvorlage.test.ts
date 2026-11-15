// **Die Vorlage der eigenen Übersetzungen ist der deutsche Katalog** —
// LOC-02 (#1688).
//
// # Was er hält
//
// `internal/app/sprachvorlage.json` ist der deutsche Katalog, flach: Pfad →
// Satz. Gegen ihn prüft Go eine eigene Übersetzung — Platzhalter,
// Auszeichnung, was veraltet ist —, an jeder Fläche gleich. Eine Datei, die
// hinter dem Katalog zurückbleibt, ließe neue Sätze als unbekannt abweisen.
//
// # Wie sie entsteht
//
//     VORLAGE_SCHREIBEN=1 npx vitest run src/sprache/sprachvorlage.test.ts
//
// **Und nur so** — von Hand geschrieben, liefe sie beim ersten Tippfehler
// auseinander.
import { readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { flach } from './flach'
import { kataloge } from './kataloge'

const datei = join(
  import.meta.dirname,
  '../../../internal/app/sprachvorlage.json',
)

// geordnet schreibt die Vorlage mit geordneten Pfaden — ein Satz je Zeile,
// damit ein Diff zeigt, welcher sich änderte.
function geordnet(saetze: Record<string, string>): string {
  const pfade = Object.keys(saetze).sort()

  return `${JSON.stringify(Object.fromEntries(pfade.map((pfad) => [pfad, saetze[pfad]])), null, 0).replace(/","/g, '",\n"')}\n`
}

describe('Die Vorlage der eigenen Übersetzungen', () => {
  it('ist der deutsche Katalog', () => {
    const soll = geordnet(flach(kataloge.de))

    if (process.env.VORLAGE_SCHREIBEN === '1') {
      writeFileSync(datei, soll)
    }

    expect(
      readFileSync(datei, 'utf8') === soll,
      'die Vorlage ist veraltet — VORLAGE_SCHREIBEN=1 npx vitest run src/sprache/sprachvorlage.test.ts',
    ).toBe(true)
  })
})
