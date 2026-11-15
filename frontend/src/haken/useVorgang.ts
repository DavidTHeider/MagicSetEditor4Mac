// Regel 4 aus AGENTS.md §7: **ein Vorgang je Datensatz.**
//
// Ein zweiter Aufruf, während der erste läuft, geht gegen einen Stand, den das
// Backend noch nicht kennt. Wer zweimal schnell auf „Speichern" klickt, schickt
// zwei Änderungen gegen dieselbe Fassung — die zweite gewinnt, und was die erste
// wollte, ist weg.
import { useCallback, useRef, useState } from 'react'

export interface Vorgang {
  // laeuft sagt, ob für diesen Datensatz gerade etwas unterwegs ist.
  laeuft: (kennung: string | number) => boolean

  // starte führt die Arbeit aus — oder gar nicht, wenn schon etwas läuft.
  //
  // Der Rückgabewert unterscheidet beides: `undefined` heißt „nicht
  // angefangen". Ohne diese Auskunft sähe ein übersprungener Aufruf wie ein
  // geglückter aus.
  starte: <T>(
    kennung: string | number,
    arbeit: () => Promise<T>,
  ) => Promise<T | undefined>
}

export function useVorgang(): Vorgang {
  const laufende = useRef(new Set<string>())
  const [, neuZeichnen] = useState(0)

  const laeuft = useCallback(
    (kennung: string | number) => laufende.current.has(String(kennung)),
    [],
  )

  const starte = useCallback(
    async <T>(kennung: string | number, arbeit: () => Promise<T>) => {
      const schluessel = String(kennung)
      if (laufende.current.has(schluessel)) {
        return undefined
      }

      laufende.current.add(schluessel)
      neuZeichnen((zahl) => zahl + 1)

      try {
        return await arbeit()
      } finally {
        // **Im `finally`.** Ein Vorgang, der bei einem Fehler eingetragen
        // bleibt, sperrt den Datensatz für den Rest der Sitzung — und niemand
        // sieht, warum die Schaltfläche nicht mehr reagiert.
        laufende.current.delete(schluessel)
        neuZeichnen((zahl) => zahl + 1)
      }
    },
    [],
  )

  return { laeuft, starte }
}

// zeigervorgang ist der Schlüssel für jeden Wechsel des Zeigers.
//
// **Eine Konstante und keine Kennung** — der Nachlauf zum Zweitblick über
// Phase 4.12: Profil- und Nutzerwechsel schreiben **dieselbe** Zeile im
// Bestand, gleich wohin sie zeigen. Wer nach dem **Ziel** schlüsselte, ließ
// zwei Klicks auf verschiedene Ziele nebeneinander laufen (§7, Regel 4).
export const zeigervorgang = 'zeiger'
