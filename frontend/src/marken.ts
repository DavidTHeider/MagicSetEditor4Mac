// Die Typen der Marken an Karten
// (#1208).
//
// **Eine eigene Datei**, wie `stilwahl.ts`: die Fassade nennt die Methoden,
// die Typen stehen bei ihrem Gegenstand (§2).
import type { Umschlag } from './api'

export interface MarkenAntwort extends Umschlag {
  // marken ist **nie `null`** — die Go-Seite gibt eine leere Liste.
  marken: string[]
}
