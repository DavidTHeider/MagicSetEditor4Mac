// Die **Schlüsselwörter eines Sets**
// (#1062).
//
// **Eine eigene Datei, wie `apibilder.ts` und `apiscryfall.ts`** — und aus
// demselben Grund: `api.ts` führt die ganze Fassade und liefe sonst über die
// Zeilengrenze (§2).

import type { SchreibAntwort } from './api'
import type {
  SchluesselwoerterAntwort,
  SchluesselwortAntwort,
} from './antworten'

// Schluesselwortapi ist der Weg zu den Schlüsselwörtern eines Sets.
export interface Schluesselwortapi {
  // Schluesselwoerter gibt die des Sets **und** die des Spiels.
  //
  // **Beide in einer Liste**, wie MSE sie zeigt: wer nur die eigenen sähe,
  // wüsste nicht, welche das Spiel schon mitbringt — `magic.mse-game` führt
  // 334 — und schriebe ein zweites »Fliegend«.
  Schluesselwoerter: (setID: string) => Promise<SchluesselwoerterAntwort>

  // SchluesselwortAnlegen hängt ein leeres ans Ende und gibt seine Kennung.
  //
  // **Leer in jedem Feld außer dem Modus** — so legt MSE eines an
  // (`AddKeywordAction`, `src/data/action/keyword.cpp`, Zeilen 20–32, Fassung
  // f9d9356 vom 2021-07-06).
  SchluesselwortAnlegen: (setID: string) => Promise<SchluesselwortAntwort>

  // SchluesselwortSetzen schreibt ein Feld eines Schlüsselworts des **Sets**.
  //
  // Die des Spiels tragen keine Kennung; der Weg dorthin ist damit gar nicht
  // erst gangbar. **`bekannt`** sind die Werte, die der Entwurf von diesem
  // Feld kennt — steht ein anderer im Bestand, ist es ein Konflikt (U02b,
  // #2179); `null` schreibt ohne Bedingung.
  SchluesselwortSetzen: (
    setID: string,
    wortID: string,
    feld: string,
    wert: string,
    bekannt: readonly string[] | null,
  ) => Promise<SchreibAntwort>

  SchluesselwortLoeschen: (
    setID: string,
    wortID: string,
  ) => Promise<SchreibAntwort>
}
