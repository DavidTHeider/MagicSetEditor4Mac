// **Wohin ein Menüeintrag führt** — aus `Anwendung.tsx` herausgezogen, weil
// die Datei an der Zeilengrenze stand (§2, Werkstück U14 von Phase 5.1,
// #1490). Vier Rückrufe, die nichts als eine Seite setzen.
import { useCallback } from 'react'

import { istSeitenschluessel, type Seitenschluessel } from '../navigation'

export interface Menueziele {
  drucken: () => void
  hilfe: () => void
  seite: (schluessel: string) => void
  ueber: () => void
}

export function useMenueziele(
  setzeAktuell: (seite: Seitenschluessel) => void,
  setzeZuUeber: (zu: boolean) => void,
): Menueziele {
  const drucken = useCallback(() => {
    setzeAktuell('druckbogen')
  }, [setzeAktuell])

  const hilfe = useCallback(() => {
    setzeAktuell('hilfe')
  }, [setzeAktuell])

  // **Ein Menüeintrag führt auf eine Lasche** (#1168, Z4-01).
  //
  // **Der Schlüssel wird geprüft und nicht geglaubt.** Er kommt über die
  // Ereignisnaht aus Go; ein blindes `as Seitenschluessel` machte aus einem
  // Tippfehler dort eine Seite, die es nicht gibt — und das Fenster stünde
  // leer. Ein Wächter hält die beiden Listen zusammen
  // (`internal/werkzeuge/menueseiten_test.go`); diese Prüfung ist die zweite
  // Hälfte davon, für den Fall, dass jemand ihn umgeht.
  const seite = useCallback(
    (schluessel: string) => {
      if (istSeitenschluessel(schluessel)) {
        setzeAktuell(schluessel)
      }
    },
    [setzeAktuell],
  )

  // **„Über" führt auf die Hilfe und rollt zum Abschnitt.**
  //
  // Das Rollen steht in der Seite und nicht hier: welchen Abschnitt es gibt,
  // weiß sie, und ein Anker in dieser Datei wäre eine zweite Antwort darauf.
  const ueber = useCallback(() => {
    setzeAktuell('hilfe')
    setzeZuUeber(true)
  }, [setzeAktuell, setzeZuUeber])

  return { drucken, hilfe, seite, ueber }
}
