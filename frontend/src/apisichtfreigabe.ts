// **Freigegebene Sammlungssichten** — Werkstück T3 von Phase 5.6
// (#1468).
//
// # Warum „Sichtfreigabe" und nicht „Freigabe"
//
// **Weil das Wort in diesem Programm schon dreimal vergeben ist**: die
// Freigabe eines **Profils** an einen zweiten Nutzer dieses Rechners
// (`apifreigaben.ts`, Phase 4.12), die Freigabe eines **Kontexts** an ein
// Modell und die Freigabe einer **Partie** für Gäste. Ein vierter Gebrauch
// unter demselben Namen wäre die Verwechslung, die man später nicht mehr
// findet.
//
// # Und warum eine Sichtkarte hier besonders ist
//
// **Sie ist die einzige Gestalt in diesem Fenster, die auch über eine Leitung
// geht.** Wer wissen will, was ein Gegenüber zu sehen bekommt, liest sie —
// und was hier nicht steht, geht nicht hinaus.
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Sichtfreigabe ist eine erteilte Sicht auf die eigene Sammlung.
export interface Sichtfreigabe {
  // empfaenger ist die Peerkennung des Gegenübers.
  empfaenger: string

  // umfang ist `auswahl`, `dubletten`, `wuensche`, `sammlung` oder
  // `digitales` — angeboten wird, was `umfaenge` nennt.
  umfang: string

  // zweck ist `teilen`, `handel` oder `partie` — **nie `paarung`**.
  zweck: string

  // auswahl sind die gewählten Druckkennungen bei `auswahl`.
  auswahl: string[]

  // behalten ist die Behaltemenge bei `dubletten`.
  behalten: number

  // ablauf ist RFC 3339 — **leer heißt unbefristet**.
  ablauf: string

  erteilt: string
}

// Sichtfreigabenantwort ist die Liste samt den möglichen Umfängen.
export interface Sichtfreigabenantwort extends Umschlag {
  freigaben: Sichtfreigabe[]

  // umfaenge sind die, die es überhaupt gibt.
  //
  // **Sie kommen aus Go und werden hier nicht nachgebaut.** Eine zweite
  // Aufzählung liefe beim sechsten Umfang auseinander.
  umfaenge: string[]

  // nachladefehler heißt: erteilt oder widerrufen **ist**, nur die Liste
  // danach war nicht zu lesen (§7, Regel 3; #1970). Der Kasten liest selbst
  // nach.
  nachladefehler?: FehlerDTO
}

// Sichtkarte ist eine Zeile, wie ein Gegenüber sie sähe.
//
// **Sechs Felder und eine Referenz** — und keines mehr. Kein Erwerbspreis,
// keine Notiz, kein Lagerort, keine Fassung, keine Bestandskennung.
export interface Sichtkarte {
  // referenz ist der **sitzungsgebundene** Name dieser Zeile — nicht ihre
  // Kennung im Bestand.
  referenz: string

  printing: string
  menge: number
  zustand: string
  sprache: string
  finish: string
  bezeichnung: string

  // wunsch heißt: gesucht und nicht angeboten — P11 (#2379). **Nur bei einem
  // Wunsch da**: eine Sicht einer älteren Fassung kennt das Feld nicht.
  wunsch?: boolean
}

// Sichtvorschauantwort ist die Vorschau samt ihrem Fehler.
export interface Sichtvorschauantwort extends Umschlag {
  zeilen: Sichtkarte[]
}

// Sichtfreigabeapi sind die vier Wege zu den erteilten Sichten.
//
// **Vier und nicht einer**: lesen sagt, was gilt; erteilen gibt etwas frei;
// widerrufen nimmt es zurück; und die Vorschau zeigt, was ein Gegenüber davon
// wirklich sähe.
export interface Sichtfreigabeapi {
  Sammlungsfreigaben: (kennung: string) => Promise<Sichtfreigabenantwort>
  SammlungFreigeben: (freigabe: Sichtfreigabe) => Promise<Sichtfreigabenantwort>
  SammlungsfreigabeWiderrufen: (
    kennung: string,
    zweck: string,
    umfang: string,
  ) => Promise<Sichtfreigabenantwort>
  Freigabesicht: (
    kennung: string,
    zweck: string,
  ) => Promise<Sichtvorschauantwort>
}
