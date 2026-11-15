// **Die eigenen Partien** —
// #1600.
//
// # Wofür es ihn gibt
//
// **Weil eine Partie im Bestand liegt und im Fenster nur im Arbeitsspeicher.**
// `usePartie` hält sie im Zustand; jeder Neustart, jeder Bestandswechsel und
// jeder **Profilwechsel** wirft ihn weg — `App.tsx` trägt `key={profilstand}`,
// und §7 Regel 1 verlangt genau das. Die Partie stand weiter im Bestand, und
// es führte kein Weg dorthin.
//
// # Warum er lädt und nicht horcht
//
// **Weil eine Liste kein Tisch ist.** Was an einem Tisch geschieht, meldet der
// Wirt (G5); **welche** Tische es gibt, ändert sich nur, wenn jemand eine
// Partie anfängt oder beendet — und dann lädt der Aufrufer neu.
//
// **Seit UXR-F16 auch, wenn jemand den Tisch verlässt**
// (#2031):
// „dran" in einer Zeile ist eine Auskunft über den Zug, und die ändert sich am
// Tisch. Gehört dem Aufrufer, nicht dem Haken — er weiß, wann ein Tisch zugeht.
import { api, auspacken } from '../api'
import { useLaden } from './useLaden'

import type { Partiezeile } from '../apipartie'

export interface Partienstand {
  // partien sind die eigenen, zuletzt geänderte zuerst — **leer, solange
  // nichts geladen ist**, und nicht `undefined`.
  partien: Partiezeile[]

  laedt: boolean

  // geladen sagt, dass die Liste einmal angekommen ist — vorher ist sie
  // unbekannt und nicht leer (UX-08.NAT-01).
  geladen: boolean
  fehler: string | undefined

  // nachladen holt die Liste erneut — für den Aufrufer, der gerade eine
  // Partie angefangen oder verlassen hat.
  nachladen: () => void
}

// **`fremdstand` zählt, was ein anderer Prozess oder die Leitung geschrieben hat**
// (#156) — Befund 3 des Gesamtreviews über 0.6.0: ohne ihn lud die Seite erst
// nach einem Wechsel nach. **Behalten**, damit die Liste dabei stehen bleibt.
export function usePartienliste(fremdstand = 0): Partienstand {
  const stand = useLaden(
    async () => {
      const antwort = await api().Partien()

      return auspacken(antwort, () => antwort.partien)
    },
    [fremdstand],
    true,
  )

  return {
    partien: stand.daten ?? [],
    laedt: stand.laedt,
    geladen: stand.daten !== undefined,
    fehler: stand.fehler,
    nachladen: stand.nachladen,
  }
}
