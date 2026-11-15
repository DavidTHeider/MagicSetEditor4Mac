// **Decks, Formate und ihre beiden Auskünfte** — Werkstück 7 von Phase 4.6
// (#1293),
// nach Vermerk 0061.
//
// **Eine eigene Datei**, damit `api.ts` unter der Zeilengrenze aus §2 bleibt —
// dieselbe Form wie `pakete.ts` und `stilwahl.ts`.
//
// **Kennungen sind Zeichenketten.** Über die Wails-Bindung käme ein `int64`
// als JavaScript-Zahl an, und die verliert oberhalb von 2⁵³ die letzten
// Stellen.

import type { FehlerDTO } from './formen'
import type { HerkunftDTO } from './apiherkunft'
import type { ExemplarDTO } from './antwortensammlung'

export interface DeckplatzDTO {
  id: string

  // Genau eines von beiden ist gefüllt. Die Null kommt als **leere**
  // Zeichenkette, nicht als '0' — sonst sähe sie aus wie eine Kennung.
  karteID: string
  printing: string

  // druckID ist die Zeile des Drucks in diesem Bestand (A2, #1441) — leer
  // für eine eigene Karte; vorlaeufig sagt, dass der Druck keinen Alias hat.
  druckID: string
  vorlaeufig: boolean

  // typzeile und manawert tragen die Auswertung (U10, #1421). Beide kommen
  // mit dem Namen; leer und null heißen **nicht zu haben** (§16, Regel 5).
  //
  // **Der Manawert ist nachgelesen**: eine zweiseitige Karte hat bei Scryfall
  // leere Kosten und trotzdem einen Wert. Für eine eigene Karte gibt es
  // keinen — dort rechnet `deckrechnung.ts` aus den Kosten.
  typzeile: string
  manawert: number

  menge: number
  bereich: string

  // bezeichnung ist der **Name** der Karte
  // (#1376).
  //
  // **Leer heißt: nicht zu haben.** Der Name eines Drucks kommt von Scryfall;
  // ohne Netz bleibt er aus (§16, Regel 5), und dann steht die Kennung da.
  bezeichnung: string

  // manakosten sind die Kosten der Karte — für die Kurve.
  //
  // **Leer heißt: nicht ausgefüllt.** `Vanguard` und `Archenemy` führen das
  // Feld und lassen es leer; wer daraus eine Kurve rechnet, lässt sie weg.
  manakosten: string
}

export interface DeckDTO {
  id: string
  fassung: string

  name: string
  spiel: string
  notiz: string

  // Leer heißt: kein Format gewählt. **Ein Deck ohne Format ist ungeprüft,
  // nicht falsch.**
  formatID: string

  // In der Übersicht leer, beim einzelnen Deck gefüllt.
  plaetze: DeckplatzDTO[]

  hauptdeck: number
  sideboard: number

  // Woher dieses Deck kam — Werkstück 12 von Phase 4.12
  // (#1351).
  //
  // **In der Übersicht leer, beim einzelnen Deck gefüllt** — wie `plaetze`.
  herkunft: HerkunftDTO
}

// Deckuebernahmeantwort trägt das übernommene Deck samt dem, was nicht mitkam.
//
// **Die Befunde stehen neben dem Deck und sind kein Fehler**: das Deck ist da.
//
// **`Deckuebernahmeantwort` und nicht `UebernahmeAntwort`**: den Namen trägt
// seit Phase 4.10 die Paketübernahme aus einer MSE2-Installation.
export interface Deckuebernahmeantwort {
  deck: DeckDTO | null
  befunde: string[]
  fehler: FehlerDTO | null
}

export interface DecksAntwort {
  decks: DeckDTO[]
  fehler: FehlerDTO | null
}

export interface DeckAntwort {
  deck: DeckDTO | null

  // nachladefehler heißt: **angelegt ist**, nur das Nachlesen danach
  // scheiterte (§7, Regel 3; #1970). `deck` trägt dann das Deck, wie es
  // angelegt wurde.
  nachladefehler?: FehlerDTO

  fehler: FehlerDTO | null
}

export interface DeckplatzAntwort {
  platz: DeckplatzDTO | null
  fehler: FehlerDTO | null
}

export interface FormatDTO {
  id: string
  fassung: string

  name: string
  spiel: string

  // **Null heißt: keine Grenze** — bei `sideboardgroesse` heißt es: kein
  // Sideboard.
  mindestgroesse: number
  hoechstgroesse: number
  sideboardgroesse: number
  hoechstzahlGleich: number

  // **Ein Sideboard ohne Obergrenze** — nicht dasselbe wie
  // `sideboardgroesse: 0`, das „kein Sideboard" heißt. CR 100.4b legt bei
  // limitiertem Spiel die übrigen Karten des Pools ins Sideboard, ohne eine
  // Zahl zu nennen.
  sideboardOffen: boolean

  // Leer heißt: Scryfall kennt dieses Format nicht, und jede Karte bleibt
  // `ungeprüft`.
  scryfallname: string
}

export interface BauartDTO {
  bauart: string
  format: FormatDTO
}

export interface FormateAntwort {
  formate: FormatDTO[]
  fehler: FehlerDTO | null
}

export interface FormatAntwort {
  format: FormatDTO | null
  fehler: FehlerDTO | null
}

export interface BauartenAntwort {
  bauarten: BauartDTO[]
  fehler: FehlerDTO | null
}

export interface KartenurteilDTO {
  platzID: string
  bezeichnung: string
  menge: number
  bereich: string

  // Eines von Scryfalls vier Wörtern — 'legal', 'not_legal', 'restricted',
  // 'banned' — oder 'ungeprüft'. **Nicht übersetzt**: der Katalog hat für die
  // fünf einen Satz.
  zulaessigkeit: string

  grund: string
  quelle: string
}

export interface GroessenbefundDTO {
  feld: string
  meldung: string
}

export interface DeckurteilDTO {
  formatname: string

  hauptdeck: number
  sideboard: number

  urteile: KartenurteilDTO[]
  groessen: GroessenbefundDTO[]

  // **`ungeprüft` steht dem nicht entgegen** — das ist die Zusage aus
  // Vermerk 0061.
  zulaessig: boolean
  ungeprueft: number
}

export interface DeckurteilAntwort {
  urteil: DeckurteilDTO | null

  // Was unterwegs schiefging — **kein Fehler**: die Prüfung lief weiter.
  befunde: string[]

  fehler: FehlerDTO | null
}

export interface HabenzeileDTO {
  printing: string

  // druckID ist die Zeile des Drucks in diesem Bestand (A2, #1441);
  // vorlaeufig sagt, dass er keinen Scryfall-Alias hat.
  druckID: string
  vorlaeufig: boolean
  bezeichnung: string
  gebraucht: number
  vorhanden: number
  fehlend: number
  grund: string
}

// VorhandenerWunschDTO ist ein Druck, der fehlt und schon gewünscht ist —
// DOM-03.AK-02 (#1698).
export interface VorhandenerWunschDTO extends HabenzeileDTO {
  gewuenscht: number
}

// FehlkartenwunschAntwort sagt, was neu gewünscht ist und was schon war.
export interface FehlkartenwunschAntwort {
  neu: ExemplarDTO[]
  vorhanden: VorhandenerWunschDTO[]
  unbekannt: number
  fehler: FehlerDTO | null
}

// ZuordnungDTO ist eine Zuordnung, wie das Fenster sie sieht (U9, #1420).
export interface ZuordnungDTO {
  id: string
  platzID: string
  exemplarID: string
  menge: number

  // bezeichnung ist, wie die Sammlung das Exemplar nennt.
  bezeichnung: string

  // abweichend sagt, dass das Exemplar einen **anderen** Druck trägt als der
  // Platz — ausdrücklich gewählt, nie über Namensgleichheit.
  abweichend: boolean
}

// ZuordnungszeileDTO ist eine Zeile des Bedarfs samt ihren Zuordnungen.
export interface ZuordnungszeileDTO extends HabenzeileDTO {
  // reserviert, zugeordnet und frei gelten für **alle** Exemplare dieses
  // Drucks — sonst sähe ein zweites Deck Stücke als frei, die vergeben sind.
  reserviert: number
  zugeordnet: number
  frei: number

  zuordnungen: ZuordnungDTO[]
}

export interface ZuordnungsstandAntwort {
  zeilen: ZuordnungszeileDTO[]
  fehlend: number
  unbekannt: number
  fehler: FehlerDTO | null
}

export interface ZuordnungAntwort {
  zuordnung: ZuordnungDTO | null
  fehler: FehlerDTO | null
}

export interface FehlkartenAntwort {
  liste: string
  fehler: FehlerDTO | null
}

// DeckbezugDTO sagt, welches Deck ein Exemplar benutzt — der Rückweg.
export interface DeckbezugDTO {
  deckID: string
  name: string
  menge: number
}

export interface DeckbezuegeAntwort {
  decks: DeckbezugDTO[]
  fehler: FehlerDTO | null
}

export interface HabenstandAntwort {
  zeilen: HabenzeileDTO[]

  // **Getrennt**: „vier fehlen" und „über vier ist nichts bekannt" sind zwei
  // Auskünfte.
  fehlend: number
  unbekannt: number

  fehler: FehlerDTO | null
}

// Deckapi ist der Teil der Fassade, der Decks und Formate bedient.
export interface Deckapi {
  Decks: (spiel: string) => Promise<DecksAntwort>
  Deck: (deckID: string) => Promise<DeckAntwort>
  DeckAnlegen: (
    name: string,
    spiel: string,
    notiz: string,
  ) => Promise<DeckAntwort>
  DeckUmbenennen: (
    deckID: string,
    fassung: string,
    name: string,
    notiz: string,
  ) => Promise<DeckAntwort>
  DeckLoeschen: (deckID: string) => Promise<SchreibAntwortDeck>

  // DeckKopieren legt eine Kopie unter neuem Namen an — DOM-03.AK-02 (#1698).
  DeckKopieren: (deckID: string, name: string) => Promise<DeckAntwort>
  // DeckAusSet legt aus den Karten eines Sets ein Deck an — FMT-04 (#1693).
  DeckAusSet: (
    setID: string,
    name: string,
    menge: number,
  ) => Promise<DeckAntwort>

  // DeckFehlkartenWuenschen setzt, was fehlt und noch nicht gewünscht ist,
  // auf die Wunschliste — DOM-03.AK-02 (#1698).
  DeckFehlkartenWuenschen: (deckID: string) => Promise<FehlkartenwunschAntwort>

  // DeckUebernehmen holt ein Deck aus einem **fremden** Profil in das eigene
  // — Werkstück 12 (#1351). Es braucht dort mindestens die Rolle
  // `mitarbeiter`.
  DeckUebernehmen: (
    quellProfil: string,
    quellDeck: string,
  ) => Promise<Deckuebernahmeantwort>

  // DeckeZurUebernahme nennt die Decks jenes Profils — zur Wahl, hinter
  // derselben Erlaubnis (UX-08.AK-04, #1680).
  DeckeZurUebernahme: (quellProfil: string) => Promise<DecksAntwort>
  DeckplatzLegen: (
    deckID: string,
    karteID: string,
    printing: string,
    menge: number,
    bereich: string,
  ) => Promise<DeckplatzAntwort>
  DeckplatzNehmen: (
    deckID: string,
    platzID: string,
  ) => Promise<SchreibAntwortDeck>

  // **Der Bereichswechsel ist ein Ruf und nicht zwei** — Befund 7 des
  // Zweitblicks über Phase 5.1: „legen, dann nehmen" ließ bei einem
  // gescheiterten zweiten Ruf dieselbe Karte in beiden Bereichen liegen.
  DeckplatzVerschieben: (
    deckID: string,
    platzID: string,
    bereich: string,
  ) => Promise<DeckplatzAntwort>
  DeckFormatWaehlen: (deckID: string, formatID: string) => Promise<DeckAntwort>
  DeckPruefen: (deckID: string) => Promise<DeckurteilAntwort>
  DeckAbgleichen: (deckID: string) => Promise<HabenstandAntwort>

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1
  // (#1420).
  // Der Bedarf reserviert nichts; zugeordnet ist, was ausdrücklich für dieses
  // Deck liegt, und `frei` rechnet über **alle** Decks.
  DeckZuordnungen: (deckID: string) => Promise<ZuordnungsstandAntwort>
  DeckZuordnen: (
    deckID: string,
    platzID: string,
    exemplarID: string,
    menge: number,
  ) => Promise<ZuordnungAntwort>
  DeckZuordnungLoesen: (
    deckID: string,
    zuordnungID: string,
  ) => Promise<SchreibAntwortDeck>
  DeckFehlkarten: (deckID: string) => Promise<FehlkartenAntwort>
  ExemplarDecks: (exemplarID: string) => Promise<DeckbezuegeAntwort>

  Bauarten: () => Promise<BauartenAntwort>
  Formate: (spiel: string) => Promise<FormateAntwort>
  FormatAusBauart: (
    bauart: string,
    name: string,
    spiel: string,
    scryfallname: string,
  ) => Promise<FormatAntwort>
  FormatSetzen: (format: FormatDTO) => Promise<FormatAntwort>
  FormatLoeschen: (formatID: string) => Promise<SchreibAntwortDeck>
}

// SchreibAntwortDeck ist dieselbe Gestalt wie `SchreibAntwort` in `api.ts`.
//
// **Hier noch einmal erklärt und nicht importiert**: `api.ts` importiert diese
// Datei, und der umgekehrte Weg wäre ein Kreis. Dieselbe Form wie in
// `apisammlung.ts`.
export interface SchreibAntwortDeck {
  fehler: FehlerDTO | null
}
