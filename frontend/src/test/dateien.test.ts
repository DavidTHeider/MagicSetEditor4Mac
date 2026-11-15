// **Quelldateien aufzählen ohne `fast-glob`** — die Wächter der Oberfläche
// lesen jede Komponente; dafür reicht Nodes `fs.globSync`.
//
// **Warum der Wechsel**: `fast-glob` zog über `micromatch` `braces` herein,
// und für `braces` gibt es seit dem 2026-10-03 einen Hinweis ohne Korrektur
// (GHSA-vfj7-8cjw-p6xm, alle Fassungen bis 3.0.3). Gemessen am selben Tag:
// beide finden dieselben 1.562 Dateien unter `src`.
import { describe, expect, it } from 'vitest'

import { dateienUnter } from './dateien'

describe('dateienUnter', () => {
  it('findet Komponenten und lässt die genannten Muster weg', () => {
    const dateien = dateienUnter(import.meta.dirname + '/..', '**/*.tsx', {
      ohne: ['**/*.test.tsx', 'test/**'],
    })

    expect(dateien).toContain('bausteine/Leerzustand.tsx')
    expect(dateien.some((datei) => datei.endsWith('.test.tsx'))).toBe(false)
    expect(dateien.some((datei) => datei.startsWith('test/'))).toBe(false)
  })

  it('nennt Pfade auf Wunsch absolut und immer geordnet', () => {
    const wurzel = import.meta.dirname
    const dateien = dateienUnter(wurzel, '*.ts', { absolut: true })

    expect(dateien.length).toBeGreaterThan(0)
    expect(dateien.every((datei) => datei.startsWith(wurzel))).toBe(true)
    expect(dateien).toEqual([...dateien].sort())
  })
})
