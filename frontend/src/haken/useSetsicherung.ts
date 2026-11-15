// **⌘S und ⇧⌘S** — was das Fenster mit dem offenen Set tut.
//
// # Warum eine eigene Datei
//
// `Anwendung.tsx` stand an der Zeilengrenze aus §2, als der Symboleditor
// seinen eigenen Verlauf bekam
// (#1340).
// Der Schnitt folgt der Frage und nicht der Zeilenzahl: *wie* gesichert wird,
// steht hier; *dass* die Menüleiste es anbietet, dort.

import { useCallback } from 'react'

import { useSetSpeichern } from './useSetSpeichern'

export interface Setsicherung {
  sichernJetzt: () => void
  sichernUnter: () => void

  // meldung und fehler reicht `useSetSpeichern` durch — der Aufrufer zeigt sie.
  speichern: ReturnType<typeof useSetSpeichern>
}

export function useSetsicherung(setID: string | undefined): Setsicherung {
  const speichern = useSetSpeichern()

  // **Ohne offenes Set tut ⌘S nichts** — so verlangt es #162 ausdrücklich, und
  // es ist die richtige Antwort: ein Fehler „kein Set gewählt" wäre eine
  // Meldung über etwas, das niemand versucht hat.
  const sichern = useCallback(
    (unter: boolean) => {
      if (setID) {
        speichern.speichern(setID, unter)
      }
    },
    [setID, speichern],
  )

  const sichernJetzt = useCallback(() => {
    sichern(false)
  }, [sichern])

  const sichernUnter = useCallback(() => {
    sichern(true)
  }, [sichern])

  return { sichernJetzt, sichernUnter, speichern }
}
