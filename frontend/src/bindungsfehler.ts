// Der strukturierte Fehler aus `internal/app`, als Fehlerobjekt.
//
// **Eine eigene Datei**, weil `api.ts` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht die Fassade, hier das, was sie ablehnt.
import type { FehlerDTO } from './formen'

// Bindungsfehler trägt den strukturierten Fehler durch eine abgelehnte Zusage.
//
// **Als Fehlerobjekt und nicht als bloße Meldung**: `code` entscheidet, ob das
// Frontend nachlädt oder hinweist, `feld` sagt, wo die Meldung erscheint (§6).
export class Bindungsfehler extends Error {
  readonly code: FehlerDTO['code']
  readonly schluessel: string
  readonly feld: string | undefined
  readonly werte: Record<string, string>

  // befunde sind die **Begleitbefunde** einer abgelehnten Antwort.
  //
  // # Warum ein Fehler sie überhaupt tragen kann
  //
  // Weil §8 es verlangt: „Ein Set, von dem 300 Karten lesbar sind und zwei
  // nicht, wird mit 300 Karten und zwei benannten Befunden geöffnet." Der Weg
  // dahinter kann trotzdem scheitern — und dann gingen die Befunde bisher
  // verloren, weil `auspacken` nur den Fehler warf.
  //
  // **Gefunden am Preisholen** (#1202):
  // Go reicht unlesbare Preisfelder ausdrücklich bis zur Antwort,
  // Kommandozeile, MCP und Dienst geben sie aus — und genau an der Fläche, an
  // der ein Mensch sitzt, fielen sie weg.
  //
  // **Leer heißt: keine.** Nicht `undefined`, damit kein Aufrufer auf zwei
  // Dinge prüft.
  readonly befunde: readonly string[]

  constructor(fehler: FehlerDTO, befunde: readonly string[] = []) {
    super(fehler.meldung)
    this.name = 'Bindungsfehler'
    this.code = fehler.code
    this.schluessel = fehler.schluessel
    this.feld = fehler.feld
    this.befunde = befunde

    // **Ein leeres Objekt und nicht `undefined`.** Der Einsetzer bekommt so
    // immer eine Tabelle, und ein Platzhalter ohne Wert bleibt stehen, statt
    // einen Zugriff auf `undefined` zu erzwingen.
    this.werte = fehler.werte ?? {}
  }
}
