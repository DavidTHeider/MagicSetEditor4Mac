// Die Typen der Exportvorlagen.
//
// **Eine eigene Datei**, weil `api.ts` sonst über die Zeilengrenze wüchse
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort die Fassade und ihre
// Methoden, hier die Gestalt dessen, was über sie kommt
// (#209).
import type { FehlerDTO } from './api'

// ExportoptionDTO ist eine Wahlmöglichkeit einer Exportvorlage.
export interface ExportoptionDTO {
  readonly art: string
  readonly name: string
  readonly beschreibung: string
  readonly vorgabe: string
  readonly auswahl: readonly string[]

  // wert ist der **behaltene** Wert — sonst die Vorgabe
  // (#492).
  //
  // **Er steht an der Vorlage und kommt nicht aus einem zweiten Aufruf**: eine
  // Vorlage und ihre Werte gehören zusammen, und zwei Antworten liefen beim
  // nächsten Wechsel auseinander.
  readonly wert: string
}

// ExportvorlageDTO ist eine Exportvorlage, wie die Oberfläche sie zeigt.
export interface ExportvorlageDTO {
  readonly paket: string
  readonly name: string
  readonly dateityp: string
  readonly optionen: readonly ExportoptionDTO[]
}

// VorlagenDTO ist die Liste der Vorlagen eines Sets.
export interface VorlagenDTO {
  readonly vorlagen: readonly ExportvorlageDTO[]
  readonly befunde: readonly string[]
  readonly fehler?: FehlerDTO | null
}

// VorlagenausfuhrDTO ist das Ergebnis eines Vorlagenlaufs.
export interface VorlagenausfuhrDTO {
  readonly pfad: string
  readonly befunde: readonly string[]
  readonly abgebrochen: boolean
  readonly fehler?: FehlerDTO | null
}
