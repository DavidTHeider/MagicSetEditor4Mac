// **Die Antworten der Massenvorschau und des Verlaufs** — Werkstück U6 von
// Phase 5.1 (#1417).
// Eine eigene Datei, weil `antworten.ts` an der Zeilengrenze aus §2 steht;
// `./antworten` reicht sie weiter.
import type { Umschlag } from './umschlag'

// FeldvorwertDTO ist der Altwert einer Karte, wie die Vorschau ihn gezeigt
// hat — und wie das Fenster ihn beim Anwenden zurückgibt (U6, #1417).
export interface FeldvorwertDTO {
  karteID: string
  wert: string
  gabEs: boolean
}

// MassenvorschauzeileDTO ist eine Zeile der Vorschau einer Massenänderung.
export interface MassenvorschauzeileDTO {
  karteID: string
  name: string
  alt: string
  gabAlt: boolean
  neu: string
  // ausschluss sagt, warum die Karte nicht geschrieben wird — leer, wenn sie
  // geschrieben wird; `gleich` heißt: sie steht schon so.
  ausschluss: string
}

// VerlaufsschrittDTO ist ein Schritt des Verlaufs — eine Massenänderung als
// ein Schritt mit Umfang, jede Zeile mit Vorher und Nachher (U6, #1417).
export interface VerlaufsschrittDTO {
  id: string
  zeitpunkt: string
  handlung: string
  karteID: string
  feld: string
  vorher: string
  gabVorher: boolean
  nachher: string
  gabNachher: boolean
  zurueckgenommen: boolean
  urheber: string
  herkunft: string
  // vorgang ist die Kennung für eine gezielte Rücknahme — leer bei einer
  // Einzeländerung.
  vorgang: string
  umfang: number
}

// VerlaufAntwort ist die Liste der Schritte, der jüngste zuerst.
export interface VerlaufAntwort extends Umschlag {
  schritte: VerlaufsschrittDTO[]
}

// MassenAntwort nennt die Karten, an denen eine Massenänderung gescheitert ist.
//
// **Nicht die Auswahl**: geschrieben wurde nichts, und nur die hier sind daran
// schuld.
export interface MassenAntwort extends Umschlag {
  betroffen: string[]
  vorgang?: string
}

// MassenvorschauAntwort ist die Vorschau einer Massenänderung.
export interface MassenvorschauAntwort extends Umschlag {
  feld: string
  wert: string
  zeilen: MassenvorschauzeileDTO[]
}
