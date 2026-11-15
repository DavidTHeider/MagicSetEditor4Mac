// **Eine Paarung, an eine Partie gebunden** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Was hier steht und was nicht
//
// **Was die Seite tut, steht hier; wie sie aussieht, in der Komponente** (§2).
//
// # Die vier Regeln aus §7
//
//  1. und 2. **Ein `key` aus der Paarungskennung** beim Aufrufer räumt beides.
//     Jede Antwort bringt den **ganzen** Stand mit, und eingetragen wird nur
//     sie — ein Nachladen daneben könnte einen älteren überschreiben.
//  3. **Geschrieben ist geschrieben**: nach einem gescheiterten Ruf bleibt der
//     gelesene Stand stehen, und es wird nicht nachgeladen.
//  4. **Ein Vorgang je Datensatz**: der Name des Vorgangs trägt die Paarung.
import { useCallback, useState } from 'react'

import { api, auspacken, Bindungsfehler } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

import type { ErgebnisDTO, MatchAntwort, Matchsitz } from '../apiturniere'

export interface Matchlauf {
  // ansehen holt den Stand — auf Zuruf und nicht von selbst.
  ansehen: () => void

  // binden sagt, dass diese Partie zu dieser Paarung gehört.
  binden: (partieID: string, zuordnung: Matchsitz[]) => void

  // melden trägt das Ergebnis **einer** Seite ein.
  melden: (von: string, kette: string, stand: ErgebnisDTO) => void

  // ausDerPartie meldet für eine Seite, was die **Partie** festgestellt hat.
  //
  // **Spielstand und Nachweis kommen von dort** — M17 (#1618). Getippt wird
  // nur, wer meldet, und das ist die Zustimmung, die G7 verlangt.
  ausDerPartie: (von: string) => void

  // uebernehmen trägt es ins Turnier ein — auf Zuruf.
  uebernehmen: () => void

  laeuft: boolean

  // stand ist `undefined`, solange nichts geholt wurde.
  stand: MatchAntwort | undefined

  fehler: string | undefined

  // standfehler ist der Grund, aus dem der Stand nach einem **geglückten**
  // Schritt fehlt (§7, Regel 3; #1970) — er gehört an den Stand und nicht
  // an den Schritt.
  standfehler: string | undefined
}

export function useTurniermatch(
  turnierID: string,
  paarungID: string,
): Matchlauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [stand, setzeStand] = useState<MatchAntwort>()
  const [fehler, setzeFehler] = useState<string>()
  const [standfehler, setzeStandfehler] = useState<string>()

  const fahre = useCallback(
    (arbeit: () => Promise<MatchAntwort>) => {
      void vorgang.starte(`turniermatch-${paarungID}`, async () => {
        setzeFehler(undefined)
        setzeStandfehler(undefined)

        try {
          const antwort = await arbeit()

          auspacken(antwort, () => undefined)

          // **Geschrieben ist geschrieben** (§7, Regel 3; #1970): die leere
          // Antwort daneben sagt nichts — sie behauptete „keine Partie
          // daran", wo eben gebunden wurde.
          if (antwort.nachladefehler) {
            setzeStand(undefined)
            setzeStandfehler(
              meldungAus(new Bindungsfehler(antwort.nachladefehler), texte),
            )

            return
          }

          setzeStand(antwort)
        } catch (ursache: unknown) {
          // **Der gelesene Stand bleibt stehen** (§7, Regel 3).
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [vorgang, texte, paarungID],
  )

  const ansehen = useCallback(() => {
    fahre(() => api().TurnierMatch(paarungID))
  }, [fahre, paarungID])

  const binden = useCallback(
    (partieID: string, zuordnung: Matchsitz[]) => {
      fahre(() =>
        api().TurnierMatchBinden(turnierID, paarungID, partieID, zuordnung),
      )
    },
    [fahre, turnierID, paarungID],
  )

  const melden = useCallback(
    (von: string, kette: string, ergebnis: ErgebnisDTO) => {
      fahre(() => api().TurnierMatchMelden(paarungID, von, kette, ergebnis))
    },
    [fahre, paarungID],
  )

  // **Und dasselbe, ohne dass jemand tippt** — M17 (#1618).
  const ausDerPartie = useCallback(
    (von: string) => {
      fahre(() => api().TurnierMatchAusDerPartie(paarungID, von))
    },
    [fahre, paarungID],
  )

  const uebernehmen = useCallback(() => {
    fahre(() => api().TurnierMatchUebernehmen(turnierID, paarungID))
  }, [fahre, turnierID, paarungID])

  return {
    ansehen,
    binden,
    melden,
    ausDerPartie,
    uebernehmen,
    laeuft: vorgang.laeuft(`turniermatch-${paarungID}`),
    stand,
    fehler,
    standfehler,
  }
}
