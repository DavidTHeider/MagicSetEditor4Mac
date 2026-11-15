// **Wie eine Karte aussieht, bevor sie es ist.**
//
// **Eine eigene Datei**, weil `api.ts` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was man mit einer Karte tut**, hier,
// **wie man sieht, was daraus würde** — ohne dass jemand etwas bestätigt (§6).
import type { VorschauAntwort } from './antworten'

export interface Vorschauapi {
  Vorschau: (
    karteID: string,
    vergroesserung: number,
  ) => Promise<VorschauAntwort>

  // Probevorschau zeichnet die Karte, als stünde in einem Feld schon der neue
  // Wert — **ohne ihn zu schreiben** (#288). Damit ist eine Änderung in der
  // Vorschau zu sehen, ohne dass jemand etwas bestätigt (§6), und der Verlauf
  // bekommt trotzdem nicht je Tastendruck einen Schritt.
  Probevorschau: (
    karteID: string,
    feld: string,
    wert: string,
    vergroesserung: number,
  ) => Promise<VorschauAntwort>

  // Stilprobevorschau zeichnet die Karte, als stünde eine Stiloption schon auf
  // dem neuen Wert — **ohne sie zu schreiben**
  // (#590).
  // Dieselbe Begründung wie eine Zeile höher, für die andere Hälfte von §6:
  // eine Auswahl schreibt beim Umstellen, ein Textfeld erst beim Verlassen.
  Stilprobevorschau: (
    karteID: string,
    name: string,
    wert: string,
    vergroesserung: number,
  ) => Promise<VorschauAntwort>
}
