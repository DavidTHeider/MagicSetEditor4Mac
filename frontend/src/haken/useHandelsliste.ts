// **Die Liste der Handelsvorgänge laden** — TRADE-01
// (#1715).
//
// **Der vorige Stand bleibt stehen, während getippt wird** (`behalten`):
// eine Liste, die bei jedem Buchstaben leer würde, springt — und eine späte
// Antwort auf „bo" überschreibt nicht die auf „bolt", denn `useLaden` zählt
// die Läufe (§7, Regel 2).
import { api } from '../fassade'
import { useSprache } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

import type { Handelslage, Handelsordnung } from '../apihandelsverzeichnis'

// Handelsauswahl ist, welche Handel die Liste zeigt — „offen" sind die
// beiden Lagen, in denen noch etwas geschieht.
export type Handelsauswahl = 'offen' | Handelslage

export const handelsauswahlen: readonly Handelsauswahl[] = [
  'offen',
  'dran',
  'wartet',
  'abgeschlossen',
  'abgebrochen',
]

// lagenFuer gibt die Lagen einer Auswahl.
export function lagenFuer(auswahl: Handelsauswahl): Handelslage[] {
  return auswahl === 'offen' ? ['dran', 'wartet'] : [auswahl]
}

export interface Handelslistenwunsch {
  suche: string
  auswahl: Handelsauswahl
  archiv: boolean
  ordnung: Handelsordnung
}

// useHandelsliste lädt die Liste — neu, wenn sich der Wunsch oder `stand`
// ändert (nach einer Änderung an einem Vorgang).
export function useHandelsliste(wunsch: Handelslistenwunsch, stand: number) {
  const sprache = useSprache()
  const { suche, auswahl, archiv, ordnung } = wunsch

  return useLaden(
    async () => {
      const antwort = await api().HandelUebersicht({
        partner: '',
        lagen: lagenFuer(auswahl),
        suche,
        marke: '',
        seit: '',
        bis: '',
        ordnung,
        archiv,
        sprache,
      })

      // **Ohne Identität ist das kein Fehler, sondern der Anfang** (#2027):
      // `null` sagt der Liste, dass der nächste Schritt eine Identität ist.
      // Jeder andere Fehler bleibt einer.
      if (antwort.fehler?.schluessel === 'ohnePeeridentitaet') {
        return null
      }

      return auspacken(antwort, () => antwort.zeilen)
    },
    [suche, auswahl, archiv, ordnung, sprache, stand],
    true,
  )
}
