// Die Antworten der Fassade — je Bindung eine Hülle um ihren Inhalt.
//
// **Eine eigene Datei, weil `api.ts` an der Zeilengrenze steht** (§2). Der
// Schnitt folgt der Frage: dort steht, **wie** das Frontend das Backend
// erreicht, hier **was** dabei herauskommt.
import type { Umschlag } from './api'
import type {
  BeispielsetDTO,
  FeldbeschreibungDTO,
  KarteDTO,
  KartenfeldwertDTO,
  KartenzeileDTO,
  PackartDTO,
  SetDTO,
  TrefferDTO,
  VorbehaltDTO,
  VorschauDTO,
  WertDTO,
} from './formen'
import type {
  PalettenEintragDTO,
  SchluesselwortDTO,
  WortlistenbereichDTO,
} from './formenworte'
import type { SetkopfDTO } from './formensetkopf'

export interface SetsAntwort extends Umschlag {
  sets: SetDTO[]
}

// KartenzeilenAntwort ist die Liste — **nie `null`** (§3).
export interface KartenzeilenAntwort extends Umschlag {
  zeilen: KartenzeileDTO[]
}

// EinzelkarteAntwort trägt **eine** Karte samt ihren Feldwerten (#891).
export interface EinzelkarteAntwort extends Umschlag {
  karte: KarteDTO
}

// KartenfeldwerteAntwort ist **ein** Feld über ein ganzes Set (#891).
//
// **Nur Karten, die das Feld haben, stehen darin.** Ein leerer Wert sagte, das
// Feld sei leer, und das ist etwas anderes als „gibt es nicht".
export interface KartenfeldwerteAntwort extends Umschlag {
  werte: KartenfeldwertDTO[]
}

export interface KartenAntwort extends Umschlag {
  karten: KarteDTO[]
}

// AnlegeAntwort trägt die Kennung einer neu angelegten Karte (#562).
export interface AnlegeAntwort extends Umschlag {
  karteID: string
}

export interface FeldbeschreibungenAntwort extends Umschlag {
  felder: FeldbeschreibungDTO[]
}

export interface SymbolpaletteAntwort extends Umschlag {
  eintraege: PalettenEintragDTO[]

  // roh sagt, ob ein Code ohne `<sym>…</sym>` eingefügt wird —
  // `always symbol:` am Feldstil.
  roh: boolean
}

// PackAntwort trägt die Packarten des Spiels hinter einem Set (#1072).
export interface PackAntwort extends Umschlag {
  arten: PackartDTO[]

  // befunde sind die Packarten, deren Filter sich nicht ausrechnen ließ.
  befunde: string[]
}

// PackziehungAntwort trägt die gezogenen Karten (#1072).
export interface PackziehungAntwort extends Umschlag {
  // karten stehen in ihrer **Ziehreihenfolge** — so wie MSE sie legt.
  karten: KarteDTO[]

  befunde: string[]
}

// KonsolenAntwort trägt das Ergebnis eines MSE-Script-Ausdrucks (#1070).
// Die Antworten der Skriptkonsole stehen in `antwortenkonsole.ts` — die Datei
// hier stand an der Zeilengrenze (§2, Werkstück U13 von Phase 5.1, #1424).
export type { KonsolenAntwort, KonsolenEintrag } from './antwortenkonsole'

// Skriptstelle ist ein Ort im Ausdruck — **eins-basiert**, wie der Zerteiler
// ihn meldet.
export interface Skriptstelle {
  zeile: number
  spalte: number
}

// Namensantwort trägt, was ein Ausdruck an dieser Karte nennen darf
// (#1164).
//
// **Aus der wirklichen Umgebung** — die Bibliothek und das, was die Vorlage
// dazulegt. Eine gepflegte Liste liefe an dem Tag auseinander, an dem die
// Bibliothek eine Funktion dazubekommt.
export interface Namensantwort extends Umschlag {
  namen: string[]
}

// SchluesselwoerterAntwort trägt die Schlüsselwörter eines Sets **und** die
// seines Spiels (#1062).
export interface SchluesselwoerterAntwort extends Umschlag {
  woerter: SchluesselwortDTO[]

  // modi sind die Namen der Modi des Spiels — die Auswahl je Schlüsselwort.
  modi: string[]
}

// SchluesselwortAntwort trägt die Kennung eines neu angelegten (#1062).
export interface SchluesselwortAntwort extends Umschlag {
  wortID: string
}

// WortlistenAntwort trägt die Auswahlbereiche eines Feldes (#1074).
export interface WortlistenAntwort extends Umschlag {
  bereiche: WortlistenbereichDTO[]
}

export interface SetangabenAntwort extends Umschlag {
  felder: FeldbeschreibungDTO[]
  werte: WertDTO[]

  // kopf sind die fünf Formatangaben (#1325).
  kopf: SetkopfDTO
}

export interface VorbehalteAntwort extends Umschlag {
  vorbehalte: VorbehaltDTO[]
}

export interface VorschauAntwort extends Umschlag {
  vorschau: VorschauDTO | null
}

// **Ein Vorschlag von Scryfall — und was er ändern würde**
// (#1005).
//
// §16, Regel 6: „Nichts wird still übernommen. Was von außen in ein Kartenfeld
// will, zeigt vorher, was sich ändert, und trägt danach seine Herkunft."
export interface Feldvorschlag {
  feld: string
  alt: string
  neu: string
}

export interface Scryfallvorschlag {
  aenderungen: Feldvorschlag[]
  befunde: string[]
  // herkunft ist die Adresse bei Scryfall — sie gehört an die Karte.
  herkunft: string
  // alterStunden sagt, wie alt die Antwort ist: 0 heißt gerade geholt.
  //
  // **Sie steht dabei, weil ein Vorschlag aus einem drei Tage alten
  // Zwischenspeicher sonst aussieht wie einer von eben** (§16, Regel 5).
  alterStunden: number

  // stand bindet den Vorschlag an die Kartenwerte, gegen die verglichen wurde
  // (#1241).
  //
  // **Er geht beim Übernehmen zurück**; gilt er nicht mehr, wird abgewiesen.
  stand: string
}

export interface ScryfallvorschlagAntwort extends Umschlag {
  vorschlag: Scryfallvorschlag | null
}

export interface SuchAntwort extends Umschlag {
  treffer: TrefferDTO[]

  // gekuerzt sagt, dass es mehr gibt, als hier steht.
  gekuerzt: boolean
}

export interface FelderAntwort extends Umschlag {
  felder: string[]
}

// BeispielsetsAntwort trägt die mitgelieferten Sets.
export interface BeispielsetsAntwort extends Umschlag {
  sets: BeispielsetDTO[]
}

// OeffnenAntwort ist der Ausgang eines Öffnens.
export interface OeffnenAntwort extends Umschlag {
  // angelegt nennt die Sets, die dabei entstanden sind — **auch bei einem
  // Abbruch** (#155): jedes Set einer AllSets-Datei liegt in einer eigenen
  // Transaktion.
  angelegt: string[]

  // unvollstaendig sagt, dass die Quelle nicht zu Ende gelesen wurde.
  unvollstaendig: boolean

  setID: string
  karten: number
  befunde: string[]
  schonOffen: boolean
  abgebrochen: boolean
}

// AusfuhrAntwort ist der Ausgang eines Exports.
export interface AusfuhrAntwort extends Umschlag {
  pfad: string
  karten: number
  befunde: string[]
  abgebrochen: boolean
}

// StandAntwort trägt den Datenstand — SQLites `data_version`.
//
// **Nur *dass* sich etwas geändert hat, nicht *was*** — der Preis steht in
// [0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md).
export interface StandAntwort extends Umschlag {
  stand: number
}

// SpeicherAntwort ist der Ausgang eines Speicherns.
//
// **Befunde und Erfolg schließen einander nicht aus** (§8): die Setdatei steht
// im Bestand und kommt heraus; was daneben lag — Bilder, Symbole —, liegt am
// Herkunftsort und kann fehlen.
export interface SpeicherAntwort extends Umschlag {
  pfad: string
  befunde: string[]
  abgebrochen: boolean
}

// BildnameDTO ist der Dateiname, unter dem eine Karte herauskommt.
export interface BildnameDTO {
  karteID: string
  nummer: number
  datei: string
}

// BildnamenAntwort trägt die Namen eines Laufs.
export interface BildnamenAntwort extends Umschlag {
  namen: BildnameDTO[]
}

// BildAntwort ist der Ausgang eines geschriebenen Bilds.
export interface BildAntwort extends Umschlag {
  // uebersprungen sagt, dass gar nichts geschrieben wurde — `altbehalten` und
  // eine Datei, die schon dalag
  // (#491).
  uebersprungen: boolean

  pfad: string
  breite: number
  hoehe: number
  befunde: string[]
}

// BogenAntwort ist der Ausgang eines gelegten Druckbogens.
export interface BogenAntwort extends Umschlag {
  pfad: string
  seiten: number
  karten: number
  befunde: string[]
  abgebrochen: boolean
}

// OrdnerAntwort trägt ein gewähltes Verzeichnis.
export interface OrdnerAntwort extends Umschlag {
  ordner: string
  abgebrochen: boolean
}

// HerkunftAntwort sagt, wessen Inhalte ein Bezug holen würde.
//
// **Angaben, kein Satz**: der Satz steht in den Sprachkatalogen (§9).
export interface HerkunftAntwort extends Umschlag {
  quelle: string
  stand: string
}

// BezugAntwort ist der Ausgang eines Paketbezugs.
export interface BezugAntwort extends Umschlag {
  ort: string
  stand: string
  pruefsumme: string
  uebergangen: string[]
}

// **Die Antworten des Symboleditors stehen in `antwortensymbol.ts`** — sie
// werden hier weitergereicht, damit jeder Aufrufer weiterhin eine Stelle hat
// (§2: der Schnitt folgt der Frage, nicht der Zeilenzahl).
export type {
  SymbolAntwort,
  SymbolgriffDTO,
  SymbolgrundformDTO,
  SymbolpunktDTO,
  SymbolpunktweiseDTO,
  SymbolteilDTO,
  SymbolzugDTO,
  SymmetriekopienAntwort,
  SymmetrieweiseDTO,
} from './antwortensymbol'

export type { Merkmalstelle, Farbantwort } from './antwortenfarben'

// **Die Technik und die Diagnose stehen in `antwortentechnik.ts`** — aus
// demselben Grund wie der Verlauf unten: die Datei bleibt unter der
// Zeilengrenze, und `./antworten` bleibt die eine Anlaufstelle (§2).
export type {
  BefehlsvorschauAntwort,
  Befehlseintrag,
  DiagnoseAntwort,
  AgentenbruchstueckDTO,
  SchalterAntwort,
  SchalterDTO,
  SchnittstelleDTO,
  SchnittstellenAntwort,
  SchnittstellenAlleAntwort,
  SchnittstellenausgangDTO,
  NetzschnittstelleDTO,
  NetzschnittstellenAntwort,
  TechnikAntwort,
  Werkzeugzeile,
} from './antwortentechnik'

// **Die Sammlung steht in `antwortensammlung.ts`** und wird hier
// weitergereicht, damit `./antworten` die eine Anlaufstelle bleibt (§2).
export type {
  Paketdateiantwort,
  Paketdateiinhalt,
  Paketdateistand,
  PaketsymbolDTO,
  PaketsymboleAntwort,
  ProbekarteDTO,
  ProbekartenAntwort,
  Skriptbefundantwort,
  SymbolverwendungDTO,
  SymbolverwendungenAntwort,
} from './antwortenvorlagen'

export type {
  Finish,
  ExemplarDTO,
  Sammlungsantwort,
  Exemplarantwort,
  PreisstandDTO,
  Preisreihe,
  Preisantwort,
  Bulkbezugsantwort,
  PortfoliozeileDTO,
  Portfolioantwort,
} from './antwortensammlung'

// **Und was fremde Märkte und Anbieter sagen, steht in `antwortenmarkt.ts`**
// (A4 und A5): eine eigene Datei, weil die Sammlung sonst über die
// Zeilengrenze aus §2 liefe.
export type {
  MarktangebotDTO,
  Marktantwort,
  Anbieterstandantwort,
  Fremdpreisantwort,
} from './antwortenmarkt'

// **Und alles, was die KI zurückgibt, sammelt `antwortenki.ts`** (A12) — aus
// zwei Dateien, damit hier eine Zeile steht.
export type * from './antwortenki'

// **Die Massenvorschau und der Verlauf stehen in `antwortenverlauf.ts`** und
// werden hier weitergereicht, damit `./antworten` die eine Anlaufstelle
// bleibt (§2) — und die Datei unter der Zeilengrenze.
export type {
  FeldvorwertDTO,
  MassenAntwort,
  MassenvorschauAntwort,
  MassenvorschauzeileDTO,
  VerlaufAntwort,
  VerlaufsschrittDTO,
} from './antwortenverlauf'
