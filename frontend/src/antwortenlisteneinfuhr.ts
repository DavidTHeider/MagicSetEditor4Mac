// Die Antworten der **Listeneinfuhr** — Werkstück 3 von Phase 4.5
// (#1292).
//
// **Eine eigene Datei neben `antwortensammlung.ts`**, weil sie eine andere
// Quelle beschreibt: die Sammlung kommt aus dem Bestand, die Deutung aus
// Scryfall.
import type { Finish } from './antwortensammlung'
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Listentreffer ist ein Druck, der zu einer Zeile passt.
export interface Listentreffer {
  printing: string
  name: string
  setcode: string
  setname: string
  nummer: string
  seltenheit: string

  // bild ist die Adresse eines kleinen Kartenbilds — leer, wenn es keines gibt.
  bild: string
}

// Listendeutungszeile ist eine gelesene Zeile samt ihren Drucken.
export interface Listendeutungszeile {
  // roh ist die Zeile, wie sie dastand.
  roh: string
  zeilennummer: number

  menge: number
  name: string
  setcode: string
  nummer: string
  finish: Finish

  // treffer sind die passenden Drucke — höchstens zwanzig.
  treffer: Listentreffer[]

  // drucke ist die Zahl der passenden Drucke — auch wenn `treffer` kürzer ist.
  //
  // **Sie steht eigens da** (§10): eine gekappte Liste ohne diese Zahl sähe
  // aus wie eine vollständige.
  drucke: number

  // eindeutig sagt, ob diese Zeile ohne Rückfrage übernommen werden kann.
  //
  // **Sie kommt von drüben und wird hier nicht gerechnet.** Die Regel „genau
  // ein Druck" gehört in eine Schicht; vier Flächen, die sie je selbst
  // rechnen, rechnen sie irgendwann verschieden.
  eindeutig: boolean

  befund: string

  // weitere sagt, dass es mehr Drucke gibt, als eine Seite nennt (G01).
  weitere: boolean

  // fehler ist der Ausfall dieser Zeile — übersetzbar; null heißt, die
  // Quelle hat geantwortet, auch mit „kein Druck" (G01).
  fehler: FehlerDTO | null
}

// Listendeutung ist, was aus einer Liste zu machen wäre.
export interface Listendeutung {
  zeilen: Listendeutungszeile[]
  befunde: string[]

  // alterSekunden ist das Alter der ältesten Antwort.
  alterSekunden: number
}

// Listendeutungsantwort trägt eine Deutung oder einen Fehler.
export interface Listendeutungsantwort extends Umschlag {
  deutung: Listendeutung | null
}

// Listeneinfuhrzeile ist eine Zeile mit entschiedenem Druck.
export interface Listeneinfuhrzeile {
  printing: string
  menge: number
  zustand: string
  sprache: string
  finish: Finish
  name: string
  setcode: string
  nummer: string
}

// Einfuhrbericht sagt, was eingelesen wurde.
export interface Einfuhrbericht {
  angelegt: number[]

  // erhoeht ist die Zahl der Zeilen, deren Menge gewachsen ist.
  erhoeht: number

  // ausgelassen zählt die Zeilen, die eine vorhandene trafen und beim
  // Auslassen stehen blieben — DOM-03.AK-01 (#1698).
  ausgelassen: number

  // stueck ist die Zahl der Stücke, die dazugekommen sind.
  stueck: number

  befunde: string[]
}

// Zusammenfuehrung sagt, was mit einer Zeile geschieht, die die Sammlung
// schon führt — DOM-03.AK-01 (#1698).
export type Zusammenfuehrung = 'addieren' | 'auslassen'

// VorschaupostenDTO ist eine Zeile der Einfuhrvorschau.
export interface VorschaupostenDTO {
  // zeile ist die Stelle in der Eingabe — beim CSV zählt die Kopfzeile mit.
  zeile: number
  bezeichnung: string
  menge: number
  zustand: string
  sprache: string
  wunsch: boolean
  art: 'neu' | 'zusammen' | 'wiederholt'

  // vorher und nachher sind die Mengen der getroffenen Zeile beim Addieren.
  vorher: number
  nachher: number
}

// SammelvorschauDTO sagt, was eine Einfuhr täte.
export interface SammelvorschauDTO {
  posten: VorschaupostenDTO[]
  befunde: string[]
}

// Sammelvorschauantwort trägt eine Vorschau oder einen Fehler.
export interface Sammelvorschauantwort extends Umschlag {
  vorschau: SammelvorschauDTO | null
}

// Einfuhrantwort trägt einen Bericht oder einen Fehler.
export interface Einfuhrantwort extends Umschlag {
  bericht: Einfuhrbericht | null
}

// Ausfuhrtextantwort trägt den ausgegebenen Text oder einen Fehler.
export interface Ausfuhrtextantwort extends Umschlag {
  form: string
  text: string
}
