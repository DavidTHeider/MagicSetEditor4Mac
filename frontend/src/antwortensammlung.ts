// Die Antworten der **Sammlung**.
//
// **Eine eigene Datei**, weil `antworten.ts` an der Zeilengrenze aus §2
// steht — und weil der Schnitt der Frage folgt: dort die Karten und Sets,
// hier der **Besitz**.
import type { HerkunftDTO } from './apiherkunft'
import type { Umschlag } from './umschlag'

// Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// Ein Exemplar hängt an einem **Druck** und an keinem Set: was jemand besitzt,
// ist ein Stück Papier, eine Karte ein Entwurf.

// Finish ist die physische Ausführung eines Drucks — Werkstück T2 von
// Phase 5.6 (#1467, 0119).
//
// **Drei und nicht zwei**: ein Wahrheitswert hielte eine geätzte Karte für
// ein Foil — mit dessen Preis.
export type Finish = 'nonfoil' | 'foil' | 'etched'

// ExemplarDTO ist ein Exemplar der eigenen Sammlung.
export interface ExemplarDTO {
  id: number

  // fassung ist die Fassung, die gelesen wurde.
  //
  // **Sie geht beim Ändern zurück.** Stimmt sie dann nicht mehr, hat jemand
  // anderes geschrieben — und statt die fremde Änderung zu überschreiben,
  // kommt ein Konflikt (§7, Regel 2). Angestrichen vom Zweitblick über
  // Phase 3.4.
  fassung: number
  printing: string

  // druckID ist die Zeile des Drucks in **diesem** Bestand — Werkstück A2 von
  // Phase 5.3 (#1441, 0086): als Zeichenkette wie jede Kennung, leer für
  // keine. vorlaeufig sagt, dass der Druck **keinen Alias** hat — die Zeile
  // kam aus einer Liste ohne Scryfall-Kennung; Preise und Bilder bleiben aus,
  // bis jemand sie auflöst. Ein Zustand, kein Fehler.
  druckID: string
  vorlaeufig: boolean

  menge: number
  zustand: string
  sprache: string

  // finish ist die physische Ausführung: `nonfoil`, `foil` oder `etched`
  // — Werkstück T2 von Phase 5.6 (#1467, 0119). **Leer kommt hier nie an**:
  // die Naht beantwortet den Nullwert und schickt das Wort.
  finish: Finish
  erwerbsdatum: string

  // erwerbspreis steht als **Text** da — `"4.17"`.
  //
  // Im Bestand ist er eine Ganzzahl in der kleinsten Einheit; JSON kennt nur
  // `number`, und das ist hier ein `double`. Was ankommt, ist die Form, in der
  // jemand ihn tippt.
  erwerbspreis: string

  erwerbswaehrung: string
  notiz: string

  // name, setcode und nummer sind der **lesbare Bezug** des Drucks
  // (#1200).
  //
  // **Der letzte gute Stand und nicht die Identität** — die ist `printing`.
  // Ohne Netz steht hier, was zuletzt bekannt war (§16, Regel 5).
  name: string
  setcode: string
  nummer: string

  // lagerort ist, wo das Stück Papier liegt — Freitext
  // ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
  lagerort: string

  // wunsch sagt, dass diese Zeile **kein Besitz** ist (0057).
  //
  // Sie zählt nicht ins Portfolio und nicht in »besitze ich diese Karte?«.
  wunsch: boolean

  // bezeichnung ist der Bezug, wie man ihn liest — „Lightning Bolt · M11
  // #149" — oder die Kennung, wenn nichts bekannt ist.
  //
  // **Gerechnet in Go und nicht hier zusammengesetzt**: drei Flächen, die sie
  // je für sich bauen, liefen auseinander (§16).
  bezeichnung: string

  // Woher diese Zeile kam — Werkstück 12 von Phase 4.12
  // (#1351).
  //
  // **Sie reist mit jeder Zeile**, weil sie über einen `LEFT JOIN` kommt und
  // nichts kostet. Eine leere `art` heißt: hier entstanden.
  herkunft: HerkunftDTO
}

// Sammlungsantwort trägt die Exemplare.
export interface Sammlungsantwort extends Umschlag {
  exemplare: ExemplarDTO[]
}

// Sammlungsfensterantwort trägt einen Ausschnitt und beide Zahlen —
// Werkstück #1500.
export interface Sammlungsfensterantwort extends Umschlag {
  exemplare: ExemplarDTO[]

  // gesamt ist die Zahl der Zeilen dieser **Sicht** — sie steht in
  // `aria-rowcount`.
  gesamt: number

  // insgesamt ist die Zahl der Exemplare ohne Sicht — „N von M".
  insgesamt: number

  // stueck ist die Summe der Mengen ohne Sicht — U15 (#2160): die Startseite
  // zählt damit, statt die ganze Liste zu holen.
  stueck: number
}

// Exemplarantwort trägt die Kennung eines angelegten Exemplars.
export interface Exemplarantwort extends Umschlag {
  id: number

  // vorhanden ist die Zeile, die es schon gibt, wenn die Aufnahme an ihr
  // scheiterte — DOM-03, S04.NAT-01 und 03 (#1698). **Wahlfrei**, weil eine
  // ältere Naht oder eine belegte Fassade sie nicht kennt.
  vorhanden?: ExemplarDTO | null
}

// PreisstandDTO ist ein Preis zu einem Zeitpunkt.
//
// **Quelle, Zeitpunkt und Alter reisen mit.** Die Quelle sagt über ihre Preise
// ausdrücklich, dass sie für keinen einsteht — eine Zahl ohne diese drei wäre
// eine Behauptung, für die niemand geradesteht.
export interface PreisstandDTO {
  printing: string
  // druckID ist die Zeile des Drucks in diesem Bestand (A2, #1441), leer für keine.
  druckID: string

  // **Die sieben Teile der Reihe** (A3, #1442) — sie sagen, **welche** Frage
  // dieser Preis beantwortet. `markt` ist `paper` oder `mtgo`, `art` ist
  // `retail` (was ein Anbieter verlangt) oder `buylist` (was er zahlt).
  quelle: string
  markt: string
  art: string
  waehrung: string
  finish: string
  zustand: string
  sprache: string

  // lage ist `genannt`, `weggefallen` oder `ungeprueft`.
  lage: string

  // wert ist leer, wo die Lage nicht `genannt` ist — dort steht kein Preis,
  // und eine `"0.00"` sähe aus wie „umsonst".
  wert: string

  // beobachtet ist der Tag, den die **Quelle** nennt — leer, wo sie keinen
  // nennt (Scryfall nennt keinen).
  beobachtet: string
  erhoben: string
  alterStunden: number
}

// Preisreihe benennt, welche Reihe ein Verlauf meint (A3, #1442).
//
// **Ohne Vorgaben**: „wenn nichts dasteht, dann Scryfall" wäre eine
// Preisstrategie, und genau die soll es nicht geben.
export interface Preisreihe {
  printing: string
  quelle: string
  markt: string
  art: string
  waehrung: string
  finish: string
  zustand: string
  sprache: string
}

// Preisantwort trägt die Preisstände eines Drucks.
export interface Preisantwort extends Umschlag {
  staende: PreisstandDTO[]

  // befunde nennt Preise, die nicht zu deuten waren (§8).
  befunde: string[]
}

// Bulkbezugsantwort trägt, was aus einem Bulkbezug wurde (A3, #1442).
export interface Bulkbezugsantwort extends Umschlag {
  // quelle ist die Familie, die gefragt wurde — heute nur `mtgjson`.
  quelle: string

  // umfang ist `heute` oder `neunzigTage`.
  umfang: string

  // drucke ist die Zahl der Drucke, nach denen gefragt wurde.
  drucke: number

  // staende ist die Zahl der abgelegten Preisstände.
  staende: number

  // befunde nennt Angaben, die nicht zu deuten waren (§8).
  befunde: string[]
}

// PortfoliozeileDTO ist die Summe einer Währung.
export interface PortfoliozeileDTO {
  waehrung: string
  wert: string
  exemplare: number
  aeltesterPreis: string
  alterStunden: number
  quelle: string
}

// Portfolioantwort trägt die Übersicht über die Sammlung.
export interface Portfolioantwort extends Umschlag {
  zeilen: PortfoliozeileDTO[]

  // ohnePreis ist die Zahl der Exemplare, für die kein Preis vorliegt.
  //
  // **Sie steht eigens da.** Eine Sammlung von 300 Karten, von denen 280 keinen
  // Preis haben, hat eine Summe, die nichts aussagt.
  ohnePreis: number
}

// BesitzDTO sagt, was von einem Druck da ist (#1292, Vermerk 0057).
export interface BesitzDTO {
  printing: string

  // stueck ist die Zahl der **besessenen** Stücke.
  stueck: number

  // gewuenscht ist die Zahl der gewünschten Stücke — getrennt gezählt.
  gewuenscht: number

  // besitzt kommt von drüben und wird hier nicht gerechnet: die Regel „ein
  // Wunsch ist kein Besitz" gehört in eine Schicht.
  besitzt: boolean

  zeilen: ExemplarDTO[]
}

// Besitzantwort trägt den Besitzstand eines Drucks.
export interface Besitzantwort extends Umschlag {
  besitz: BesitzDTO | null
}

// LagerortDTO ist ein Lagerort samt dem, was dort liegt.
export interface LagerortDTO {
  // lagerort ist der Freitext — leer heißt „ohne Angabe".
  lagerort: string
  zeilen: number
  stueck: number
}

// Lagerortantwort trägt die Lagerorte.
export interface Lagerortantwort extends Umschlag {
  lagerorte: LagerortDTO[]
}

// DublettenDTO ist ein Exemplar, von dem mehr da ist als gewollt.
export interface DublettenDTO {
  exemplar: ExemplarDTO

  // ueberzaehlig ist, wie viele über der behaltenen Zahl liegen.
  ueberzaehlig: number
}

// Dublettenantwort trägt die überzähligen Exemplare.
export interface Dublettenantwort extends Umschlag {
  dubletten: DublettenDTO[]
}

// MassenzeileDTO ist ein Exemplar samt seiner gelesenen Fassung (#1292).
//
// **Die Fassung geht mit** (§7, Regel 2): ein Massenvorgang über bloße
// Kennungen überschriebe jede Änderung seit dem Laden — fünfzigfach auf einmal.
export interface MassenzeileDTO {
  id: number
  fassung: number
}

// MassenaenderungDTO sagt, was an den genannten Zeilen zu ändern ist.
//
// **`null` heißt „unverändert".** Ein leerer Lagerort ist eine gültige
// Angabe — „nimm die Ortsangabe weg" —, und ohne diese Unterscheidung wäre sie
// von „lass ihn in Ruhe" nicht zu trennen.
export interface MassenaenderungDTO {
  zeilen: MassenzeileDTO[]
  menge: number | null
  zustand: string | null
  lagerort: string | null
  wunsch: boolean | null
}

// MassenberichtDTO sagt, was geschehen ist.
export interface MassenberichtDTO {
  geaendert: number
  befunde: string[]

  // vorher nennt je geänderter Zeile den Stand davor samt neuer Fassung —
  // damit sich die Änderung zurücknehmen lässt (DOM-03.AK-03, #1698).
  vorher: MassenstandDTO[]
}

// MassenstandDTO ist eine Zeile vor der Änderung, mit der Fassung danach.
export interface MassenstandDTO {
  id: number
  fassung: number
  menge: number
  zustand: string
  lagerort: string
  wunsch: boolean
}

// Massenantwort trägt den Bericht eines Massenvorgangs.
export interface Massenantwort extends Umschlag {
  bericht: MassenberichtDTO | null
}

// PreisschubDTO sagt, was ein Schub geholt hat und was bleibt.
export interface PreisschubDTO {
  geholt: number

  // offen ist die Zahl der Drucke, die DANACH noch fehlen — der Fortschritt.
  offen: number

  // fehlgeschlagen ist die Zahl der Versuche, die nichts eingetragen haben.
  //
  // **Sie steht neben `offen`, damit die keine Auskunft erfinden muss**: ein
  // Fehlschlag ließ `offen` sinken, obwohl der Druck offen blieb, und ein
  // ganzer Schub voller Fehlschläge meldete »fertig« (Nachlauf über Phase
  // 4.5, Befund 7).
  fehlgeschlagen: number

  // fertig kommt von drüben und wird hier nicht gerechnet.
  fertig: boolean

  // steht sagt, dass dieser Schub nichts bewegt hat, obwohl Arbeit bleibt —
  // der Satz, an dem ein Aufrufer aufhört zu rufen. Kommt ebenfalls von
  // drüben.
  steht: boolean

  befunde: string[]
}

// Preisschubantwort trägt den Schub oder einen Fehler.
export interface Preisschubantwort extends Umschlag {
  schub: PreisschubDTO | null
}
