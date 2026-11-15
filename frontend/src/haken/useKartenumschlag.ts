// **Eine Karte aus einem Umschlag annehmen** —
// #1639.
//
// # Warum sie hier steht und nicht im Bereich Handeln
//
// **Weil eine Karte in ein Set gehört und das Set hier offen ist.** Der
// Umschlag nennt kein Ziel; wer eine Karte annimmt, muss sagen, wohin — und
// das ist die Frage, die dieses Fenster ohnehin schon beantwortet hat.
// Dieselbe Überlegung hat `Deckumschlag` zu den Decks gestellt.
//
// # Und warum die angenommene Karte eine neue ist
//
// **Weil die Kennung aus einem fremden Bestand hier keine ist.** Wer sie
// übernähme, schriebe über eine Karte, von der der Absender nichts weiß.
//
// # Warum ein Haken und kein Knopf
//
// **Seit UXR-05 (#1913) steht der Weg im Menü „Mehr"** neben „Karte
// anlegen" (`seiten/Kartenwerkzeuge.tsx`), und ein Menüeintrag ist kein
// Bauteil, das seine Meldungen selbst zeichnet. Bis dahin war das hier der
// Baustein `Kartenumschlag` mit eigenem Knopf; was er tat, tut der Haken
// unverändert — die Zeile darüber zeigt, was er meldet.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Kartenumschlaggriff {
  // annehmen öffnet den Umschlag und legt die Karte im offenen Set an.
  annehmen: () => Promise<void>

  // laeuft hält den Eintrag an, solange ein Umschlag offen ist (§7, Regel 4):
  // zweimal denselben anzunehmen legte zwei Karten an.
  laeuft: boolean

  // fehler und satz sind, was der letzte Lauf zu sagen hatte.
  fehler: string | undefined
  satz: string | undefined
}

export function useKartenumschlag(
  // setID ist das offene Set — dorthin kommt die Karte.
  setID: string,
  // angenommen läuft, nachdem eine Karte hereingekommen ist.
  angenommen: () => void,
): Kartenumschlaggriff {
  const texte = useTexte()
  const eigene = texte.seiten.karten
  const lauf = useVorgang()

  const [fehler, setzeFehler] = useState<string>()
  const [satz, setzeSatz] = useState<string>()

  const kennzeichen = `kartenumschlag/${setID}`

  const annehmen = (): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeSatz(undefined)

      try {
        const antwort = await api().UmschlagKarte(setID)

        const kennung = auspacken(antwort, () => antwort.karteID)

        // **Ein abgebrochener Dialog ist keine Meldung wert** (§6) — und er
        // ist an der leeren Kennung zu erkennen.
        if (kennung !== '') {
          setzeSatz(eigene.umschlagAngenommen)
          angenommen()
        }
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return { annehmen, laeuft: lauf.laeuft(kennzeichen), fehler, satz }
}
