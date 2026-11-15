// **Die Turnierseite der Fassade** — Werkstück 5 von Phase 4.7
// (#1294),
// nach Vermerk 0062.
//
// **Eine eigene Datei neben `apidecks.ts`**, aus demselben Grund: `api.ts`
// steht an der Zeilengrenze aus §2. Der Schnitt folgt der Frage — dort steht
// die **Sammlung**, hier der **Verlauf**.
//
// **Jede Kennung ist Text.** JavaScript kennt nur `float64`; eine Kennung
// jenseits von 2^53 käme gerundet an.
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// TurnierDTO ist ein Turnier samt Teilnehmern und Verlauf.
export interface TurnierDTO {
  id: string
  fassung: string

  name: string
  spiel: string
  notiz: string

  // **Ein Match aus einem Spiel** — dann sagen Spielquote und
  // Gegnerspielquote nichts (MTR §3.1), weil sie Kopien der Matchzahlen sind.
  // Gerechnet werden sie trotzdem; dieses Feld entscheidet die Anzeige.
  einzelspiel: boolean

  // formatID ist das Format, nach dem gespielt wird — **leer, wenn keines
  // gewählt ist** (DOM-04.AK-01, #1699).
  formatID: string

  // abgeschlossen ist das Datum des Abschlusses — **leer heißt offen**
  // (DOM-04.AK-04, #1699, Vermerk 0139).
  abgeschlossen: string

  teilnehmer: TeilnehmerDTO[]
  runden: RundeDTO[]

  // **Die Rundenzahl aus Appendix E der Magic Tournament Rules** — null heißt:
  // keine Empfehlung. Ein Vorschlag und keine Vorschrift.
  empfohleneRunden: number
}

// TeilnehmerDTO ist ein Teilnehmer samt seiner eingereichten Liste.
export interface TeilnehmerDTO {
  id: string
  name: string

  // **Die Herkunft der Liste, nicht ihre Quelle** — leer, wenn keines
  // hinterlegt wurde.
  deckID: string

  // Die Liste, wie sie beim Anmelden aussah.
  deckabzug: AbzugsplatzDTO[]

  ausgestiegen: boolean
  stelle: number
}

// AbzugsplatzDTO ist eine Zeile der eingereichten Liste.
//
// **Text und keine Verweise**: sie überlebt jedes Löschen. Bei einem echten
// Druck bleibt `bezeichnung` leer — wie er heißt, weiß nur Scryfall.
export interface AbzugsplatzDTO {
  id: string
  bezeichnung: string
  printing: string
  // druckID ist die Zeile des Drucks in diesem Bestand (A2, #1441), leer bei
  // einer eigenen Karte.
  druckID: string
  menge: number
  bereich: string
}

// RundeDTO ist eine Runde samt ihren Paarungen.
export interface RundeDTO {
  id: string
  nummer: number
  paarungen: PaarungDTO[]
}

// PaarungDTO ist ein Match — oder ein Bye.
export interface PaarungDTO {
  id: string
  a: string

  // **Leer bei einem Bye** — und nicht „0". Eine Null sähe wie ein Gegner aus.
  b: string

  // **`null`, solange nicht gespielt ist.** Eine offene Paarung zählt nicht.
  ergebnis: ErgebnisDTO | null
}

// ErgebnisDTO ist ein Spielstand, kein Sieger.
//
// **Weil Appendix C der Turnierregeln Spielpunkte braucht**, und die folgen
// aus dem Stand. Wer nur den Sieger hielte, könnte zwei der vier Tiebreaker
// nicht rechnen.
export interface ErgebnisDTO {
  spieleA: number
  spieleB: number
  unentschieden: number
}

// RangDTO ist eine Zeile der Rangliste.
//
// **Die Quoten kommen gerundet an, gerechnet wird ungerundet** — vier
// Nachkommastellen, weil zwei Teilnehmer, die sich in der vierten
// unterscheiden, nicht gleichauf sind.
export interface RangDTO {
  teilnehmerID: string
  name: string
  platz: number

  siege: number
  niederlagen: number
  unentschieden: number
  byes: number

  matchpunkte: number

  matchquote: number
  gegnermatchquote: number
  spielquote: number
  gegnerspielquote: number

  ausgestiegen: boolean
}

export interface TurniereAntwort {
  turniere: TurnierDTO[]
  fehler: FehlerDTO | null
}

export interface TurnierAntwort {
  turnier: TurnierDTO | null
  fehler: FehlerDTO | null
}

export interface TeilnehmerAntwort {
  teilnehmer: TeilnehmerDTO | null
  fehler: FehlerDTO | null
}

export interface RundeAntwort {
  runde: RundeDTO | null
  fehler: FehlerDTO | null
}

export interface PaarungenAntwort {
  paarungen: PaarungDTO[]
  fehler: FehlerDTO | null
}

export interface RanglisteAntwort {
  rangliste: RangDTO[]
  fehler: FehlerDTO | null
}

// SchreibAntwortTurnier ist dieselbe Gestalt wie `SchreibAntwort` in `api.ts`.
//
// **Hier noch einmal erklärt und nicht importiert**: `api.ts` importiert diese
// Datei, und der umgekehrte Weg wäre ein Kreis. Dieselbe Form wie in
// `apidecks.ts`.
export interface SchreibAntwortTurnier {
  fehler: FehlerDTO | null
}

// KorrekturfolgenDTO sagt, was eine Berichtigung verschöbe — DOM-04.AK-02
// (#1699). Die späteren Runden bleiben, wie sie sind; sie stehen hier, damit
// es gesagt wird.
export interface KorrekturfolgenDTO {
  runde: number
  spaetereRunden: number[]
  verschoben: PlatzwechselDTO[]
}

// PlatzwechselDTO ist ein Teilnehmer vor und nach der Berichtigung.
export interface PlatzwechselDTO {
  teilnehmerID: string
  name: string
  platzVorher: number
  platzNachher: number
  punkteVorher: number
  punkteNachher: number
}

export interface KorrekturfolgenAntwort {
  folgen: KorrekturfolgenDTO | null
  fehler: FehlerDTO | null
}

// ProtokolleintragDTO ist eine Zeile des Protokolls — DOM-04.AK-04. Die
// Angaben sind sprachfrei; den Satz baut das Fenster aus `vorgang`.
export interface ProtokolleintragDTO {
  zeit: string
  vorgang: string
  angaben: string
  paarungID: string
}

export interface ProtokollAntwort {
  eintraege: ProtokolleintragDTO[]
  fehler: FehlerDTO | null
}

// TurnierdateiAntwort ist der Ausgang von Ausgabe, Einlesen und Bericht —
// DOM-04.AK-03 (#1699). `abgebrochen` heißt: der Dialog wurde weggeklickt.
export interface TurnierdateiAntwort {
  kennung: string
  pfad: string
  inhalt: string
  abgebrochen: boolean
  befunde: string[]
  fehler: FehlerDTO | null
}

// Turnierapi ist, was das Fenster über Turniere rufen kann.
export interface Turnierapi {
  Turniere: (spiel: string) => Promise<TurniereAntwort>
  Turnier: (turnierID: string) => Promise<TurnierAntwort>
  TurnierAnlegen: (
    name: string,
    spiel: string,
    notiz: string,
    einzelspiel: boolean,
    formatID: string,
  ) => Promise<TurnierAntwort>
  // **Abschluss und Protokoll** — DOM-04.AK-04 (#1699): gegen die Fassung.
  TurnierAbschliessen: (
    turnierID: string,
    fassung: string,
  ) => Promise<SchreibAntwortTurnier>
  TurnierWiedereroeffnen: (
    turnierID: string,
    fassung: string,
  ) => Promise<SchreibAntwortTurnier>
  TurnierProtokoll: (turnierID: string) => Promise<ProtokollAntwort>
  // **Datei und Bericht** — DOM-04.AK-03 (#1699): jeder fragt nach einem Ort.
  TurnierAusgabeSpeichern: (turnierID: string) => Promise<TurnierdateiAntwort>
  TurnierBerichtSpeichern: (turnierID: string) => Promise<TurnierdateiAntwort>
  TurnierEinlesenOeffnen: () => Promise<TurnierdateiAntwort>
  // **Was eine Berichtigung verschöbe** — DOM-04.AK-02 (#1699): ein Lesen,
  // `null` fragt nach dem Zurücknehmen.
  TurnierKorrekturfolgen: (
    turnierID: string,
    paarungID: string,
    stand: ErgebnisDTO | null,
  ) => Promise<KorrekturfolgenAntwort>
  // **Nach einer Vorlage** — DOM-04 (#1699): Einstellungen und auf Wunsch
  // die Namen, nie die Decks.
  TurnierKopieren: (
    turnierID: string,
    name: string,
    mitTeilnehmenden: boolean,
  ) => Promise<TurnierAntwort>
  TurnierLoeschen: (turnierID: string) => Promise<SchreibAntwortTurnier>

  TurnierAnmelden: (
    turnierID: string,
    name: string,
    deckID: string,
  ) => Promise<TeilnehmerAntwort>
  // **Turnier und Teilnehmer.** Die Kennung eines Teilnehmers ist global;
  // ohne das Turnier daneben änderte ein Ruf jemanden in einem anderen.
  TurnierAussteigen: (
    turnierID: string,
    teilnehmerID: string,
    ausgestiegen: boolean,
  ) => Promise<SchreibAntwortTurnier>

  // **Der Vorschlag schafft keine Tatsachen** — ein Turnierleiter soll
  // hinsehen können, bevor er eine Runde legt.
  TurnierPaarungsvorschlag: (turnierID: string) => Promise<PaarungenAntwort>
  TurnierRundeLegen: (turnierID: string) => Promise<RundeAntwort>

  // **`null` heißt: doch nicht gespielt.** Berichtigen ist am Turniertisch der
  // Normalfall.
  //
  // **Die Fassung reist mit** (§7, Regel 2): zwei Leute am selben Tisch,
  // beide haben die Runde offen — wer als Zweiter einträgt, überschriebe
  // sonst die Berichtigung des Ersten.
  TurnierErgebnisSetzen: (
    turnierID: string,
    fassung: string,
    paarungID: string,
    stand: ErgebnisDTO | null,
  ) => Promise<SchreibAntwortTurnier>

  TurnierRangliste: (turnierID: string) => Promise<RanglisteAntwort>

  // **Eine Paarung, an eine Partie gebunden** (G7,
  // #1464).
  //
  // **Vier Wege, und keiner trägt allein ein Ergebnis ein**: binden sagt,
  // welche Partie dazugehört; melden ist die Aussage **einer** Seite;
  // übernehmen trägt ein, wenn beide dasselbe sagen. `TurnierErgebnisSetzen`
  // daneben bleibt — wer am Küchentisch führt, tippt weiter.
  TurnierMatch: (paarungID: string) => Promise<MatchAntwort>
  TurnierMatchBinden: (
    turnierID: string,
    paarungID: string,
    partieID: string,
    zuordnung: Matchsitz[],
  ) => Promise<MatchAntwort>
  TurnierMatchMelden: (
    paarungID: string,
    von: string,
    kette: string,
    stand: ErgebnisDTO,
  ) => Promise<MatchAntwort>
  // TurnierMatchAusDerPartie meldet, was die **Partie** festgestellt hat —
  // M17 (#1618). Spielstand und Nachweis kommen von dort; übrig bleibt, wer
  // meldet.
  TurnierMatchAusDerPartie: (
    paarungID: string,
    von: string,
  ) => Promise<MatchAntwort>
  TurnierMatchUebernehmen: (
    turnierID: string,
    paarungID: string,
  ) => Promise<MatchAntwort>
}

// Matchsitz ordnet einen Teilnehmer seinem Sitz zu.
//
// **Drei Angaben, und die dritte ist der Punkt**: ein Teilnehmer ist ein
// **Name**, und ein Name ist kein authentifizierter Spieler. Der Abdruck
// kommt aus der Partie und wird nicht mitgeschickt.
export interface Matchsitz {
  teilnehmerID: string
  sitz: string

  // peer ist die Peerkennung — leer heißt: an diesem Rechner.
  peer: string

  // deckabdruck kommt zurück und geht nicht hinaus.
  deckabdruck: string
}

// Matchmeldung ist, was eine Seite gemeldet hat.
export interface Matchmeldung {
  von: string
  ergebnis: ErgebnisDTO

  // kette ist der letzte Kettenabdruck des Ereignisstroms — der Nachweis.
  //
  // **Was er belegt**: dass beide Seiten denselben Strom gesehen haben.
  // **Was er nicht belegt**: dass richtig gespielt wurde.
  kette: string
}

// MatchAntwort trägt die Bindung einer Paarung und ihre Meldungen.
export interface MatchAntwort extends Umschlag {
  // gebunden sagt, ob an dieser Paarung überhaupt eine Partie hängt.
  //
  // **`false` ist kein Fehler** — die meisten Paarungen werden von Hand
  // eingetragen.
  gebunden: boolean

  paarungID: string
  partieID: string
  fassung: string

  sitze: Matchsitz[]
  meldungen: Matchmeldung[]

  // uebernommen ist das Ergebnis, das übernommen wurde — `null`, solange
  // keines übernommen ist.
  uebernommen: ErgebnisDTO | null

  // nachladefehler heißt: **der Schritt gilt**, nur der Stand danach war
  // nicht zu lesen (§7, Regel 3; #1970). Die Listen sind dann leer und
  // sagen nichts.
  nachladefehler?: FehlerDTO
}
