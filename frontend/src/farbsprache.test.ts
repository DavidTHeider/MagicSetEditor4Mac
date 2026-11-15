// **Farbe und Rundung kommen aus Merkmalen, nicht aus Klassen** — AP-2
// (#2359),
// Befunde B-07 und B-09; Vermerk 0065, Festlegungen 2 und 3.
//
// Festlegung 2 sagt es seit Phase 4.10: *ein Bauteil, das eine Farbe
// unmittelbar schreibt, hat die Sprache umgangen.* Gezählt am 2026-10-16
// standen trotzdem **568** Tailwind-Palettenklassen in 196 Dateien — und
// Fehlermeldungen in **zwei** Farbsystemen: 77-mal die Fehlerfarbe der
// Gestaltungssprache, die damals nach ihrer Variablen hieß, und 214-mal
// Tailwinds Rot in 800 und 300 — beide an `role="alert"`.
//
// # Was er hält
//
// - **Keine Palettenfarbe** außerhalb der Tests — weder `text-red-800` noch
//   `fill-sky-200/60` noch `bg-white`. Was eine Farbe braucht, nimmt ein
//   Merkmal aus `stil.css` (`fehler`, `hinweis`, `erfolg`, `auswahl`, …), und
//   dessen Wert steht je Farbwelt hell und dunkel dort, wo der Kontrast
//   gerechnet wird (`themen.test.ts`).
// - **Drei Rundungen und rund** — `rounded-sm`, `rounded-md`, `rounded-lg`,
//   `rounded-full` und `rounded-none`; ihre Werte stehen in `@theme`.
// - **Abstände nur aus der Skala** — kein `p-[13px]`, kein `gap-[0.7rem]`:
//   Innen- und Außenabstände kommen aus Tailwinds `--spacing`. Gezählt am
//   2026-10-16 stand keiner; der Wächter hält, dass es so bleibt.
//
// **Kommentare liest er nicht**: sie erzählen, was früher dastand, und
// das darf eine Palettenklasse nennen. Ausgeschnitten wird über den Parser
// und nicht über einen Ausdruck — `accept="image/*"` ist kein Kommentar.
//
// # Was er nicht hält
//
// **Eine Farbe, die als Wert geschrieben ist** — `style={{ color: '#991b1b' }}`
// oder ein `#…` in einer CSS-Datei außerhalb der Merkmalsblöcke. Das prüft
// `kontrast.ts` nicht und dieser Wächter auch nicht; gezählt am 2026-10-16
// stand keines im Quelltext der Bauteile.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'
import { ohneKommentare } from './test/kommentare'

const quellen = join(import.meta.dirname)

const palette =
  /\b(?:bg|text|border|ring|fill|stroke|outline|divide|accent|decoration|shadow|placeholder|from|via|to)-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-\d{2,3}\b|\b(?:bg|text|border|ring|fill|stroke|outline|divide)-(?:white|black)\b/g

const erlaubteRundungen = new Set([
  'rounded-sm',
  'rounded-md',
  'rounded-lg',
  'rounded-full',
  'rounded-none',
])

const abstand =
  /(?<![\w-])-?(?:p[xytblrse]?|m[xytblrse]?|gap(?:-[xy])?|space-[xy])-\[[^\]\s]+\]/g

const rundung = /(?<![\w-])rounded(?:-[a-z0-9]+)*(?:-\[[^\]]+\])?/g

function lesen(datei: string): string {
  return ohneKommentare(datei, readFileSync(join(quellen, datei), 'utf8'))
}

function produktdateien(): string[] {
  return dateienUnter(quellen, ['**/*.ts', '**/*.tsx', '**/*.css'], {
    ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**'],
  })
}

describe('die Werkzeuge des Wächters', () => {
  it('nimmt Zeilen-, Block- und JSX-Kommentare, aber keine Zeichenkette', () => {
    const text = [
      '// text-red-800',
      "const a = 'image/*' // bg-white",
      '/* fill-sky-200 */ const b = "https://beispiel.invalid"',
      'const c = <p>{/* stroke-slate-600 */}</p>',
    ].join('\n')

    const rest = ohneKommentare('probe.tsx', text)

    expect(rest).not.toMatch(palette)
    expect(rest).toContain("'image/*'")
    expect(rest).toContain('"https://beispiel.invalid"')
  })

  it('erkennt einen willkürlichen Abstand, aber keinen aus der Skala', () => {
    expect('p-[13px] gap-x-[0.7rem] -mt-[2px]'.match(abstand)).toEqual([
      'p-[13px]',
      'gap-x-[0.7rem]',
      '-mt-[2px]',
    ])
    expect('p-2 gap-x-3 -mt-1 h-[1.875rem]'.match(abstand)).toBeNull()
  })

  it('nimmt Blockkommentare aus CSS', () => {
    expect(ohneKommentare('probe.css', '/* text-white */ a {}')).not.toMatch(
      palette,
    )
  })
})

describe('die Farbsprache', () => {
  it('kennt keine Palettenfarbe außerhalb der Tests', () => {
    const funde = produktdateien().flatMap((datei) =>
      [...lesen(datei).matchAll(palette)].map(
        (treffer) => `${datei}: ${treffer[0]}`,
      ),
    )

    expect(funde).toEqual([])
  })

  it('rundet nur mit den Stufen aus @theme', () => {
    const funde = produktdateien().flatMap((datei) =>
      [...lesen(datei).matchAll(rundung)]
        .map((treffer) => treffer[0])
        .filter((klasse) => !erlaubteRundungen.has(klasse))
        .map((klasse) => `${datei}: ${klasse}`),
    )

    expect(funde).toEqual([])
  })

  it('nimmt Abstände nur aus der Skala', () => {
    const funde = produktdateien().flatMap((datei) =>
      [...lesen(datei).matchAll(abstand)].map(
        (treffer) => `${datei}: ${treffer[0]}`,
      ),
    )

    expect(funde).toEqual([])
  })
})
