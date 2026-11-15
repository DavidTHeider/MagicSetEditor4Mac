// **„Karte erstellen" aus der Befehlspalette** — UX-02.AK-02
// (#1674).
//
// # Warum auf der Kartenseite und nicht in der Palette
//
// **Weil die Karte dort entsteht, wo sie immer entsteht**: mit dem Vorgang
// der Kartenseite (§7, Regel 4 — ein zweiter Wunsch während des ersten tut
// nichts), mit ihrem Fehler am Feld (§6) und mit der neuen Karte
// aufgeschlagen danach. Die Palette sagt nur, **dass** eine entstehen soll;
// ein zweiter Schreibweg daneben wäre eine zweite Fassung derselben Regeln.
//
// # Warum ein Effekt und kein `useAufgenommen`
//
// **Weil Anlegen ein Ruf nach draußen ist** und nicht ins Zeichnen gehört.
// Was schon erledigt ist, merkt sich eine Referenz, und der Wunsch wird
// verbraucht: er soll genau **eine** Karte ergeben, auch über eine neu
// gebaute Seite hinweg.
import { useEffect, useRef } from 'react'

import { useAnschluss } from '../anschlussraum'

import type { Anschluss } from '../anschluss'

export function useKartenwunsch(
  setID: string,
  anlegen: () => Promise<void>,
): void {
  const anschluss = useAnschluss()
  const wunsch = anschluss.offen?.art === 'karte' ? anschluss.offen : undefined
  const erledigt = useRef<Anschluss | undefined>(undefined)

  useEffect(() => {
    if (wunsch === undefined || erledigt.current === wunsch) {
      return
    }

    erledigt.current = wunsch

    // **Verbraucht, sobald entschieden ist** — angelegt oder verworfen. Die
    // Referenz allein hielt ihn nur für diese eine Seite: nach einem
    // Setwechsel entsteht die Kartenseite unter `key={setID}` neu, und die
    // neue legte bei jeder Rückkehr eine weitere Karte an (Zweitblick über
    // Phase 6.3, S7).
    anschluss.verbrauche()

    // **Nur im Set, für das er gedacht war** (§7, Regel 2): wer inzwischen
    // ein anderes geöffnet hat, bekäme sonst dort eine Karte.
    if (wunsch.setID === setID) {
      void anlegen()
    }
  })
}
