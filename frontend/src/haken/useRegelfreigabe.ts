// **Was vor „Bereit" durchgesetzt wird** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum die Frage vor dem Anfangen steht
//
// **Das Issue verlangt es**: „Spieler sehen vor Start, welche Regeln
// automatisch durchgesetzt werden." Danach wäre es zu spät — dann steht die
// Partie, und die Frage war, ob sie so stehen soll.
//
// # Die Regeln aus §7
//
//  1. **Ein Wechsel räumt** (Regel 1): wer das Spiel oder die Zahl der Sitze
//     ändert, bekommt eine neue Antwort; die alte gehört zu einem Stand, den
//     es nicht mehr gibt.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2) —
//     das hält `useLaden` über seine Abhängigkeiten.
import { useState } from 'react'

import { useLaden } from './useLaden'

import { api, auspacken } from '../api'

import type { RegelfreigabeDTO } from '../apiregeln'

// Regelfreigabelauf ist, was die Seite bekommt.
export interface Regelfreigabelauf {
  freigabe: RegelfreigabeDTO | undefined

  laeuft: boolean
  fehler: string | undefined
}

export function useRegelfreigabe(
  spiel: string,
  strenge: string,
  sitze: number,

  // modus und abgewaehlt kamen mit PLAY-02
  // (#1710).
  //
  // **Beide gehören in die Abhängigkeiten** (§7, Regel 2): wer den Modus
  // wechselt, bekommt eine neue Antwort — die alte gehört zu einem Stand,
  // den es nicht mehr gibt.
  modus = '',
  abgewaehlt: readonly string[] = [],
): Regelfreigabelauf {
  const stand = useLaden(async () => {
    // **Ohne Spiel wird nicht gefragt.** Solange kein Deck gewählt ist, gibt
    // es kein Spiel — und eine Anfrage darüber bekäme eine Absage auf eine
    // Frage, die niemand gestellt hat. Genau daran hing schon einmal ein
    // Fehler im Inspector eines frisch angelegten Decks
    // (`useDeckpruefung`, 2026-09-09).
    if (spiel === '') {
      return undefined
    }

    const antwort = await api().Regelfreigabe(spiel, strenge, sitze, modus, [
      ...abgewaehlt,
    ])

    return auspacken(antwort, () => antwort.freigabe)

    // **Die Abwahl geht als Zeichenkette in die Abhängigkeiten**: ein Feld
    // ist bei jedem Zeichnen ein neues, und `useLaden` verglich dann bei
    // jedem Strich neu (§7, Regel 2).
  }, [spiel, strenge, String(sitze), modus, abgewaehlt.join(',')])

  // **Eine Wahl an der Tafel räumt die Tafel nicht** — P1a
  // (#2335),
  // gefunden an der Hüllenprobe `p1a`. Während die Antwort auf die neue
  // Abwahl unterwegs ist, gab `useLaden` nichts zurück: die Regeltafel
  // verschwand für einen Augenblick, und danach war die Gruppe wieder zu
  // und der Fokus fort.
  //
  // **Behalten wird nur die Antwort desselben Spiels** mit derselben
  // Strenge und Sitzzahl (§7, Regel 1): Modus und Abwahl sind eine Wahl
  // **an** dieser Antwort und kein neues Bezugsobjekt. Wer das Deck und
  // damit das Spiel wechselt, bekommt erst die neue.
  //
  // **Gemerkt wird im Zustand und nicht in einer Referenz**: die Antwort ist
  // Teil dessen, was gezeichnet wird, und React liest beim Zeichnen keine
  // Referenz (`react-hooks/refs`).
  const bezug = [spiel, strenge, String(sitze)].join('|')
  const [zuletzt, setzeZuletzt] = useState<
    { bezug: string; freigabe: RegelfreigabeDTO } | undefined
  >(undefined)

  if (stand.daten !== undefined && zuletzt?.freigabe !== stand.daten) {
    setzeZuletzt({ bezug, freigabe: stand.daten })
  }

  const behalten =
    stand.laedt && zuletzt?.bezug === bezug ? zuletzt.freigabe : undefined

  return {
    freigabe: stand.daten ?? behalten,
    laeuft: stand.laedt,
    fehler: stand.fehler,
  }
}
