// **Die Tafel der Kennungsfelder stimmt mit dem Quelltext** — #2389.
//
// # Was er prüft
//
// Jede Datei in `docs/FELDPRUEFUNG.md`, Abschnitt „Kennungsfelder", trägt
// mindestens so oft `{...kennungseingabe}`, wie die Tafel Felder nennt — und
// jede Datei, die es trägt, steht in der Tafel.
//
// # Was er nicht prüft
//
//   - **Ob ein Feld eine Kennung nimmt** — das entscheidet ein Mensch beim
//     Eintragen; ein neues Feld für einen Code fällt hier nicht auf, bis es
//     jemand einträgt.
//   - **Ob WebKit die Eigenschaften beachtet** — das misst die Hüllenprobe.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'

function tafel(): Map<string, number> {
  const text = readFileSync(join('..', 'docs', 'FELDPRUEFUNG.md'), 'utf8')
  const abschnitt = text.split('## Kennungsfelder')[1]?.split('\n## ')[0] ?? ''
  const zeilen = new Map<string, number>()

  for (const treffer of abschnitt.matchAll(/^\| `([^`]+)` \| (\d+) \|/gmu)) {
    zeilen.set(treffer[1] ?? '', Number(treffer[2]))
  }

  return zeilen
}

describe('die Kennungsfelder', () => {
  it('stehen in der Tafel, und die Tafel steht im Quelltext', () => {
    const erwartet = tafel()

    expect(erwartet.size).toBeGreaterThan(10)

    const gefunden = new Map<string, number>()

    for (const datei of dateienUnter('src', '**/*.tsx', {
      ohne: ['**/*.test.tsx'],
    })) {
      const anzahl =
        readFileSync(join('src', datei), 'utf8').split('{...kennungseingabe}')
          .length - 1

      if (anzahl > 0) {
        gefunden.set(datei, anzahl)
      }
    }

    const fehlend = [...erwartet].filter(
      ([datei, felder]) => (gefunden.get(datei) ?? 0) < felder,
    )
    const unbenannt = [...gefunden.keys()].filter(
      (datei) => !erwartet.has(datei),
    )

    expect(fehlend).toEqual([])
    expect(unbenannt).toEqual([])
  })
})
