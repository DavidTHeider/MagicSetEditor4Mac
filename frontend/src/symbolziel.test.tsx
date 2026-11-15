// **Ein Kartensymbolziel überlebt keinen Setwechsel** — Befund 3 des
// Zweitblicks über Phase 4.11.
//
// # Der Befund
//
// `symbolziel` nennt Karte und Feld, aber **nicht das Set**. Geräumt wurde es
// nur beim Aufschlagen einer anderen Karte; die vier Setwahlwege und
// `setSchliessen` taten es nicht.
//
// Danach verband der Komponentenschlüssel die **neue** Set-ID mit Karte und
// Feld des **alten** Sets. Beim Speichern wird zuerst die Symboldatei ins neue
// Set geschrieben und erst danach das Feld der alten Karte gesetzt — die
// Kartenaufsicht weist das zu Recht ab, und die Datei liegt schon da.
//
// **Das verletzt §7 Regel 1** und lässt einen halben Stand zurück.
//
// # Warum die Lösung eine Ableitung ist und kein Effekt
//
// Ein Effekt liefe **nach** dem Zeichnen: für einen Anstrich stünde das alte
// Ziel neben der neuen Set-ID. Das Ziel trägt deshalb sein Set bei sich, und
// wer es liest, prüft es — dieselbe Gestalt wie „eine Antwort gehört zu dem
// Stand, aus dem sie stammt" (§7, Regel 2), nur für einen Zustand.
import { describe, expect, it } from 'vitest'

import { geraeumtesZiel, zielImSet } from './seiten/symbolablegen'

describe('Ein Symbolziel gehört zu seinem Set', () => {
  it('gilt im Set, in dem es angelegt wurde', () => {
    expect(
      zielImSet({ setID: '7', karteID: '3', feld: 'symbol' }, '7'),
    ).toEqual({
      karteID: '3',
      feld: 'symbol',
    })
  })

  it('gilt in einem anderen Set nicht', () => {
    expect(
      zielImSet({ setID: '7', karteID: '3', feld: 'symbol' }, '9'),
    ).toBeUndefined()
  })

  // **Und ohne offenes Set erst recht nicht** — ⌘W schließt das Set, und die
  // Symbolseite darf danach nicht auf eine Karte zeigen, die niemand sieht.
  it('gilt ohne offenes Set nicht', () => {
    expect(
      zielImSet({ setID: '7', karteID: '3', feld: 'symbol' }, undefined),
    ).toBeUndefined()
  })

  // **Kein Ziel bleibt kein Ziel** — das Setsymbol.
  it('bleibt leer, wenn keines gesetzt ist', () => {
    expect(zielImSet(undefined, '7')).toBeUndefined()
  })
})

// **Und es kehrt nicht zurück** — Nachlauf über Phase 4.11.
//
// # Der Befund
//
// `zielImSet` **verbirgt** ein fremdes Ziel; es räumte den Zustand aber nicht.
// Wer Set A verließ und später wiederkam, fand das alte Ziel vor — und nach
// einem Zurückspielen des Bestands kann dieselbe Kennung inzwischen zu einer
// **anderen** Karte gehören.
//
// **Verbergen ist nicht Räumen** (§7, Regel 1).
describe('Ein fremdes Symbolziel wird geräumt', () => {
  it('bleibt, solange das Set dasselbe ist', () => {
    expect(
      geraeumtesZiel({ setID: '7', karteID: '3', feld: 'symbol' }, '7'),
    ).toEqual({ setID: '7', karteID: '3', feld: 'symbol' })
  })

  it('fällt weg, sobald ein anderes Set offen ist', () => {
    expect(
      geraeumtesZiel({ setID: '7', karteID: '3', feld: 'symbol' }, '9'),
    ).toBeUndefined()
  })

  it('fällt weg, wenn gar kein Set offen ist', () => {
    expect(
      geraeumtesZiel({ setID: '7', karteID: '3', feld: 'symbol' }, undefined),
    ).toBeUndefined()
  })

  it('bleibt leer, wenn keines gesetzt war', () => {
    expect(geraeumtesZiel(undefined, '7')).toBeUndefined()
  })
})
