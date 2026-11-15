// **Eine Kennung ohne Leerraum** —
// #1278,
// Befund 5.
import { describe, expect, it } from 'vitest'

import { fehlerkennung, feldkennung, hilfekennung } from './feldkennung'

describe('feldkennung', () => {
  // **Der Normalfall bleibt, wie er ist.** Ein Feldname ohne Sonderzeichen
  // soll im DOM stehen, wie er heißt.
  it('lässt einen unbedenklichen Namen unverändert', () => {
    expect(feldkennung('name')).toBe('name')
    expect(hilfekennung('name')).toBe('hilfe-name')
    expect(fehlerkennung('name')).toBe('fehler-name')
  })

  // **Und der echte Fall aus dem Pack.** `archenemy.mse-game` nennt sein
  // Textfeld `rule text`; aus `hilfe-rule text` wurden zwei Verweise, und
  // keiner zeigte auf einen Knoten.
  it('ersetzt Leerzeichen — sonst zerfällt aria-describedby', () => {
    expect(hilfekennung('rule text')).not.toContain(' ')
    expect(hilfekennung('casting cost')).not.toContain(' ')
    expect(fehlerkennung('card number')).not.toContain(' ')
  })

  // **Und alles andere, was eine Auswahl zerlegt.** Ein fremdes Paket darf
  // jeden Namen führen.
  it('ersetzt auch Punkt, Doppelpunkt und Schrägstrich', () => {
    for (const roh of ['a.b', 'a:b', 'a/b', 'a#b', 'a[b]']) {
      expect(feldkennung(roh)).toMatch(/^[A-Za-z0-9_-]+$/)
    }
  })

  // **Zwei Felder, zwei Kennungen — und zwar für jedes Paar.**
  //
  // Der erste Anlauf ersetzte jede Folge von Sonderzeichen durch `_` und
  // hängte die Länge an; damit wurden `a b`, `a.b`, `a:b`, `a/b` und `a#b`
  // **alle** zu `a_b_3`. Der Lauf dazu verglich nur Paare mit verschiedener
  // Länge und ging durch. Angestrichen vom Nachlauf über die Nacharbeit
  // (#1282,
  // Befund 6).
  it('gibt verschiedenen Namen verschiedene Kennungen', () => {
    const namen = [
      'a b',
      'a.b',
      'a:b',
      'a/b',
      'a#b',
      'a-b',
      'a_b',
      'a  b',
      'ab',
      'a-20-b',
      'rule text',
      'rule-text',
    ]

    const kennungen = namen.map(feldkennung)

    expect(new Set(kennungen).size).toBe(namen.length)
  })

  // **Und jede ist eine gültige Kennung.**
  it('gibt nur unbedenkliche Zeichen aus', () => {
    for (const roh of ['a b', 'a.b', 'ä ö', 'a\tb', '🙂 x', 'a/b']) {
      expect(feldkennung(roh)).toMatch(/^[A-Za-z0-9_-]+$/)
    }
  })
})
