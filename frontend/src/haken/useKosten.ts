// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Wechsel des Kontos räumt den Stand ab** (Regel 1): die Zahlen
//     gehören einem Konto, und die eines anderen unter dem neuen Namen wären
//     die gefährlichste Anzeige.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (Regel 2): das
//     Laden geht über `useLaden`, und der zählt die Wechsel mit.
//  3. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist das Konto.
//  4. **Geschrieben ist geschrieben** (Regel 3): nach einer gesetzten Grenze
//     wird nachgeladen — scheitert das, bleibt die Grenze gesetzt, und der
//     Fehler steht daneben.
import { useCallback, useState } from 'react'

import type { HaushaltsstandAntwort, KostenpostenAntwort } from '../apikosten'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Kostenlauf {
  stand: HaushaltsstandAntwort | undefined
  posten: KostenpostenAntwort | undefined

  laedt: boolean
  laeuft: boolean

  fehler: string | undefined

  // setzeGrenze schreibt die Grenze und lädt danach nach.
  setzeGrenze: (wert: number, waehrung: string) => void
}

export function useKosten(
  vorgang: Vorgang,
  konto: string,
  zeitraum: string,
): Kostenlauf {
  const texte = useTexte()
  const [fehler, setzeFehler] = useState<string>()

  const stand = useLaden(
    () => api().Kostenstand(konto, zeitraum),
    [konto, zeitraum],
  )

  const posten = useLaden(
    () => api().Kostenposten(konto, zeitraum),
    [konto, zeitraum],
  )

  const setzeGrenze = useCallback(
    (wert: number, waehrung: string) => {
      void vorgang.starte(`kosten:${konto}`, async () => {
        setzeFehler(undefined)

        try {
          const antwort = await api().KostenbudgetSetzen(
            konto,
            zeitraum,
            wert,
            waehrung,
          )

          auspacken(antwort, () => undefined)

          // **Nachladen ist Bestenfalls-Arbeit** (§7, Regel 3): die Grenze
          // steht, auch wenn das Lesen danach scheitert.
          stand.nachladen()
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [konto, stand, texte, vorgang, zeitraum],
  )

  return {
    stand: stand.daten,
    posten: posten.daten,
    laedt: stand.laedt || posten.laedt,
    laeuft: vorgang.laeuft(`kosten:${konto}`),
    fehler: fehler ?? stand.fehler ?? posten.fehler,
    setzeGrenze,
  }
}
