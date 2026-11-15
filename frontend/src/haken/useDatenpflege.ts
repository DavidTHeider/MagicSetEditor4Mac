// **Nachsehen, wählen, aufräumen** — Werkstück O8 von Phase 5.2
// (#1432).
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist der Bestand,
//     und es gibt genau einen — Nachsehen und Aufräumen teilen ihn sich.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einer Pflege bleibt die
//     Bilanz stehen, auch wenn das Nachsehen danach scheiterte. Die Vorschau
//     wird **nicht** stillschweigend neu geholt — wer noch einmal aufräumen
//     will, sieht zuerst noch einmal nach.
//  3. **Regel 1 und 2 greifen nicht**: es gibt kein Bezugsobjekt, das wechseln
//     könnte.
import { useCallback, useState } from 'react'

import type {
  PflegebilanzAntwort,
  PflegevorschauAntwort,
  PflegewahlDTO,
} from '../apipflege'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Pflegelauf {
  // nachsehen holt die Vorschau — sie ändert nichts.
  nachsehen: () => void

  // pflegen führt die gewählten Maßnahmen aus.
  //
  // **`erstellt` kommt vom Aufrufer und nicht aus dem Zustand hier.** Es steht
  // dort ohnehin — die Vorschau liegt vor ihm —, und ein Griff in den eigenen
  // Zustand brauchte eine Wache gegen „noch nichts nachgesehen", die kein
  // Aufruf betreten kann (§3). Dieselbe Entscheidung wie bei
  // `useSicherungZurueckspielen.zurueckspielen`.
  pflegen: (erstellt: string, wahl: PflegewahlDTO[]) => void

  laeuft: boolean

  vorschau: PflegevorschauAntwort | undefined
  bilanz: PflegebilanzAntwort | undefined
  fehler: string | undefined

  raeumen: () => void
}

export function useDatenpflege(vorgang: Vorgang): Pflegelauf {
  const texte = useTexte()

  const [vorschau, setzeVorschau] = useState<PflegevorschauAntwort>()
  const [bilanz, setzeBilanz] = useState<PflegebilanzAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const raeumen = useCallback(() => {
    setzeVorschau(undefined)
    setzeBilanz(undefined)
    setzeFehler(undefined)
  }, [])

  const nachsehen = useCallback(() => {
    void vorgang.starte('pflege', async () => {
      raeumen()

      try {
        const antwort = await api().Pflegevorschau()

        setzeVorschau(auspacken(antwort, () => antwort))
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [raeumen, texte, vorgang])

  const pflegen = useCallback(
    (erstellt: string, wahl: PflegewahlDTO[]) => {
      void vorgang.starte('pflege', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await api().Pflegen(erstellt, wahl)

          // **Die Bilanz wird gesetzt, bevor der Fehler geworfen wird.**
          // Gepflegt ist gepflegt (§7, Regel 3): ein Fehler beim Nachsehen
          // danach darf die Zeilen nicht verschlucken.
          setzeBilanz(antwort)
          auspacken(antwort, () => undefined)

          // **Die Vorschau verschwindet erst nach einer geglückten Pflege.**
          // Sie beschreibt einen Stand, den es nicht mehr gibt; wer noch
          // einmal aufräumen will, sieht zuerst noch einmal nach.
          setzeVorschau(undefined)
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  return {
    nachsehen,
    pflegen,
    laeuft: vorgang.laeuft('pflege'),
    vorschau,
    bilanz,
    fehler,
    raeumen,
  }
}
