// Die Antworten der **Paketwerkstatt** (#1164).
//
// **Eine eigene Datei**, weil `antworten.ts` an der Zeilengrenze aus §2 steht
// — derselbe Schnitt wie bei `antwortensammlung.ts`.
import type { Skriptstelle } from './api'
import type { Umschlag } from './umschlag'

// Paketdateiantwort trägt die bearbeitbaren Dateien eines Pakets (#1164) und
// sein Beiwerk — Bilder und Schriften, die sich nicht als Text lesen lassen
// (FMT-02, #1691).
export interface Paketdateiantwort extends Umschlag {
  dateien: string[]
  beiwerk: string[]
}

// Paketdateieinfuegung sagt, unter welchem Namen eine Datei aus dem Dialog
// ins Paket kam — oder dass niemand eine gewählt hat (FMT-02, #1691).
export interface Paketdateieinfuegung extends Umschlag {
  datei: string
  abgebrochen: boolean
}

// Paketdateiinhalt trägt eine Paketdatei samt ihrem Stand (#1164).
export interface Paketdateiinhalt extends Umschlag {
  inhalt: string

  // stand ist die SHA-256 des Inhalts — sie geht beim Speichern zurück.
  stand: string
}

// Paketdateistand trägt den Stand nach dem Speichern (#1164).
export interface Paketdateistand extends Umschlag {
  stand: string
}

// Skriptbefundantwort trägt die Stelle eines Syntaxfehlers in einer
// Paketdatei — Werkstück B von Phase 4.9
// (#1324).
//
// **Ohne `Umschlag`**, weil die Prüfung keinen Fehlerweg hat: sie liest einen
// Text, den der Aufrufer mitbringt, und schreibt nichts. Ein Umschlag um eine
// Antwort ohne Fehler wäre ein Zweig, den kein Lauf betreten kann (§3).
export interface Skriptbefundantwort {
  // stelle ist der Ort — `null`, wenn kein Skriptwert einen Fehler hat.
  stelle: Skriptstelle | null

  // meldung ist der deutsche Satz (§9); das Fenster setzt über den Schlüssel
  // seinen eigenen ein.
  meldung: string

  // schluessel benennt die Meldung — derselbe Katalogschlüssel wie bei der
  // Konsole.
  schluessel: string

  // werte füllt die Platzhalter des Satzes.
  werte: string[]
}

// ProbekarteDTO ist eine wählbare Testkarte der Werkstatt — Werkstück U7 von
// Phase 5.1 (#1418).
export interface ProbekarteDTO {
  id: string

  // name ist der Wert des Feldes `name` — leer, wenn die Karte keinen hat.
  name: string

  // nummer ist die Nummer im Set — die Aufschrift, wenn der Name fehlt.
  nummer: number
}

// ProbekartenAntwort trägt die Karten des ersten Sets, das den Stil des
// Pakets benutzt — leer, wenn es keines gibt.
export interface ProbekartenAntwort extends Umschlag {
  karten: ProbekarteDTO[]
}

// PaketsymbolDTO ist ein Name, den ein Paket definiert, mit seiner Stelle
// (U7, #1418).
export interface PaketsymbolDTO {
  name: string

  // art ist der definierende Block — `card field`, `set field`,
  // `styling field`, `extra card field` — oder `script` für eine Zuweisung.
  art: string

  datei: string
  zeile: number
  spalte: number
}

// PaketsymboleAntwort trägt jede Definition eines Pakets.
export interface PaketsymboleAntwort extends Umschlag {
  symbole: PaketsymbolDTO[]
}

// SymbolverwendungDTO ist eine Zeile, die einen Namen nennt.
export interface SymbolverwendungDTO {
  datei: string
  zeile: number

  // text ist die Zeile ohne Einrückung — zum Zeigen.
  text: string
}

// SymbolverwendungenAntwort trägt die Referenzen eines Namens.
export interface SymbolverwendungenAntwort extends Umschlag {
  verwendungen: SymbolverwendungDTO[]
}
