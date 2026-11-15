// **Das Betriebsband fragt von selbst** — OPS02
// (#2348).
//
// # Höchstens alle 30 Sekunden
//
// **Das Kriterium nennt die Zahl**, und sie passt zu dem, was gemessen wird:
// Platz und Sicherung ändern sich in Minuten, ein Dienst, der endet, soll
// sich nicht erst nach dem nächsten Start melden. Jeder Abruf misst seinen
// Aufwand selbst (`aufwandMikros`).
//
// # Still, wie der Bestandsstand
//
// **Ein Fehler bleibt still und lässt die letzten Hinweise stehen** — wie in
// `useBestandsstand`: die Abfrage hat niemand angefordert, und 30 Sekunden
// später kommt die nächste. Gefragt wird nur, wenn jemand hinsieht, und nur
// eine zur Zeit.
import { useEffect, useState } from 'react'

import { api } from '../api'
import type { BetriebshinweisDTO } from '../apibetriebsband'

// betriebstakt ist der Abstand zweier Abfragen.
export const betriebstakt = 30_000

// useBetriebsband gibt die Hinweise des letzten gelungenen Abrufs — leer, bis
// einer gelang.
export function useBetriebsband(): BetriebshinweisDTO[] {
  const [hinweise, setzeHinweise] = useState<BetriebshinweisDTO[]>([])

  useEffect(() => {
    let laeuft = true
    let unterwegs = false

    const fragen = async () => {
      if (unterwegs || document.visibilityState === 'hidden') {
        return
      }

      unterwegs = true

      try {
        const antwort = await api().Betriebsband()

        if (laeuft && !antwort.fehler) {
          setzeHinweise(antwort.hinweise)
        }
      } catch {
        // Ein Abbruch über die Naht — die letzten Hinweise bleiben.
      } finally {
        unterwegs = false
      }
    }

    void fragen()

    const uhr = setInterval(() => {
      void fragen()
    }, betriebstakt)

    return () => {
      laeuft = false
      clearInterval(uhr)
    }
  }, [])

  return hinweise
}
