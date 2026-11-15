// **Was ein abgeschalteter Bereich im Fenster verbirgt** — OPS-10
// (#1908).
import { describe, expect, it } from 'vitest'

import { verborgeneSeiten } from './abschaltung'

function mit(werte: Record<string, string>) {
  return verborgeneSeiten((schluessel) => werte[schluessel] ?? '')
}

describe('verborgene Seiten', () => {
  it('verbirgt ohne Wahl nichts', () => {
    expect(mit({}).size).toBe(0)
  })

  it('verbirgt jede Seite eines abgeschalteten Bereichs', () => {
    expect([...mit({ 'schalter.spielen': 'aus' })].sort()).toEqual([
      'decks',
      'spielen',
      'turniere',
    ])
    expect([...mit({ 'schalter.ressourcen': 'aus' })].sort()).toEqual([
      'pakete',
      'vorlagen',
    ])
  })

  it('verbirgt die Seite einer abgeschalteten Funktion, die wie sie heißt', () => {
    expect([...mit({ 'schalter.turniere': 'aus' })]).toEqual(['turniere'])
  })

  it('lässt an, was „an“ sagt', () => {
    expect(mit({ 'schalter.spielen': 'an' }).size).toBe(0)
  })

  // **Der Weg zurück bleibt** — auch dann, wenn ein Schlüssel ankäme, den Go
  // nie annimmt: ohne die Einstellungen schaltete niemand wieder ein.
  it('verbirgt die Dienstzone nie', () => {
    expect(
      mit({ 'schalter.programm': 'aus', 'schalter.einstellungen': 'aus' }).size,
    ).toBe(0)
  })
})
