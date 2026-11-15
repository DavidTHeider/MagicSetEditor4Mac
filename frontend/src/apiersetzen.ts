// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).
//
// **Eine eigene Datei**, wie `apiablage.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).
import type { Umschlag } from './api'

// ErsetzungsfrageDTO ist die Frage, wie das Fenster sie stellt.
//
// **Ein Verbund und keine fünf Stellen**: fünf Werte hintereinander sagen am
// Aufrufort nichts mehr, und JavaScript hat keine benannten Argumente.
export interface ErsetzungsfrageDTO {
  text: string
  ersatz: string

  // feld beschränkt auf ein Kartenfeld; leer heißt: alle Textfelder.
  feld: string

  grossKleinBeachten: boolean
  ganzesWort: boolean

  // stand ist der Fingerabdruck, den die Vorschau mitgab —
  // #1241.
  //
  // **Leer heißt: ohne Prüfung.** Beim Ausführen ohne vorherige Vorschau gibt
  // es nichts zu binden; wer bestätigt, gibt ihn zurück, und der Dienst weist
  // ab, wenn er nicht mehr gilt.
  stand: string
}

// ErsetzungsstelleDTO ist ein Feldwert, der sich ändern würde.
export interface ErsetzungsstelleDTO {
  karteID: string

  // nummer ist die Stelle der Karte im Set — **eins-basiert**, wie die
  // Kartenliste zählt.
  nummer: number

  feld: string

  // vorher und nachher stehen **ohne Auszeichnungen** da: ein `<i>` in einer
  // Vorschau sieht wie ein Fehler aus.
  vorher: string
  nachher: string

  treffer: number
}

// ErsetzungsAntwort ist, was sich ändern würde — oder geändert hat.
export interface ErsetzungsAntwort extends Umschlag {
  // stellen ist **begrenzt**; karten und treffer zählen alles.
  stellen: ErsetzungsstelleDTO[]

  karten: number
  treffer: number

  // weitere sagt, dass die Grenze gegriffen hat.
  weitere: boolean

  // stand ist der Fingerabdruck der betroffenen Stellen (#1241) — er geht beim
  // Bestätigen zurück.
  stand: string
}

// Ersetzungsapi sind die beiden Wege — und der erste schreibt nicht.
export interface Ersetzungsapi {
  // ErsetzungVorschau sagt, was sich ändern würde. ÄNDERT NICHTS.
  ErsetzungVorschau: (
    setID: string,
    frage: ErsetzungsfrageDTO,
  ) => Promise<ErsetzungsAntwort>

  // ErsetzungAusfuehren ersetzt und gibt dieselbe Übersicht zurück.
  //
  // **Dieselbe Übersicht**, damit der Aufrufer nicht raten muss, ob die
  // Vorschau noch galt: was hier zurückkommt, ist das, was geschrieben wurde.
  ErsetzungAusfuehren: (
    setID: string,
    frage: ErsetzungsfrageDTO,
  ) => Promise<ErsetzungsAntwort>
}
