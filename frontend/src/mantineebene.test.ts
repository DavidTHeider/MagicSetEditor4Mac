// **Mantine liegt in einer Ebene, zwischen Preflight und den Utilities** —
// #2451,
// gefunden in AP-3 (#2360).
//
// Tailwind 4 legt Preflight und Utilities in Kaskadenebenen, und **eine
// ungeschichtete Regel gewinnt gegen jede geschichtete**, gleich wo sie
// steht. `@mantine/core/styles.css` hat keine Ebene: seine Regel
// `input, button, textarea, select { font: inherit }` schlug jede Klasse für
// Schrift an Knöpfen und Feldern — gemessen in `ap3gestalt`, `fontWeight` 400
// an einem Knopf mit `font-medium`.
//
// # Was er hält
//
// - **Mantine kommt aus seiner geschichteten Fassung** (`styles.layer.css`,
//   alles in `@layer mantine`), nie aus der ungeschichteten — in keiner
//   Stildatei der Oberfläche.
// - **Die Reihenfolge ist erklärt, bevor Tailwind sie erklärt**:
//   `theme, base, mantine, components, utilities` — Mantine über Preflight,
//   die Utilities über Mantine. So steht es in Mantines eigener Anleitung
//   (mantine.dev, *Mantine styles*, Abschnitt *CSS layers*, abgerufen am
//   2026-10-16: `@layer base, mantine, components;`).
//
// # Was er nicht hält
//
// **Ob die Bauteile danach aussehen wie vorher** — das vergleicht
// `mantineebene` am nativen Fenster. Und eine Regel ohne Ebene, die nicht von
// Mantine kommt: die Auswahlpfeile stehen mit Absicht ohne
// (`erscheinungsmerkmale.css`).
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'
import { ohneKommentare } from './test/kommentare'

const quellen = join(import.meta.dirname)

const lesen = (datei: string): string =>
  ohneKommentare(datei, readFileSync(join(quellen, datei), 'utf8'))

describe('Mantine in einer Ebene', () => {
  it('bindet Mantine nirgends ohne Ebene ein', () => {
    const funde = dateienUnter(quellen, ['**/*.css']).filter((datei) =>
      /@import\s+['"]@mantine\/[a-z-]+\/styles\.css['"]/.test(lesen(datei)),
    )

    expect(funde).toEqual([])
    expect(lesen('stil.css')).toMatch(
      /@import\s+'@mantine\/core\/styles\.layer\.css';/,
    )
  })

  it('erklärt die Reihenfolge der Ebenen, bevor Tailwind es tut', () => {
    const stil = lesen('stil.css')
    const erklaerung = /@layer\s+([a-z,\s]+);/.exec(stil)

    expect(erklaerung?.[1]?.split(/\s*,\s*/)).toEqual([
      'theme',
      'base',
      'mantine',
      'components',
      'utilities',
    ])
    // **Als Ausdruck und nicht wörtlich**: der Wächter über die Einfuhren
    // (`abhaengigkeiten_test.go`) läse den Satz sonst als eine zweite Stelle,
    // die Tailwind einführt.
    expect(erklaerung?.index).toBeLessThan(
      stil.search(/@import\s+'tailwindcss'/),
    )
  })
})
