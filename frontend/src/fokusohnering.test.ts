// **Kein Tastaturfokus ohne Ring** — angestrichen vom Zweitblick über Phase
// 6.3 (M3): Palette, Menüs und Skriptverlauf zeigten den Fokus nur als
// Flächenton, `focus-visible:bg-betont focus-visible:outline-none` — 1,29 : 1
// gegen die Fläche im Hellen, 1,13 im Dunklen, und derselbe Ton wie beim
// Zeigen mit der Maus. WCAG 2.4.7 verlangt einen sichtbaren Fokus, 1.4.11
// für ihn 3 : 1.
//
// **Gelesen werden `.tsx` und `.ts`** — Klassenlisten stehen hier bewusst
// auch in `.ts`-Dateien (Nachlauf über Phase 6.3, N13), und **drei
// Schreibweisen**, mit denen Tailwind 4 den Ring wegnimmt.
//
// # Was er nicht hält
//
// **Ob der Ring, der dasteht, 3 : 1 hält** — das hält `docs/KONTRASTE.md`
// und die Sperre in `kontrast.ts` (Ring auf Grund, Fläche und betonter
// Fläche). **Eine vierte Schreibweise** — ein `ring-0`, eine Klasse, die
// aus Teilen zusammengesetzt wird — und einen Fokus, den ein Bauteil in
// `style` wegnimmt.
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const quelle = import.meta.dirname

function bauteile(ort: string): string[] {
  return readdirSync(ort).flatMap((name) => {
    const pfad = join(ort, name)
    if (statSync(pfad).isDirectory()) {
      return bauteile(pfad)
    }

    return /\.tsx?$/.test(name) && !name.includes('.test.') ? [pfad] : []
  })
}

describe('Der Tastaturfokus', () => {
  it('nimmt nirgends den Ring weg', () => {
    const funde = bauteile(quelle).filter((pfad) =>
      /focus-visible:outline-(none|hidden|0)\b/.test(
        readFileSync(pfad, 'utf8'),
      ),
    )

    expect(funde.map((pfad) => relative(quelle, pfad))).toEqual([])
  })
})
