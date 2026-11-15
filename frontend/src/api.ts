// Die Fassade zum Go-Teil.
//
// # Warum von Hand und nicht erzeugt
//
// Unter Wails 2 lag daneben ein **Generat** (`wailsjs/`), das beim Bau
// entstand und in einem frischen Klon nicht da war; ein `import` daraus machte
// `tsc` und `vitest` davon abhängig, dass vorher jemand gebaut hat. Wails 3
// erzeugte Module an derselben Stelle — diese Fassade bleibt die eine
// Beschreibung, und `anbindung.ts` ruft über den Namen (ENG-02, 0132).
//
// # Warum jede Bindung einen Umschlag zurückgibt
//
// Wails nimmt bei zwei Rückgabewerten den zweiten **nur dann** als Fehler, wenn
// er Gos `error` erfüllt. Ein strukturierter Fehler tut das nicht — er wäre
// stillschweigend verworfen worden. Und ihn zu einem `error` zu machen hieße,
// ihn als bloße Zeichenkette zu bekommen: Code, Schlüssel und Feldbezug gingen
// verloren, und genau die verlangt §8.
//
// `auspacken` macht daraus wieder eine abgelehnte Zusage — mit dem Fehler als
// Wert, nicht als Text.

import { Bindungsfehler } from './bindungsfehler'
import { auspacken, type Umschlag } from './umschlag'

import type { Anbietergeheimnisapi, Zugangsapi } from './apizugaenge'
import type { Platzapi } from './apiplatz'
import type { Bilderapi } from './apibilder'
import type { Scryfallapi } from './apiscryfall'
import type { Feldapi } from './apifeld'
import type { Bestandsapi } from './apibestand'
import type { Ablageapi } from './apiablage'
import type { Listenapi } from './apiliste'
import type { Markenapi } from './apimarken'
import type { Stilapi } from './apistil'
import type { Ersetzungsapi } from './apiersetzen'
import type { Autoersetzungsapi } from './apiautoersetzung'
import type { FeldAntwort, Kartenapi } from './apikarten'
import type { Diagnoseapi } from './apidiagnose'
import type { Objektmarkenapi } from './apiobjektmarken'
import type { Kontaktapi } from './apikontakte'
import type { Dateiansichtapi } from './apidateiansicht'
import type { Einstellungsapi } from './apieinstellungen'
import type { Gesundheitsapi } from './apigesundheit'
import type { Metrikenapi } from './apimetriken'
import type { Betriebsapi } from './apibetrieb'
import type { Zwischenspeicherapi } from './apizwischenspeicher'
import type { Diagnosepaketapi } from './apidiagnosepaket'
import type { Setzugangapi } from './apisetzugang'
import type { Pflegeapi } from './apipflege'
import type { Netzapi } from './apinetzgruppe'
import type { Indexapi } from './apiindex'
import type { Rechercheapi } from './apirecherche'
import type { Ruecksetzungsapi } from './apiruecksetzung'
import type { Kostenapi } from './apikosten'
import type { Rueckkehrapi } from './apirueckkehr'
import type { Vertraeglichkeitapi } from './apivertraeglichkeit'
import type { Verlaufsapi } from './apiverlauf'
import type { Weiterarbeitsapi } from './apiweiterarbeiten'
import type { Aufmerksamkeitsapi } from './apiaufmerksamkeit'
import type { Befehlsapi } from './apibefehle'

export type { KartenbildAntwort } from './apikarten'
export type { EinfuegeAntwort } from './apiablage'
export type {
  KartenfilterAntwort,
  KartenlisteAntwort,
  ListenspalteDTO,
  ListenzeileDTO,
} from './apiliste'
export type {
  ErsetzungsAntwort,
  ErsetzungsfrageDTO,
  ErsetzungsstelleDTO,
} from './apiersetzen'
export type { AutoersetzungAntwort } from './apiautoersetzung'
export type { UmschlagAntwort, UmschlagobjektDTO } from './apiumschlag'
export type { NetzstandAntwort, NetzfeldDTO } from './apinetz'
export type { DienstAntwort, DienstZustand } from './apidienste'
import type { Partieapi } from './apipartie'
import type { Partieverwaltungapi } from './apipartieverwaltung'
import type { Regelapi } from './apiregeln'
import type { Texturapi } from './apitextur'
import type { Tischbildapi } from './apitischbild'

// **Die Partie als Stern und nicht als Aufzählung** — zehn Namen von einer
// Fläche, und die Aufzählung hob diese Datei über die Zeilengrenze aus §2.
// Derselbe Weg wie bei `apileute` eine Zeile weiter unten.
export type * from './apipartie'
export type * from './apipartieverwaltung'
export type * from './apileihe'
export type * from './apinebenbei'
export type * from './apitextur'
import type { Symbolapi } from './apisymbol'
import type { Symbolbibliotheksapi } from './apisymbolbibliothek'
import type { Sprachapi } from './apisprache'
import type { Schluesselwortapi } from './apischluesselwoerter'
import type { Wortlistenapi } from './apiwortlisten'
import type { Deckapi } from './apidecks'
import type { Deckeinfuhrapi } from './apideckeinfuhr'
import type { Turnierapi } from './apiturniere'
import type { Sammlungsapi } from './apisammlung'
import type { Aktualisierungsapi } from './apiaktualisierung'
import type { Ausfuhrapi } from './apiausfuhr'
import type {
  BeispielsetDTO,
  BewegungAntwort,
  FehlerDTO,
  FeldbeschreibungDTO,
  KarteDTO,
  KartenzeileDTO,
  PackartDTO,
  SetDTO,
  TrefferDTO,
  VorbehaltDTO,
  VorschauDTO,
  WertDTO,
} from './formen'

// **Weitergereicht, damit `../api` die eine Anlaufstelle bleibt.** Der Schnitt
// nach `formen.ts` und `formensetkopf.ts` ist eine Frage der Datei, keine der
// Aufrufer; wer jede Seite umstellte, machte aus einer Zeilengrenze eine
// Umbauaktion.
export type { SetkopfDTO } from './formensetkopf'
export type { HerkunftDTO } from './apiherkunft'
export type { Deckuebernahmeantwort } from './apidecks'
export { keineHerkunft } from './apiherkunft'
export type {
  PalettenEintragDTO,
  SchluesselwortDTO,
  WortlistenbereichDTO,
  WortmenueeintragDTO,
} from './formenworte'
export type {
  BeispielsetDTO,
  BewegungAntwort,
  FehlerDTO,
  FeldbeschreibungDTO,
  KarteDTO,
  KartenzeileDTO,
  PackartDTO,
  SetDTO,
  TrefferDTO,
  VorbehaltDTO,
  VorschauDTO,
  WertDTO,
}

export type { AuswahlDTO, ZuschnittAntwort } from './formen'
export type {
  Rueckspielantwort,
  SetAnlegeAntwort,
  Sicherungsantwort,
  Bilderlage,
  Sicherungsauswahl,
  Startbefund,
  Verschluesselungsantwort,
  FassungsAntwort,
  LadeAntwort,
} from './antwortenbestand'

// Bindungsfehler — **weitergereicht**, damit `../api` die eine Anlaufstelle
// bleibt. Der Schnitt nach `bindungsfehler.ts` ist eine Frage der Datei (§2,
// Zeilengrenze), keine der Aufrufer.
export { Bindungsfehler }

// Umschlag und auspacken — **weitergereicht**, damit `../api` die eine
// Anlaufstelle bleibt. Der Schnitt nach `umschlag.ts` ist eine Frage der Datei
// (§2, Zeilengrenze), keine der Aufrufer.
export { auspacken, type Umschlag }

export type SchreibAntwort = Umschlag

// Api ist die Fassade, wie das Frontend sie benutzt — schon ausgepackt.
//
// **Die Ausfuhrwege stehen in `apiausfuhr.ts`** (§2, Zeilengrenze); ihre
// Gestalten werden von hier weitergereicht, damit `../api` die eine
// Anlaufstelle bleibt.
export type { VorlagenausfuhrDTO, VorlagenDTO } from './vorlagen'
// **Ein Stern statt fünfzig Namen** — `antworten.ts` ist **eine** Fläche, und
// die Aufzählung hob diese Datei über die Zeilengrenze aus §2. Derselbe Weg
// wie bei `apileute` und `apipartie` weiter unten.
export type * from './antworten'
export type * from './antwortenstatistik'
import type {
  AnlegeAntwort,
  FeldbeschreibungenAntwort,
  FelderAntwort,
  MassenAntwort,
  MassenvorschauAntwort,
  FeldvorwertDTO,
  SetangabenAntwort,
  SuchAntwort,
  VorbehalteAntwort,
} from './antworten'
export type { BezugAntwort, VorschauAntwort } from './antworten'
export type {
  AbhaengigkeitDTO,
  AbhaengigkeitenAntwort,
  Anlegbarkeit,
  OrteAntwort,
  PaketAntwort,
  BezugsquellenAntwort,
  PaketDTO,
  PaketeAntwort,
  Paketherkunftsart,
  PaketplanAntwort,
  PaketplanDTO,
  Paketvorgang,
  QuelleDTO,
  UebernahmeAntwort,
} from './pakete'
import type { Nutzerapi } from './apinutzer'
import type { Tokenapi } from './apitoken'
import type { Vorgabenapi } from './apivorgaben'
import type { Freigabenapi } from './apifreigaben'
import type { Faktorapi } from './apifaktor'
import type { Sitzungsapi } from './apisitzung'
import type { Vorschauapi } from './apivorschau'
import type { Statistikapi } from './apistatistik'
import type { Paketapi } from './apipakete'
export type * from './apileute'
export type { MarkenAntwort } from './marken'
export type { EinstellungDTO, EinstellungenAntwort } from './einstellungen'
export type {
  EinfuhrAntwort,
  Einfuhrbefund,
  Einfuhrzeile,
  EinfuhrvorschauAntwort,
  Einstellungsbericht,
  Einstellungssicht,
} from './einstellungsaustausch'
export type { PlatzpostenDTO } from './apiplatz'
export type {
  FolgenAntwort,
  FolgenDTO,
  StilDTO,
  StileAntwort,
  StiloptionDTO,
  StiloptionenAntwort,
  WahlbildDTO,
  WahlbilderAntwort,
  Wahlmoeglichkeit,
} from './stilwahl'
export interface Api
  extends
    Bilderapi,
    Scryfallapi,
    Wortlistenapi,
    Schluesselwortapi,
    Feldapi,
    Symbolapi,
    Symbolbibliotheksapi,
    Sprachapi,
    Bestandsapi,
    Aktualisierungsapi,
    Ausfuhrapi,
    Deckapi,
    Deckeinfuhrapi,
    Turnierapi,
    Partieapi,
    Partieverwaltungapi,
    Regelapi,
    Texturapi,
    Tischbildapi,
    Sammlungsapi,
    Kartenapi,
    Ablageapi,
    Listenapi,
    Markenapi,
    Stilapi,
    Ersetzungsapi,
    Autoersetzungsapi,
    Zugangsapi,
    Anbietergeheimnisapi,
    Platzapi,
    Paketapi,
    Nutzerapi,
    Tokenapi,
    Vorgabenapi,
    Freigabenapi,
    Faktorapi,
    Sitzungsapi,
    Vorschauapi,
    Statistikapi,
    Einstellungsapi,
    Gesundheitsapi,
    Metrikenapi,
    Betriebsapi,
    Zwischenspeicherapi,
    Diagnosepaketapi,
    Setzugangapi,
    Pflegeapi,
    Ruecksetzungsapi,
    Indexapi,
    Netzapi,
    Rechercheapi,
    Kostenapi,
    Rueckkehrapi,
    Vertraeglichkeitapi,
    Verlaufsapi,
    Weiterarbeitsapi,
    Aufmerksamkeitsapi,
    Befehlsapi,
    Diagnoseapi,
    Dateiansichtapi,
    Objektmarkenapi,
    Kontaktapi {
  // KarteAnlegen hängt eine leere Karte an ein Set und gibt ihre Kennung —
  // wer eine Karte anlegt, will sie aufschlagen
  // (#562).
  KarteAnlegen: (setID: string) => Promise<AnlegeAntwort>

  // KarteLoeschen entfernt eine Karte samt ihren Feldwerten
  //
  // trotzdem bestätigt ein Löschen, das einen Deckplatz mitnimmt (#1369).
  // (#621).
  //
  // **Zurückzunehmen mit ⌘Z** — der Verlauf trägt seit #621 eine eigene Art
  // für ganze Karten, und die Rücknahme bringt dieselbe Kennung zurück. Ohne
  // sie wäre das hier Datenverlust mit einem Klick (#562).
  //
  // **Jede Kartenänderung nennt das Set** — seit
  // #808.
  // Nicht, weil das Fenster sich vertun könnte, sondern weil es das
  // nachweislich tut: die vier Regeln aus §7 stehen in der Leitlinie, weil
  // eine stehengebliebene Kennung nach einem Setwechsel vorgekommen ist.
  KarteKopieren: (setID: string, karteID: string) => Promise<AnlegeAntwort>
  KarteLoeschen: (
    setID: string,
    karteID: string,
    trotzdem: boolean,
  ) => Promise<SchreibAntwort>
  KartenTauschen: (
    setID: string,
    eineID: string,
    andereID: string,
  ) => Promise<SchreibAntwort>

  // Kartenfelder gibt die **Form** jedes Kartenfeldes — Art, Mehrzeiligkeit,
  // Wahlen, Beschreibung. Ohne sie stellte der Editor jedes Feld als
  // einzeiliges Textfeld dar, auch den Regeltext und die Seltenheit
  // (#560).
  //
  // **Getrennt von den Werten**: die ändern sich bei jedem Tastendruck, die
  // Beschreibungen nicht.
  Kartenfelder: (setID: string) => Promise<FeldbeschreibungenAntwort>

  // Setangaben gibt die Angabenfelder eines Sets samt ihren Werten — Titel,
  // Setcode, Copyright, Symbol
  // (#567).
  //
  // **Beides in einer Antwort**, anders als bei den Karten: ein Set ist **ein**
  // Gegenstand, und es gibt keine Liste, durch die die Beschreibungen
  // mitreisen müssten.
  Setangaben: (setID: string) => Promise<SetangabenAntwort>
  // **`bekannt`** wie bei `FeldSetzen` (U02b, #2179).
  SetwertSetzen: (
    setID: string,
    feld: string,
    wert: string,
    bekannt: readonly string[] | null,
  ) => Promise<SchreibAntwort>
  Vorbehalte: (setID: string) => Promise<VorbehalteAntwort>

  // BestandSichern schreibt eine Sicherung der ganzen Datenbank
  // (#799).
  //

  // FeldSetzen schreibt ein Kartenfeld. **`bekannt`** sind die Werte, die der
  // Entwurf von diesem Feld kennt — Ausgangswert, selbst geschickte, gesehene.
  // Steht im Bestand ein anderer, hat jemand anderes geschrieben, und es ist
  // ein Konflikt am Feld (`fremderWert`, U02b, #2179); `null` schreibt ohne
  // Bedingung.
  FeldSetzen: (
    setID: string,
    karteID: string,
    feld: string,
    wert: string,
    bekannt: readonly string[] | null,
  ) => Promise<FeldAntwort>

  FelderVorschau: (
    setID: string,
    karteIDs: string[],
    feld: string,
    wert: string,
  ) => Promise<MassenvorschauAntwort>
  FelderSetzenMitErwartung: (
    setID: string,
    erwartet: FeldvorwertDTO[],
    feld: string,
    wert: string,
  ) => Promise<MassenAntwort>
  Suchen: (text: string, feld: string, setID: string) => Promise<SuchAntwort>
  Suchfelder: (setID: string) => Promise<FelderAntwort>
}

// **Die Fassade steht in `fassade.ts`** und wird hier weitergereicht (§2).
export { api, fassade } from './fassade'
