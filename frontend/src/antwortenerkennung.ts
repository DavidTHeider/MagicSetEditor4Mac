// **Was aus einem Kartenfoto wird** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Eine eigene Datei
//
// **Weil `antwortenmarkt.ts` bei 395 Zeilen stand** — der Schnitt folgt aber
// der Frage und nicht der Zahl (§2): dort steht, was ein Modell über eine
// **vorhandene** Karte sagt, hier, was es auf einem **Foto** liest.
import type { Umschlag } from './umschlag'
import type { KIBelegbefundDTO } from './antwortenmarkt'

// KIBeobachtungDTO ist **eine** gelesene Stelle.
//
// **Nicht „die Karte"**: was das Modell beisteuert, ist der Suchbegriff, und
// den Druck entscheidet die Suche.
export interface KIBeobachtungDTO {
  // art ist eine der sechs: name, setcode, nummer, typzeile, layout,
  // kunstwerk.
  art: string

  // wert ist, was dastand — roh.
  wert: string

  // lesbar sagt, ob die Stelle überhaupt zu sehen war.
  //
  // **Ein eigener Zustand neben dem leeren Wert**: „die Fußzeile ist
  // abgeschnitten" und „dort steht nichts" sind zwei Auskünfte.
  lesbar: boolean
}

// KIBildkandidatDTO ist ein Druck, der zu den Beobachtungen passt.
export interface KIBildkandidatDTO {
  kennung: string
  name: string
  setcode: string
  setname: string
  nummer: string
  seltenheit: string
  bild: string

  // gestuetzt sind die Evidenzarten, die **diesen** Kandidaten stützen.
  //
  // **Je Kandidat und nicht einmal für alle**: der Name passt auf jeden, das
  // Set nur auf einige — daran sieht man, was die Auswahl noch offenlässt.
  gestuetzt: string[]
}

// KIErkennungAntwort ist, was aus dem Bild wurde.
//
// **Sie legt nichts an.** Weder Druckbindung noch Sammlungsmenge; die Auswahl
// trifft ein Mensch.
export interface KIErkennungAntwort extends Umschlag {
  beobachtungen: KIBeobachtungDTO[]

  // nichtGelesen sind die Stellen, zu denen nichts zu sagen war.
  nichtGelesen: string[]

  befunde: KIBelegbefundDTO[]

  kandidaten: KIBildkandidatDTO[]

  // drucke ist die Zahl der passenden Drucke — **die Unsicherheit**.
  //
  // **Gezählt statt geschätzt**: eine Karte, auf die 72 Drucke passen, ist
  // nicht zu 14 % erkannt — sie ist erkannt, und der Druck ist offen.
  drucke: number

  // weitere sagt, dass es hinter der Seite weitergeht.
  weitere: boolean

  // eindeutig heißt: genau ein Druck passt, und dahinter liegt nichts.
  //
  // **Es heißt nicht „übernimm ihn".**
  eindeutig: boolean

  alter: string
  anbieter: string
  modell: string

  // abgebrochen sagt, dass niemand eine Datei gewählt hat.
  //
  // **Ein eigener Zustand neben der leeren Antwort**: „nichts gelesen" und
  // „gar nicht gefragt" sind zwei Auskünfte. Ohne ihn stünde nach einem
  // geschlossenen Dialog „kein Druck passt" da.
  abgebrochen: boolean
}

// KIErkennungsbezug ist, was für eine Erkennung gebraucht wird.
export interface KIErkennungsbezug {
  // pfad ist die Bilddatei auf diesem Rechner.
  //
  // **Leer heißt am Fenster: frag nach.** Die Bindung öffnet dann den
  // Dateidialog; an den Flächen ohne Fenster ist leer ein Fehler.
  pfad: string

  anbieter: string
  alias: string

  // nachweis ist die Freigabe eines Menschen — leer heißt: keine.
  nachweis: string
}
