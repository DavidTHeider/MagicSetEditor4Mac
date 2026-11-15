// **Abgestufte Rücksetzung** — Werkstück O4 von Phase 5.2
// (#1428).
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist der Bestand,
//     und Nachsehen wie Zurücksetzen teilen sich denselben Vorgang.
//  2. **Geschrieben ist geschrieben** (Regel 3): die Bilanz bleibt stehen,
//     auch wenn danach etwas schiefgeht — beim Neuanfang ist sie die einzige
//     Auskunft darüber, wo der alte Bestand liegt.
//  3. **Ein Wechsel der Art räumt die Vorschau ab** (Regel 1): sie beschreibt
//     eine andere Rücksetzung, und eine stehengebliebene wäre die gefährlichste
//     Antwort. Die Seite hängt dafür ein `key` an die Art.
import { useCallback, useState } from 'react'
import { epocheGesehen } from '../bestandsepoche'

import type {
  RuecksetzungsbilanzAntwort,
  RuecksetzungsvorschauAntwort,
} from '../apiruecksetzung'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Ruecksetzungslauf {
  // nachsehen holt die Vorschau — sie ändert nichts.
  nachsehen: (art: string) => void

  // zuruecksetzen führt sie aus.
  //
  // **`art` kommt vom Aufrufer und nicht aus dem Zustand hier** — sie steht
  // dort ohnehin, und ein Griff in den eigenen Zustand brauchte eine Wache
  // gegen „noch nichts nachgesehen", die kein Aufruf betreten kann (§3).
  zuruecksetzen: (
    art: string,
    name: string,
    nachweis: string,
    revision: string,
  ) => void

  laeuft: boolean

  vorschau: RuecksetzungsvorschauAntwort | undefined
  bilanz: RuecksetzungsbilanzAntwort | undefined
  fehler: string | undefined

  raeumen: () => void
}

// **`gewechselt` läuft nach einer Rücksetzung, die das Bezugsobjekt nimmt**
// (#2292) — Neuanfang, Profil, Mensch. Sie meldet es nach oben wie ein
// Zurückspielen; die Bilanz bleibt, denn sie steht hier und nicht dort.
//
// **Nach jeder anderen läuft `veraendert`**: sie lässt Set und Karte stehen,
// ändert aber, was das Fenster gelesen hat — dasselbe Neulesen wie nach dem
// Zurücksetzen der Darstellung aus der Menüleiste. Eine Art, die hier niemand
// kennt, liest damit neu, statt still im alten Stand zu bleiben.
const nimmtDasBezugsobjekt = new Set(['anwendung', 'profil', 'nutzer'])

export function useRuecksetzung(
  vorgang: Vorgang,
  gewechselt: () => void,
  veraendert: () => void,
): Ruecksetzungslauf {
  const texte = useTexte()

  const [vorschau, setzeVorschau] = useState<RuecksetzungsvorschauAntwort>()
  const [bilanz, setzeBilanz] = useState<RuecksetzungsbilanzAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const raeumen = useCallback(() => {
    setzeVorschau(undefined)
    setzeBilanz(undefined)
    setzeFehler(undefined)
  }, [])

  const nachsehen = useCallback(
    (art: string) => {
      void vorgang.starte('ruecksetzung', async () => {
        raeumen()

        try {
          const antwort = await api().Ruecksetzungsvorschau(art, '', '')

          setzeVorschau(auspacken(antwort, () => antwort))
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [raeumen, texte, vorgang],
  )

  const zuruecksetzen = useCallback(
    (art: string, name: string, nachweis: string, revision: string) => {
      void vorgang.starte('ruecksetzung', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await api().Zuruecksetzen(
            art,
            '',
            '',
            name,
            nachweis,
            revision,
          )

          // **Die Epoche zuerst, auch vor einem Fehler** (J05, #2269): ein
          // gescheiterter Neuanfang kann den Bestand schon getauscht haben.
          epocheGesehen(antwort.bestandsepoche)

          // **Die Bilanz wird gesetzt, bevor der Fehler geworfen wird**
          // (§7, Regel 3): beim Neuanfang steht dort, wo der alte Bestand
          // liegt — ohne diese Zeile fände ihn niemand wieder.
          setzeBilanz(antwort)
          auspacken(antwort, () => undefined)

          // **Die Vorschau verschwindet erst nach einer geglückten
          // Rücksetzung.** Sie beschreibt einen Stand, den es nicht mehr
          // gibt; wer noch einmal zurücksetzen will, sieht zuerst nach.
          setzeVorschau(undefined)

          if (nimmtDasBezugsobjekt.has(art)) {
            gewechselt()
          } else {
            veraendert()
          }
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [gewechselt, texte, veraendert, vorgang],
  )

  return {
    nachsehen,
    zuruecksetzen,
    laeuft: vorgang.laeuft('ruecksetzung'),
    vorschau,
    bilanz,
    fehler,
    raeumen,
  }
}
