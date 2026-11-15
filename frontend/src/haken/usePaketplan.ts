// **Erst der Plan, dann die Änderung** — DOM-01.AK-02 und AK-03
// (#1696).
//
// Was eine Paketänderung bewirkt — betroffene Sets, gebrochene
// Abhängigkeiten, eine lokale Änderung —, steht **vor** ihr da. Der Haken
// fragt den Plan, hält ihn fest und führt **genau den geplanten Vorgang** aus:
// eingebunden und ersetzt wird mit der Quelle aus dem Plan, entfernt und
// zurückgesetzt mit Art und Namen, zu denen er gefragt wurde. Eine zweite
// Wahl dazwischen gibt es nicht.
//
// # Die vier Regeln aus §7
//
// 1. **Ein Wechsel des Bezugsobjekts** räumt den Plan — das tut der Aufrufer
//    mit [Paketplanung.verwerfen], denn nur er weiß, wann gewechselt wird.
// 2. **Eine Antwort gehört zu ihrem Stand**: jedes Planen und jedes
//    Verwerfen zählt den Lauf weiter, und eine überholte Antwort wird nicht
//    eingetragen.
// 3. **Nachgeladen wird auch nach einem Fehlschlag** — und das ist die
//    begründete Abweichung, dieselbe wie in `Paketliste.tsx`: Entfernen
//    löscht vor dem Ausbuchen, Ersetzen verschiebt vor dem Abgleich. Ein
//    Fehler danach lässt eine geänderte Platte zurück.
// 4. **Ein Vorgang zur Zeit** — geplant und ausgeführt wird je unter einer
//    Kennung.
import { useCallback, useState } from 'react'

import {
  api,
  auspacken,
  type PaketDTO,
  type PaketplanAntwort,
  type PaketplanDTO,
} from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useLauf } from './useLauf'
import { useVorgang } from './useVorgang'

export interface Paketplanung {
  plan: PaketplanDTO | undefined

  // fehler ist die Meldung zum Plan — beim Planen oder beim Ausführen.
  fehler: string | undefined

  laeuft: boolean

  planen: (vorgang: 'entfernen' | 'zurueck', paket: PaketDTO) => Promise<void>

  // waehlen öffnet den Verzeichnisdialog und plant das Einbinden — oder das
  // Ersetzen, wenn es das Paket schon gibt.
  waehlen: () => Promise<void>

  // ausfuehren führt den geplanten Vorgang aus; `ueberschreiben` bestätigt
  // das Ersetzen eines lokal geänderten Pakets.
  ausfuehren: (ueberschreiben: boolean) => Promise<void>

  verwerfen: () => void
}

// Geplant ist ein Plan samt dem Paket, zu dem er gefragt wurde.
interface Geplant {
  plan: PaketplanDTO
  art: string
  name: string
}

export function usePaketplan(geaendert: () => void): Paketplanung {
  const texte = useTexte()
  const lauf = useLauf()
  const vorgang = useVorgang()

  const [geplant, setzeGeplant] = useState<Geplant>()
  const [fehler, setzeFehler] = useState<string>()

  const fragen = useCallback(
    async (
      frage: () => Promise<PaketplanAntwort>,
      art: string,
      name: string,
    ) => {
      const stand = lauf.naechster()
      setzeGeplant(undefined)
      setzeFehler(undefined)

      await vorgang.starte('plan', async () => {
        try {
          const antwort = await frage()
          const plan = auspacken(antwort, () => antwort.plan)

          if (lauf.istAktuell(stand) && plan !== null) {
            setzeGeplant({ plan, art, name })
          }
        } catch (ursache: unknown) {
          if (lauf.istAktuell(stand)) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    [lauf, vorgang, texte],
  )

  const planen = useCallback(
    (art: 'entfernen' | 'zurueck', paket: PaketDTO) =>
      fragen(
        () => api().PaketPlan(art, paket.art, paket.name, ''),
        paket.art,
        paket.name,
      ),
    [fragen],
  )

  const waehlen = useCallback(
    () => fragen(() => api().PaketPlanWaehlen(), '', ''),
    [fragen],
  )

  const ausfuehren = useCallback(
    async (ueberschreiben: boolean) => {
      if (geplant === undefined) {
        return
      }

      const { plan, art, name } = geplant

      await vorgang.starte('ausfuehren', async () => {
        setzeFehler(undefined)

        try {
          const antwort = await {
            entfernen: () => api().PaketEntfernen(art, name),
            einbinden: () => api().PaketEinbindenVon(plan.quelle),
            ersetzen: () => api().PaketErsetzen(plan.quelle, ueberschreiben),
            zurueck: () => api().PaketZurueck(art, name),
          }[plan.vorgang]()

          auspacken(antwort, () => undefined)
          lauf.naechster()
          setzeGeplant(undefined)
        } catch (ursache: unknown) {
          setzeFehler(meldungAus(ursache, texte))
        } finally {
          geaendert()
        }
      })
    },
    [geplant, vorgang, lauf, texte, geaendert],
  )

  const verwerfen = useCallback(() => {
    lauf.naechster()
    setzeGeplant(undefined)
    setzeFehler(undefined)
  }, [lauf])

  return {
    plan: geplant?.plan,
    fehler,
    laeuft: vorgang.laeuft('plan') || vorgang.laeuft('ausfuehren'),
    planen,
    waehlen,
    ausfuehren,
    verwerfen,
  }
}
