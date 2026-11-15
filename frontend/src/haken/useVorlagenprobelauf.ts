// **Was die Exportvorlage ergäbe** — UXR-F7
// (#2015).
//
// Der Probelauf wertet die Vorlage aus — mit **denselben Wahlen**, mit denen der
// Knopf danach ausführt. Eine Auswertung läuft über alle gewählten Karten,
// deshalb kommt die Frage erst, wenn die Wahl zur Ruhe gekommen ist: wer in
// einem Textfeld tippt, ändert die Option mit jedem Anschlag, und jede Frage
// wäre ein Lauf über das ganze Set (`useVerzoegert`).
//
// **Verworfen wird die überholte Antwort** (§7, Regel 2): `useLaden` zählt mit.
// Und `veraltet` sagt, dass die Anzeige noch zu einer **früheren** Wahl gehört
// — in der Ruhefrist und während der Auswertung —, damit sie sich nicht als
// Antwort auf die heutige ausgibt.
import { useCallback } from 'react'

import { api } from '../api'
import type { VorlagenprobelaufAntwort } from '../antwortenprobelauf'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'
import { useVerzoegert } from './useVerzoegert'

// probelauffrist ist die Ruhe, die eine Wahl braucht, bevor gefragt wird.
export const probelauffrist = 250

export interface Vorlagenprobelauf extends Ladezustand<VorlagenprobelaufAntwort> {
  veraltet: boolean
}

export function useVorlagenprobelauf(
  setID: string,
  paket: string,
  wahl: Readonly<Record<string, string>>,
  auswahl: readonly string[],
): Vorlagenprobelauf {
  // **Als Text, damit der Vergleich nach Wert geht**: ein Objekt, das bei jedem
  // Zeichnen neu entsteht, setzte den Zeitgeber ständig zurück.
  const jetzt = JSON.stringify([setID, paket, wahl, auswahl])
  const frage = useVerzoegert(jetzt, probelauffrist)

  const holen = useCallback(async () => {
    const [kennung, name, gewaehlt, karten] = JSON.parse(frage) as [
      string,
      string,
      Record<string, string>,
      string[],
    ]

    const antwort = await api().VorlageProbelauf(
      kennung,
      name,
      gewaehlt,
      karten,
    )

    return auspacken(antwort, () => antwort)
  }, [frage])

  const stand = useLaden(holen, [frage], true)

  return { ...stand, veraltet: stand.laedt || frage !== jetzt }
}
