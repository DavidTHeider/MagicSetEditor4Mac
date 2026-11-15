// **Der Musterkatalog hält gegen den Quelltext** — UX-08.AK-01
// (#1680).
//
// `docs/UX-ARCHITEKTUR.md` führt unter „Muster — wann welches" acht Muster,
// je mit dem Baustein, der es trägt. **Dieser Lauf hält**, dass es jeden
// genannten Baustein gibt und dass ihn mindestens eine andere Datei benutzt —
// ein Muster, dessen Baustein niemand einsetzt, stünde nur auf dem Papier.
//
// **Was er nicht hält**: ob eine Seite das richtige Muster gewählt hat. Das
// ist eine Frage an die Durchsicht und an die Tafel selbst.
import { existsSync, readFileSync } from 'node:fs'
import { basename, join } from 'node:path'

import { dateienUnter } from './test/dateien'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

function tafel(): { muster: string; baustein: string }[] {
  const datei = readFileSync(
    join(quellen, '../../docs/UX-ARCHITEKTUR.md'),
    'utf8',
  )
  const abschnitt =
    datei.split('## Muster — wann welches')[1]?.split('\n## ')[0] ?? ''

  return [...abschnitt.matchAll(/^\| ([^|]+) \| [^|]+ \| `([^`]+)` \|/gm)].map(
    (zeile) => ({ muster: (zeile[1] ?? '').trim(), baustein: zeile[2] ?? '' }),
  )
}

describe('Der Musterkatalog', () => {
  it('nennt acht Muster, von der Zusatzhilfe bis zur Unterbrechung', () => {
    expect(tafel().map((zeile) => zeile.muster)).toEqual([
      'Tooltip',
      'Aufklappen',
      'Popover',
      'Rückfrage',
      'Inspector',
      'Schublade',
      'Assistent',
      'Modal',
    ])
  })

  it.each(tafel())(
    '$muster: $baustein liegt da und wird benutzt',
    ({ baustein }) => {
      expect(existsSync(join(quellen, baustein)), baustein).toBe(true)

      const name = basename(baustein).replace(/\.tsx?$/, '')
      const nutzer = dateienUnter(quellen, '**/*.tsx', {
        ohne: ['**/*.test.tsx', 'test/**'],
      })
        .filter((datei) => datei !== baustein)
        .filter((datei) =>
          readFileSync(join(quellen, datei), 'utf8').includes(`/${name}'`),
        )

      expect(nutzer.length, baustein).toBeGreaterThan(0)
    },
  )
})
