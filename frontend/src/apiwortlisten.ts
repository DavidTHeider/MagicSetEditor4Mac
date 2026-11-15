// Die **Auswahllisten** eines Kartenfeldes
// (#1074).
//
// **Eine eigene Datei, wie `apibilder.ts` und `apiscryfall.ts`** — und aus
// demselben Grund: `api.ts` führt die ganze Fassade und liefe sonst über die
// Zeilengrenze (§2).

import type { SchreibAntwort } from './api'
import type { WortlistenAntwort } from './antworten'

// Wortlistenapi ist der Weg zu den Auswahllisten eines Feldes.
//
// **Zwei Schritte und nicht einer**, wie bei Scryfall (§16, Regel 6): erst
// zeigen, was zur Wahl steht, dann wählen.
export interface Wortlistenapi {
  // Wortlisten gibt, was ein Feld zur Auswahl anbietet — die Wörter, die das
  // Spiel für diese Stelle kennt.
  //
  // Ein Feld kann mehrere Bereiche haben: `<word-list-type-en>Creature</…>`
  // und `<word-list-class-en>Bear</…>` stehen im selben Typfeld.
  Wortlisten: (karteID: string, feld: string) => Promise<WortlistenAntwort>

  // WortWaehlen schreibt den Wert, der nach einem Klick ins Menü dasteht.
  //
  // **Rechnen und Schreiben in einem Aufruf** (§7, Regel 4): der neue Wert
  // kommt aus `service`, nicht aus dem Fenster — wer ihn hier zusammensetzte,
  // baute MSEs Auswahllogik ein zweites Mal nach.
  //
  // **Der Weg zählt die Einträge von oben**: `[3]` ist der vierte Eintrag des
  // Menüs, `[2, 0]` der erste im Untermenü des dritten.
  //
  // **`bekannt` kommt vom Entwurf des Feldes** (U02b, #2179): steht im Bestand
  // ein Wert, den er nicht kennt, schreibt die Wahl nicht.
  WortWaehlen: (
    setID: string,
    karteID: string,
    feld: string,
    bereich: number,
    weg: number[],
    bekannt: readonly string[] | null,
  ) => Promise<SchreibAntwort>
}
