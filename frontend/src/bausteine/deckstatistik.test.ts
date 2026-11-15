// **Typen, Länder, Durchschnitt und Quellen** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Gegen handgerechnete Beispiele**, wie das Issue es verlangt: jede Zahl
// hier ist von Hand nachgerechnet und steht im Fall dabei.
import { describe, expect, it } from 'vitest'

import {
  durchschnitt,
  laenderzahl,
  quellen,
  typverteilung,
} from './deckstatistik'

import type { Statistikposten } from './deckstatistik'

// posten baut einen Posten — kurz, weil die Fälle davon leben.
function posten(
  typzeile: string,
  kosten: string,
  menge = 1,
  manawert?: number,
): Statistikposten {
  // **`exactOptionalPropertyTypes`**: ein Feld mit `undefined` ist nicht
  // dasselbe wie ein fehlendes — und die Naht gibt es nur, wo es eines gibt.
  return manawert === undefined
    ? { typzeile, kosten, menge }
    : { typzeile, kosten, menge, manawert }
}

describe('die Typverteilung', () => {
  it('zählt den ersten Obertyp und fasst gleiche zusammen', () => {
    // **Zwei Kreaturen (3 + 1), ein Land (4), ein Spontanzauber (2).**
    expect(
      typverteilung([
        posten('Creature — Elf Druid', '{G}', 3),
        posten('Legendary Creature — Beast', '{4}{G}', 1),
        posten('Basic Land — Forest', '', 4),
        posten('Instant', '{R}', 2),
      ]),
    ).toEqual([
      ['Creature', 4],
      ['Land', 4],
      ['Instant', 2],
    ])
  })

  it('lässt eine Karte ohne Typzeile weg — sie ist keine Aussage', () => {
    expect(typverteilung([posten('', '{G}', 3)])).toEqual([])
  })

  it('nimmt den Obertyp ohne die Auszeichnungen davor', () => {
    // **`Legendary`, `Basic`, `Snow` und `Artifact` sind Auszeichnungen**;
    // gezählt wird der Kartentyp, sonst stünde jede Legende für sich.
    expect(
      typverteilung([
        posten('Snow Land — Forest', '', 1),
        posten('Artifact Creature — Golem', '{3}', 1),
      ]),
    ).toEqual([
      ['Land', 1],
      ['Creature', 1],
    ])
  })
})

describe('die Länderzahl', () => {
  it('zählt Stücke und nicht Zeilen', () => {
    expect(
      laenderzahl([
        posten('Basic Land — Forest', '', 9),
        posten('Land', '', 2),
        posten('Creature — Elf', '{G}', 4),
      ]),
    ).toBe(11)
  })
})

describe('der Durchschnitt', () => {
  it('rechnet über die Nichtländer und rundet auf zwei Stellen', () => {
    // **(1×2 + 2×3 + 1×0) / 4 = 2** — das Land zählt nicht mit.
    expect(
      durchschnitt([
        posten('Creature — Elf', '{1}{G}', 1),
        posten('Instant', '{2}{R}', 2),
        posten('Artifact', '{0}', 1),
        posten('Basic Land — Forest', '', 20),
      ]),
    ).toBe(2)
  })

  it('nimmt den nachgelesenen Manawert, wo es einen gibt', () => {
    // **Eine zweiseitige Karte hat leere Kosten und trotzdem einen Wert.**
    // (3 + 1) / 2 = 2
    expect(
      durchschnitt([
        posten('Creature — Werewolf', '', 1, 3),
        posten('Instant', '{R}', 1),
      ]),
    ).toBe(2)
  })

  it('ist ohne Nichtländer nicht null, sondern keine Zahl', () => {
    expect(
      durchschnitt([posten('Basic Land — Forest', '', 20)]),
    ).toBeUndefined()
    expect(durchschnitt([])).toBeUndefined()
  })
})

describe('die Quellen', () => {
  it('zählt je Farbe die Länder, die sie geben — nach dem Regeltext', () => {
    // **Vier Wälder geben Grün, zwei Inseln Blau; die Kreatur zählt nicht.**
    expect(
      quellen([
        { ...posten('Basic Land — Forest', '', 4), farben: ['gruen'] },
        { ...posten('Basic Land — Island', '', 2), farben: ['blau'] },
        { ...posten('Creature — Elf', '{G}', 4), farben: ['gruen'] },
      ]),
    ).toEqual([
      ['blau', 2],
      ['gruen', 4],
    ])
  })

  it('zählt ein Zweifarbenland in beiden Farben', () => {
    expect(
      quellen([
        {
          ...posten('Land — Forest Island', '', 3),
          farben: ['blau', 'gruen'],
        },
      ]),
    ).toEqual([
      ['blau', 3],
      ['gruen', 3],
    ])
  })
})

// **Der Schnitt zählt nur, was er kennt** — Befund der Sichtprüfung zu U10
// (2026-09-09): im Messdeck stand ein Platz auf einen Druck, den Scryfall
// nicht kennt. Er ging als Nullkostenkarte in die Rechnung und zog den
// Schnitt von 1,00 auf 0,92.
describe('der Durchschnitt und das Unbekannte', () => {
  it('lässt eine Karte aus, deren Wert niemand kennt', () => {
    // Blitz 4× zu 1 und Goyf 4× zu 2 sind 12 auf 8 Stück — 1,5.
    // Der unbekannte Platz zählt nicht mit; mit ihm wären es 12 auf 12.
    expect(
      durchschnitt([
        { typzeile: 'Instant', kosten: '{R}', menge: 4, manawert: 1 },
        { typzeile: 'Creature — Lhurgoyf', kosten: '{1}{G}', menge: 4 },
        { typzeile: '', kosten: '', menge: 4 },
      ]),
    ).toBe(1.5)
  })

  it('nimmt eine Karte mit, deren Wert nachgelesen ist', () => {
    // Zweiseitig: keine Kosten, aber ein Wert. Zusammen mit 1× zu 1 macht
    // das 4 auf 2 Stück — 2,00.
    expect(
      durchschnitt([
        { typzeile: 'Sorcery // Land', kosten: '', menge: 1, manawert: 3 },
        { typzeile: 'Instant', kosten: '{R}', menge: 1 },
      ]),
    ).toBe(2)
  })

  it('gibt nichts, wenn kein Nichtland einen bekannten Wert hat', () => {
    expect(
      durchschnitt([
        { typzeile: 'Basic Land — Forest', kosten: '', menge: 12 },
        { typzeile: '', kosten: '', menge: 4 },
      ]),
    ).toBeUndefined()
  })
})

// **Die Vorderseite entscheidet** — CR 712.8a. Gefunden beim Bau der Läufe
// oben (2026-09-09): `Sorcery // Land` galt als Land, weil die ganze Zeile
// nach dem Wort abgesucht wurde.
describe('zweiseitige Karten', () => {
  it('zählt eine Zauberei mit Landrückseite nicht als Land', () => {
    const posten = [
      { typzeile: 'Sorcery // Land', kosten: '', menge: 4, manawert: 6 },
    ]

    expect(laenderzahl(posten)).toBe(0)
    expect(typverteilung(posten)).toEqual([['Sorcery', 4]])
  })

  it('zählt ein Land mit Landrückseite als Land', () => {
    const posten = [
      { typzeile: 'Land // Land', kosten: '', menge: 2, manawert: 0 },
    ]

    expect(laenderzahl(posten)).toBe(2)
  })
})
