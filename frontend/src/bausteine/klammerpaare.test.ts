// **Das Klammerpaar unter dem Schreibzeiger** — Werkstück B von Phase 4.9
// (#1324).
//
// Wie `wortstelle.test.ts`: reine Rechnung, kein DOM, keine Vorrichtung.
import { describe, expect, it } from 'vitest'

import { klammerpaar } from './klammerpaare'
import type { Merkmalstelle } from '../antworten'

// merkmale baut die Merkmalliste einer **einzeiligen** Quelle nach.
//
// **Von Hand und nicht aus dem Lexer.** Ein Test, der die Bindung ruft, prüft
// zwei Dinge zugleich; hier geht es um die Rechnung darüber.
//
// **Und sie führt auch Operatoren, die keine Klammern sind.** Der erste
// Anlauf legte nur Klammern ab — und ließ damit den häufigsten Fall der
// Wirklichkeit aus: in `f(1, 2)` ist das Komma ein Operator. Der Zweig, der
// ihn aussortiert, blieb unbetreten, und die Abdeckung hat es angestrichen.
function merkmale(quelle: string): Merkmalstelle[] {
  const stellen: Merkmalstelle[] = []
  let zeile = 1
  let spalte = 1

  for (const zeichen of quelle) {
    if (zeichen === '\n') {
      zeile += 1
      spalte = 1
      continue
    }

    if ('()[]{},+:'.includes(zeichen)) {
      stellen.push({ art: 'Operator', zeile, spalte, breite: 1 })
    }

    spalte += 1
  }

  return stellen
}

describe('klammerpaar', () => {
  it('findet die schließende zu der Klammer unter dem Zeiger', () => {
    const quelle = 'f(a, b)'

    // Der Zeiger steht **hinter** dem `(` — Stelle 2.
    const paar = klammerpaar(merkmale(quelle), quelle, 2)

    expect(paar).toEqual({
      hier: { art: 'Operator', zeile: 1, spalte: 2, breite: 1 },
      partner: { art: 'Operator', zeile: 1, spalte: 7, breite: 1 },
    })
  })

  it('findet die öffnende zu der Klammer unter dem Zeiger', () => {
    const quelle = 'f(a, b)'

    // Hinter dem `)` — Stelle 7.
    const paar = klammerpaar(merkmale(quelle), quelle, 7)

    expect(paar?.hier.spalte).toBe(7)
    expect(paar?.partner?.spalte).toBe(2)
  })

  it('zählt Schachtelung mit', () => {
    const quelle = 'f(g(1), 2)'

    expect(klammerpaar(merkmale(quelle), quelle, 2)?.partner?.spalte).toBe(10)
    expect(klammerpaar(merkmale(quelle), quelle, 4)?.partner?.spalte).toBe(6)
  })

  it('unterscheidet die drei Arten', () => {
    const quelle = '[a, (b), {c}]'

    expect(klammerpaar(merkmale(quelle), quelle, 1)?.partner?.spalte).toBe(13)
    expect(klammerpaar(merkmale(quelle), quelle, 5)?.partner?.spalte).toBe(7)
    expect(klammerpaar(merkmale(quelle), quelle, 10)?.partner?.spalte).toBe(12)
  })

  it('geht über Zeilen', () => {
    const quelle = 'f(\n  1,\n  2\n)'

    const paar = klammerpaar(merkmale(quelle), quelle, 2)

    expect(paar?.partner).toEqual({
      art: 'Operator',
      zeile: 4,
      spalte: 1,
      breite: 1,
    })
  })

  // **Der Fall, für den es die Anzeige gibt.** Ohne Partner ist die Klammer
  // der Befund — und genau das soll man sehen, bevor der Zerteiler es sagt.
  it('gibt keinen Partner, wenn keiner da ist', () => {
    const quelle = 'f(a, b'

    expect(klammerpaar(merkmale(quelle), quelle, 2)).toEqual({
      hier: { art: 'Operator', zeile: 1, spalte: 2, breite: 1 },
      partner: null,
    })
  })

  it('gibt keinen Partner für eine schließende ohne öffnende', () => {
    const quelle = 'a, b)'

    expect(klammerpaar(merkmale(quelle), quelle, 5)?.partner).toBeNull()
  })

  // **Eine falsch geschlossene Klammer hat keinen Partner.** `(a]` ist kein
  // Paar; wer hier eines zeigte, behauptete, der Text sei in Ordnung.
  it('paart nicht über die Art hinweg', () => {
    const quelle = '(a]'

    expect(klammerpaar(merkmale(quelle), quelle, 1)?.partner).toBeNull()
  })

  // **Ein Operator über zwei Zeichen ist keine Klammer** — auch dann nicht,
  // wenn er mit einer anfinge.
  //
  // Der Lexer gibt so etwas heute nicht her: seine mehrzeichigen Operatoren
  // sind `:=`, `==`, `!=`, `<=` und `>=` (`internal/msescript/lexer.go`), und
  // keiner beginnt mit einer Klammer. **Die Prüfung steht trotzdem da**, weil
  // sie sonst erst auffiele, wenn die Sprache einen bekäme — und dann als
  // falsch hervorgehobenes Paar, nicht als Fehler.
  it('nimmt keinen mehrzeichigen Operator für eine Klammer', () => {
    const quelle = '(= 1)'
    const alsOperator: Merkmalstelle[] = [
      { art: 'Operator', zeile: 1, spalte: 1, breite: 2 },
      { art: 'Operator', zeile: 1, spalte: 5, breite: 1 },
    ]

    expect(klammerpaar(alsOperator, quelle, 0)).toBeNull()
  })

  it('gibt nichts, wenn keine Klammer am Zeiger steht', () => {
    const quelle = 'f(a, b)'

    expect(klammerpaar(merkmale(quelle), quelle, 4)).toBeNull()
    expect(klammerpaar(merkmale(quelle), quelle, -1)).toBeNull()
  })

  // **Der Zeiger vor der Klammer zählt auch.** Wer an den Anfang der Zeile
  // springt und die Klammer rechts von sich hat, meint sie.
  it('nimmt auch die Klammer rechts vom Zeiger', () => {
    const quelle = '(a)'

    expect(klammerpaar(merkmale(quelle), quelle, 0)?.hier.spalte).toBe(1)
  })

  // **Eine Klammer in einer Zeichenkette ist keine.** Sie steht nicht in der
  // Merkmalliste — der Lexer in Go hat sie schon verschluckt, und genau
  // deshalb rechnet diese Datei über die Merkmale und nicht über den Text.
  it('sieht nur, was der Lexer als Operator führt', () => {
    const quelle = 'f("(")'
    const ohneZeichenkette: Merkmalstelle[] = [
      { art: 'Operator', zeile: 1, spalte: 2, breite: 1 },
      { art: 'Zeichenkette', zeile: 1, spalte: 3, breite: 3 },
      { art: 'Operator', zeile: 1, spalte: 6, breite: 1 },
    ]

    expect(klammerpaar(ohneZeichenkette, quelle, 2)?.partner?.spalte).toBe(6)
  })
})
