// **Das Wort unter dem Schreibzeiger** —
// #1164.
import { describe, expect, it } from 'vitest'

import { passende, wortstelle } from './wortstelle'

describe('Die Wortstelle', () => {
  it('findet das Wort links vom Zeiger', () => {
    expect(wortstelle('to_up', 5)).toEqual({ wort: 'to_up', von: 0, bis: 5 })
  })

  // **Der Punkt trennt.** In `card.na` ist `na` das Wort — wer dort
  // vervollständigt, meint ein Feld und nicht die Karte.
  it('trennt am Punkt', () => {
    expect(wortstelle('card.na', 7)).toEqual({ wort: 'na', von: 5, bis: 7 })
  })

  it('gibt nichts, wenn links kein Wort steht', () => {
    expect(wortstelle('1 + ', 4)).toEqual({ wort: '', von: 4, bis: 4 })
  })

  // **Nur was links steht.** Wer mitten in einem Wort tippt, meint den Teil
  // vor dem Zeiger.
  it('nimmt nur den Teil links vom Zeiger', () => {
    expect(wortstelle('length', 3)).toEqual({ wort: 'len', von: 0, bis: 3 })
  })

  it('kennt Ziffern und Unterstriche als Wortzeichen', () => {
    expect(wortstelle('card.type_1', 11).wort).toBe('type_1')
  })
})

describe('Die passenden Namen', () => {
  const namen = ['to_upper', 'to_lower', 'to_int', 'card', 'length']

  it('gibt die Namen, die das Wort fortsetzen', () => {
    expect(passende(namen, 'to_', 8)).toEqual([
      'to_upper',
      'to_lower',
      'to_int',
    ])
  })

  it('deckelt die Zahl', () => {
    expect(passende(namen, 'to_', 2)).toEqual(['to_upper', 'to_lower'])
  })

  // **Groß und klein zählen beim Suchen nicht.** MSE-Bezeichner sind durchweg
  // klein; wer `TO_` tippt, meint dieselben.
  it('sucht ohne Rücksicht auf Groß- und Kleinschreibung', () => {
    expect(passende(namen, 'TO_I', 8)).toEqual(['to_int'])
  })

  it('gibt nichts ohne Wort', () => {
    expect(passende(namen, '', 8)).toEqual([])
  })

  // **Ein fertig getipptes Wort schlägt sich nicht selbst vor.**
  //
  // Sonst stünde die Liste offen, sobald jemand `card` geschrieben hat — und
  // stünde damit dauerhaft im Weg.
  it('schlägt ein fertiges Wort nicht sich selbst vor', () => {
    expect(passende(namen, 'card', 8)).toEqual([])
  })

  it('gibt nichts, wenn nichts passt', () => {
    expect(passende(namen, 'zzz', 8)).toEqual([])
  })
})
