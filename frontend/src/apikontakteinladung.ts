// **Einladen, prüfen, annehmen** — NET-01
// (#1705).
//
// **Eine eigene Datei neben `apipeerkontakt.ts`**, weil sie eine andere Frage
// beantwortet: dort, wen dieses Profil kennt — hier, wie jemand dazukommt.
import type { Peerkontakteantwort } from './apipeerkontakt'
import type { Umschlag } from './umschlag'

// Kontakteinladung ist eine ausgestellte Einladung.
export interface Kontakteinladung {
  code: string

  // ablauf ist RFC 3339.
  ablauf: string

  // fingerabdruck ist der eigene — für die technische Angabe.
  fingerabdruck: string
}

// Kontakteinladungsantwort ist die Einladung samt ihrem Fehler.
export interface Kontakteinladungsantwort extends Umschlag {
  einladung: Kontakteinladung
}

// Einladungsvorschau ist, was eine Einladung sagt, bevor jemand annimmt.
export interface Einladungsvorschau {
  kennung: string
  name: string
  zweck: string
  adresse: string
  ablauf: string
  fingerabdruck: string

  // vergleichscode ist, was beide Seiten laut vergleichen.
  vergleichscode: string

  // lage ist die eines vorhandenen Kontakts — `unbekannt`, wenn es keinen
  // gibt.
  lage: string

  // vorgaenger ist der Name des Kontakts, dessen neuer Schlüssel diese
  // Einladung sein will — leer, wenn es ihn hier nicht gibt.
  vorgaenger: string

  angenommen: boolean

  // einloesbar sagt, dass sich dieses Profil nach der Annahme beim
  // Einladenden melden wird (NET-01b) — dann wird nach dem Namen gefragt,
  // der hinausgeht.
  einloesbar: boolean
}

// Einladungsvorschauantwort ist die Vorschau samt ihrem Fehler.
export interface Einladungsvorschauantwort extends Umschlag {
  vorschau: Einladungsvorschau
}

// Einladungsapi sind die drei Wege.
export interface Einladungsapi {
  KontakteinladungErstellen: (
    name: string,
    zweck: string,
  ) => Promise<Kontakteinladungsantwort>
  KontakteinladungPruefen: (code: string) => Promise<Einladungsvorschauantwort>
  KontakteinladungAnnehmen: (
    code: string,
    eigenerName: string,
    verglichen: string,
    freigeben: boolean,
    meinName: string,
  ) => Promise<Peerkontakteantwort>
}
