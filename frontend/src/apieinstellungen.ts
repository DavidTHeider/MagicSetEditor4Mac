// **Was die Anwendung sich über Läufe hinweg merkt** — die Einstellungen und
// die Sprache der Oberfläche.
//
// Eigene Datei, weil `api.ts` mit ihr über 400 Zeilen ginge (§2). Der Schnitt
// folgt der Frage und nicht der Zeilenzahl: die übrigen Bindungen dort
// arbeiten an einem **Set** oder einer **Karte** und leben mit ihr; diese drei
// überdauern beide.
//
// **Die Sprache steht hier und nicht bei den Karten**, obwohl sie über eine
// eigene Bindung kommt: sie ist eine Einstellung wie jede andere, nur eine,
// die auch die nativen Teile brauchen — Menüleiste und Dateidialog entstehen
// in Go.
import type { Darstellungsstand } from './darstellungsstand'
import type { Objekthandlungen } from './objekthandlungen'
import type { Umschlag } from './umschlag'
import type { EinstellungenAntwort } from './einstellungen'
import type { DienstkatalogAntwort } from './dienstkatalog'
import type { GrenzenAntwort } from './grenzen'
import type {
  EinfuhrAntwort,
  EinfuhrvorschauAntwort,
} from './einstellungsaustausch'

export interface Einstellungsapi {
  Einstellungen: () => Promise<EinstellungenAntwort>
  Dienstkatalog: () => Promise<DienstkatalogAntwort>

  // **Die Grenzen dieses Prozesses** (#1429) — dieselbe Auskunft wie über
  // MCP, auf der Kommandozeile und im Dienst.
  Grenzen: () => Promise<GrenzenAntwort>
  EinstellungSetzen: (schluessel: string, wert: string) => Promise<Umschlag>

  // SpracheSetzen meldet dem Go-Teil, in welcher Sprache die Oberfläche
  // gerade steht.
  SpracheSetzen: (sprache: string) => Promise<Umschlag>

  // DarstellungsstandMelden sagt der nativen Menüleiste, wie Seitenleiste,
  // Navigator und Inspector dastehen — UXR-F14 (#2029). **Nach `SpracheSetzen`
  // gebaut**, aus demselben Grund: die Leiste entsteht in Go, und was sie
  // zeigt, weiß nur das Fenster.
  DarstellungsstandMelden: (stand: Darstellungsstand) => Promise<Umschlag>

  // ObjekthandlungenMelden sagt der Menüleiste, welche Handlungen das
  // gewählte Objekt hat — GESTE (#2358). Dieselbe Liste wie im Rechtsklick
  // und im Überlauf; ein Klick kommt als `ereignisObjekthandlung` zurück.
  ObjekthandlungenMelden: (meldung: Objekthandlungen) => Promise<Umschlag>

  // FensterfehlerMelden schreibt einen Fehler beim Zeichnen mit seiner
  // Kennung ins Protokoll — F01 (#2340). Die Fehlerfläche nennt dieselbe
  // Kennung; ohne diese Zeile gäbe es zu ihr nichts zu finden.
  FensterfehlerMelden: (fehler: Fensterfehler) => Promise<Umschlag>

  // **Der Austausch** — Werkstück O3 von Phase 5.2
  // (#1427).
  //
  // **`EinstellungenSichern` fragt selbst nach dem Ziel**, und
  // `EinstellungenEinlesen` selbst nach der Quelle: beides sind native
  // Dialoge, und ein nachgebauter wäre eine schlechtere Kopie (§6).
  EinstellungenSichern: (abschnitte: string[]) => Promise<Umschlag>
  EinstellungenEinlesen: () => Promise<EinfuhrvorschauAntwort>

  // EinstellungenUebernehmen nimmt den Text, die **Revision** aus der
  // Vorschau und die gewählten Zeilen als »ebene:schluessel«.
  EinstellungenUebernehmen: (
    inhalt: string,
    revision: string,
    nur: string[],
  ) => Promise<EinfuhrAntwort>
}

// Fensterfehler ist ein Fehler, den eine Fehlergrenze gefangen hat (F01).
export interface Fensterfehler {
  kennung: string
  ebene: 'bereich' | 'wurzel' | 'ruf' | 'meldung'
  bereich: string
  meldung: string
}
