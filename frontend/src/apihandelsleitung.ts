// **Ein Handel von der Eröffnung bis zur Übergabe** — Werkstück T1 von Phase
// 5.6 (#1466),
// angeschlossen mit
// #1637.
//
// **Vier Wege für den Vertrag und drei für den Ausgang.** Eröffnen legt einen
// Vorgang an, Vorschlagen eine Revision, Schreiten bewegt den Zustand, Stand
// fragt. Schicken, Bestätigen und Wiederaufnahme gehören dem **Ausgang** —
// das ist eine andere Frage als der Vorgang.
import type { FehlerDTO } from './formen'
import type { Sichtvorschauantwort } from './apisichtfreigabe'
import type { Umschlag } from './umschlag'

// dieLeitungsschritte sind die neun Schritte, die über die Leitung gehen.
//
// **Ohne die drei des Abschlusses** — `verbuchen`, `abgleichen` und
// `bestaetigen` bewegen Bestand und Beleg und verlangen eine Freigabe; sie
// stehen in `apihandelsverbuchung.ts`.
export const dieLeitungsschritte = [
  'anbieten',
  'gegenangebot',
  'zustimmen',
  'vorbereiten',
  'uebergeben',
  'ablehnen',
  'abbrechen',
  'ablaufen',
  'klaeren',
] as const

// Handelsposition ist eine Zeile eines Angebots.
export interface Handelsposition {
  // von ist die Partei, deren Bestand die Position verlässt.
  von: string

  printing: string
  druckId: number
  menge: number

  // guete ist der Zustand des Exemplars — `NM`, `LP`, …
  guete: string
  sprache: string
  finish: string

  // abbild ist der Abdruck des Exemplarstands beim Anbieten.
  abbild: string

  // bewertung gilt nur mit bewertungGeteilt.
  bewertung: number
}

// Handelsvorgang ist **eine** Revision.
export interface Handelsvorgang {
  vorgang: string
  revision: number

  // eltern ist der Abdruck der vorigen Revision — leer in der ersten.
  eltern: string

  // abdruck ist die Kennung **dieser** Revision. Eine Zustimmung nennt ihn,
  // statt „ja" zu sagen.
  abdruck: string

  art: string
  initiator: string
  parteien: string[]

  // ich ist die Partei **dieses** Profils — die, deren Stimme `naechster`
  // nennt (#2026).
  // Leer, wo es keine Identität gibt. **Gesetzt vom Dienst**, nicht hier aus
  // den Parteien gerechnet.
  ich: string
  zustand: string

  positionen: Handelsposition[]
  bewertungGeteilt: boolean
  zustimmungen: string[]

  // ende sagt, ob aus diesem Zustand nichts mehr folgt.
  ende: boolean

  // beschluss ist der Schritt, dem die Zustimmungen gelten — leer, wenn
  // keiner offen ist (TRADE-03, #1717).
  beschluss: string

  // lage, naechster, moegliche und phase sagen, wo der Vorgang für
  // **dieses** Profil steht und was es jetzt tun darf (TRADE-02, #1716) —
  // aus `domain`, damit dieses Fenster die Matrix nicht nachrechnet.
  lage: string
  naechster: string
  moegliche: string[]
  phase: string
}

// Handelsvorgangantwort trägt den Vorgang samt seinem Fehler.
export interface Handelsvorgangantwort extends Umschlag {
  vorgang: Handelsvorgang
}

// Handelsvorgaengeantwort sind die Handel mit einem Gegenüber — je Vorgang
// die letzte Revision, zuletzt geänderte zuerst (UX-08.AK-04,
// #1680).
export interface Handelsvorgaengeantwort extends Umschlag {
  vorgaenge: Handelsvorgang[]
}

// Handelswunsch ist, womit ein Vorgang eröffnet wird.
//
// **Ohne Initiator** — er kommt aus dem Ausweis dieses Rechners.
export interface Handelswunsch {
  vorgang: string
  art: string
  gegenueber: string[]
  positionen: Handelsposition[]
  bewertungGeteilt: boolean
}

// Versandantwort nennt die Nachricht, die nun im Ausgang liegt.
export interface Versandantwort extends Umschlag {
  nachrichtId: string
}

// Handelsfortschritt ist, was eine Seite als bestätigt kennt.
export interface Handelsfortschritt {
  vorgang: string
  revision: number
  abdruck: string
  folge: number
}

// Nachlieferung ist eine Nachricht, die noch einmal hinausmuss.
export interface Nachlieferung {
  nachrichtId: string
  vorgang: string
  empfaenger: string
  folge: number
  art: string
}

// Wiederaufnahmeantwort sagt, was nach einem Abriss fehlt.
export interface Wiederaufnahmeantwort extends Umschlag {
  nachzuliefern: Nachlieferung[]

  // klaerung heißt: zwei Verläufe stehen da, und das lösen zwei Menschen.
  klaerung: boolean
}

// Zustellungsbericht sagt, was über die Leitung ging.
//
// **„Getragen" und nicht „zugestellt"**: ob es angekommen ist, sagt die
// Bestätigung von drüben und nicht diese Zahl.
export interface Zustellungsbericht extends Umschlag {
  gegenueber: string
  getragen: number
}

// Handelszusage ist eine festgehaltene Besitzzeile.
export interface Handelszusage {
  exemplarId: number
  menge: number
  fassung: number
  revision: number
  frist?: string
  uebergeben: boolean
}

// Zusagenantwort ist, was für einen Vorgang festgehalten ist.
export interface Zusagenantwort extends Umschlag {
  vorgang: string
  zusagen: Handelszusage[]

  // nachladefehler heißt: **festgehalten ist**, nur die Liste danach war
  // nicht zu lesen (§7, Regel 3; #1970). `zusagen` ist dann leer und sagt
  // nichts.
  nachladefehler?: FehlerDTO
}

// Handelsprotokollzeile ist ein Zustandswechsel (TRADE-03.AK-03, #1717).
export interface Handelsprotokollzeile {
  revision: number

  // von ist leer, wenn die Revision neu ist.
  von: string
  nach: string

  // zeit ist RFC 3339.
  zeit: string
}

// Handelsprotokollantwort ist das Protokoll samt seinem Fehler.
export interface Handelsprotokollantwort extends Umschlag {
  zeilen: Handelsprotokollzeile[]
}

// Anhaltspreis ist ein Betrag in einer Währung (TRADE-02b, #1951) — in der
// kleinsten Einheit; bei einer Summe ist `erhoben` der älteste Bestandteil.
export interface Anhaltspreis {
  waehrung: string
  wert: number
  quelle: string
  erhoben: string
}

// Handelsdruck ist, was eine Handelszeile über ihre Karte sagt — P10
// (#2378). **Leer heißt unbekannt**; `eigen` ist ein Druck ohne
// Scryfall-Kennung, und der hat nie ein Bild.
export interface Handelsdruck {
  name: string
  setcode: string
  nummer: string

  // bild ist die Adresse eines kleinen Kartenbilds — nur, was Scryfall schon
  // einmal beantwortet hat.
  bild: string
  eigen: boolean
}

// Anhaltszeile ist der Preis einer Position; leer heißt „kein Preis".
export interface Anhaltszeile {
  preise: Anhaltspreis[]
  druck: Handelsdruck
}

// Anhaltsseite ist, was eine Partei gibt.
export interface Anhaltsseite {
  von: string
  summen: Anhaltspreis[]
  ohnePreis: number
}

// Anhaltsdifferenz ist erhalten weniger gegeben, aus eigener Sicht.
export interface Anhaltsdifferenz {
  waehrung: string
  wert: number
  vollstaendig: boolean
}

// Preisanhalt ist, was die Seiten eines Tauschs ungefähr wert sind.
export interface Preisanhalt {
  zeilen: Anhaltszeile[]
  seiten: Anhaltsseite[]
  differenzen: Anhaltsdifferenz[]
  hauptwaehrung: string
}

// Preisanhaltantwort ist der Anhalt samt seinem Fehler.
export interface Preisanhaltantwort extends Umschlag {
  anhalt: Preisanhalt
}

// Handelsleitungapi sind die zehn Wege des Vertrags, seines Ausgangs und
// seiner Zusagen.
export interface Handelsleitungapi {
  HandelProtokoll: (vorgang: string) => Promise<Handelsprotokollantwort>
  HandelPreisanhalt: (
    positionen: Handelsposition[],
  ) => Promise<Preisanhaltantwort>
  HandelEroeffnen: (wunsch: Handelswunsch) => Promise<Handelsvorgangantwort>
  HandelVorschlagen: (
    vorgang: string,
    positionen: Handelsposition[],
  ) => Promise<Handelsvorgangantwort>
  // `gesehen` ist der Abdruck der Revision, die der Schritt meint (#2496);
  // leer heißt: nicht gefragt.
  HandelSchritt: (
    vorgang: string,
    schritt: string,
    wer: string,
    gesehen: string,
  ) => Promise<Handelsvorgangantwort>
  HandelStand: (vorgang: string) => Promise<Handelsvorgangantwort>
  HandelVorgaenge: (gegenueber: string) => Promise<Handelsvorgaengeantwort>

  HandelSchicken: (
    vorgang: string,
    empfaenger: string,
  ) => Promise<Versandantwort>
  HandelPostbestaetigung: (nachrichtId: string) => Promise<Versandantwort>
  HandelWiederaufnahme: (
    vorgang: string,
    empfaenger: string,
    seiner: Handelsfortschritt,
  ) => Promise<Wiederaufnahmeantwort>

  HandelZustellen: (kennung: string) => Promise<Zustellungsbericht>

  // HandelSichtAnfragen fragt ein Gegenüber über die Leitung nach dem, was es
  // diesem Profil zum Handeln freigibt — P11 (#2379).
  HandelSichtAnfragen: (kennung: string) => Promise<Sichtvorschauantwort>

  HandelFesthalten: (
    vorgang: string,
    exemplarId: number,
    menge: number,
    fassung: number,
  ) => Promise<Zusagenantwort>
  Handelszusagen: (vorgang: string) => Promise<Zusagenantwort>
}
