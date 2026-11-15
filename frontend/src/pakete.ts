// Die Typen des Paketkatalogs (#54).
//
// **Eine eigene Datei**, weil `api.ts` an der Zeilengrenze steht (§2) — und
// weil es dieselbe Trennung ist wie bei `vorlagen.ts`: die Fassade nennt die
// Methoden, die Typen stehen bei ihrem Gegenstand.
import type { Umschlag } from './api'

// PaketDTO ist ein installiertes Vorlagenpaket.
export interface PaketDTO {
  art: string
  name: string

  // anzeigename ist, was über dem Paket steht — der Kurzname, großgeschrieben
  // wie MSE2; ohne ihn der Name (UX-05, S07.NAT-02, #1677). **Er kommt fertig
  // aus Go** (`domain.Paket.Anzeigename`): die Regel ist portiert, und eine
  // zweite Fassung hier liefe mit ihr auseinander.
  anzeigename: string
  version: string
  pfad: string
  herkunft: string

  // pruefsumme ist, was beim Bezug herauskam — leer bei einem Paket, das
  // niemand bezogen hat.
  pruefsumme: string

  // braucht sind die Abhängigkeiten als fertige Zeilen — „paket version".
  braucht: string[]

  // gruppe ist `installer group:` — der Pfad, unter dem die Verwaltung dieses
  // Paket zeigt, mit `/` als Trenner (#244).
  gruppe: string

  // rang ist `position hint:` — die Ordnung innerhalb einer Gruppe. Ohne
  // Angabe eine Million, also hinter allem.
  rang: number

  // artrang ist MSEs `PackageType` — die Reihenfolge der sieben Arten.
  artrang: number

  // **Die Übersicht** — DOM-01.AK-01 (#1696): woher das Paket stammt
  // (`mitgeliefert`, `bezogen`, `uebernommen`, `eigen`), ob es seit dem
  // Einbinden lokal geändert wurde, wer es braucht und wie viele Sets dieses
  // Profils es nennen — gezählt nur bei Spielen und Stilen.
  herkunftsart: Paketherkunftsart
  geaendert: boolean
  gebrauchtVon: string[]
  benutzt: number

  // hauptdatei ist die Datei, die das Paket ausmacht — `style`, `game` …;
  // sie kommt aus Go (`msepaket.Art.Hauptdatei`), FMT-02 (#1691).
  hauptdatei: string
}

// Paketherkunftsart sagt, woher ein Paket stammt.
export type Paketherkunftsart =
  'mitgeliefert' | 'bezogen' | 'uebernommen' | 'eigen'

// PaketplanDTO sagt **vor** einer Änderung, was sie bewirkt — DOM-01.AK-02.
//
// `vorgang` ist `entfernen`, `einbinden`, `ersetzen` (ein Einbinden über ein
// vorhandenes Paket) oder `zurueck`. `quelle` trägt den Ort, aus dem
// eingebunden oder ersetzt würde; leer beim Entfernen und beim Zurück.
export interface PaketplanDTO {
  vorgang: Paketvorgang
  paket: string
  quelle: string
  von: string
  zu: string
  geaendert: boolean
  moeglich: boolean
  sets: string[]
  konflikte: AbhaengigkeitDTO[]
}

// Paketvorgang ist, was ein Plan plant.
export type Paketvorgang = 'entfernen' | 'einbinden' | 'ersetzen' | 'zurueck'

// PaketplanAntwort trägt einen Plan — oder einen Abbruch am Dialog.
export interface PaketplanAntwort extends Umschlag {
  plan: PaketplanDTO | null
  abgebrochen: boolean
}

// QuelleDTO ist eine Bezugsquelle, die die Anwendung anbietet (#263).
//
// **Angaben, kein Satz** (§9): `schluessel` benennt den erklärenden Satz in den
// Sprachkatalogen.
export interface QuelleDTO {
  name: string
  adresse: string
  stand: string
  unterverzeichnis: string
  schluessel: string

  // megabyte ist die Größe des **Repositoriums**, wie GitHub sie meldet —
  // nicht die des Downloads; den meldet `codeload` nicht.
  megabyte: number
}

// BezugsquellenAntwort trägt die angebotenen Quellen.
export interface BezugsquellenAntwort extends Umschlag {
  quellen: QuelleDTO[]
}

// PaketeAntwort trägt den Katalog samt Befunden.
export interface PaketeAntwort extends Umschlag {
  pakete: PaketDTO[]
  befunde: string[]
}

// PaketAntwort trägt ein einzelnes Paket — oder einen Abbruch.
export interface PaketAntwort extends Umschlag {
  paket: PaketDTO | null
  abgebrochen: boolean
}

// OrteAntwort nennt die vorgeschlagenen MSE2-Verzeichnisse.
//
// **Nur die, die es gibt** — und gelesen wird keines davon, bevor jemand es
// wählt (#55).
export interface OrteAntwort {
  orte: string[]
}

// UebernahmeAntwort ist der Ausgang einer Übernahme aus einer MSE2-Installation.
// Paketweitergabe sagt, wohin ein Paket als `.mse-installer` hinausging —
// FMT-02.AK-04 (#1691).
export interface Paketweitergabe extends Umschlag {
  datei: string
  abgebrochen: boolean
}

export interface UebernahmeAntwort extends Umschlag {
  uebernommen: PaketDTO[]

  // uebergangen sind die Pakete, die schon dalagen — übergangen und nicht
  // überschrieben.
  uebergangen: string[]
  befunde: string[]
  abgebrochen: boolean
}

// AbhaengigkeitDTO ist ein Vorbehalt gegen eine Abhängigkeit (#57).
//
// **`art` ist ein Schlüssel, kein Satz** (§9): `fehlt`, `zualt`, `unlesbar`
// oder `zyklus`. Den Satz baut die Oberfläche in ihrer Sprache.
export interface AbhaengigkeitDTO {
  paket: string
  braucht: string
  art: string
  verlangt: string
  da: string
}

export interface AbhaengigkeitenAntwort extends Umschlag {
  befunde: AbhaengigkeitDTO[]
}

// Anlegbarkeit ist eine Paketart, die sich anlegen lässt
// (#1350).
//
// **Aus der Anwendung und nicht aus einer Liste hier.** Zwei von sieben
// Paketarten lassen sich leer anlegen, und welche das sind, entscheidet
// `internal/service` — eine Aufzählung im Frontend wäre eine zweite Wahrheit.
export interface Anlegbarkeit {
  art: string

  // brauchtSpiel sagt, ob ein Spielname mitzugeben ist: ein Stil legt die
  // Felder **seines** Spiels aus und kann ohne es nichts.
  brauchtSpiel: boolean
}
