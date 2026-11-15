// **Modal bleibt die Ausnahme** — UX-08.AK-01
// (#1680),
// nach dem Artboard *BelegMuster*: „nur bei begründeter Unterbrechung".
//
// # Die Regel
//
// **Genau drei Dinge unterbrechen**: die Befehlspalette, die Einführung beim
// ersten Start und die Kommandozentrale (`docs/UX-ARCHITEKTUR.md`, Dialoge
// und Rückfragen). Jede Bestätigung im Datenbereich ist eine Rückfrage an
// der Klickstelle (`Entscheidung`), jedes Detail eine Schublade oder der
// Inspector.
//
// # Was dieser Lauf hält
//
// **Wer den `Dialog` benutzt** — und dass nur zwei Dateien Mantines Modal
// und Drawer anfassen: `Dialog.tsx` und `Schublade.tsx`. Ein viertes Modal
// fiele damit auf, ebenso ein Modal, das am `Dialog` vorbei gebaut wird.
//
// # Und was nicht
//
// **Ein handgebautes `role="dialog"`.** Das sucht der zweite Lauf — über
// den Quelltext, nicht über das, was gezeichnet wird: wer die Rolle aus
// einer Variablen setzt, geht durch.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { dateienUnter } from './test/dateien'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

function dateien(): string[] {
  return dateienUnter(quellen, '**/*.tsx', {
    ohne: ['**/*.test.tsx', 'test/**'],
  })
}

// lies gibt den Quelltext ohne Kommentarzeilen — ein Kommentar, der
// `<Dialog …>` erklärt, benutzt ihn nicht.
function lies(datei: string): string {
  return readFileSync(join(quellen, datei), 'utf8')
    .split('\n')
    .filter((zeile) => !/^\s*(\/\/|\*|\{\/\*)/.test(zeile))
    .join('\n')
}

describe('Modal bleibt die Ausnahme', () => {
  it('unterbricht an genau drei Stellen', () => {
    const nutzer = dateien()
      .filter((datei) => /<Dialog\b/.test(lies(datei)))
      .sort()

    expect(nutzer).toEqual([
      'bausteine/Einfuehrung.tsx',
      'rahmen/Kommandozentrale.tsx',
      'rahmen/Palette.tsx',
    ])
  })

  it('baut Modal und Schublade nur an einer Stelle je', () => {
    const mantine = dateien()
      .filter((datei) =>
        /\b(Modal|Drawer)\b[^\n]*from '@mantine/.test(lies(datei)),
      )
      .sort()

    expect(mantine).toEqual(['bausteine/Dialog.tsx', 'bausteine/Schublade.tsx'])
  })

  // **Ein Attributselektor liest eine Rolle und setzt keine** — AP-7
  // (#2364): der Objekthandlungsraum fragt mit `closest('[role="dialog"]')`,
  // ob eine Taste in einem Dialog fällt. Gesucht wird, was eine Rolle
  // **vergibt**; was in eckigen Klammern steht, ist eine Frage nach ihr.
  const vergibtDialogrolle =
    /(?<!\[)role=["']{0,1}(alert)?dialog|(?<!\[)aria-modal/

  it('erkennt eine vergebene Rolle und übergeht einen Selektor', () => {
    expect(vergibtDialogrolle.test('<div role="dialog">')).toBe(true)
    expect(vergibtDialogrolle.test("<div role='alertdialog'>")).toBe(true)
    expect(vergibtDialogrolle.test('<div aria-modal="true">')).toBe(true)
    expect(vergibtDialogrolle.test('closest(\'[role="dialog"]\')')).toBe(false)
  })

  it('setzt die Rolle eines Dialogs nicht von Hand', () => {
    const befunde = dateien().filter((datei) =>
      vergibtDialogrolle.test(lies(datei)),
    )

    expect(befunde).toEqual([])
  })
})
