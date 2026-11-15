// **Keine feste Weiß- oder Schwarzschrift** — angestrichen vom Zweitblick
// über Phase 6.3 (M6): „Anwenden" der Massenänderung trug `text-white` auf
// der Marke, 2,25 : 1 auf dunklem Pergament und 1,71 in der dunklen Moos-Welt.
// `--mse-marketext` gibt es genau dafür (`stil.css`); eine feste Farbe folgt
// weder dem Erscheinungsbild noch einer Farbwelt noch eigenen Farben.
//
// **Gelesen werden `.tsx` und `.ts`** — Klassenlisten stehen hier bewusst
// auch in `.ts`-Dateien (Nachlauf über Phase 6.3, N13).
//
// # Was er nicht hält
//
// **Ob die Merkmalfarbe dort genug Kontrast hat** — das hält
// `docs/KONTRASTE.md` —, und Farben, die über `style` gesetzt werden.
// **Bunte feste Farben** — Rot für Fehler, Grün, Bernstein: sie folgen dem
// Erscheinungsbild über `dark:` und sind je Paar in `docs/KONTRASTE.md`
// gemessen, aber keiner Farbwelt und keinen eigenen Farben. Sie sagen etwas
// (Fehler, Erfolg, Warnung), und ein Merkmal je Bedeutung gibt es heute nur
// für die Warnung. Das ist die Grenze, die bleibt (Phase 6.10, N13).
//
// **Die grauen hält er seit N13** (unten): ein `text-slate-600 dark:text-
// slate-400` ahmte `text-leise` nach und stünde auf einem eigenen Grund von
// #4a4a4a bei 3,37 : 1; ein `bg-slate-100` ahmte `bg-betont` nach, und
// Merkmalschrift darauf maß niemand.
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

describe('Die Schriftfarbe', () => {
  it('ist nie fest weiß oder schwarz', () => {
    const funde = bauteile(quelle).flatMap((pfad) =>
      [...readFileSync(pfad, 'utf8').matchAll(/\btext-(?:white|black)\b/g)].map(
        () => relative(quelle, pfad),
      ),
    )

    expect(funde).toEqual([])
  })
})

// **Kein festes Grau** — N13 aus dem Nachlauf über Phase 6.3 (#1664), Phase
// 6.10 (#1671). Ein Grau ist hier immer ein Ersatz für ein Merkmal —
// `leise`, `betont`, `kante` —, und das Merkmal folgt Farbwelt und eigenen
// Farben, das Grau nicht. **Kommentare zählen nicht**: dort nennen Sätze die
// Klasse, um zu erklären, warum sie fehlt.
// **Eine benannte Ausnahme**: der Zuschnittrahmen liegt über einem Bild,
// dessen Farben niemand kennt — Weiß über fast Schwarz ist dort das einzige
// Paar, das auf jedem Untergrund sichtbar bleibt (Werkstück J von Phase
// 4.10). Ein Merkmal folgte der Farbwelt und verschwände auf einer Karte
// derselben Farbe.
const ueberEinemBild = new Set(['seiten/Zuschnittrahmen.tsx: ring-slate-900'])

describe('Ein Grau', () => {
  it('steht nie fest, sondern als Merkmal', () => {
    const ohneKommentar = (text: string) =>
      text.replace(/\/\*[\s\S]*?\*\//g, '').replace(/\/\/.*$/gm, '')
    const funde = bauteile(quelle).flatMap((pfad) =>
      [
        ...ohneKommentar(readFileSync(pfad, 'utf8')).matchAll(
          /\b(?:text|bg|border|divide|outline|ring)-(?:slate|gray|zinc|neutral|stone)-\d{2,3}\b/g,
        ),
      ]
        .map(([klasse]) => `${relative(quelle, pfad)}: ${klasse}`)
        .filter((fund) => !ueberEinemBild.has(fund)),
    )

    expect(funde).toEqual([])
  })
})
