// **Womit ein Gast sich an einen fremden Tisch setzt** — P9c
// (#2377),
// Artboard `Netzpartie09`, Vermerk 0199.
//
// **Vorgewählt ist das eigene Deck, mit dem zuletzt gearbeitet wurde** —
// dieselbe Regel wie für den ersten Sitz am eigenen Tisch (P6, `vorbelegung`).
// Ein Sitz, der schon ein Deck hat, lässt das Mitgebrachte liegen; es
// mitzuschicken schadet also nie.
import { leihwert } from './sitzwahl'

import type { Partiemitbring } from '../apipartieweg'

// mitbringvorgabe gibt die Kennung des vorgewählten Decks — leer ohne eines.
export function mitbringvorgabe(
  decks: readonly { id: string }[],
  zuletzt: readonly string[],
): string {
  const eigen = (id: string) => decks.some((deck) => deck.id === id)

  return zuletzt.find(eigen) ?? decks[0]?.id ?? ''
}

// mitbringAus macht aus dem Wert der Deckwahl, was mitgeht — ein Deck, eine
// Leihe oder nichts.
export function mitbringAus(wert: string): Partiemitbring {
  const leihe = leihwert('')

  return wert.startsWith(leihe)
    ? { deckID: '', leiheID: wert.slice(leihe.length) }
    : { deckID: wert, leiheID: '' }
}
