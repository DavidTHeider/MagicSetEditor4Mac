// **Was von fremden Märkten und Anbietern zurückkommt** — die Werkstücke A4
// und A5 von Phase 5.3
// (#1443,
// #1444).
//
// **Eine eigene Datei**, seit `antwortensammlung.ts` mit ihnen über die
// Zeilengrenze aus §2 lief. Der Schnitt folgt der Frage: dort steht, was aus
// der **eigenen** Sammlung kommt, hier, was **andere** über sie sagen.
import type { BelegDTO } from './apirecherche'
import type { KorpusstandDTO } from './apiindex'
import type { Umschlag } from './umschlag'

// MarktangebotDTO ist ein Posten auf einem fremden Marktplatz (A4, #1443).
//
// **Kein Preisstand und kein Exemplar**: es ist die Aussage eines fremden
// Verkäufers — jemand sagt, er habe etwas und gebe es zu einem Preis ab.
export interface MarktangebotDTO {
  quelle: string

  // kennung ist die Nummer dieses Postens beim Anbieter — nicht die der
  // Kartenfassung.
  kennung: string

  // wert ist der Preis als Text — `"2.17"`. Als Zahl verlöre er Cent.
  wert: string
  waehrung: string

  menge: number
  buendel: number

  finish: string

  // zustand steht so da, wie der Anbieter ihn nennt — er wird **nicht** in die
  // Skala dieser Anwendung übersetzt (0095).
  zustand: string
  sprache: string

  signiert: boolean
  veraendert: boolean
  gradiert: boolean

  // imUrlaub sagt, dass der Verkäufer gerade nicht versendet.
  imUrlaub: boolean

  verkaeufer: string
  land: string
}

// Marktantwort trägt, was ein Marktabruf ergeben hat (A4, #1443).
export interface Marktantwort extends Umschlag {
  quelle: string

  // fassung ist die Nummer der Kartenfassung beim Anbieter.
  fassung: string

  // zugeordnet sagt, ob die Fassung gerade erst gefunden wurde.
  zugeordnet: boolean

  angebote: MarktangebotDTO[]

  // befunde nennt Posten, die nicht zu deuten waren (§8).
  befunde: string[]
}

// Anbieterstandantwort ist das Kontingent an einem Anbieterkonto (A5, #1444).
export interface Anbieterstandantwort extends Umschlag {
  dienst: string

  // gesamt, rest und verbraucht sind Kredite.
  gesamt: number
  rest: number
  verbraucht: number

  // ueberzugssatz ist Text — Geld je Kredit; als Zahl verlöre er Stellen.
  ueberzugssatz: string

  // hoechstensAltMinuten sagt, wie alt die Zahl beim Anbieter sein kann.
  //
  // **Sie steht in jeder Antwort**: wer den Rest liest und die Frist nicht,
  // hält ihn für eine Sperre.
  hoechstensAltMinuten: number
}

// Fremdpreisantwort ist, was ein Verlaufsbezug ergeben hat (A5, #1444).
export interface Fremdpreisantwort extends Umschlag {
  quelle: string

  // fassung ist die Kennung der Karte beim Anbieter.
  fassung: string

  // zugeordnet sagt, ob sie gerade erst gefunden wurde.
  zugeordnet: boolean

  // staende ist die Zahl der abgelegten Preisstände.
  staende: number

  befunde: string[]
}

// KIBelegDTO ist eine belegte Fähigkeit eines KI-Modells (A6, #1445).
export interface KIBelegDTO {
  faehigkeit: string
  datum: string

  // beleg sagt, **wie** geprüft wurde — „geprüft" allein ist eine Behauptung.
  beleg: string
}

// KIGliedDTO ist ein Glied der Herkunftskette.
export interface KIGliedDTO {
  art: string
  kennung: string
  aus: string
  quelle: string
  revision: string
  lizenz: string
}

// KIModellDTO ist ein KI-Modell samt dem, was über es feststeht.
export interface KIModellDTO {
  anbieter: string
  alias: string
  laufzeit: string

  // ort ist `lokal` oder `cloud` — **wie der Anbieter es erklärt**. Eine
  // Adresse auf `127.0.0.1` beweist nichts: sie kann weiterleiten.
  ort: string

  // faehigkeiten sind die **belegten** — jede mit Datum und Beleg.
  faehigkeiten: KIBelegDTO[]

  // herkunft ist leer, wo sie unbekannt ist — und das ist eine Auskunft.
  herkunft: KIGliedDTO[]

  bindung: string

  // kontextfenster ist, wie viele Tokens das Modell auf einmal liest — 0
  // heißt nicht angegeben; kontextbeleg sagt, woher die Zahl stammt
  // (AI-01.AK-02, #1742).
  kontextfenster: number
  kontextbeleg: string
}

// KICloudDTO ist der Cloudstand samt Fundstelle.
//
// **Drei Stände und nicht zwei** (A7, #1446): `aus`, `an`, `ungewiss`. „Keine
// Serverdatei" heißt ungewiss — die zweite dokumentierte Abschaltung steht in
// der Umgebung eines fremden Prozesses.
export interface KICloudDTO {
  stand: string
  beleg: string
}

// KIKatalogantwort ist der Katalog samt der Regel, die für ihn gilt.
export interface KIKatalogantwort extends Umschlag {
  nurLokal: boolean

  // einrichtung ist einer der fünf Stände: `aus`, `unerreichbar`,
  // `ohnemodell`, `ohnefaehigkeit`, `bereit` (A7, #1446). Er sagt, **warum**
  // die Liste leer ist.
  einrichtung: string

  // anbieter sind die Quellen, jede mit ihrem eigenen Stand.
  //
  // **Eine Liste seit A8** (#1447). A7 trug Einrichtung, Fassung und Adresse
  // **flach**; mit dem zweiten Anbieter sprächen sie von einem von zweien
  // und sähen aus, als sprächen sie von beiden.
  anbieter: KIAnbieterDTO[]

  cloud: KICloudDTO

  // befunde sind die Sätze, die erklären, was fehlt.
  befunde: string[]

  modelle: KIModellDTO[]

  // vorrang nennt, wen ein Vorschlag mit Vorrang für Name oder Flavortext
  // fragt, als `anbieter/modell` — #2054: es steht am Formular.
  vorrang: string[]

  // vorrangbefunde sind die Sätze, die die Reihe betreffen — Z4.
  vorrangbefunde: string[]
}

// KIAnbieterDTO ist ein Anbieter samt dem, was über ihn feststeht (A8, #1447).
export interface KIAnbieterDTO {
  name: string

  // einrichtung ist einer der fünf Stände: `aus`, `unerreichbar`,
  // `ohnemodell`, `ohnefaehigkeit`, `bereit`.
  einrichtung: string

  // laufzeit ist die Fassung — bei Anthropic die der Schnittstelle, bei
  // Ollama die des Programms. Leer heißt: es hat nichts geantwortet.
  laufzeit: string

  // adresse ist, wohin gefragt wurde — leer bei einem Anbieter mit fester.
  adresse: string

  // ort ist `lokal` oder `cloud` — **wo dieser Anbieter rechnet** (#1539).
  // Eine Messung an einem Cloudanbieter kostet Geld, und das muss die Fläche
  // wissen, **bevor** sie ruft.
  ort: string

  // befunde sind die Sätze **dieses** Anbieters.
  befunde: string[]

  // messbar sind die Modelle, die er **nennt**, ohne etwas über sie zu sagen
  // (A9, #1448). Sie stehen nicht im Katalog — ohne Beleg gehört keines
  // hinein —, und ohne diese Liste wären sie unerreichbar.
  messbar: string[]

  // gemessen sind die Modelle, für die eine Messung im Bestand liegt — sie
  // lässt sich zurücknehmen.
  gemessen: string[]

  // sitzungAus sagt, dass er **für diese Sitzung** abgeschaltet ist — nur
  // Apples Gerätemodell kennt das; nach einem Neustart gilt wieder die
  // Einstellung (AI-05, #2042).
  sitzungAus: boolean
}

// KIGeraetSitzungAntwort sagt, ob das Gerätemodell in dieser Sitzung gefragt
// wird (AI-05, #2042).
export interface KIGeraetSitzungAntwort extends Umschlag {
  an: boolean
}

// KIProbeDTO ist eine gemessene Fähigkeit (A9, #1448).
export interface KIProbeDTO {
  anbieter: string
  alias: string
  faehigkeit: string
  datum: string
  beleg: string

  // modell ist, was wirklich geantwortet hat — nicht der Alias.
  modell: string
}

// KIProbeAntwort trägt die Messungen an einem Modell.
export interface KIProbeAntwort extends Umschlag {
  proben: KIProbeDTO[]
}

// KIFeldwertDTO ist ein vorgeschlagener Feldwert.
export interface KIFeldwertDTO {
  feld: string
  werte: string[]

  // bisher ist, was **heute** im Feld steht — leer heißt: nichts.
  //
  // **Damit sichtbar ist, was sich ändert** (A11, #1450): ein Vorschlag, der
  // nur den neuen Wert zeigt, verlangt von jemandem, dass er den alten im
  // Kopf hat.
  bisher: string
}

// KIFeldherkunftDTO sagt, woher der Wert eines Feldes stammt (A11, #1450).
export interface KIFeldherkunftDTO {
  feld: string
  anbieter: string
  modell: string

  // aufgabe ist, was ein Mensch getippt hat — **nicht** der Rohprompt.
  aufgabe: string

  // auftrag bindet die Felder **einer** Übernahme zusammen.
  auftrag: string

  zeitpunkt: string

  // veraendert sagt, ob am Feld seither von Hand gearbeitet wurde.
  veraendert: boolean
}

// KIFeldherkunftAntwort trägt die Herkünfte einer Karte.
export interface KIFeldherkunftAntwort extends Umschlag {
  herkunft: KIFeldherkunftDTO[]
}

// KIVorschlagAntwort ist ein Kartenvorschlag (A7, #1446).
//
// **Er ändert die Karte nicht.** Erst `KIVorschlagUebernehmen` schreibt.
// KIVorschlagsbezug ist, was für einen Vorschlag gebraucht wird (A13.4,
// #1452).
//
// **Ein Verbund und keine fünf Zeichenketten**: mit Belegfrage und Freigabe
// wären es sieben, und wer zwei davon vertauscht, merkt es an keiner Stelle.
export interface KIVorschlagsbezug {
  anbieter: string
  alias: string

  // auftrag ist, was das Modell tun soll.
  auftrag: string

  // nachweis ist die Freigabe eines Menschen — leer heißt: keine (#1539).
  // Ein Modell auf diesem Rechner braucht keine; jedes andere schon.
  nachweis: string

  // belegfrage geht an den Suchindex — **leer heißt: gar nicht fragen**.
  //
  // **Wo sie steht, sieht das Modell dieselben Belege wie eine
  // Einschätzung** — aus demselben Index, mit demselben Manifest.
  belegfrage: string

  // freigegeben nennt die **privaten** Korpora, die dabei mitgesucht werden
  // dürfen. Leer heißt: keine.
  freigegeben: string[]

  // nur nennt, **wonach** gefragt wird — fehlt es, alle Felder. Mit
  // `anbieter: 'vorrang'` machen Name und Flavortext eine kleine Aufgabe für
  // Apples Gerätemodell (AI-05, #2042).
  nur?: string[]
}

export interface KIVorschlagAntwort extends Umschlag {
  kennung: string

  // anbieter und modell stehen beide da — das Issue verlangt es: „Provider-
  // /Modellname im Vorschlag sichtbar" (A8, #1447). Ein Alias allein sagt
  // nicht, wer geantwortet hat, und zwei Anbieter dürfen denselben führen.
  anbieter: string
  modell: string
  felder: KIFeldwertDTO[]

  // befunde sind die Felder, die das Schema **nicht** durchgelassen hat.
  befunde: string[]

  ablauf: string

  // uebernommen ist die Zahl der geschriebenen Felder — null, solange
  // niemand übernommen hat.
  uebernommen: number

  // belege sind die Stellen, die das Modell **gesehen** hat (A13.4, #1452).
  //
  // **Leer, wo keine Belegfrage gestellt wurde.** Sie stehen **neben** den
  // Feldern und nicht als deren Beleg: was das Modell vorschlägt, bleibt
  // eine Vermutung.
  belege: BelegDTO[]

  korpusstaende: KorpusstandDTO[]

  // einschraenkungen sagen, was **nicht** gesucht wurde — etwa der eigene
  // Bestand ohne Freigabe.
  einschraenkungen: string[]

  // weg ist `gewaehlt`, `vorrang` oder `rueckfall` (AI-05, #2042).
  weg: string

  // ausgang ist leer, oder `abgelehnt`: das Modell hat verweigert, und
  // rueckfaelle nennt, wer es versuchen darf — lokal, nie die Cloud.
  ausgang: string
  rueckfaelle: string[] | null
}

// KIBelegbefundDTO ist ein Mangel am Quellenanspruch einer Aussage.
export interface KIBelegbefundDTO {
  // art ist `unbekannt`, `ohnebeleg`, `regelnummer` oder `adresse`.
  art: string

  // stelle ist, was beanstandet wird — die Kennung, die Nummer, die Adresse.
  stelle: string

  meldung: string
}

// KIUebernahmeAntwort sagt, was übernommen wurde.
export interface KIUebernahmeAntwort extends Umschlag {
  felder: number
}
