// **Der Kampf in Zeilen** — Artboard *Tisch*: „Aschenwächter → Hallenwart",
// „drei weitere ungeblockt → Bo".
//
// # Warum es sie gibt
//
// **Weil der Kampf im Kern stand und nirgends zu lesen war** (`M6`, CR 508.1
// und CR 509.1, Fassung 2026-09-25). Die Felder, mit denen man ihn erklärt,
// zeigen den Stand nur dem, der gerade erklärt; wer blocken will, muss
// wissen, wer angreift, und wer den Schaden verteilt, wer geblockt ist.
//
// # Was hier entschieden wird und was im Fenster
//
// **Hier: wer gegen wen**, mit den Namen, die die Sicht führt. **Dort: wie
// es sich liest** — ob drei ungeblockte eine Zeile bekommen oder drei.
import { kartenwort } from './feldwerte'

import type { Partie } from '../apipartie'

// Kampfzeile ist ein geblockter Angreifer — oder die ungeblockten eines
// Ziels.
export type Kampfzeile =
  | { art: 'geblockt'; angreifer: string; blocker: string[] }
  | { art: 'ungeblockt'; angreifer: string[]; ziel: string }

// kampfzeilen gibt den Kampf, wie er gerade steht: erst die geblockten
// Angreifer in der Reihenfolge des Angriffs, dann die ungeblockten je Ziel.
export function kampfzeilen(partie: Partie): Kampfzeile[] {
  // **Mehrere Blocker an einem Angreifer** — CR 509.1a lässt jeden Blocker
  // einen Angreifer wählen, und zwei dürfen denselben wählen.
  const blocker = new Map<string, string[]>()

  for (const block of partie.zug.bloecke) {
    const angreifer = block.angreifer ?? ''
    blocker.set(angreifer, [...(blocker.get(angreifer) ?? []), block.kreatur])
  }

  const geblockt: Kampfzeile[] = []
  const ungeblockt = new Map<string, string[]>()

  for (const angriff of partie.zug.angriffe) {
    const gegen = blocker.get(angriff.kreatur)

    if (gegen !== undefined) {
      geblockt.push({
        art: 'geblockt',
        angreifer: name(partie, angriff.kreatur),
        blocker: gegen.map((eine) => name(partie, eine)),
      })

      continue
    }

    const ziel = angriff.ziel ?? ''
    ungeblockt.set(ziel, [...(ungeblockt.get(ziel) ?? []), angriff.kreatur])
  }

  return [
    ...geblockt,
    ...[...ungeblockt].map(([ziel, angreifer]): Kampfzeile => ({
      art: 'ungeblockt',
      angreifer: angreifer.map((eine) => name(partie, eine)),
      ziel: name(partie, ziel),
    })),
  ]
}

// name gibt den Namen eines Objekts oder Sitzes — **die Kennung, wo die Sicht
// beides nicht führt** (§8). Ein Ziel ist ein Spieler oder eine Karte
// (CR 508.1b).
function name(partie: Partie, kennung: string): string {
  const objekt = partie.objekte.find((eines) => eines.kennung === kennung)
  if (objekt !== undefined) return kartenwort(objekt)

  return partie.sitze.find((sitz) => sitz.kennung === kennung)?.name ?? kennung
}
