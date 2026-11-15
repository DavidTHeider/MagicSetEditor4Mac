// **Eine Karte kopieren** — der Haken hinter dem Menüeintrag.
//
// Bis Werkstück U2 von Phase 5.1
// (#1413) war
// das ein Bauteil mit eigenem Knopf (`seiten/Kartenkopieren.tsx`). Seit die
// Kartenhandlungen ein Menü sind, ist der Knopf ein Eintrag — und was er
// **tut**, steht hier: der Ruf, der Fehler, die Sperre gegen den zweiten
// Klick (§7, Regel 4).
//
// **Ein Wechsel der Karte räumt den Fehler** (§7, Regel 1): ein Fehler der
// vorigen Karte an der nächsten wäre eine falsche Aussage.
import { useState } from 'react'

import { api, auspacken } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { type Vorgang } from './useVorgang'

interface Kopiergriff {
  kopieren: () => void
  laeuft: boolean
  fehler: string | undefined
}

export function useKartenkopieren(
  setID: string,
  karteID: string,
  vorgang: Vorgang,
  aufKopiert: (neueID: string) => void,
): Kopiergriff {
  const texte = useTexte()
  const [fehler, setzeFehler] = useState<string>()

  const [zuletzt, setzeZuletzt] = useState(karteID)
  if (zuletzt !== karteID) {
    setzeZuletzt(karteID)
    setzeFehler(undefined)
  }

  const kopieren = () => {
    void vorgang.starte(`kopieren:${karteID}`, async () => {
      try {
        const antwort = await api().KarteKopieren(setID, karteID)
        const neue = auspacken(antwort, () => antwort.karteID)

        setzeFehler(undefined)
        aufKopiert(neue)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return { kopieren, laeuft: vorgang.laeuft(`kopieren:${karteID}`), fehler }
}
