// **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Was hier steht und was nicht
//
// **Was die Seite tut, steht hier; wie sie aussieht, in der Komponente** (§2).
//
// # Die vier Regeln aus §7
//
//  1. und 2. **Ein Wechsel des Profils räumt dieses Bauteil ab** — der
//     Aufrufer trägt den `key`. Jede Antwort bringt den **ganzen** Stand mit,
//     und eingetragen wird nur sie; ein Nachladen daneben könnte einen
//     älteren überschreiben.
//  3. **Geschrieben ist geschrieben**: nach einem gescheiterten Ruf bleibt die
//     gelesene Liste stehen, und es wird nicht nachgeladen. **Und ein
//     Nachladefehler ist kein gescheiterter Ruf** (#1960): der Schritt gilt,
//     die Liste ist unbekannt, und ihr Grund steht an ihr — in
//     `listenfehler` und nicht in `fehler`.
//  4. **Ein Vorgang je Datensatz**: `useVorgang` lässt den zweiten Ruf nicht
//     daneben laufen.
import { useCallback, useState } from 'react'

import {
  api,
  auspacken,
  Bindungsfehler,
  type DeckleiheAntwort,
  type Leihe,
  type Leihwunsch,
} from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Leihlauf {
  // ansehen holt die Leihen — auf Zuruf und nicht von selbst.
  ansehen: () => void

  // anbieten stellt eine Leihe aus und legt sie als Umschlag ab.
  anbieten: (wunsch: Leihwunsch, ziel: string) => void

  // annehmen liest einen Umschlag.
  annehmen: (pfad: string) => void

  // widerrufen nimmt eine ausgestellte zurück.
  widerrufen: (kennung: string) => void

  // zurueckgeben trägt ein, dass eine geliehene nicht mehr gilt.
  zurueckgeben: (kennung: string) => void

  laeuft: boolean

  // leihen ist `undefined`, solange nichts geholt wurde; eine leere Liste
  // heißt: geholt, und es gibt keine.
  leihen: Leihe[] | undefined

  // abgelegt ist der Pfad des zuletzt geschriebenen Umschlags.
  abgelegt: string | undefined

  fehler: string | undefined

  // listenfehler ist der Grund, aus dem die Liste nach einem **geglückten**
  // Schritt nicht zu lesen war — er gehört an die Liste und nicht an den
  // Schritt.
  listenfehler: string | undefined
}

// geaendert meldet jedes geglückte Schreiben an einer Leihe — P1b (#2370):
// die Deckwahl am Sitz liest die Leihen selbst und erführe sonst nichts.
export function useDeckleihe(geaendert?: () => void): Leihlauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [leihen, setzeLeihen] = useState<Leihe[]>()
  const [abgelegt, setzeAbgelegt] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()
  const [listenfehler, setzeListenfehler] = useState<string>()

  const fahre = useCallback(
    (arbeit: () => Promise<DeckleiheAntwort>) => {
      void vorgang.starte('deckleihe', async () => {
        setzeFehler(undefined)
        setzeListenfehler(undefined)

        try {
          const antwort = await arbeit()

          auspacken(antwort, () => undefined)

          // **Geschrieben ist geschrieben** (§7, Regel 3) — auch wenn die
          // Liste danach nicht zu lesen war: die Seite lädt selbst nach.
          geaendert?.()

          if (antwort.pfad !== '') {
            setzeAbgelegt(antwort.pfad)
          }

          // **Geschrieben ist geschrieben** (§7, Regel 3; #1960): die leere
          // Liste daneben sagt nichts. Sie einzutragen hieße „keine Leihe"
          // zu behaupten; die alte stehen zu lassen, einen Stand von vorher.
          if (antwort.nachladefehler) {
            setzeLeihen(undefined)
            setzeListenfehler(
              meldungAus(new Bindungsfehler(antwort.nachladefehler), texte),
            )

            return
          }

          setzeLeihen(antwort.leihen)
        } catch (ursache: unknown) {
          // **Die gelesene Liste bleibt stehen** (§7, Regel 3).
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [vorgang, texte, geaendert],
  )

  const ansehen = useCallback(() => {
    fahre(() => api().Deckleihen())
  }, [fahre])

  const anbieten = useCallback(
    (wunsch: Leihwunsch, ziel: string) => {
      fahre(() => api().DeckleiheAnbieten(wunsch, ziel))
    },
    [fahre],
  )

  const annehmen = useCallback(
    (pfad: string) => {
      fahre(() => api().DeckleiheAnnehmen(pfad))
    },
    [fahre],
  )

  const widerrufen = useCallback(
    (kennung: string) => {
      fahre(() => api().DeckleiheWiderrufen(kennung))
    },
    [fahre],
  )

  const zurueckgeben = useCallback(
    (kennung: string) => {
      fahre(() => api().DeckleiheZurueckgeben(kennung))
    },
    [fahre],
  )

  return {
    ansehen,
    anbieten,
    annehmen,
    widerrufen,
    zurueckgeben,
    laeuft: vorgang.laeuft('deckleihe'),
    leihen,
    abgelegt,
    fehler,
    listenfehler,
  }
}
