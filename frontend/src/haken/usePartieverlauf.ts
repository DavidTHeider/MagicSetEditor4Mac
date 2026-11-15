// **Die Rückschau im Fenster** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum ein eigener Haken und nicht ein Feld in `usePartie`
//
// **Weil er eine andere Frage beantwortet** (§2): dort steht, was eine Partie
// **tut** — anlegen, befehlen, pausieren —, hier, was sie **war**. Der Schnitt
// ist derselbe wie auf der Go-Seite, wo `partieverlauf.go` neben `partiewege.go`
// steht.
//
// # Die vier Regeln aus §7
//
//  1. **Ein Wechsel räumt den offenen Kontext** und
//  2. **eine Antwort gehört zu dem Stand, aus dem sie stammt** — beide löst
//     **ein** `key` beim Aufrufer ein, aus Partie und Sitz. Das ist die
//     sauberste Gestalt für beide zugleich, und sie steht so in §7: React baut
//     das Bauteil neu auf, und damit ist ein noch laufender Ruf für den alten
//     Sitz ins Leere gelaufen statt in den neuen Stand. **Wer den `key`
//     entfernt, nimmt beide Regeln weg.**
//
//     **Und deshalb steht hier kein `useLauf` daneben.** Er wäre ein dritter
//     Riegel, den kein Lauf betreten kann: der Datensatz ist je Einbau fest,
//     und zwei Rufe nebeneinander verhindert Regel 4. Eine Absicherung, die
//     keine Lage erreicht, sieht aus wie eine Zusage und ist keine (§10).
//  3. **Geschrieben ist geschrieben** — trifft hier nicht zu, denn dieser Weg
//     schreibt nicht. Er steht trotzdem da, weil eine Aufzählung mit einer
//     Lücke wie eine Zusage aussieht (§10).
//  4. **Ein Vorgang je Datensatz**: der Datensatz ist die Partie. `useVorgang`
//     lässt den zweiten Ruf nicht daneben laufen.
//
// # Warum er nicht von selbst lädt
//
// **Weil ein Verlauf mit jedem Zug wächst.** Nach jedem Befehl neu zu laden
// hieße, ihn in einer Partie mit hundert Zügen hundertmal zu holen, und der
// letzte Lauf trüge alles. Er wird geholt, wenn jemand ihn sehen will — und
// dann sagt der Stand, wie alt er ist.
import { useCallback, useState } from 'react'

import { api, auspacken, type Spielereignis } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Verlaufslauf {
  // ansehen holt den Verlauf — auf Zuruf und nicht von selbst.
  ansehen: () => void

  laeuft: boolean

  // verlauf ist der Strom. **`undefined` heißt: noch nicht geholt**, eine
  // leere Liste heißt: geholt, und es ist nichts geschehen. Zwei Zustände,
  // die eine Oberfläche verschieden erklären muss (§6).
  verlauf: Spielereignis[] | undefined

  fehler: string | undefined
}

export function usePartieverlauf(kennung: string, fuer: string): Verlaufslauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [verlauf, setzeVerlauf] = useState<Spielereignis[]>()
  const [fehler, setzeFehler] = useState<string>()

  const ansehen = useCallback(() => {
    void vorgang.starte(`partie-verlauf-${kennung}`, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().PartieVerlauf(kennung, fuer)

        auspacken(antwort, () => undefined)

        setzeVerlauf(antwort.verlauf)
      } catch (ursache: unknown) {
        // **Der alte Stand bleibt stehen.** Ein gescheitertes Nachladen ist
        // kein Grund, das Gelesene wegzuwerfen — es wäre dieselbe Sorte
        // Schaden wie ein Lesefehler, der eine geglückte Änderung zurücknimmt
        // (§7, Regel 3).
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [vorgang, texte, kennung, fuer])

  return {
    ansehen,
    laeuft: vorgang.laeuft(`partie-verlauf-${kennung}`),
    verlauf,
    fehler,
  }
}
