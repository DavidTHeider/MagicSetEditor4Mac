// **Womit ein Gast sich setzt** — P9c
// (#2377):
// vorgewählt ist das Deck, mit dem zuletzt gearbeitet wurde — wie am eigenen
// Tisch (P6) —, und ein Wert der Wahl wird zu dem, was mitgeht.
import { describe, expect, it } from 'vitest'

import { mitbringAus, mitbringvorgabe } from './mitbringen'
import { leihwert } from './sitzwahl'

const decks = [{ id: '1' }, { id: '2' }]

describe('das Mitgebrachte', () => {
  it('wählt das zuletzt geöffnete eigene Deck vor', () => {
    expect(mitbringvorgabe(decks, ['9', '2', '1'])).toBe('2')
  })

  it('nimmt sonst das erste Deck, und ohne Deck keines', () => {
    expect(mitbringvorgabe(decks, ['9'])).toBe('1')
    expect(mitbringvorgabe([], ['9'])).toBe('')
  })

  it('macht aus dem Wert, was mitgeht', () => {
    expect(mitbringAus('2')).toEqual({ deckID: '2', leiheID: '' })
    expect(mitbringAus(leihwert('l1'))).toEqual({ deckID: '', leiheID: 'l1' })
    expect(mitbringAus('')).toEqual({ deckID: '', leiheID: '' })
  })
})
