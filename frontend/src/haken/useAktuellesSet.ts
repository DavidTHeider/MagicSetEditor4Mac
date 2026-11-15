// **Regel 2 aus AGENTS.md §7, für die Kartenlasche** —
// #716.
//
// # Wofür
//
// `Karten` bleibt beim Setwechsel **dieselbe** React-Komponente. Eine Handlung,
// die währenddessen läuft — anlegen, kopieren, umordnen, löschen —, trug ihre
// Antwort danach in das **neue** Set ein: die Karte des alten stand
// aufgeschlagen da, die Liste wurde nachgeladen und der Bestand als geändert
// gemeldet. Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// # Warum über den Rückruf und nicht über einen Zähler
//
// Ein Rückruf kennt das Set, aus dem er stammt, ohnehin: er entsteht in dem
// Zeichenvorgang, in dem geklickt wurde, und schließt dessen `setID` ein. Der
// Vergleich mit dem **heutigen** Stand genügt — und er ist billiger als ein
// Zähler, den jede Stelle mitführen müsste.
//
// **Der Ref wird im Effekt geschrieben und im Rückruf gelesen**, nie beim
// Zeichnen: React verbietet das, und der Grund ist derselbe wie unsere Regel —
// was während des Zeichnens geschieht, gehört nicht zum Zeichnen.
import { useCallback, useEffect, useRef } from 'react'

export interface Aktuellheit {
  // nochAktuell sagt, ob ein Bezugsobjekt noch das heutige ist.
  nochAktuell: (bezug: string) => boolean

  // nurAktuell lässt einen Rückruf nur durch, solange sein Bezugsobjekt gilt.
  //
  // **Variadisch**, damit dieselbe Wache für Rückrufe mit und ohne Wert gilt —
  // zwei Fassungen wären zwei Zweige, von denen die Tests nur einen beträten
  // (§3).
  nurAktuell: <A extends unknown[]>(
    bezug: string,
    tun: (...args: A) => void,
  ) => (...args: A) => void
}

export function useAktuellesSet(bezug: string): Aktuellheit {
  const heute = useRef(bezug)

  useEffect(() => {
    heute.current = bezug
  }, [bezug])

  const nochAktuell = useCallback(
    (gemeint: string) => heute.current === gemeint,
    [],
  )

  const nurAktuell = useCallback(
    <A extends unknown[]>(gemeint: string, tun: (...args: A) => void) =>
      (...args: A) => {
        if (heute.current !== gemeint) {
          return
        }

        tun(...args)
      },
    [],
  )

  return { nochAktuell, nurAktuell }
}
