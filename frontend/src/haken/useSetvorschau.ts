// **Vor dem Einlesen eines Sets nachsehen** — FMT-03.AK-01
// (#1692).
//
// *Set öffnen …* und ⌘O fragen im Dialog nach der Datei und lassen Go sagen,
// was ihr Einfahren täte. **Gibt es nichts zu sagen, wird gleich
// eingelesen** — ein Klick mehr für jede saubere Datei wäre eine Hürde ohne
// Auskunft. Sonst steht die Vorschau da, bis jemand einliest oder verwirft:
// Befunde und fehlende Pakete **vor** dem Einlesen, nicht danach.
import { useCallback, useState } from 'react'

import { api, auspacken } from '../api'
import type { Setvorschau } from '../antwortenbestand'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Setvorschaulauf {
  vorschau: Setvorschau | undefined
  fehler: string | undefined
  laeuft: boolean
  waehlen: () => Promise<void>
  einlesen: (pfad: string) => void
  verwerfen: () => void
}

// einlesen bekommt den Pfad; wie eingelesen wird, weiß die Seite.
export function useSetvorschau(
  einlesen: (pfad: string) => void,
): Setvorschaulauf {
  const texte = useTexte()
  const vorgang = useVorgang()
  const [vorschau, setzeVorschau] = useState<Setvorschau>()
  const [fehler, setzeFehler] = useState<string>()

  const waehlen = useCallback(async () => {
    await vorgang.starte('vorschau', async () => {
      setzeFehler(undefined)
      setzeVorschau(undefined)

      try {
        const antwort = await api().SetVorschauWaehlen()
        const gelesen = auspacken(antwort, () => antwort.vorschau)

        // **Ein abgebrochener Dialog ist keine Meldung wert.**
        if (gelesen === null) {
          return
        }

        if (gelesen.befunde.length === 0 && gelesen.bedarf.length === 0) {
          einlesen(gelesen.pfad)

          return
        }

        setzeVorschau(gelesen)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [einlesen, texte, vorgang])

  return {
    vorschau,
    fehler,
    laeuft: vorgang.laeuft('vorschau'),
    waehlen,
    einlesen: (pfad: string) => {
      // **Erst weg, dann einlesen** — die Vorschau gehört zu einer Frage, die
      // mit dem Einlesen beantwortet ist.
      setzeVorschau(undefined)
      einlesen(pfad)
    },
    verwerfen: () => {
      setzeVorschau(undefined)
    },
  }
}
