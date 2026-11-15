// **Was die KI zurückgibt** — herausgezogen für Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum es diese Datei gibt
//
// **Weil `antworten.ts` die eine Anlaufstelle bleibt und trotzdem unter der
// Zeilengrenze aus §2.** Die KI-Gestalten liegen in zwei Dateien — den
// Textvorschlag bei den Märkten und Anbietern, das Bild in `antwortenbild.ts`
// —, und sie hier zu bündeln macht aus zwölf Zeilen dort eine.
export type {
  KIBelegDTO,
  KIGliedDTO,
  KICloudDTO,
  KIKatalogantwort,
  KIFeldwertDTO,
  KIVorschlagAntwort,
  KIProbeAntwort,
  KIProbeDTO,
  KIUebernahmeAntwort,
  KIModellDTO,
  KIFeldherkunftDTO,
  KIFeldherkunftAntwort,
} from './antwortenmarkt'

export type {
  KIBildangabeDTO,
  KIBildvorschlagAntwort,
  KIBildherkunftAntwort,
} from './antwortenbild'
