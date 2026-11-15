// **Eine Partie für Gäste öffnen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum das ein eigener Haken ist
//
// **Weil erreichbar sein und mitspielen lassen zwei Entscheidungen sind** —
// und weil die Freigabe eine andere Frage beantwortet als `usePartie`: dort
// steht, was eine Partie **tut**, hier, wer an ihr **mitspielen darf**.
//
// # Die vier Regeln aus §7
//
//  1. **Ein Wechsel räumt den offenen Kontext.** Der Aufrufer hängt diesen
//     Haken an einen `key` aus der Partiekennung; wer die Partie wechselt,
//     baut ihn neu auf — und damit ist auch Regel 2 eingelöst.
//  2. Siehe oben: **ein `key` löst beide zugleich**, und das ist die
//     sauberste Gestalt dafür (§7).
//  3. **Geschrieben ist geschrieben**: die Antwort **ist** der neue Stand —
//     es wird nicht nachgeladen, und deshalb kann dabei auch nichts
//     schiefgehen.
//  4. **Ein Vorgang je Datensatz**: der Datensatz ist die Freigabe.
//     `useVorgang` lässt den zweiten Ruf nicht daneben laufen — zwei Klicks
//     auf „freigeben" öffneten sonst zwei Tische.
import { useCallback, useEffect, useState } from 'react'

import { api, auspacken, type PartiefreigabeAntwort } from '../api'
import { meldungAus } from '../meldung'
import { horcheNachholend } from '../nachholen'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Freigabelauf {
  // freigeben öffnet die Partie für die genannten Gastsitze.
  freigeben: (gastsitze: string[]) => void

  // beenden schließt sie wieder.
  beenden: () => void

  // ansehen holt den Stand — für den ersten Aufbau und zum Nachsehen, wer
  // inzwischen sitzt.
  ansehen: () => void

  laeuft: boolean

  stand: PartiefreigabeAntwort | undefined

  fehler: string | undefined
}

// ereignisPartiefreigabe ist, womit Go meldet, dass sich die Besetzung eines
// eigenen Tisches geändert hat — **wörtlich dasselbe wie in
// `internal/app/partiebeitritt.go`**; `fensterereignisse_test.go` hält beide
// Seiten zusammen.
export const ereignisPartiefreigabe = 'partie:freigabe'

export function usePartiefreigabe(kennung: string, fuer: string): Freigabelauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [stand, setzeStand] = useState<PartiefreigabeAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const schluessel = `partie-freigabe-${kennung}`

  const fahre = useCallback(
    (arbeit: () => Promise<PartiefreigabeAntwort>) => {
      void vorgang.starte(schluessel, async () => {
        setzeFehler(undefined)

        try {
          const antwort = await arbeit()

          auspacken(antwort, () => undefined)

          setzeStand(antwort)
        } catch (ursache: unknown) {
          // **Der alte Stand bleibt stehen** (§7, Regel 3): ein
          // gescheitertes Freigeben nimmt eine bestehende Freigabe nicht
          // zurück.
          setzeFehler(meldungAus(ursache, texte))
        }
      })
    },
    [vorgang, texte, schluessel],
  )

  const freigeben = useCallback(
    (gastsitze: string[]) => {
      fahre(() => api().PartieFreigeben(kennung, fuer, gastsitze))
    },
    [fahre, kennung, fuer],
  )

  const beenden = useCallback(() => {
    fahre(() => api().PartieFreigabeBeenden())
  }, [fahre])

  const ansehen = useCallback(() => {
    fahre(() => api().PartieFreigabeStand())
  }, [fahre])

  // **Die Besetzung kommt von selbst** — gefunden mit zwei Fenstern
  // (`nat04`, S12.NAT-04): bis dahin sagte die Freigabe „Noch sitzt
  // niemand", während der Gast schon zog. **Nachholend**, denn verborgen
  // baut `<Activity>` den Horcher ab (`nachholen.ts`).
  useEffect(
    () =>
      horcheNachholend(ereignisPartiefreigabe, 'partie-freigabe', (daten) => {
        if (daten[0] === kennung) ansehen()
      }),
    [kennung, ansehen],
  )

  return {
    freigeben,
    beenden,
    ansehen,
    laeuft: vorgang.laeuft(schluessel),
    stand,
    fehler,
  }
}
