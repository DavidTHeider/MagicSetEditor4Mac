// Das Sichern eines Sets — hinter den Knöpfen der Setliste **und** hinter ⌘S.
//
// # Warum ein Haken und keine zweite Kopie
//
// ⌘S gilt im ganzen Fenster und steht deshalb in `App`; die Knöpfe stehen je
// Zeile in der Setliste (#162).
// Zwei Aufrufer, ein Vorgang — und zwei Kopien liefen an dem Tag auseinander,
// an dem eine von ihnen eine Regel aus §7 lernt und die andere nicht.
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist das Set. Zweimal
//     ⌘S auf dasselbe Set schriebe zweimal dieselbe Datei — der zweite Lauf
//     ginge gegen einen Stand, den das Backend noch nicht kennt.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einem Fehler bleibt die
//     Meldung stehen; sie ist das, was zu sehen ist.
//  3. **Eine Antwort gehört zu ihrem Set** (Regel 2 und 1): sie trägt seit
//     #749 die
//     Kennung mit, aus der sie stammt. Wer sie anzeigt, vergleicht.
//
// **Warum die Kennung mitreist und der Haken nicht selbst räumt.** Er weiß
// nicht, was „das offene Set" ist — ⌘S kennt es, die Setliste nicht. Und ein
// Aufräumen beim Wechsel liefe **vor** der Antwort, nicht nach ihr: ein Lauf,
// der eine Sekunde später zurückkommt, schriebe danach wieder hinein.
import { useCallback, useState } from 'react'

import { api, auspacken, type SpeicherAntwort } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Speicherlauf {
  speichern: (setID: string, unter: boolean) => void
  laeuft: (setID: string) => boolean

  speicherung: SpeicherAntwort | undefined
  fehler: string | undefined

  // fuer ist die Kennung des Sets, zu dem die Meldung gehört (#749).
  fuer: string | undefined

  raeumen: () => void
}

export function useSetSpeichern(): Speicherlauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [speicherung, setzeSpeicherung] = useState<SpeicherAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const [fuer, setzeFuer] = useState<string>()

  const raeumen = useCallback(() => {
    setzeSpeicherung(undefined)
    setzeFehler(undefined)
    setzeFuer(undefined)
  }, [])

  const speichern = useCallback(
    (setID: string, unter: boolean) => {
      void vorgang.starte(`speichern-${setID}`, async () => {
        setzeFehler(undefined)
        setzeSpeicherung(undefined)
        setzeFuer(setID)

        try {
          const antwort = unter
            ? await api().SetSpeichernUnter(setID)
            : await api().SetSpeichern(setID)

          auspacken(antwort, () => undefined)

          // **Ein abgebrochener Dialog ist keine Meldung wert.** Wer das
          // Sichern selbst abbricht, weiß, dass nichts geschrieben wurde.
          if (antwort.abgebrochen) {
            return
          }

          setzeSpeicherung(antwort)
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [vorgang, texte],
  )

  const laeuft = useCallback(
    (setID: string) => vorgang.laeuft(`speichern-${setID}`),
    [vorgang],
  )

  return { speichern, laeuft, speicherung, fehler, fuer, raeumen }
}
