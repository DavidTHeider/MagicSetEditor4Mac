// **Die begrenzten Metriken** — Werkstück O9 von Phase 5.2
// (#1433),
// seit OPS-05 (#1722)
// mit Laufzeiten, Verbrauch und dem Leistungsprofil.
// Die Gestalt kommt aus `internal/app/metriken.go` und
// `internal/app/metrikenbetrieb.go`.
import type { Umschlag } from './umschlag'

// MessreiheDTO ist eine Zeile des Zählwerks.
//
// **Drei Merkmale, und alle drei sind Aufzählungen** — Fläche, Wirkungsklasse,
// Ausgang. Keine Karten-, Set- oder Profilkennung: davon gäbe es beliebig
// viele, und die Tabelle wüchse mit dem Bestand.
export interface MessreiheDTO {
  flaeche: string
  wirkung: string
  ausgang: string
  anzahl: number
}

// LaufzeitreiheDTO ist die Verteilung für eine Fläche und Art — `eimer`
// zählt je Grenze aus `eimerMs` und einen offenen Eimer darüber.
export interface LaufzeitreiheDTO {
  flaeche: string
  art: string
  anzahl: number
  abgewiesen: number
  gescheitert: number
  summeMs: number
  eimer: number[]
}

// SpurpunktDTO ist ein Eintrag der Ablaufspur — der Weg als Muster, nie der
// rohe Pfad.
export interface SpurpunktDTO {
  zeit: string
  vorgang: string
  flaeche: string
  weg: string
  art: string
  dauerMs: number
  ausgang: string
}

// SpurDTO sagt, was die Spur hält und wie vollständig es ist.
export interface SpurDTO {
  an: boolean
  abtastung: number
  grenze: number
  aufbewahrungS: number
  gesehen: number
  aufgenommen: number
  verdraengt: number
  abgelaufen: number
  punkte: SpurpunktDTO[]
}

// SpeicherDTO sind die Zahlen der Zwischenspeicher fremder Dienste.
export interface SpeicherDTO {
  treffer: number
  geholt: number
  veraltet: number
  fehlgeschlagen: number
}

// DatenbankDTO sind die Zahlen der Verbindung zum Bestand.
export interface DatenbankDTO {
  offen: number
  inBenutzung: number
  frei: number
  gewartet: number
  wartezeitMs: number
}

// WebprozessDTO ist ein WebKit-Prozess dieser Anwendung (OPS02, #2348).
export interface WebprozessDTO {
  art: 'webinhalt' | 'gpu' | 'netz'
  fussabdruckBytes: number
}

// ProzessDTO ist der Verbrauch dieses Prozesses.
//
// **Der Fußabdruck ist die Zahl der Aktivitätsanzeige** (OPS02, #2348), der
// Heap nur ein Teil davon; `web` ist leer am Dienst und am Werkzeugserver,
// und `webUngemessen` sagt, warum die WebKit-Prozesse nicht gemessen sind.
export interface ProzessDTO {
  heapBytes: number
  systemBytes: number
  goroutinen: number
  cpuMs: number
  fussabdruckBytes: number
  web: WebprozessDTO[]
  webUngemessen: string
}

// MetrikenAntwort trägt den Stand dieses Prozesses.
export interface MetrikenAntwort extends Umschlag {
  reihen: MessreiheDTO[]

  // grenze ist die Zahl der Reihen, die es überhaupt geben kann.
  grenze: number

  laufzeiten: LaufzeitreiheDTO[]
  eimerMs: number[]
  spur: SpurDTO
  speicher: SpeicherDTO

  // datenbank ist `null`, wo dieser Prozess keinen Bestand hält.
  datenbank: DatenbankDTO | null
  prozess: ProzessDTO

  // seit und bis sind der Erhebungszeitraum, aufwandMikros was das Erheben
  // gekostet hat, ungemessen was fehlt (OPS-05.AK-04).
  seit: string
  bis: string
  aufwandMikros: number
  ungemessen: string[]
}

// LeistungsprofilAntwort sagt, wo das Profil liegt.
export interface LeistungsprofilAntwort extends Umschlag {
  datei: string
  sekunden: number
  bytes: number
}

export interface Metrikenapi {
  Metriken: () => Promise<MetrikenAntwort>
  Leistungsprofil: (sekunden: number) => Promise<LeistungsprofilAntwort>
}
