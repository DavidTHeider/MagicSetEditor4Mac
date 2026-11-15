// **Die Antworten rund um die Bestandsdatei** — sichern, zurückspielen, und
// was beim Start geschah
// (#940,
// #962).
//
// **Eine eigene Datei**, weil `antworten.ts` über die Zeilengrenze wuchs (§2).
// Der Schnitt folgt der Frage: dort die Antworten **im** Bestand, hier die
// über den Bestand selbst.

import type { AbhaengigkeitDTO } from './pakete'
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Sicherungsantwort ist der Ausgang einer Bestandssicherung
// (#799).
//
// **Der Pfad kommt mit.** Wer sichert, will wissen, wohin — und bei einem
// Abbruch ist er leer, was „nichts getan" von „getan" unterscheidet.
export interface Sicherungsantwort extends Umschlag {
  pfad: string
}

// Sicherungsstand ist das Alter der letzten Sicherung — OPS01
// (#2347).
// **`gesichert` ist leer, wenn es keinen Vermerk gibt**: unbekannt heißt
// unbekannt, und das Fenster erfindet keine Zeit.
export interface Sicherungsstand {
  stufe: 'unbekannt' | 'frisch' | 'alt' | 'unstimmig'
  gesichert: string
  tage: number
  anlass: string
  flaeche: string
  datei: string
  verschluesselt: boolean
  plan: Sicherungsplanstand
}

// Sicherungsplanstand ist der freiwillige Plan — `ordner` leer heißt: keiner.
export interface Sicherungsplanstand {
  ordner: string
  letzter: string
  faellig: boolean
}

// SicherungsstandAntwort trägt den Stand und, woran der Plan zuletzt
// scheiterte — nur das Fenster führt ihn aus.
export interface SicherungsstandAntwort extends Umschlag {
  stand: Sicherungsstand
  planfehler?: FehlerDTO | null
}

// Sicherungsauswahl ist, was das Fenster über eine gewählte Sicherung erfährt
// (#940).
//
// **Vor der Übernahme**, damit jemand entscheiden kann. `sets` und `karten`
// sind **−1**, wenn die Sicherung die Tabelle nicht kennt — das ist etwas
// anderes als 0.
export interface Sicherungsauswahl extends Umschlag {
  pfad: string
  angelegtAm: string
  schemastand: string
  sets: number
  karten: number
  bilder: Bilderlage
  abgebrochen: boolean

  // verschluesselt sagt, ob die Sicherung ein Kennwort trägt — auch dann,
  // wenn es fehlte und `fehler` deshalb gesetzt ist (OPS-06, #1723).
  verschluesselt: boolean

  // paketeEnthalten sind die eigenen Pakete darin, paketeFehlen die
  // bezogenen, die hier nicht liegen und neu zu beziehen sind (OPS-06).
  paketeEnthalten: string[]
  paketeFehlen: string[]
}

// Bilderlage sagt, was beim Zurückspielen mit den eingesetzten Bildern
// geschieht (#1162).
//
// **Drei Fälle und nicht zwei**: „keine dabei“ heißt bei einem Archiv etwas
// anderes als bei einer nackten `.db` aus 0.1.0 oder 0.2.0. Das eine
// beschreibt einen Stand ohne Bilder, das andere sagt über sie **nichts** —
// und deshalb wird einmal geleert und einmal stehengelassen.
export type Bilderlage = 'ersetzt' | 'geleert' | 'unberuehrt'

// Rueckspielantwort sagt, ob der Bestand ersetzt wurde — und wohin der alte kam.
//
// **Und nach einem Fehler, ob der bisherige Stand wieder ganz dasteht** —
// `unveraendert`, die ausdrückliche Rückkehrmeldung aus A05 (#2219).
export interface Rueckspielantwort extends Umschlag {
  ersetzt: boolean
  beiseite: string
  unveraendert: boolean

  // bestandsepoche ist die Epoche des Bestands danach (J05, #2269): eine
  // Kennung aus einer älteren verwirft das Fenster.
  bestandsepoche?: number
}

// Startbefund sagt, ob der Bestand offen ist (#962).
//
// **Kein `Umschlag`.** Er ist selbst die Auskunft über einen Fehler; ein
// Fehlerfeld daneben wäre die Frage „ist der Fehler fehlerhaft?".
export interface Startbefund {
  offen: boolean

  // grund ist die Meldung des gescheiterten Starts, oder leer.
  //
  // **Im Klartext und als Einzelheit** — seit Werkstück 3 von Phase 4.13
  // (#1352)
  // steht der Satz, den jemand liest, in `schluessel`.
  grund: string

  // schluessel ordnet den Grund einer Klasse zu — oder ist leer.
  schluessel: string

  // fassung ist die laufende Fassung dieser Anwendung (#1168).
  //
  // **Sie kommt mit dem Startbefund**, weil der ohnehin einmal beim Aufbau
  // gelesen wird und sich die Fassung zur Laufzeit nicht ändert.
  fassung: string

  // rueckspielrest sagt, dass ein abgebrochenes Zurückspielen beim Start nicht
  // abgeschlossen werden konnte — sonst leer (#1170).
  //
  // **Warum es hier steht**: nach einem solchen Rest läuft die Anwendung mit
  // der **neuen** Datenbank und den **alten** Bildern. Stünde das nur im
  // Protokoll, sagte das Fenster nichts davon.
  //
  // **Im Klartext**, wie `grund` und aus demselben Grund.
  rueckspielrest: string

  // rueckweg nennt die Sicherung, mit der diese Fassung wieder arbeiten kann
  // — oder ist `null`.
  //
  // **Weil der Rat sonst ins Leere zeigt** (Werkstück 5 von Phase 4.13,
  // #1352):
  // neben dem Bestand liegen bis zu dreißig Dateien, und welche aufgeht, weiß
  // nur, wer sie öffnet. Das hat Go getan.
  rueckweg: Rueckweg | null
}

// Rueckweg ist die Sicherung, die diese Fassung wieder öffnen kann.
//
// **Der Dateiname und nicht der Pfad** — ein Pfad aus dem Inneren ist genau
// das, was die Abnahme von Phase 4.13 verbietet.
export interface Rueckweg {
  datei: string
  schemastand: string
  sets: number
  karten: number
}

// SetAnlegeAntwort ist der Ausgang eines Anlegens (#891).
//
// **Ein eigener Verbund und nicht `AnlegeAntwort`**: jene nennt eine
// **Karte**, und ein Feld `karteID` mit einer Setkennung darin ist eine
// Falschangabe, die niemand beim Lesen bemerkt.
export interface SetAnlegeAntwort extends Umschlag {
  id: string
}

// SetLoeschAntwort ist der Ausgang eines Löschens
// (#1218).
//
// **Befunde ohne Fehler sind der Regelfall** (§7, Regel 3): das Set ist weg,
// und was daneben liegen blieb — eingesetzte Bilder auf der Platte — steht
// hier. Wer beides in ein Feld legte, machte aus „aufgeräumt bis auf die
// Bilder" ein „fehlgeschlagen".
export interface SetLoeschAntwort extends Umschlag {
  befunde: string[]
}

// FassungsAntwort ist, was das Fenster über eine neuere Fassung erfährt
// (#1100).
export interface FassungsAntwort extends Umschlag {
  // gefragt sagt, ob überhaupt gefragt wurde.
  //
  // **Getrennt von `neuer`**, weil „aus" und „nichts Neues" verschiedene
  // Dinge sind.
  gefragt: boolean
  neuer: boolean
  fassung: string
  seite: string
  // holen sagt, ob die Stufe das Laden erlaubt — Go entscheidet, das Fenster
  // zeigt den Knopf nur dann (AP-12, #2369).
  holen: boolean
}

// LadeAntwort ist der Ausgang eines Ladevorgangs (#1100, Stufe 2).
export interface LadeAntwort extends Umschlag {
  neuer: boolean
  fassung: string

  // pfad steht erst da, wenn die Prüfsumme stimmt.
  pfad: string
}

// Verschluesselungsantwort sagt, wie der Bestand auf der Platte liegt —
// Werkstück 14 von Phase 4.12
// (#1351).
//
// **Der Stand kommt auf jedem Weg zurück**, auch nach dem Umschalten (§7,
// Regel 3): sonst müsste die Seite nachladen, und zwischen Antwort und
// Nachladen läge ein Stand, den niemand kennt.
export interface Verschluesselungsantwort extends Umschlag {
  verschluesselt: boolean
}

// Setvorschau sagt, was das Einfahren einer Datei täte (FMT-03.AK-01,
// #1692).
//
// **`schonOffen` ist eine Kennung oder leer**: liegt dieselbe Datei schon im
// Bestand, gäbe das Einfahren dieses Set zurück.
export interface Setvorschau {
  pfad: string
  name: string
  spiel: string
  stil: string
  karten: number
  befunde: string[]
  schonOffen: string
  bedarf: AbhaengigkeitDTO[]
}

export interface SetvorschauAntwort extends Umschlag {
  vorschau: Setvorschau | null
  abgebrochen: boolean
}
