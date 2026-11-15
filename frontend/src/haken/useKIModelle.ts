// **Der Modellkatalog, einmal je Kasten** — herausgezogen für Werkstück A12
// von Phase 5.3
// (#1451).
//
// # Warum ein Haken und nicht ein Bauteil
//
// **Weil ein Bauteil abgeräumt wird und ein Haken nicht.** Die Wahl steht in
// `KIVorschlag` innerhalb des aufgeklappten Teils; klappt jemand zu, fällt
// sie aus dem DOM — und ein Katalog, der **in ihr** hinge, würde beim
// nächsten Aufklappen erneut fragen.
//
// **Genau das ist beim ersten Anlauf passiert** (2026-09-12): der Lauf
// „räumt seinen Inhalt ab, wenn er zugeht" sah zwei Rufe statt einem. Das
// war kein Fehler des Laufs, sondern ein Ruf nach draußen, den niemand
// ausgelöst hat (§16, Regel 6).
import { useMemo } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { KIModellDTO } from '../antworten'

// useKIModelle liest den Katalog, sobald `gefragt` gilt.
export function useKIModelle(gefragt: boolean): {
  modelle: KIModellDTO[]
  laedt: boolean
  fehler: string | undefined
} {
  const katalog = useLaden(async () => {
    if (!gefragt) {
      return undefined
    }

    const antwort = await api().KIModelle()

    return auspacken(antwort, () => antwort)
  }, [gefragt])

  const modelle = useMemo(() => katalog.daten?.modelle ?? [], [katalog.daten])

  return { modelle, laedt: katalog.laedt, fehler: katalog.fehler }
}
