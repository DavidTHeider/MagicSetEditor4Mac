// Die Seitenliste — **die einzige Quelle**.
//
// `AGENTS.md` §3 sagt es für die Barrierefreiheitsprüfung, und es gilt genauso
// für die Navigation und die Befehlspalette: eine zu kurze Aufzählung ist grün,
// und niemand sieht es. Wer hier eine Seite einträgt, hat sie in der Navigation,
// in der Palette und in der axe-Prüfung — oder in keinem der drei.
import type { Katalog } from './sprache/kataloge'

// **Zwei Listen, ein Typ.**
//
// Fünf Seiten sprechen über das **offene Set**, vier über die Anwendung — und
// bis #288
// standen alle in derselben Spalte. „Statistik" neben „Einstellungen" sagt
// nicht, dass die eine das offene Set meint und die andere das Programm.
//
// **Der Unterschied steht im Typ und nicht in einem Feld.** Mit zwei getrennten
// Unionen weiß der Übersetzer, dass eine Setlasche ein Set hat: `setID` ist
// dort eine Zeichenkette, und die Wächter „kein Set gewählt", die in fünf
// Laschen standen, sind weg — es waren Zweige, die kein Aufruf betreten konnte
// (§3).
export const anwendungsschluessel = [
  'start',
  'sets',
  'pakete',
  'vorlagen',

  // **Die Sammlung ist eine Anwendungsseite und keine Setlasche** (#1165):
  // was jemand besitzt, hängt an keinem Set
  // ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).
  'sammlung',

  // **Decks stehen neben der Sammlung** (#1293, Vermerk 0061): ein Deck zieht
  // aus mehreren Sets zugleich und hängt an keinem — dieselbe Überlegung wie
  // dort.
  'decks',

  // **Turniere stehen neben den Decks** (#1294, Vermerk 0062): ein Turnier
  // führt über die Zeit, wozu ein Deck der Anlass ist — und es hängt an
  // keinem Set.
  'turniere',

  // **Spielen steht neben den Turnieren** — Werkstück G1 von Phase 5.5
  // (#1458).
  // Eine Partie zieht ihre Decks aus der Deckverwaltung und hängt an keinem
  // Set; dieselbe Überlegung wie dort, und die Reihenfolge der Leiste folgt
  // dem Weg: erst ein Deck, dann ein Turnier oder ein Tisch.
  'spielen',

  // **Handeln steht neben dem Spielen** — Werkstück T9 von Phase 5.6
  // (#1474).
  // Es hängt an keinem Set und an keinem Deck, sondern an einem **Gegenüber**
  // — und es kommt nach dem Spielen, weil man mit dem handelt, mit dem man
  // gespielt hat.
  'handeln',

  // **Nutzer und Profile** — Phase 4.12
  // (#1351).
  // Sie stehen bei den Einstellungen, weil sie derselben Frage gelten: wem
  // gehört, was hier liegt.
  'nutzer',

  'hilfe',
  'einstellungen',
] as const

export const setschluessel = [
  'karten',
  'setangaben',
  'schluesselwoerter',
  'konsole',
  'pack',
  'symbol',
  'stil',
  'suche',

  // **Ersetzen steht neben der Suche** (#1168): jene findet über alle Sets,
  // dieses ändert in **einem**. In MSE2 ist es derselbe Dialog; hier sind es
  // zwei Laschen, weil die Reichweite verschieden ist.
  'ersetzen',
  'statistik',
  'bilder',
  'druckbogen',
  'ausgeben',
] as const

// seitenschluessel ist beides zusammen — **abgeleitet und nicht abgeschrieben**.
export const seitenschluessel = [
  ...anwendungsschluessel,
  ...setschluessel,
] as const

export type Anwendungsschluessel = (typeof anwendungsschluessel)[number]
export type Setschluessel = (typeof setschluessel)[number]
export type Seitenschluessel = (typeof seitenschluessel)[number]

// istSetseite ist die Wache, die den Typ verengt.
// istSeitenschluessel sagt, ob eine Zeichenkette eine Seite dieser Anwendung
// benennt.
//
// **Für Werte, die von außen kommen** — die Menüleiste gibt ihren Schlüssel
// über die Ereignisnaht aus Go (#1168). Ein blindes `as Seitenschluessel`
// machte aus einem Tippfehler dort ein leeres Fenster.
export function istSeitenschluessel(wert: string): wert is Seitenschluessel {
  return (seitenschluessel as readonly string[]).includes(wert)
}

export function istSetseite(seite: Seitenschluessel): seite is Setschluessel {
  return (setschluessel as readonly string[]).includes(seite)
}

// inhaltsflaeche gibt die Klassen, die den Inhalt einer Seite mittig setzen
// und kappen.
//
// **Eine Arbeitsfläche ist kein Fließtext**
// (#697).
//
// Der Rahmen kappte jede Seite bei `6xl` — 1.152 Punkte —, und der Grund
// daneben war richtig: ohne Kappung liefe eine Textzeile über einen 34-Zoll-
// Bildschirm, und niemand fände die nächste wieder.
//
// **Für die Laschen des Sets stimmt er nicht.** Dort liegen Kartenliste,
// Formular und Vorschau nebeneinander, und jede hat ihre eigene Breite; ein
// Fließtext steht auf keiner von ihnen. Gemessen an der laufenden Anwendung
// blieben bei 1.152 für ein Eingabefeld der Karte **176 Punkte**, während auf
// einem 1.728 Punkte breiten Fenster 336 ungenutzt danebenlagen.
//
// **100 rem und nicht `none`**: auch eine Arbeitsfläche hat eine Grenze — die
// Vorbehalteliste und die Statistiktabelle sind Text, und sie stehen auf
// denselben Laschen.
// **Beide Klassen zusammen und nicht nur die Breite**: der Aufrufer setzte
// sie sonst in eine Zeichenkette mit Platzhalter — und `keintextimjsx` kann
// einem Klassennamen nicht ansehen, dass er keine Aufschrift ist.
export function inhaltsflaeche(seite: Seitenschluessel): string {
  return istSetseite(seite) ? 'mx-auto max-w-[100rem]' : 'mx-auto max-w-6xl'
}

export interface Seite {
  schluessel: Seitenschluessel

  // pfad ist die Adresse in der Anwendung, mit führendem Schrägstrich.
  pfad: string

  // kuerzel ist die Taste, die zusammen mit ⌘ auf diese Seite führt; leer, wenn
  // es keine gibt.
  //
  // **Die erste Reihe gehört der Seitenleiste** — Werkstück C von Phase 4.10
  // (#1325).
  // ⌘1 bis ⌘6 sind ihre sechs Bereiche, jede Ziffer an der ersten Seite ihres
  // Bereichs (`bereiche.ts`); ⌘? und ⌘, die beiden macOS-Eigenheiten aus §6.
  //
  // **Die Regel ist ein Satz: ⌘ + Ziffer ist die Leiste, ⌥⌘ + Ziffer sind die
  // Laschen des offenen Sets.** Eine Tastenreihe, die an zwei verschiedene
  // Orte führt, ergäbe keine Regel, die man sich merken kann.
  kuerzel: string

  // zweitkuerzel ist die Ziffer der **zweiten** Reihe — ⌥⌘1 … ⌥⌘0
  // (#1291).
  //
  // **Sie gehört den Laschen eines offenen Sets**, und sie zählt in der
  // Reihenfolge, in der die Laschenleiste sie zeigt.
  //
  // **Zehn Ziffern für dreizehn Laschen.** Drei tragen keine, und welche das
  // sind, steht bei ihnen — es sind die drei, deren Zeile schon vorher sagte,
  // dass niemand sie im Minutentakt aufschlägt. **Das ist der Preis der
  // Regel**, und er gehört ausgesprochen: sie hatten vorher eine.
  //
  // **Warum ⌥⌘ und nicht ⇧⌘**: macOS nimmt ⇧⌘3, ⇧⌘4 und ⇧⌘5 für die
  // Bildschirmfotos, und die erreichen keine Anwendung. Eine Reihe mit drei
  // Löchern ist keine Reihe.
  //
  // **Leer heißt: keine Taste dieser Reihe.** Kein Eintrag trägt beide.
  zweitkuerzel: string
}

export const seiten: readonly Seite[] = [
  // **Die Seitenleiste: ⌘1 bis ⌘8**, in der Reihenfolge, in der sie dasteht.
  { schluessel: 'start', pfad: '/', kuerzel: '1', zweitkuerzel: '' },
  { schluessel: 'sets', pfad: '/sets', kuerzel: '2', zweitkuerzel: '' },
  { schluessel: 'pakete', pfad: '/pakete', kuerzel: '6', zweitkuerzel: '' },
  { schluessel: 'vorlagen', pfad: '/vorlagen', kuerzel: '', zweitkuerzel: '' },
  { schluessel: 'sammlung', pfad: '/sammlung', kuerzel: '3', zweitkuerzel: '' },
  { schluessel: 'decks', pfad: '/decks', kuerzel: '4', zweitkuerzel: '' },
  { schluessel: 'turniere', pfad: '/turniere', kuerzel: '', zweitkuerzel: '' },

  // **⌘9 war für diesen Bereich reserviert, und jetzt gibt es ihn** (G1,
  // #1458). Die Reservierung stand seit Phase 4.10 in `reservierteKuerzel`;
  // sie ist dort weg, weil eine Taste nicht zugleich frei und vergeben sein
  // kann.
  { schluessel: 'spielen', pfad: '/spielen', kuerzel: '', zweitkuerzel: '' },

  // **⌘0 war für diesen Bereich reserviert, und jetzt gibt es ihn** (T9,
  // #1474). Die Reservierung stand seit Phase 4.10 in `reservierteKuerzel`;
  // sie ist dort weg, weil eine Taste nicht zugleich frei und vergeben sein
  // kann — derselbe Satz wie bei ⌘9 eine Zeile höher.
  //
  // **Die letzte Ziffer schließt die Reihe**: zehn Tasten, zehn Seiten der
  // Leiste, und ⌘0 steht auf der Tastatur rechts neben der ⌘9 wie der
  // Bereich rechts neben dem Spielen.
  { schluessel: 'handeln', pfad: '/handeln', kuerzel: '5', zweitkuerzel: '' },

  // **Die Hilfe trägt `?`**
  // (#679):
  // unter macOS ist das die Taste für Hilfe — im Menü steht sie als ⌘? .
  { schluessel: 'hilfe', pfad: '/hilfe', kuerzel: '?', zweitkuerzel: '' },

  // **Und die Einstellungen `,`** — die macOS-Eigenheit aus §6.
  {
    schluessel: 'einstellungen',
    pfad: '/einstellungen',
    kuerzel: ',',
    zweitkuerzel: '',
  },

  // **Nutzer und Profile: ⌘'** — ein Zeichen und keine Zahl, auf Anweisung
  // des Auftraggebers vom 2026-09-04.
  //
  // # Warum der Apostroph
  //
  // Die Zahlen sind vergeben (1 bis 8), und die beiden freien Zeichen, die
  // macOS ohnehin belegt, sind es auch: `,` sind die Einstellungen, `?` ist
  // die Hilfe. Was übrig bleibt, darf mit **keinem** Systemkürzel kollidieren
  // — und das schließt aus: `.` (Abbrechen), `-` und `+` (Zoom), `[` und `]`
  // (Navigation), `/` (in vielen Programmen die Hilfe).
  //
  // **Der Apostroph ist frei**, auf US-Belegung ungeschaltet und auf
  // deutscher über ⇧#. Dass ein geschaltetes Zeichen geht, zeigt `?`: es
  // steht seit jeher für die Hilfe und braucht dort ebenfalls ⇧.
  {
    schluessel: 'nutzer',
    pfad: '/nutzer',
    kuerzel: "'",
    zweitkuerzel: '',
  },

  // **Die Laschen eines offenen Sets: ⌥⌘1 bis ⌥⌘0**, in der Reihenfolge der
  // Laschenleiste (`laschengruppen` unten): bauen, gestalten, das Set,
  // Ausgabe, Technik.
  // **Die Karten tragen ⌘3 und stehen in der Seitenleiste** — Werkstück I von
  // Phase 4.10
  // (#1325).
  //
  // Sie sind die Fläche, auf der gearbeitet wird, und sie lagen hinter dem
  // Weg *Sets → ein Set öffnen → erste Lasche*. Der Arbeitsplatz mit seinen
  // dreizehn Laschen wirkte dadurch überladen, und das Wichtigste steckte
  // darin.
  //
  // **Sie bleiben eine Setseite** — der Rahmen zeigt den Arbeitsplatz, und
  // ohne offenes Set steht dort der Leerzustand, der schon da war. Was sich
  // ändert, ist der Weg dorthin, nicht die Seite.
  { schluessel: 'karten', pfad: '/karten', kuerzel: '', zweitkuerzel: '' },

  { schluessel: 'bilder', pfad: '/bilder', kuerzel: '', zweitkuerzel: '1' },
  { schluessel: 'suche', pfad: '/suche', kuerzel: '', zweitkuerzel: '2' },
  { schluessel: 'ersetzen', pfad: '/ersetzen', kuerzel: '', zweitkuerzel: '3' },
  { schluessel: 'stil', pfad: '/stil', kuerzel: '', zweitkuerzel: '4' },
  { schluessel: 'symbol', pfad: '/symbol', kuerzel: '', zweitkuerzel: '5' },
  {
    schluessel: 'setangaben',
    pfad: '/setangaben',
    kuerzel: '',
    zweitkuerzel: '6',
  },

  // **Die Schlüsselwörter haben ⌥⌘7 zurück** (#1325): seit die Karten in die
  // Seitenleiste gewandert sind, ist eine Ziffer der zweiten Reihe frei — und
  // von den dreien ohne Taste ist dies die, an der jemand wirklich arbeitet.
  {
    schluessel: 'schluesselwoerter',
    pfad: '/schluesselwoerter',
    kuerzel: '',
    zweitkuerzel: '7',
  },

  {
    schluessel: 'statistik',
    pfad: '/statistik',
    kuerzel: '',
    zweitkuerzel: '8',
  },

  // **Das Zufallspack trägt keine** (#1072): ein Werkzeug zum Ausprobieren,
  // keine Ansicht, an der jemand arbeitet. Es hatte ⌥⌘6.
  { schluessel: 'pack', pfad: '/pack', kuerzel: '', zweitkuerzel: '' },

  { schluessel: 'ausgeben', pfad: '/ausgeben', kuerzel: '', zweitkuerzel: '9' },
  {
    schluessel: 'druckbogen',
    pfad: '/druckbogen',
    kuerzel: '',
    zweitkuerzel: '0',
  },

  // **Die Konsole trägt keine** (#1070): ein Werkzeug zum Nachsehen. Sie
  // hatte ⌥⌘5.
  { schluessel: 'konsole', pfad: '/konsole', kuerzel: '', zweitkuerzel: '' },
]

// beschriftung holt den Namen einer Seite aus dem Katalog.
//
// Über den Katalog und nicht über ein Feld an `Seite`: ein Name in dieser Datei
// wäre ein Text im Quelltext, und `keintextimjsx.test.ts` weist ihn ab — zu
// Recht, denn er wäre nicht übersetzbar.
export function beschriftung(
  texte: Katalog,
  schluessel: Seitenschluessel,
): string {
  return texte.seiten[schluessel].titel
}

// nachbarlasche gibt die Seite neben der gegebenen — zyklisch.
//
// **Als eigene Funktion und nicht als Ausdruck in der Komponente.** Sie hat
// zwei Antworten, die niemand am Bildschirm auseinanderhält: „die nächste" und
// „gar keine". Hier lassen sich beide prüfen; in einer Tastenbehandlung wäre
// der zweite Fall ein Zweig, den kein Test betritt.
//
// **`undefined`, wenn `aktuell` nicht in der Reihe steht** — oder die Reihe
// leer ist. Beides heißt: es gibt keinen Nachbarn, und die Pfeiltaste tut
// nichts.
export function nachbarlasche(
  reihe: readonly Seitenschluessel[],
  aktuell: Seitenschluessel,
  richtung: 1 | -1,
): Seitenschluessel | undefined {
  const stelle = reihe.indexOf(aktuell)

  if (stelle < 0) {
    return undefined
  }

  return reihe.at((stelle + richtung + reihe.length) % reihe.length)
}
