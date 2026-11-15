// **Dunkel gewählt ist dasselbe Bild wie dunkel vom System** — ein Befund der
// Abnahme von 0.7.0 (ABN-01, #2175).
//
// Der dunkle Satz steht in jedem Stilblatt zweimal: einmal unter
// `prefers-color-scheme: dark` für ein Fenster, das dem System folgt, und
// einmal unter `[data-erscheinungsbild='dunkel']` für eine ausdrückliche
// Wahl. In `stil.css` fehlten dem zweiten die acht Merkmale des Tischs: wer
// auf einem hellen System „dunkel" wählte, bekam den dunklen Holzgrund aus
// dem Tischbild, aber Schrift, Kante, Auswahl und Kartenrücken des hellen
// Satzes — der Zonenname „Stapel" stand bei 1,03 : 1 auf seinem Grund
// (gemessen am nativen Fenster, `abn01tisch`).
//
// **Warum kein Wächter es sah**: `themen.test.ts` liest Merkmale mit
// `--mse-[a-z]+`, und das trifft `--mse-tisch-grund` nicht. Hier zählt jedes
// Merkmal, auch mit Bindestrich im Namen.
//
// **Was er nicht prüft**: ob ein Satz gut aussieht. Die Kontraste der Paare
// stehen in `docs/KONTRASTE.md`; hier steht, dass beide Wege dieselben
// Werte bekommen.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const lies = (datei: string) =>
  readFileSync(join(import.meta.dirname, datei), 'utf8')

// merkmale gibt die Merkmale des Rumpfs, der nach `anfang` mit `{` beginnt —
// bis zur passenden Klammer, auch über einen eingeschachtelten Block hinweg.
function merkmale(quelle: string, anfang: number): Record<string, string> {
  const auf = quelle.indexOf('{', anfang)
  let tiefe = 0
  let ende = auf
  for (; ende < quelle.length; ende += 1) {
    if (quelle[ende] === '{') tiefe += 1
    if (quelle[ende] === '}') tiefe -= 1
    if (tiefe === 0) break
  }

  const rumpf = quelle.slice(auf, ende).replace(/\/\*[\s\S]*?\*\//g, '')
  const werte: Record<string, string> = {}
  for (const [, name, wert] of rumpf.matchAll(/(--mse-[a-z-]+):\s*([^;]+);/g)) {
    if (name !== undefined && wert !== undefined) werte[name] = wert.trim()
  }

  return werte
}

// mitGrund gibt den ersten Block mit diesem Selektor, der einen Grund setzt —
// andere Blöcke desselben Selektors regeln anderes (Farbschema, Fasern).
function mitGrund(quelle: string, selektor: string): Record<string, string> {
  for (
    let stelle = quelle.indexOf(selektor);
    stelle >= 0;
    stelle = quelle.indexOf(selektor, stelle + 1)
  ) {
    const werte = merkmale(quelle, stelle)
    if (werte['--mse-grund'] !== undefined) return werte
  }

  throw new Error(`kein Block mit Grund für ${selektor}`)
}

const paare = [
  {
    name: 'Pergament (stil.css)',
    datei: 'stil.css',
    gewaehlt: "html[data-erscheinungsbild='dunkel'] {",
    system:
      'html:not([data-erscheinungsbild]) {\n    @media (prefers-color-scheme: dark)',
  },
  ...(['klar', 'schiefer', 'moos'] as const).map((welt) => ({
    name: `${welt} (themen.css)`,
    datei: 'themen.css',
    gewaehlt: `html[data-farbwelt='${welt}'][data-erscheinungsbild='dunkel'] {`,
    system: `html[data-farbwelt='${welt}']:not([data-erscheinungsbild]) {`,
  })),
]

describe('Dunkel gewählt und dunkel vom System', () => {
  for (const paar of paare) {
    it(`${paar.name}: tragen dieselben Merkmale mit denselben Werten`, () => {
      const quelle = lies(paar.datei)
      const gewaehlt = mitGrund(quelle, paar.gewaehlt)
      const system = mitGrund(quelle, paar.system)

      expect(Object.keys(system).length).toBeGreaterThan(9)
      expect(gewaehlt).toEqual(system)
    })
  }
})
