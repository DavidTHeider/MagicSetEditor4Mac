// **Die Zwischenablage über Karten** — Werkstück 1 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, wie `apikarten.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).
import type { SchreibAntwort, Umschlag } from './api'

// EinfuegeAntwort trägt die Kennungen der eingefügten Karten.
export interface EinfuegeAntwort extends Umschlag {
  // karten sind die neuen Karten — **auch wenn das Füllen scheitert**. Eine
  // Karte, die entstanden ist, verschwindet nicht; wer sie nicht nennt, lässt
  // sie im Bestand zurück, ohne dass jemand davon weiß.
  karten: string[]

  // befunde sind die Felder, die nicht angekommen sind — mit ihrem Namen.
  //
  // **Ein Set aus einem fremden Spiel wird ganz abgewiesen** (§8); was hier
  // steht, sind die Fälle *innerhalb* desselben Spiels: ein Feld, das die
  // Vorlage des Zielsets nicht kennt.
  befunde: string[]
}

// Ablageapi sind die drei Wege der Zwischenablage.
//
// `KarteKopieren` dupliziert **innerhalb** eines Sets; diese drei tragen
// zwischen zwei Sets — und über jede Grenze, die Text kennt.
export interface Ablageapi {
  // AblageKopieren legt die genannten Karten als Text auf das Klemmbrett.
  AblageKopieren: (setID: string, karteIDs: string[]) => Promise<SchreibAntwort>

  // AblageAusschneiden tut dasselbe und entfernt sie danach aus dem Set.
  //
  // **In dieser Reihenfolge**, wie MSE2: `canCut()` ist `canDelete()`, und
  // wer zuerst löscht, hat verloren, was er wegnehmen wollte.
  AblageAusschneiden: (
    setID: string,
    karteIDs: string[],
  ) => Promise<SchreibAntwort>

  // AblageEinfuegen legt die Karten vom Klemmbrett im Set an.
  //
  // **Ohne Kartenliste**: was hineinkommt, steht im Klemmbrett, und das liest
  // die Go-Seite selbst — ein Text durch die Naht wäre derselbe Inhalt auf
  // einem zweiten Weg.
  AblageEinfuegen: (setID: string) => Promise<EinfuegeAntwort>
}
