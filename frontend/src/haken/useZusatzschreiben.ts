// **Ein Zusatzwert wird geschrieben wie ein Feldwert** — mit Vorgang,
// Speicherstand und Meldung am Feld
// (#2298).
//
// **Ein Vorgang für alle Zusatzfelder einer Karte** (§7, Regel 4): sie liegen
// zusammen im Feld `extra data`, und der Dienst setzt einen Wert, indem er den
// Block liest, ändert und schreibt. Zwei Läufe zugleich schrieben den zweiten
// über den ersten.
import type { Katalog } from '../sprache/kataloge'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus, verursachendesFeld } from '../meldung'
import { schreibenBeginnt, schreibenEndet } from '../speicherstand'
import type { Vorgang } from './useVorgang'

// zusatzdaten ist das Kartenfeld, in dem die Zusatzwerte liegen.
export const zusatzdaten = 'extra data'

interface Zusatzwege {
  setID: string
  karteID: string
  vorgang: Vorgang
  texte: Katalog
  melden: (befund: { feld: string; meldung: string } | undefined) => void
  geglueckt: () => void
}

export function useZusatzschreiben({
  setID,
  karteID,
  vorgang,
  texte,
  melden,
  geglueckt,
}: Zusatzwege): (feld: string, wert: string) => Promise<void> {
  return async (feld: string, wert: string) => {
    const ergebnis = await vorgang.starte(
      `${karteID}/${zusatzdaten}`,
      async () => {
        schreibenBeginnt(setID)
        try {
          auspacken(
            await api().ZusatzwertSetzen(setID, karteID, feld, wert),
            () => undefined,
          )
          schreibenEndet(setID, karteID, zusatzdaten, true)
          return true
        } catch (ursache: unknown) {
          schreibenEndet(setID, karteID, zusatzdaten, false)
          melden({
            feld: verursachendesFeld(ursache, feld),
            meldung: meldungAus(ursache, texte),
          })
          return false
        }
      },
    )
    if (ergebnis === true) {
      geglueckt()
    }
  }
}
