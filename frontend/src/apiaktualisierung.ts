// **Nach einer neueren Fassung sehen, und sie holen**
// (#1100).
//
// **Eine eigene Datei**, wie `apibestand.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).

import type { FassungsAntwort, LadeAntwort } from './antwortenbestand'

// Aktualisierungsapi ist der Weg zu den Freigaben dieses Programms.
export interface Aktualisierungsapi {
  // NachNeuerFassungSehen fragt nach einer neueren Fassung.
  //
  // **Ohne Parameter, und die Stufe entscheidet Go.** Ob überhaupt gefragt
  // wird, hängt an der Einstellung — das Fenster trifft die Entscheidung
  // nicht, es zeigt sie (§2).
  NachNeuerFassungSehen: () => Promise<FassungsAntwort>

  // NeueFassungLaden holt die Freigabe ins Downloadverzeichnis.
  //
  // **Der Fortschritt kommt nicht hierüber**, sondern über das Ereignis
  // `aktualisierung:fortschritt` (`fenster.ts`): ein Abbild von 188 MB dauert
  // länger als ein Aufruf.
  NeueFassungLaden: () => Promise<LadeAntwort>
}
