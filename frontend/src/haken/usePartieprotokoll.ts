// **Was von einer Partie übrig bleibt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Zwei Fragen, ein Haken
//
// **Wer darf das Protokoll lesen, und wann ist es weg?** Beide handeln von
// demselben Gegenstand, beide ändern ihn, und beide zeigen danach denselben
// Stand. Zwei Haken nebeneinander hätten zwei Listen derselben
// Einwilligungen — und die erste, die nicht nachlädt, zeigt eine alte.
//
// # Die vier Regeln aus §7
//
//  1. und 2. **Ein `key` aus der Partiekennung** beim Aufrufer räumt beides:
//     wer die Partie wechselt, baut diesen Haken neu auf. Zusätzlich steht
//     die Kennung im Namen des Vorgangs.
//  3. **Geschrieben ist geschrieben**: die Antwort trägt den neuen Stand
//     schon mit — es wird **nicht** nachgeladen, und ein Fehler beim
//     Schreiben lässt den alten Stand stehen. **Ein Nachladefehler ist kein
//     Fehler beim Schreiben** (#1970): der Schritt gilt, und der Grund steht
//     an der Liste statt einer leeren.
//  4. **Ein Vorgang je Datensatz**: `useVorgang` lässt den zweiten Ruf nicht
//     daneben laufen.
import { useCallback, useState } from 'react'

import {
  api,
  auspacken,
  Bindungsfehler,
  type Offenlegung,
  type PartieoffenlegungAntwort,
} from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Protokolllauf {
  // ansehen holt die Einwilligungen — auf Zuruf und nicht von selbst.
  ansehen: () => void

  // freigeben ersetzt die Einwilligung eines Sitzes.
  freigeben: (von: string, an: string[]) => void

  // zuruecknehmen schließt die Tür wieder.
  zuruecknehmen: (von: string) => void

  // vergessen löscht das Protokoll — nur bei einer beendeten Partie.
  vergessen: () => void

  laeuft: boolean

  // offenlegungen ist `undefined`, solange nichts geholt wurde; eine leere
  // Liste heißt: geholt, und niemand hat etwas freigegeben.
  offenlegungen: Offenlegung[] | undefined

  // istVergessen heißt: das Protokoll ist gelöscht.
  istVergessen: boolean

  fehler: string | undefined

  // listenfehler ist der Grund, aus dem die Liste nach einem **geglückten**
  // Schritt fehlt (§7, Regel 3; #1970) — er gehört an die Liste und nicht an
  // den Schritt.
  listenfehler: string | undefined
}

export function usePartieprotokoll(kennung: string): Protokolllauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [offenlegungen, setzeOffenlegungen] = useState<Offenlegung[]>()
  const [istVergessen, setzeVergessen] = useState(false)
  const [fehler, setzeFehler] = useState<string>()
  const [listenfehler, setzeListenfehler] = useState<string>()

  const fahre = useCallback(
    (arbeit: () => Promise<PartieoffenlegungAntwort>, danach = false) => {
      void vorgang.starte(`partie-protokoll-${kennung}`, async () => {
        setzeFehler(undefined)
        setzeListenfehler(undefined)

        try {
          const antwort = await arbeit()

          auspacken(antwort, () => undefined)

          // **Geschrieben ist geschrieben** (§7, Regel 3; #1970): die leere
          // Liste daneben sagt nichts — sie behauptete „niemand".
          if (antwort.nachladefehler) {
            setzeOffenlegungen(undefined)
            setzeListenfehler(
              meldungAus(new Bindungsfehler(antwort.nachladefehler), texte),
            )

            return
          }

          setzeOffenlegungen(antwort.offenlegungen)

          if (danach) {
            setzeVergessen(true)
          }
        } catch (ursache: unknown) {
          // **Der alte Stand bleibt stehen** (§7, Regel 3): ein gescheiterter
          // Ruf ist kein Grund, das Gelesene wegzuwerfen.
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [vorgang, texte, kennung],
  )

  const ansehen = useCallback(() => {
    fahre(() => api().PartieOffenlegungen(kennung))
  }, [fahre, kennung])

  const freigeben = useCallback(
    (von: string, an: string[]) => {
      fahre(() => api().PartieOffenlegen(kennung, von, an))
    },
    [fahre, kennung],
  )

  const zuruecknehmen = useCallback(
    (von: string) => {
      fahre(() => api().PartieOffenlegungZuruecknehmen(kennung, von))
    },
    [fahre, kennung],
  )

  const vergessen = useCallback(() => {
    fahre(() => api().PartieVergessen(kennung), true)
  }, [fahre, kennung])

  return {
    ansehen,
    freigeben,
    zuruecknehmen,
    vergessen,
    laeuft: vorgang.laeuft(`partie-protokoll-${kennung}`),
    offenlegungen,
    istVergessen,
    fehler,
    listenfehler,
  }
}
