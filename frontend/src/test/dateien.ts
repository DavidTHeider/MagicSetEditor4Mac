// dateienUnter zählt Dateien unter einem Ordner auf — für die Wächter, die
// jede Komponente lesen.
//
// **Nodes `fs.globSync` und nicht `fast-glob`** (2026-10-03): `fast-glob`
// zog über `micromatch` `braces` herein, und dafür gibt es einen Hinweis ohne
// Korrektur (GHSA-vfj7-8cjw-p6xm). Beide finden dieselben Dateien.
//
// **Geordnet**: `fast-glob` gab die Reihenfolge des Dateisystems; ein Wächter,
// der einen Fund nennt, soll auf zwei Rechnern dieselbe Liste nennen.
import { globSync } from 'node:fs'
import { join } from 'node:path'

interface Wahl {
  // ohne sind Muster, die nicht mitzählen — relativ zum Ordner.
  ohne?: string[]

  // absolut gibt volle Pfade statt Pfade relativ zum Ordner.
  absolut?: boolean
}

export function dateienUnter(
  ordner: string,
  muster: string | string[],
  { ohne = [], absolut = false }: Wahl = {},
): string[] {
  const gefunden = globSync(muster, { cwd: ordner, exclude: ohne })

  return (
    absolut ? gefunden.map((datei) => join(ordner, datei)) : gefunden
  ).sort()
}
