// **Wie die Leitung steht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum auf Zuruf und nicht im Takt
//
// **Weil eine Diagnose eine Frage ist und kein Gerät.** Wer sie im Sekundentakt
// zöge, hielte eine Naht warm, die sonst schweigt — und die Zahl, die er
// dabei misst, wäre zum Teil seine eigene. Sie wird geholt, wenn jemand
// hinsieht.
//
// # Und warum die Bildzeit hier entsteht
//
// **Weil sie im Fenster gemessen wird und nirgends sonst.** Das Issue nennt
// „Framezeit" in einem Atemzug mit Route und RTT; die Go-Seite hat sie nicht
// und dürfte sie deshalb auch nicht nennen (§10). Hier steht sie — aus
// `requestAnimationFrame`, also aus der Quelle, die sie kennt.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api, auspacken, type PartiediagnoseAntwort } from '../api'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

// Bildproben ist die Zahl der Bilder, über die gemittelt wird.
//
// **Gemittelt und nicht das letzte**: ein einzelnes Bild trifft jede
// Unregelmäßigkeit des Rechners und sagt über den Tisch nichts.
export const Bildproben = 30

export interface Diagnoselauf {
  // ansehen holt den Stand.
  ansehen: () => void

  laeuft: boolean

  // stand ist `undefined`, solange nichts geholt wurde.
  stand: PartiediagnoseAntwort | undefined

  // bildmillis ist die gemessene Bildzeit — `undefined`, solange zu wenige
  // Bilder vorliegen.
  bildmillis: number | undefined

  fehler: string | undefined
}

export function usePartiediagnose(): Diagnoselauf {
  const vorgang = useVorgang()
  const texte = useTexte()

  const [stand, setzeStand] = useState<PartiediagnoseAntwort>()
  const [bildmillis, setzeBildmillis] = useState<number>()
  const [fehler, setzeFehler] = useState<string>()

  // **Gemessen wird, wenn jemand nachsieht — dreißig Bilder, dann Schluss.**
  //
  // Eine Messung, solange dieser Haken lebt, wäre eine Schleife mit 60
  // Anforderungen je Sekunde, solange der Tisch aufgeschlagen ist — auch nach
  // einem verlorenen Grafikkontext, wenn es nichts mehr zu messen gibt
  // (gemessen am 2026-09-25 in der WebKit-Probe). PLAY-06.AK-04
  // (#1714)
  // schließt genau das aus. **Der Preis**: die ersten Bilder nach dem Klick
  // gehen mit in den Mittelwert; eine Leitung zu prüfen zeichnet den Tisch
  // aber nicht neu, und was sie kostet, ist die Auskunft, die man sucht.
  const bilder = useRef<number[]>([])
  const messung = useRef<(() => void) | undefined>(undefined)

  useEffect(() => () => messung.current?.(), [])

  const ansehen = useCallback(() => {
    messung.current?.()
    bilder.current = []
    messung.current = bildzeitMessen(bilder, setzeBildmillis)

    void vorgang.starte('partie-diagnose', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().Partiediagnose()

        auspacken(antwort, () => undefined)

        setzeStand(antwort)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [vorgang, texte])

  return {
    ansehen,
    laeuft: vorgang.laeuft('partie-diagnose'),
    stand,
    bildmillis,
    fehler,
  }
}

// bildzeitMessen mittelt den Abstand zwischen zwei Bildern.
//
// **Der Abstand und nicht die Arbeit.** Was ein Bild kostet, weiß niemand von
// außen; was man messen kann, ist, wie viel Zeit zwischen zwei Bildern liegt
// — und genau das merkt ein Mensch als Ruckeln.
function bildzeitMessen(
  bilder: { current: number[] },
  setze: (millis: number) => void,
): () => void {
  let vorher = 0
  let angefordert = 0

  const schlag = (jetzt: number) => {
    if (vorher !== 0) {
      bilder.current = [...bilder.current, jetzt - vorher]

      if (bilder.current.length === Bildproben) {
        setze(
          bilder.current.reduce((summe, eines) => summe + eines, 0) /
            Bildproben,
        )

        return
      }
    }

    vorher = jetzt
    angefordert = requestAnimationFrame(schlag)
  }

  angefordert = requestAnimationFrame(schlag)

  return () => {
    cancelAnimationFrame(angefordert)
  }
}
