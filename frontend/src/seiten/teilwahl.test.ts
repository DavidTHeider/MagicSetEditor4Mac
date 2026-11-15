// **Die Mehrfachauswahl in der Teilliste** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// Portiert aus `SymbolPartsSelection::select`
// (`src/gui/symbol/selection.cpp`, Zeilen 25–56, Fassung `f9d9356` vom
// 2021-07-06).
import { describe, expect, it } from 'vitest'

import { gewaehltNach, nachEinemWegfall, type Wahl } from './teilwahl'

describe('Die Wahl in der Teilliste', () => {
  // **Ein gewöhnlicher Klick wählt genau eines** — `SELECT_OVERRIDE`.
  it('ersetzt die Wahl bei einem gewöhnlichen Klick', () => {
    expect(gewaehltNach([0, 2], 1, false)).toEqual([1])
  })

  // **Und lässt sie stehen, wenn schon genau dieses gewählt ist.**
  //
  // MSE gibt dafür `false` zurück (Zeile 35) und zeichnet nichts neu. Hier ist
  // es dieselbe Liste — React zeichnet dann ebenfalls nichts neu.
  it('lässt eine Einzelwahl auf sich selbst stehen', () => {
    const vorher: Wahl = [1]

    expect(gewaehltNach(vorher, 1, false)).toBe(vorher)
  })

  // **⌘-Klick nimmt hinzu** — `SELECT_TOGGLE`.
  it('nimmt bei gedrückter Befehlstaste hinzu', () => {
    expect(gewaehltNach([0], 2, true)).toEqual([0, 2])
  })

  // **Und nimmt wieder weg.**
  it('nimmt ein gewähltes Teil wieder aus der Wahl', () => {
    expect(gewaehltNach([0, 1, 2], 1, true)).toEqual([0, 2])
  })

  // **Die Wahl steht in Listenreihenfolge, nicht in Klickreihenfolge.**
  //
  // MSE führt ein `set` — es hat keine Klickreihenfolge, und alles, was die
  // Wahl liest, läuft ohnehin über die Liste. Hier ist es eine sortierte
  // Liste, damit dasselbe gilt, ohne dass jemand daran denken muss.
  it('hält die Wahl in Listenreihenfolge', () => {
    expect(gewaehltNach([2], 0, true)).toEqual([0, 2])
  })

  // **Die letzte Wahl lässt sich nicht wegnehmen.**
  //
  // Das ist **nicht** MSE — dort darf die Wahl leer sein. Hier hängt die
  // Zeichenfläche daran: sie zeigt die Griffe des gewählten Teils, und ohne
  // Wahl gäbe es keine, mit denen man weiterarbeitet. Ein Klick, der die
  // Bedienung wegnimmt, ist keine Bedienung.
  it('behält das letzte gewählte Teil', () => {
    const vorher: Wahl = [1]

    expect(gewaehltNach(vorher, 1, true)).toBe(vorher)
  })
})

// **Ein gelöschtes Teil rückt die Wahl nach** — §7, Regel 1 im Kleinen: die
// Stellen hinter dem gefallenen rutschen um eins vor, und eine Wahl, die
// stehen bliebe, zeigte danach auf ein anderes Teil.
describe('Die Wahl nach einem gelöschten Teil', () => {
  it('rückt die Stellen dahinter nach', () => {
    expect(nachEinemWegfall([0, 2, 3], 1)).toEqual([0, 1, 2])
  })

  it('nimmt das gelöschte selbst heraus', () => {
    expect(nachEinemWegfall([1, 2], 1)).toEqual([1])
  })

  // **Und lässt niemanden ohne Wahl zurück.**
  it('fällt auf die erste Stelle zurück', () => {
    expect(nachEinemWegfall([0], 0)).toEqual([0])
  })
})
