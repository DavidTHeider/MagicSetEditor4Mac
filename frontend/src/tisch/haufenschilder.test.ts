// **Die Zahl steht am Haufen** — PLAY-03.AK-04, grafische Hälfte
// (#1711).
//
// # Der Befund
//
// **Sprachlich sagt der Tisch seit heute, was eine Zahl zählt** — „7 auf
// einem Haufen", „3 Marken +1/+1", „Spielstein". **Auf der Leinwand war die
// Haufenzahl weiterhin nirgends**: die Marken sitzen als Punkte an der Karte
// (`markenpunkte.ts`), und sieben Wälder übereinander sahen aus wie einer.
//
// Das Kriterium verlangt beides: „sprachlich **und grafisch**
// unterscheidbar".
//
// # Warum ein Schild und keine Punkte
//
// **Weil es etwas anderes zählt.** Marken gehören **zu** einer Karte und
// sitzen auf ihr; eine Haufenzahl zählt **Karten** und gehört an den Stapel,
// nicht auf die oberste. Wer beides als Punkte zeichnete, machte aus zwei
// Fragen eine.
//
// # Was er nicht prüft
//
// **Wie es aussieht.** `pixibeiwerk.ts` steht in `.covignore`, weil es
// **zeichnet**; und die Browservorschau verliert den WebGL-Kontext (Roadmap,
// G3). Geprüft ist, **wo** ein Schild sitzt und **was** darauf steht.
import { describe, expect, it } from 'vitest'

import { haufenschilder } from './haufenschilder'

import type { Tischlage, Tischplan } from './tischplan'

function eineLage(teile: Partial<Tischlage>): Tischlage {
  return {
    kennung: 'k1',
    x: 10,
    y: 20,
    breite: 63,
    hoehe: 88,
    drehung: 0,
    hoehenrang: 1,
    verdeckt: false,
    zone: 'tisch',
    sitz: 'sitz-1',
    seite: 'vorne',
    karteID: '7',
    name: 'Wald',
    geheim: false,
    zaehler: {},
    haengtAn: '',
    haufen: '',
    untergelegt: false,
    art: 'karte',
    haufenzahl: 1,
    ...teile,
  }
}

function einPlan(lagen: Tischlage[]): Tischplan {
  return {
    lagen,
    felder: [],
    sitze: [],
    zahlen: [],
    heimatlos: [],
    breite: 0,
    hoehe: 0,
  }
}

describe('haufenschilder', () => {
  it('nennt die Zahl eines Haufens', () => {
    const schilder = haufenschilder(
      einPlan([
        eineLage({ kennung: 'a', haufen: 'a', haufenzahl: 7, hoehenrang: 1 }),
        eineLage({ kennung: 'b', haufen: 'a', haufenzahl: 7, hoehenrang: 2 }),
      ]),
    )

    expect(schilder).toHaveLength(1)
    expect(schilder[0]?.text).toBe('7')
  })

  // **Am obersten und nicht am untersten.** Ein Haufen liegt versetzt; die
  // oberste Karte ist die, die man sieht, und an ihrer Ecke steht die Zahl.
  it('hängt am obersten Blatt des Haufens', () => {
    const schilder = haufenschilder(
      einPlan([
        eineLage({ kennung: 'a', haufen: 'a', haufenzahl: 2, x: 10, y: 20 }),
        eineLage({
          kennung: 'b',
          haufen: 'a',
          haufenzahl: 2,
          x: 14,
          y: 24,
          hoehenrang: 9,
        }),
      ]),
    )

    // **Auf die Stelle genau und nicht »größer als«.** Der erste Anlauf
    // prüfte `toBeGreaterThan(14)` — und blieb **grün**, als der Code statt
    // des obersten einfach das **erste** Blatt nahm: dessen Ecke liegt
    // ebenfalls rechts von 14. Eine Gegenprobe, die nichts ändert, ist keine
    // (§10); hier war der Lauf zu schwach und nicht der Eingriff wirkungslos.
    //
    // Die Ecke des obersten Blatts: 14 + 63 − 3 und 24 + 88 − 3.
    expect(schilder[0]?.x).toBe(74)
    expect(schilder[0]?.y).toBe(109)
  })

  // **Eine Karte ohne Haufen bekommt keine Zahl.** „1" an jeder Karte wäre
  // eine Auskunft, die man abzählen kann.
  it('schweigt, wo kein Haufen ist', () => {
    expect(haufenschilder(einPlan([eineLage({})]))).toEqual([])
  })

  // **Und jeder Haufen bekommt genau eines** — auch wenn zwei nebeneinander
  // liegen.
  it('gibt je Haufen ein Schild', () => {
    const schilder = haufenschilder(
      einPlan([
        eineLage({ kennung: 'a', haufen: 'a', haufenzahl: 2 }),
        eineLage({ kennung: 'b', haufen: 'a', haufenzahl: 2 }),
        eineLage({ kennung: 'c', haufen: 'c', haufenzahl: 3 }),
      ]),
    )

    expect(schilder).toHaveLength(2)
    expect(schilder.map((eines) => eines.text).sort()).toEqual(['2', '3'])
  })

  // **Die Kennung trägt den Haufen** — der Zeichner hält seine Textobjekte
  // daran, wie bei den Zonenschildern. Ohne sie baute er bei jedem Zug neue.
  it('trägt den Haufen in der Kennung', () => {
    const schilder = haufenschilder(
      einPlan([eineLage({ kennung: 'a', haufen: 'a', haufenzahl: 2 })]),
    )

    expect(schilder[0]?.kennung).toContain('a')
  })
})
