// **Ein Ort, den ein Dateidialog wählt** — UXR-F15
// (#2030).
//
// # Was hier steht und was nicht
//
// **Was die Wahl tut, steht hier; wie sie aussieht, in `Pfadwahl`** (§2). Der
// Dialog gehört der Go-Seite — das Fenster bekäme einen Namen und keinen
// Pfad —, und dieser Haken hält, was danach gilt: der gewählte Pfad, der Grund
// eines gescheiterten Dialogs, und ob einer offen ist.
//
// # Drei Zusagen
//
//   - **Abbrechen lässt den vorigen Stand stehen.** Ein weggeklickter Dialog
//     ist keine Wahl und keine Meldung; was gewählt war, bleibt. Dasselbe
//     gilt für einen gescheiterten: der Grund steht am Feld, der Pfad bleibt
//     (§7, Regel 3 — der letzte gute Stand wird nicht für einen Fehler
//     hergegeben).
//   - **Ein Dialog je Wahl** (§7, Regel 4): `useVorgang` lässt keinen zweiten
//     daneben aufgehen.
//   - **Der nächste Versuch räumt die Meldung des vorigen.**
//
// **Die Regeln 1 und 2 löst der Aufrufer**: wer das Profil wechselt, baut die
// Seite neu auf (`key`), und eine Antwort, die danach kommt, trägt in ein
// Bauteil ein, das es nicht mehr gibt.
import { useCallback, useState } from 'react'

import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken, type Umschlag } from '../umschlag'
import { useVorgang } from './useVorgang'

// Wahlantwort ist, was jeder Dialog zurückgibt, der einen Pfad wählt.
export interface Wahlantwort extends Umschlag {
  pfad: string
  abgebrochen: boolean
}

export interface Wahlstand {
  // pfad ist der gewählte Ort — leer, solange keiner gewählt wurde.
  pfad: string

  // fehler ist der Grund, aus dem der letzte Dialog scheiterte.
  fehler: string | undefined

  // laeuft sagt, dass gerade ein Dialog offen ist.
  laeuft: boolean

  // waehle öffnet den Dialog.
  waehle: () => void
}

export function usePfadwahl(
  frage: () => Promise<Wahlantwort>,
  kennung: string,
): Wahlstand {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [pfad, setzePfad] = useState('')
  const [fehler, setzeFehler] = useState<string>()

  const waehle = useCallback(() => {
    void vorgang.starte(kennung, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await frage()

        auspacken(antwort, () => undefined)

        // **Ein Abbruch ist keine Wahl**: der vorige Pfad bleibt.
        if (!antwort.abgebrochen) {
          setzePfad(antwort.pfad)
        }
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [vorgang, kennung, frage, texte])

  return { pfad, fehler, laeuft: vorgang.laeuft(kennung), waehle }
}
