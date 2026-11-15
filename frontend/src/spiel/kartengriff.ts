// **Was ein Griff an einer Karte tut** — P4
// (#2372),
// Befund C-02: am gezeichneten Tisch wählte ein Klick nur, und jede Handlung
// verlangte den Blick in die Spalte rechts. Jetzt löst ein Klick die
// **Primärhandlung** aus und ein Doppelklick die **Zweithandlung** — welche
// das sind, sagt der Kern (P3, `handlungen`), nicht das Fenster.
//
// # Ohne doppelte Primärwirkung
//
// **Ein Doppelklick besteht aus zwei Klicks.** Hat ein Objekt keine
// Zweithandlung, wirkt der erste sofort und der zweite nicht noch einmal.
// Hat es eine, wartet der erste kurz, ob ein zweiter kommt — sonst tappte
// ein Doppelklick auf ein Grundland es erst und wollte es dann für Mana
// tappen.
//
// # Auswahl bleibt
//
// **⌘-Klick wählt**, wie bisher ein Klick: die Handlungsleiste mit allen
// Formen hängt an der gewählten Karte, und ohne Primärhandlung wählt auch
// ein einfacher Klick.
import type { Objekthandlung } from '../apipartie'
import { befehlsformen } from './befehlsformen'

// doppelklickfrist ist, wie lange ein erster Klick auf einen zweiten wartet —
// nur bei einem Objekt mit Zweithandlung.
export const doppelklickfrist = 250

// handlungZu gibt die Auskunft des Kerns zu einem Objekt.
export function handlungZu(
  handlungen: readonly Objekthandlung[],
  kennung: string,
): Objekthandlung | undefined {
  return handlungen.find((eine) => eine.objekt === kennung)
}

// ohneFeld sagt, ob eine Befehlsart ohne ein weiteres Feld gestellt wird.
export function ohneFeld(art: string): boolean {
  return befehlsformen.find((form) => form.art === art)?.felder.length === 0
}

// Klick ist, was vom Ereignis zählt.
export interface Klick {
  // detail ist die Zahl der Klicks in Folge — 0 von der Tastatur.
  detail: number
  metaKey: boolean
}

// Griff ist, was ein Klick auslöst.
export type Griff =
  | { art: 'waehlen' }
  | { art: 'befehl'; befehl: string }
  | { art: 'warten'; befehl: string }
  | { art: 'zweit'; befehl: string }
  | { art: 'nichts' }

// griffFuer sagt, was ein Klick an diesem Objekt auslöst.
export function griffFuer(
  handlung: Objekthandlung | undefined,
  klick: Klick,
): Griff {
  if (klick.metaKey || handlung === undefined || handlung.primaer === '') {
    return klick.detail > 1 ? { art: 'nichts' } : { art: 'waehlen' }
  }

  if (handlung.zweit === '') {
    return klick.detail > 1
      ? { art: 'nichts' }
      : { art: 'befehl', befehl: handlung.primaer }
  }

  switch (klick.detail) {
    case 0:
      return { art: 'befehl', befehl: handlung.primaer }
    case 1:
      return { art: 'warten', befehl: handlung.primaer }
    case 2:
      return { art: 'zweit', befehl: handlung.zweit }
  }

  return { art: 'nichts' }
}
