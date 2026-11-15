// **Kein Haken ohne Aufrufer** — SEC04
// (#2357),
// Befund Z-02: der Wächter für Bausteine (`bausteineohneaufrufer.test.ts`)
// nahm die Haken ausdrücklich aus — dieselbe Klasse, ein anderer Ort.
//
// # Was er prüft
//
// Jede Datei unter `src/haken/`, die etwas exportiert, wird von **mindestens
// einer** Datei außerhalb der Tests importiert.
//
// # Was er nicht prüft
//
//   - **Ob der Aufrufer sinnvoll ist** — ein Import genügt ihm.
//   - **Ob jeder Export einer Datei benutzt wird**, nur die Datei.
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const wurzel = 'src'

function quellen(): string[] {
  const gefunden: string[] = []

  const gehe = (ort: string) => {
    for (const eintrag of readdirSync(ort, { withFileTypes: true })) {
      const pfad = join(ort, eintrag.name)

      if (eintrag.isDirectory()) {
        gehe(pfad)

        continue
      }

      if (
        /\.tsx?$/u.test(eintrag.name) &&
        !/\.test\.tsx?$/u.test(eintrag.name) &&
        !pfad.includes('/test/')
      ) {
        gefunden.push(pfad)
      }
    }
  }

  gehe(wurzel)

  return gefunden
}

describe('die Haken dieses Fensters', () => {
  it('haben alle einen Aufrufer', () => {
    const alle = quellen()
    const haken = alle.filter((pfad) => pfad.includes('/haken/'))

    expect(haken.length).toBeGreaterThan(20)

    const ohne = haken.filter((pfad) => {
      const name =
        pfad
          .split('/')
          .pop()
          ?.replace(/\.tsx?$/u, '') ?? ''

      return !alle.some(
        (andere) =>
          andere !== pfad &&
          new RegExp(`from '[^']*/${name}'`, 'u').test(
            readFileSync(andere, 'utf8'),
          ),
      )
    })

    expect(ohne).toEqual([])
  })
})
