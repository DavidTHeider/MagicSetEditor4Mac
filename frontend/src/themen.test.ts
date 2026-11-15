// **Jede Farbwelt ist gerechnet** — UX-09.AK-01 und AK-03
// (#1681):
// *„klar unterscheidbare gut lesbare Standards"*.
//
// **Gelesen aus `themen.css` selbst**, nicht aus einer Kopie der Werte: eine
// zweite Liste liefe beim ersten neuen Farbton auseinander. Gerechnet werden
// dieselben Paare wie für Pergament (`lesbarePaare`), und dazu der
// ungünstigste Bildpunkt unter dem Schleier des Fensterbilds.
//
// **Was dieser Lauf nicht sieht**: wie die Farben im Fenster wirken. Das
// misst `ux09-gestaltung` am gerenderten Element; hier steht, dass die
// Zahlen stimmen, die dort ankommen.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

// **Aus der Datei gelesen und nicht über `?raw`**: Vitest verarbeitet keine
// Stilblätter und gäbe eine leere Zeichenkette zurück — und jeder Block
// fehlte, ohne dass eine Farbe falsch wäre.
const themen = readFileSync(join(import.meta.dirname, 'themen.css'), 'utf8')

// **Pergament steht in `stil.css`** — es ist die Vorgabe und keine Farbwelt
// in `themen.css`; bis zum Nachlauf über Phase 6.3 rechnete dieser Lauf es
// deshalb gar nicht (G3, N4).
const stil = readFileSync(join(import.meta.dirname, 'stil.css'), 'utf8')
import {
  kontrast,
  lesbarePaare,
  lesbarkeitsbefunde,
  statusfarben,
} from './kontrast'

// block gibt die Merkmale einer Regel, die mit diesem Selektor beginnt.
function block(selektor: string, quelle = themen): Record<string, string> {
  const anfang = quelle.indexOf(selektor + ' {')
  if (anfang < 0) {
    throw new Error(`kein Block für ${selektor}`)
  }

  const rumpf = quelle.slice(anfang, quelle.indexOf('}', anfang))
  const werte: Record<string, string> = {}
  for (const [, name, wert] of rumpf.matchAll(
    /(--mse-[a-z]+):\s*(#[0-9a-f]+);/g,
  )) {
    if (name !== undefined && wert !== undefined) {
      werte[name] = wert
    }
  }

  return werte
}

// mische gibt `anteil` der ersten Farbe über der zweiten — wie
// `color-mix(in srgb, …)`.
function mische(eins: string, zwei: string, anteil: number): string {
  const kanaele = (farbe: string) =>
    [1, 3, 5].map((stelle) => parseInt(farbe.slice(stelle, stelle + 2), 16))
  const [a, b] = [kanaele(eins), kanaele(zwei)]

  return (
    '#' +
    a
      .map((wert, stelle) =>
        Math.round(wert * anteil + (b[stelle] ?? 0) * (1 - anteil))
          .toString(16)
          .padStart(2, '0'),
      )
      .join('')
  )
}

const schleier = Number(/--mse-schleier: (\d+)%/.exec(themen)?.[1]) / 100

// pergament gibt die Merkmale des ersten Blocks mit diesem Selektor in
// `stil.css`, der einen Grund trägt — vor ihm stehen Blöcke desselben
// Selektors mit anderem Inhalt (die Fasern, das Farbschema).
function pergament(selektor: string): Record<string, string> {
  for (
    let anfang = stil.indexOf(selektor + ' {');
    anfang >= 0;
    anfang = stil.indexOf(selektor + ' {', anfang + 1)
  ) {
    const werte = block(selektor, stil.slice(anfang))
    if (werte['--mse-grund'] !== undefined) {
      return werte
    }
  }

  throw new Error(`kein Block mit Grund für ${selektor}`)
}

const farbwelten = ['klar', 'schiefer', 'moos'] as const

describe('Die Farbwelten', () => {
  it('haben einen Schleier, der gelesen werden kann', () => {
    expect(schleier).toBeGreaterThan(0.5)
    expect(schleier).toBeLessThan(1)
  })

  for (const welt of farbwelten) {
    const hell = block(`html[data-farbwelt='${welt}']`)
    const dunkel = block(
      `html[data-farbwelt='${welt}'][data-erscheinungsbild='dunkel']`,
    )

    for (const [bild, werte] of [
      ['hell', hell],
      ['dunkel', dunkel],
    ] as const) {
      it(`${welt}, ${bild}: nennt alle zehn Merkmale`, () => {
        expect(Object.keys(werte)).toHaveLength(10)
      })

      // **Mit den Farben der Meldungen des Erscheinungsbilds** — und unter
      // dem Schleier, wie ihn die Sperre bei eigenen Farben rechnet
      // (Zweitblick über Phase 6.3, S4).
      it(`${welt}, ${bild}: hält jedes Paar über seiner Schwelle`, () => {
        const lies = (merkmal: string) => werte[merkmal] ?? ''

        expect(lesbarkeitsbefunde(lies, { dunkel: bild === 'dunkel' })).toEqual(
          [],
        )
        expect(
          lesbarkeitsbefunde(lies, { dunkel: bild === 'dunkel', schleier }),
        ).toEqual([])
        expect(lesbarePaare.length).toBeGreaterThan(10)
      })

      // **Unter dem Fensterbild** liegt der Grund zu `schleier` über einem
      // beliebigen Bildpunkt — der ungünstigste ist Schwarz oder Weiß.
      it(`${welt}, ${bild}: bleibt unter dem Schleier lesbar`, () => {
        for (const extrem of ['#000000', '#ffffff']) {
          const grund = mische(werte['--mse-grund'] ?? '', extrem, schleier)
          for (const [vorn, schwelle] of [
            ['--mse-text', 4.5],
            ['--mse-leise', 4.5],
            ['--mse-marke', 4.5],
            ['--mse-warnung', 4.5],
            ['--mse-kante', 3],
          ] as const) {
            expect(
              kontrast(werte[vorn] ?? '', grund),
              `${vorn} über ${extrem}`,
            ).toBeGreaterThanOrEqual(schwelle)
          }

          for (const [vorn, farbe] of statusfarben[bild]) {
            expect(
              kontrast(farbe, grund),
              `${vorn} über ${extrem}`,
            ).toBeGreaterThanOrEqual(4.5)
          }
        }
      })
    }

    // **Klar unterscheidbar**: kein Grund einer Farbwelt gleicht dem einer
    // anderen.
    it(`${welt}: unterscheidet sich von Pergament`, () => {
      expect(hell['--mse-grund']).not.toBe('#f2e8d5')
      expect(dunkel['--mse-grund']).not.toBe('#15120e')
    })
  }

  // **Pergament, die Vorgabe** — aus `stil.css`, mit denselben Paaren, den
  // Farben der Meldungen und dem Schleier. Bei 88 % stand die Warnung über
  // einem schwarzen Bildpunkt bei 4,45 : 1 (N4).
  for (const [bild, werte] of [
    ['hell', pergament(':root')],
    ['dunkel', pergament("html[data-erscheinungsbild='dunkel']")],
  ] as const) {
    it(`pergament, ${bild}: hält jedes Paar, auch unter dem Schleier`, () => {
      const lies = (merkmal: string) => werte[merkmal] ?? ''

      expect(
        lesbarkeitsbefunde(lies, { dunkel: bild === 'dunkel', schleier }),
      ).toEqual([])
    })
  }

  // **Das dunkle Bild gilt auch, wenn das System dunkel ist** und nichts
  // gewählt wurde — derselbe Block steht ein zweites Mal im Medienblock.
  it('führt jedes dunkle Bild auch für das dunkle System', () => {
    for (const welt of farbwelten) {
      expect(
        block(`    html[data-farbwelt='${welt}']:not([data-erscheinungsbild])`),
      ).toEqual(
        block(`html[data-farbwelt='${welt}'][data-erscheinungsbild='dunkel']`),
      )
    }
  })
})
