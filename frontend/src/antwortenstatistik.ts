// **Die Antworten der Statistik** — aus `antworten.ts` herausgelöst, als
// `WertzahlDTO` die Datei über die Zeilengrenze aus §2 hob. Hier stehen die
// Zahlen, die ein Set über sich sagt: Kennzahlen, Belegung, Verteilung.
import type { Umschlag } from './api'

// ZahlDTO ist eine Zeile der Belegung.
export interface ZahlDTO {
  name: string
  karten: number
}

// WertzahlDTO ist eine Zeile der Verteilung — samt den Karten des Werts, UXR-F6
// (#2013).
//
// **Nur die Verteilung nennt sie.** Eine Zeile der Belegung ist ein Feld, und
// eine Karte steht in vielen — das wären Feldzahl mal Kartenzahl Kennungen für
// eine Frage, die niemand stellt. Aufsteigend, und jede Karte einmal.
export interface WertzahlDTO extends ZahlDTO {
  kartenIDs: string[]
}

export interface StatistikDTO {
  karten: number
  felder: number
  leere: number
  belegung: ZahlDTO[]
}

export interface StatistikAntwort extends Umschlag {
  statistik: StatistikDTO
}

export interface VerteilungAntwort extends Umschlag {
  verteilung: WertzahlDTO[]
}
