// **Das Symbol eines Sets** —
// #890.
//
// **Eine eigene Datei**, wie `apifeld.ts` und `apibilder.ts`, und aus
// demselben Grund: `api.ts` führt die ganze Fassade und liefe sonst über die
// Zeilengrenze (§2).

import type {
  SymbolAntwort,
  SymbolgriffDTO,
  SymbolgrundformDTO,
  SymbolpunktweiseDTO,
  SymbolteilDTO,
  SymmetriekopienAntwort,
  SymmetrieweiseDTO,
} from './antworten'

// Symbolapi ist der Weg zum gezeichneten Setsymbol.
export interface Symbolapi {
  // SymbolSetzen legt eine gezeichnete Zeichnung neben das Set und gibt ihren
  // Dateinamen.
  //
  // **Sie wird dabei nicht zum Setsymbol.** Das ist ein Feldwert, und
  // `SetwertSetzen` schreibt ihn — eine zweite Stelle, die dasselbe Feld
  // setzt, liefe mit der ersten auseinander (§7, Regel 4).
  // **`unter` nennt die Datei, die überschrieben werden soll.** Leer heißt:
  // eine neue anlegen. Wer zum zweiten Mal ablegt, reicht den Namen von
  // vorhin — sonst bliebe die erste Datei als verwaistes Beiwerk liegen
  // (#890, Befund 1 des Nachlaufs).
  SymbolSetzen: (
    setID: string,
    teile: SymbolteilDTO[],
    unter: string,
  ) => Promise<SymbolAntwort>

  // Symbol liest eine abgelegte Zeichnung zurück.
  Symbol: (setID: string, name: string) => Promise<SymbolAntwort>

  // Grundform gibt eine der vier Grundformen, mittig und in voller Größe.
  //
  // **Gerechnet wird auf der Go-Seite.** Der Kreis benutzt eine Konstante, die
  // MSE selbst „magic" nennt, das Vieleck eine Herleitung über drei
  // Gleichungen. Eine zweite Rechnung hier liefe mit jener auseinander.
  Grundform: (form: SymbolgrundformDTO) => Promise<SymbolAntwort>

  // SymbolNachzeichnen macht aus einer Bilddatei die Teile einer Zeichnung.
  //
  // **Ohne Argumente**: der Dateidialog steckt in der Bindung, wie bei
  // `BildSetzen`. Nur deshalb können Kommandozeile, MCP und Dienst dieselbe
  // Fähigkeit anbieten (§16) — dort steht der Pfad im Aufruf.
  //
  // **Ein Abbruch gibt eine leere Liste ohne Fehler.** Wer den Dialog
  // schließt, hat nichts gewählt.
  SymbolNachzeichnen: () => Promise<SymbolAntwort>

  // PunktEinfuegen setzt einen Punkt in die **Mitte** der Strecke dahinter.
  //
  // **Die Zeichnung geht hin und zurück.** Sie ist klein, und die Rechnung
  // gehört auf die Go-Seite (§2): ein Zustand dort wäre ein zweiter neben dem
  // der Seite, und zwei Zustände laufen auseinander.
  PunktEinfuegen: (
    teile: SymbolteilDTO[],
    weise: SymbolpunktweiseDTO,
  ) => Promise<SymbolAntwort>

  // GriffZiehen bewegt einen Bézierhandgriff und zieht den anderen nach, wie
  // die Sperre es verlangt.
  //
  // **Was die Sperre daraus macht, entscheidet die Go-Seite**: `free` lässt
  // den anderen stehen, `size` macht ihn zum Gegenteil, `direction` dreht ihn
  // und behält seine Länge.
  GriffZiehen: (
    teile: SymbolteilDTO[],
    griff: SymbolgriffDTO,
  ) => Promise<SymbolAntwort>

  // PunktLoeschen entfernt einen Punkt — unter drei geht es nicht.
  PunktLoeschen: (
    teile: SymbolteilDTO[],
    weise: SymbolpunktweiseDTO,
  ) => Promise<SymbolAntwort>

  // **Die Symmetrie** —
  // #1092.
  //
  // SymmetrieAnlegen umschließt das gewählte Teil, statt ein leeres anzulegen:
  // eine Symmetrie ohne Kinder zeichnet nichts.
  SymmetrieAnlegen: (
    teile: SymbolteilDTO[],
    stelle: number,
  ) => Promise<SymbolAntwort>

  // SymmetrieAufloesen macht aus ihr ihre Kinder — sie bleiben stehen.
  SymmetrieAufloesen: (
    teile: SymbolteilDTO[],
    stelle: number,
  ) => Promise<SymbolAntwort>

  // SymmetrieSetzen ändert Art, Kopienzahl, Mitte oder Griff.
  SymmetrieSetzen: (
    teile: SymbolteilDTO[],
    weise: SymmetrieweiseDTO,
  ) => Promise<SymbolAntwort>

  // **Gruppen und Reihenfolge** —
  // #1168.
  //
  // TeileGruppieren fasst die genannten Stellen zusammen; die Gruppe steht,
  // wo das erste gewählte Teil stand.
  TeileGruppieren: (
    teile: SymbolteilDTO[],
    stellen: number[],
  ) => Promise<SymbolAntwort>

  // GruppeAufloesen setzt die Kinder an die Stelle ihrer Gruppe.
  GruppeAufloesen: (
    teile: SymbolteilDTO[],
    stellen: number[],
  ) => Promise<SymbolAntwort>

  // TeilVerschieben ändert die Reihenfolge — und damit, was über was liegt.
  //
  // **Die Zielstelle zählt ohne das bewegte Teil**; die Rechnung steht auf
  // der Go-Seite, weil sie von MSE kommt.
  TeilVerschieben: (
    teile: SymbolteilDTO[],
    von: number,
    nach: number,
  ) => Promise<SymbolAntwort>

  // Symmetriekopien gibt die Züge, die eine Symmetrie zeichnet.
  //
  // **Von der Go-Seite und nicht hier gerechnet**: die Abbildung hat drei
  // Fallen, die man ohne den C++-Quelltext verfehlt — die
  // Multiplikationsrichtung, die Spiegelachse bei `(a + b)/2`, das Abrunden
  // der Kopienzahl auf eine gerade. Eine zweite Fassung hier liefe mit jener
  // auseinander, und dann zeichnete der Editor etwas anderes als der Rasterer.
  Symmetriekopien: (
    teile: SymbolteilDTO[],
    stelle: number,
  ) => Promise<SymmetriekopienAntwort>
}
