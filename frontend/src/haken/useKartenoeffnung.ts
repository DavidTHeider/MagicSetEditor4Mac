// **Eine Karte aufschlagen — und wohin das Fenster dabei geht.**
//
// Aus `Anwendung.tsx` gezogen, als die Schubladen dazukamen (UX-08,
// #1680):
// seitdem hängt der Seitenwechsel davon ab, **wo** man gerade steht, und das
// ist eine Entscheidung mit eigenem Grund.
import { useCallback, type Dispatch, type SetStateAction } from 'react'

import type { Seitenschluessel } from '../navigation'
import { schubladeVon } from '../schubladen'
import type { Symbolzielsteuerung } from './useSymbolziel'

interface Kartenoeffnungsziele {
  setzeSetID: (setID: string) => void
  setzeKarteID: (karteID: string) => void
  symbolziel: Pick<Symbolzielsteuerung, 'oeffne' | 'schliesse'>
  setzeAktuell: Dispatch<SetStateAction<Seitenschluessel>>
}

export interface Kartenoeffnung {
  // oeffneKarte merkt sich Set und Karte und führt zu den Karten.
  oeffneKarte: (setID: string, karteID: string) => void

  // oeffneSymbol führt zum Symboleditor für ein Kartenfeld (#1339).
  oeffneSymbol: (karteID: string, feld: string) => void
}

export function useKartenoeffnung({
  setzeSetID,
  setzeKarteID,
  symbolziel,
  setzeAktuell,
}: Kartenoeffnungsziele): Kartenoeffnung {
  const { oeffne, schliesse } = symbolziel

  const oeffneKarte = useCallback(
    (gewaehltesSet: string, gewaehlteKarte: string) => {
      setzeSetID(gewaehltesSet)
      setzeKarteID(gewaehlteKarte)

      // **Eine andere Karte räumt das Symbolziel mit** (§7, Regel 1).
      schliesse()

      // **Eine offene Schublade bleibt offen.** Die Karten darunter melden
      // ihre Wahl auch, wenn niemand sie angeklickt hat — die erste Karte
      // beim Öffnen (#668, `useKartenmeldung`). Ein Wechsel auf *Karten*
      // schlösse die Schublade, die gerade aufging; von überall sonst — der
      // Suche etwa — führt die Karte zu den Karten.
      setzeAktuell((jetzt) =>
        schubladeVon(jetzt) === undefined ? 'karten' : jetzt,
      )
    },
    [setzeSetID, setzeKarteID, schliesse, setzeAktuell],
  )

  const oeffneSymbol = useCallback(
    (gewaehlteKarte: string, feld: string) => {
      oeffne(gewaehlteKarte, feld)
      setzeAktuell('symbol')
    },
    [oeffne, setzeAktuell],
  )

  return { oeffneKarte, oeffneSymbol }
}
