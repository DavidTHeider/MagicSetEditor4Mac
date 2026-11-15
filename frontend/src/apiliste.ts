// **Die Gestalt der Kartenliste** — Werkstück 3 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, wie `apiablage.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).
import type { Umschlag } from './api'

// ListenspalteDTO ist eine Spalte, wie die Liste sie anbietet.
export interface ListenspalteDTO {
  // feld ist der Feldname, wie das Spielpaket ihn schreibt — nicht übersetzt.
  feld: string

  // aufschrift ist `card list name:` — **leer bleibt leer**, und das Fenster
  // entscheidet, was es dann zeigt: den Feldnamen.
  aufschrift: string

  sichtbar: boolean

  // ausrichtung ist `links`, `rechts` oder `mitte`.
  ausrichtung: string
}

// ListenzeileDTO ist eine Karte, wie die Liste sie zeigt.
export interface ListenzeileDTO {
  karteID: string

  // nummer ist die Stelle im **Set**, eins-basiert — sie bleibt, wenn jemand
  // anders sortiert.
  nummer: number

  // werte trägt **jede** angebotene Spalte, auch die unsichtbaren: wer eine
  // Spalte einblendet, soll nicht auf eine neue Antwort warten müssen.
  werte: Record<string, string>

  // marken sind die Marken dieser Karte, alphabetisch
  // (#1208).
  //
  // **In der Zeile und nicht in einem zweiten Ruf**: sie werden gezeigt und es
  // wird nach ihnen gefiltert, und eine Abfrage je Zeile wären bei 3.788
  // Karten 3.788 Abfragen.
  marken: string[]
}

// KartenlisteAntwort ist der Aufbau der Liste.
// KartenfilterAntwort trägt die Kennungen der Karten, die durch den
// Schnellfilter kommen.
export interface KartenfilterAntwort extends Umschlag {
  kartenIDs: string[]
}

export interface KartenlisteAntwort extends Umschlag {
  spalten: ListenspalteDTO[]
  zeilen: ListenzeileDTO[]

  // sortierfeld ist die **wirksame** Ordnung — leer heißt: die Reihenfolge des
  // Sets. Zeigt die gemerkte Wahl auf ein Feld, das es nicht mehr gibt, steht
  // hier nichts.
  sortierfeld: string
  aufsteigend: boolean

  // ohneVorlage sagt, dass die Kartenfelder des Spiels nicht zu lesen waren —
  // dann kommen die Spalten aus den Karten, ohne Aufschrift und ohne die
  // Reihenfolge des Pakets.
  ohneVorlage: boolean
}

// Listenapi sind die drei Wege zur Gestalt der Liste.
export interface Listenapi {
  // Kartenlistenaufbau gibt Spalten, Zeilen und die wirksame Ordnung.
  Kartenlistenaufbau: (setID: string) => Promise<KartenlisteAntwort>

  // KartenFiltern legt den Schnellfilter über die Karten eines Sets und gibt,
  // wer bleibt — Werkstück U5 von Phase 5.1 (#1416). Die Liste selbst hat das
  // Fenster schon.
  KartenFiltern: (
    setID: string,
    anfrage: string,
  ) => Promise<KartenfilterAntwort>

  // ListenspaltenSetzen schreibt die Wahl fort — **die ganze Wahl**.
  //
  // **Die Reihenfolge der Liste ist die Reihenfolge der Spalten.** Eine
  // Stellenangabe daneben wäre eine zweite Antwort auf dieselbe Frage.
  ListenspaltenSetzen: (
    setID: string,
    spalten: ListenspalteDTO[],
  ) => Promise<Umschlag>

  // ListensortierungSetzen schreibt sie fort — leeres Feld heißt: gar nicht.
  ListensortierungSetzen: (
    setID: string,
    feld: string,
    aufsteigend: boolean,
  ) => Promise<Umschlag>
}
