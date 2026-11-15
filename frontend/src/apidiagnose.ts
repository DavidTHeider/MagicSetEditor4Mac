// **Die Diagnose** — Werkstück 4 von Phase 3.5
// (#1166).
//
// **Eine eigene Datei**, wie `apikarten.ts`, und aus demselben Grund: `api.ts`
// führt die ganze Fassade und liefe sonst über die Zeilengrenze (§2).

import type { DiagnoseAntwort } from './antworten'
import type { Umschlag } from './umschlag'

// Diagnoseapi sind die zwei Wege zum Text für einen Fehlerbericht.
//
// **Zwei und nicht einer.** Anzeigen und Sichern sind zwei Handlungen, und
// dazwischen steht die Zustimmung: der Text wird gelesen, bevor er die
// Anwendung verlässt.
export interface Diagnoseapi {
  // Diagnose gibt den Text — **ohne Terminal**.
  //
  // Er trägt Fassung, Rechnerart, den Ort des Bestands und das Ende des
  // Protokolls. Keine Kartentexte, keine Feldwerte, keine Setinhalte; das
  // Benutzerverzeichnis erscheint als `~`.
  Diagnose: () => Promise<DiagnoseAntwort>

  // DiagnoseSichern schreibt sie in eine Datei, die jemand wählt.
  //
  // **Ein Dialog und kein fester Ort**: die Datei gehört dem, der sie
  // weitergibt.
  DiagnoseSichern: () => Promise<Umschlag>
}
