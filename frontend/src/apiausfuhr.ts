// **Was ein Set verlässt** — Exportvorlagen, MTG JSON, der Druckbogen.
//
// **Eine eigene Datei**, wie `apibestand.ts` und `apiaktualisierung.ts`, und
// aus demselben Grund: `api.ts` führt die ganze Fassade und liefe sonst über
// die Zeilengrenze (§2).
//
// # Der Schnitt folgt der Frage
//
// Die vier Wege haben dieselbe Gestalt: ein Set geht in eine Datei, die
// jemand anderes liest — ein HTML aus einer Vorlage, ein MTG JSON, ein PDF.
// Was sie unterscheidet, ist das Format am anderen Ende.

import type { AusfuhrAntwort, BogenAntwort, OeffnenAntwort } from './antworten'
import type {
  BogenprobelaufAntwort,
  VorlagenprobelaufAntwort,
} from './antwortenprobelauf'
import type { SchreibAntwort } from './api'
import type { VorlagenDTO, VorlagenausfuhrDTO } from './vorlagen'

// Ausfuhrapi ist der Weg aus einem Set heraus.
export interface Ausfuhrapi {
  // ExportoptionSetzen merkt eine Wahl an einer Exportvorlage.
  ExportoptionSetzen: (
    vorlage: string,
    feld: string,
    wert: string,
  ) => Promise<SchreibAntwort>

  // Exportvorlagen nennt die Vorlagen, die zu diesem Set passen.
  Exportvorlagen: (setID: string) => Promise<VorlagenDTO>

  // VorlageAusfuehren schreibt das Set über eine Vorlage.
  VorlageAusfuehren: (
    setID: string,
    paket: string,
    dateityp: string,
    wahl: Record<string, string>,

    // auswahl sind die Kartenkennungen — **leer heißt: das ganze Set**
    // (#1167).
    auswahl: string[],
  ) => Promise<VorlagenausfuhrDTO>

  // VorlageProbelauf wertet eine Vorlage aus, ohne zu schreiben — UXR-F7
  // (#2015).
  // **Dieselben Wahlen wie `VorlageAusfuehren`** und kein Dialog.
  VorlageProbelauf: (
    setID: string,
    paket: string,
    wahl: Record<string, string>,
    auswahl: string[],
  ) => Promise<VorlagenprobelaufAntwort>

  // Die drei Wege über MTG JSON — das Kartenformat, das
  // `magic-set-generator` liest (§11).
  MTGJSONOeffnen: (spiel: string) => Promise<OeffnenAntwort>
  MTGJSONVerzeichnisOeffnen: (spiel: string) => Promise<OeffnenAntwort>
  MTGJSONSpiele: () => Promise<string[]>
  MTGJSONSchreiben: (setID: string, spiel: string) => Promise<AusfuhrAntwort>

  // BogenLegen setzt die Karten eines Sets auf Druckseiten.
  BogenLegen: (
    setID: string,
    papier: string,
    anschnitt: number,
    dpi: number,
    marken: boolean,

    // auswahl sind die Kartenkennungen — **leer heißt: das ganze Set** (#1167).
    // MSE2 stellt die Frage vor **vier** seiner sieben Ausgabewege —
    // Drucken, Druckvorschau, HTML und mehrere Bilder
    // (`ExportCardSelectionType`, `SetWindow::selectionChoices`).
    auswahl: string[],
  ) => Promise<BogenAntwort>

  // BogenProbelauf rechnet den Bogen durch, ohne ihn zu legen — UXR-F7
  // (#2015). **Dieselben Wahlen wie `BogenLegen`** und kein Dialog.
  BogenProbelauf: (
    setID: string,
    papier: string,
    anschnitt: number,
    dpi: number,
    marken: boolean,
    auswahl: string[],
  ) => Promise<BogenprobelaufAntwort>
}
