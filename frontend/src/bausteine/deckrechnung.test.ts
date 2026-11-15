// **Die Rechnung eines Decks, tabellengetrieben** — Werkstück I von Phase 4.10
// (#1325).
import { describe, expect, it } from 'vitest'

import { farbenVon, farbverteilung, kurve, manawert } from './deckrechnung'

describe('manawert', () => {
  it.each([
    ['', 0],
    ['R', 1],
    ['UU', 2],
    ['3WW', 5],
    ['10', 10],
    ['2U', 3],
    ['X', 0],
    ['XR', 1],
    // **Hybride Symbole zählen, was zwischen den Schrägen steht** — dieselbe
    // Auslegung wie Scryfalls `cmc`.
    ['2/W', 3],
    // **Kleinbuchstaben zählen wie große.** MSE schreibt sie groß, aber ein
    // Paket von 2007 muss sich daran nicht gehalten haben (§8: fremde Daten
    // sind der Normalfall).
    ['3ww', 5],
  ])('rechnet %s zu %i', (kosten, erwartet) => {
    expect(manawert(kosten)).toBe(erwartet)
  })
})

describe('farbenVon', () => {
  it('gibt jede Farbe einmal, auch bei doppeltem Symbol', () => {
    expect(farbenVon('UU')).toEqual(['blau'])
  })

  it('gibt beide Farben einer zweifarbigen Karte', () => {
    expect(farbenVon('1WU')).toEqual(['weiss', 'blau'])
  })

  it('gibt nichts bei farblosen Kosten', () => {
    expect(farbenVon('3')).toEqual([])
  })
})

describe('kurve', () => {
  it('zählt über die Mengen und nicht über die Plätze', () => {
    expect(
      kurve([
        { kosten: 'R', menge: 4 },
        { kosten: 'R', menge: 2 },
      ]),
    ).toEqual([
      [0, 0],
      [1, 6],
    ])
  })

  // **Ohne Lücken**, damit ein Balkenbild nicht lügt.
  it('füllt fehlende Werte mit null', () => {
    expect(
      kurve([
        { kosten: 'R', menge: 4 },
        { kosten: '3WW', menge: 2 },
      ]),
    ).toEqual([
      [0, 0],
      [1, 4],
      [2, 0],
      [3, 0],
      [4, 0],
      [5, 2],
    ])
  })

  it('gibt nichts für ein leeres Deck', () => {
    expect(kurve([])).toEqual([])
  })
})

describe('farbverteilung', () => {
  // **WUBRG und nicht alphabetisch** — die Reihenfolge, in der Magic seine
  // Farben immer nennt.
  it('hält die Reihenfolge WUBRG', () => {
    expect(
      farbverteilung([
        { kosten: 'G', menge: 1 },
        { kosten: 'W', menge: 1 },
        { kosten: 'B', menge: 1 },
      ]),
    ).toEqual([
      ['weiss', 1],
      ['schwarz', 1],
      ['gruen', 1],
    ])
  })

  // **Eine mehrfarbige Karte zählt in jeder ihrer Farben** — die Summe ist
  // damit größer als das Deck, und das ist die Aussage.
  it('zählt eine zweifarbige Karte doppelt', () => {
    expect(farbverteilung([{ kosten: 'WU', menge: 3 }])).toEqual([
      ['weiss', 3],
      ['blau', 3],
    ])
  })

  it('nennt keine Farbe, die nicht vorkommt', () => {
    expect(farbverteilung([{ kosten: '3', menge: 4 }])).toEqual([])
  })
})

// **Der nachgelesene Wert schlägt den gerechneten** — Befund der Sichtprüfung
// zu U10 (2026-09-09): eine zweiseitige Karte hat bei Scryfall **leere**
// Kosten und trotzdem ein `cmc`. Die Kurve rechnete nur aus den Kosten und
// stellte sie damit in die Spalte null — oder ließ sie ganz weg.
describe('die Kurve und der nachgelesene Manawert', () => {
  it('nimmt den nachgelesenen Wert, wo die Kosten leer sind', () => {
    expect(
      kurve([
        { kosten: '', menge: 2, manawert: 3 },
        { kosten: 'R', menge: 4 },
      ]),
    ).toEqual([
      [0, 0],
      [1, 4],
      [2, 0],
      [3, 2],
    ])
  })

  it('lässt einen Posten weg, dessen Wert niemand kennt', () => {
    // **Unbekannt ist nicht null.** Ein Platz auf einen Druck, den Scryfall
    // nicht kennt, hat weder Kosten noch Wert; ihn in die Spalte null zu
    // stellen behauptete, er koste nichts.
    expect(
      kurve([
        { kosten: '', menge: 4 },
        { kosten: '2', menge: 1 },
      ]),
    ).toEqual([
      [0, 0],
      [1, 0],
      [2, 1],
    ])
  })

  it('gibt nichts, wenn kein Posten einen bekannten Wert hat', () => {
    expect(kurve([{ kosten: '', menge: 4 }])).toEqual([])
  })
})
