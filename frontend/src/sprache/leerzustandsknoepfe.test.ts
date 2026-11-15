// **Ein Leerzustand nennt den Knopf, der daneben steht — in jeder Sprache** —
// G02 (#2249).
//
// Wo die Handlung schon als Hauptknopf über der leeren Liste steht, wäre ein
// zweiter derselbe Knopf noch einmal; der Satz soll ihn **nennen**, mit der
// Beschriftung, die dasteht. Bis G02 sagte „Noch kein Deck" nur, wozu ein
// Deck ein Spiel braucht, und „Noch kein Turnier" ebenso — welcher Knopf
// eines beginnt, stand nirgends.
//
// **Was er nicht hält**: dass der Knopf wirklich daneben steht. Das sagen die
// Bilder `leer-<seite>.png` der Zustandsprobe (`scripts/huellenprobe/zustaende.mjs`).
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'

type Katalog = (typeof kataloge)['de']

const paare: [string, (k: Katalog) => string, (k: Katalog) => string][] = [
  ['Decks', (k) => k.seiten.decks.leer, (k) => k.seiten.decks.neu],
  [
    'Turniere, anlegen',
    (k) => k.seiten.turniere.leer,
    (k) => k.seiten.turniere.neuOeffnen,
  ],
  [
    'Turniere, einlesen',
    (k) => k.seiten.turniere.leer,
    (k) => k.seiten.turniere.einlesen,
  ],
  [
    'Handeln',
    (k) => k.seiten.handeln.liste.leer,
    (k) => k.seiten.handeln.neuerTausch,
  ],
]

describe('Leerzustände nennen ihren Knopf', () => {
  for (const [sprache, katalog] of Object.entries(kataloge)) {
    for (const [stelle, leer, knopf] of paare) {
      it(`${stelle} (${sprache})`, () => {
        expect(leer(katalog)).toContain(knopf(katalog))
      })
    }
  }
})
