// **Eine Liste mit fremder Rolle ist keine Liste** — gefunden mit axe in
// WebKit am nativen Fenster (`tischhand09`, P4,
// #2372):
// drei Stellen trugen `role="log"`, `"status"` oder `"alert"` an der `ul`
// selbst. Damit verliert sie ihre Listenrolle, und ihre Einträge stehen ohne
// Liste da (Regel `listitem`). Unter jsdom fiel es nicht auf, weil die
// geprüften Seiten diese Listen leer zeigten.
//
// # Was dieser Wächter hält
//
// **Keine Live-Region-Rolle an `ul` oder `ol`** — die Rolle gehört an einen
// Rahmen um die Liste. Ein Menü, ein Baum und eine Gruppe darin bleiben
// erlaubt: dort tragen die Einträge ihre eigene Rolle.
//
// # Was er nicht hält
//
// **Eine Rolle, die als Ausdruck kommt** (`role={…}`) liest er nicht, und
// eine Liste, die ein anderes Bauteil zeichnet, auch nicht — beides prüft axe
// am gezeichneten Baum.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import { dateienUnter } from './test/dateien'

const quellen = join(import.meta.dirname)

// liveRollen sind die Rollen, die eine Liste zur Ansage machen wollen.
const liveRollen = ['alert', 'log', 'status', 'marquee', 'timer']

// listenMitRolle findet `<ul …>` und `<ol …>` samt ihrer festen Rolle.
export function listenMitRolle(quelltext: string): string[] {
  return [...quelltext.matchAll(/<(ul|ol)\b[^>]*?\brole="([a-z]+)"/gs)]
    .filter((treffer) => liveRollen.includes(treffer[2] ?? ''))
    .map((treffer) => `${treffer[1] ?? ''} role="${treffer[2] ?? ''}"`)
}

describe('Listenrollen', () => {
  it('findet eine Live-Region an einer Liste, und nur sie', () => {
    expect(listenMitRolle('<ul role="status" className="x">')).toEqual([
      'ul role="status"',
    ])
    expect(listenMitRolle('<ol\n  role="log"\n>')).toEqual(['ol role="log"'])
    expect(listenMitRolle('<ul role="menu"><li role="menuitem">')).toEqual([])
    expect(listenMitRolle('<div role="status"><ul>')).toEqual([])
  })

  it('kommt im Quelltext nicht vor', () => {
    const funde = dateienUnter(quellen, '**/*.tsx', {
      ohne: ['**/*.test.tsx'],
    }).flatMap((datei) =>
      listenMitRolle(readFileSync(join(quellen, datei), 'utf8')).map(
        (fund) => `${datei}: ${fund}`,
      ),
    )

    expect(funde).toEqual([])
  })
})
