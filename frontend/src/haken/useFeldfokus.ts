// **Den Fokus in dasselbe Feld der nächsten Karte zurückholen** —
// #1290.
//
// Warum es ihn gibt und warum er einen Beobachter benutzt, steht im Lauf
// daneben.

import { useEffect } from 'react'

export function useFeldfokus(
  wunsch: string | undefined,
  wuenschen: (feld: string | undefined) => void,
): void {
  useEffect(() => {
    if (wunsch === undefined) {
      return undefined
    }

    // **Nur beschriftete Eingabefelder.** Ein `aria-label` steht auch an
    // Knöpfen; den Fokus auf einen Knopf zu setzen, weil er zufällig `name`
    // heißt, wäre schlimmer als gar nichts.
    const suchen = () =>
      document.querySelector<HTMLElement>(
        `input[aria-label="${wunsch}"], textarea[aria-label="${wunsch}"]`,
      )

    const holen = () => {
      const feld = suchen()
      if (feld === null) {
        return false
      }

      // **Nur, wenn niemand den Fokus inzwischen selbst gesetzt hat.**
      //
      // Befund 6 des Zweitblicks über Phase 4.3: der Haken fokussierte sein
      // verspätetes Ziel, ohne zu prüfen, wo der Fokus gerade steht. Wer
      // während des Ladens weiterklickt oder weitertabbt, bekam ihn
      // **rückwirkend** weggenommen — eine alte Absicht schlug eine neue.
      //
      // `document.body` heißt „niemand hat ihn": genau der Zustand nach dem
      // Abräumen des alten Feldes.
      if (
        document.activeElement !== null &&
        document.activeElement !== document.body
      ) {
        wuenschen(undefined)

        return true
      }

      feld.focus()
      wuenschen(undefined)

      return true
    }

    if (holen()) {
      return undefined
    }

    const beobachter = new MutationObserver(() => {
      if (holen()) {
        beobachter.disconnect()
      }
    })

    beobachter.observe(document.body, { childList: true, subtree: true })

    return () => {
      beobachter.disconnect()
    }
  }, [wunsch, wuenschen])
}
