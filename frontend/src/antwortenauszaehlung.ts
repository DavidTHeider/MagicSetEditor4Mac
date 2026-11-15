// **Die Antworten der Auszählung** —
// #1338.
//
// **Eine eigene Datei**, wie `antwortenkartenskripte.ts` und aus demselben
// Grund (§2, Zeilengrenze): hier stehen die Auskünfte des **Spielpakets** über
// das, wonach es auszählen lässt.
import type { Umschlag } from './api'

// AuszaehldimensionDTO ist ein `statistics dimension:`-Block des Spielpakets.
//
// **Ohne den Skripttext.** Er rechnet den Wert je Karte und ist für niemanden
// lesbar, der ihn nicht ausführt.
export interface AuszaehldimensionDTO {
  name: string
  beschreibung: string

  // numerisch sagt, dass die Gruppen eine Achse bilden und keine Aufzählung.
  numerisch: boolean
}

// AuszaehlgruppeDTO ist ein Balken: ein Wert und wie viele Karten ihn tragen.
export interface AuszaehlgruppeDTO {
  name: string
  karten: number

  // farbe ist die im Paket erklärte Farbe, in MSEs Schreibweise — leer, wo
  // keine dasteht.
  farbe: string

  // kartenIDs sind die Karten des Balkens — UXR-F6
  // (#2013),
  // aufsteigend. Eine leere Gruppe nennt eine leere Liste.
  kartenIDs: string[]
}

// Auszaehldimensionsantwort trägt die Dimensionen des Spielpakets.
export interface Auszaehldimensionsantwort extends Umschlag {
  dimensionen: AuszaehldimensionDTO[]
}

// Auszaehlantwort trägt eine Auszählung.
//
// **Flach und nicht um ein geschachteltes DTO herum** — dieselbe Zusage wie
// bei `Kartenskriptlaufantwort`: beide Listen sind auf jedem Weg da (§3).
export interface Auszaehlantwort extends Umschlag {
  dimension: string
  beschreibung: string
  numerisch: boolean
  gruppen: AuszaehlgruppeDTO[]
  befunde: string[]
}
