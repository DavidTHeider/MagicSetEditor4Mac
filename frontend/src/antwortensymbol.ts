// **Die Antworten des Symboleditors** —
// #890.
//
// **Eine eigene Datei**, weil `antworten.ts` mit der Symmetrie über die
// Zeilengrenze aus §2 wuchs
// (#1092).
// Der Schnitt folgt der Frage: dort steht, was der **Bestand** antwortet, hier,
// was der **Editor** antwortet.

import type { Umschlag } from './umschlag'

// SymbolpunktDTO ist ein Kontrollpunkt einer Symbolzeichnung (#890).
//
// **Die Griffe stehen relativ zur Stelle**, so wie in der Datei — wer sie
// absolut reichte, müsste sie bei jedem Ziehen umrechnen, und die Umrechnung
// stünde dann auf beiden Seiten der Naht.
export interface SymbolpunktDTO {
  x: number
  y: number

  griffDavorX: number
  griffDavorY: number
  griffDanachX: number
  griffDanachY: number

  kurveDanach: boolean

  // sperre ist `free`, `direction` oder `size` — sie bindet die beiden Griffe
  // aneinander, während jemand zeichnet.
  sperre: string
}

// SymbolteilDTO ist ein Teil einer Zeichnung (#890).
export interface SymbolteilDTO {
  // art ist `Form`, `Symmetrie` oder `Gruppe`.
  art: string
  name: string

  kombination: string
  punkte: SymbolpunktDTO[]

  spiegelart: string
  kopien: number
  mitteX: number
  mitteY: number
  griffX: number
  griffY: number

  teile: SymbolteilDTO[]
}

// SymbolzugDTO ist ein geschlossener Zug — eine Kopie eines Kindteils (#1092).
export interface SymbolzugDTO {
  punkte: SymbolpunktDTO[]
}

// SymmetriekopienAntwort trägt die Züge, die eine Symmetrie zeichnet (#1092).
export interface SymmetriekopienAntwort extends Umschlag {
  kopien: SymbolzugDTO[][]
}

// SymmetrieweiseDTO sagt, welche Symmetrie geändert wird und wie (#1092).
//
// **`mitteSetzen` und `griffSetzen` statt fehlender Felder**: die 0 ist eine
// gültige Stelle — die linke obere Ecke —, und ein Aufruf, der nur die Art
// ändert, dürfte die Mitte nicht dorthin ziehen.
export interface SymmetrieweiseDTO {
  teil: number
  spiegelart: string
  kopien: number
  mitteX: number
  mitteY: number
  griffX: number
  griffY: number
  mitteSetzen: boolean
  griffSetzen: boolean
}

// SymbolgrundformDTO nennt, welche Grundform gewünscht ist (#890).
export interface SymbolgrundformDTO {
  // art ist `Kreis`, `Rechteck`, `Vieleck` oder `Stern`.
  art: string

  // ecken gilt für Vieleck und Stern.
  ecken: number

  // eng heißt gleichseitig — Kreis statt Ellipse, Quadrat statt Rechteck.
  eng: boolean
}

// SymbolpunktweiseDTO nennt, welcher Punkt gemeint ist (#890).
export interface SymbolpunktweiseDTO {
  teil: number
  punkt: number
}

// SymbolgriffDTO nennt, welcher Griff wohin gezogen wird (#890).
export interface SymbolgriffDTO {
  teil: number
  punkt: number

  // davor sagt, welcher der beiden Griffe gemeint ist.
  davor: boolean

  // x und y sind die **neue Stelle der Griffspitze**, nicht eine
  // Verschiebung — eine Stelle veraltet nicht, eine Differenz gegen einen
  // gemerkten Anfangswert schon.
  x: number
  y: number
}

// SymbolAntwort trägt eine gezeichnete Zeichnung (#890).
export interface SymbolAntwort extends Umschlag {
  teile: SymbolteilDTO[]

  // name ist der Dateiname im Set — leer, solange keine abgelegt ist.
  name: string
}
