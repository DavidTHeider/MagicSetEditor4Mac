// **Abgestufte Rücksetzung** — Werkstück O4 von Phase 5.2
// (#1428).
// Die Gestalt kommt aus `internal/app/ruecksetzung.go`.
import type { NachweisAntwort } from './apikosten'
import type { Umschlag } from './umschlag'

// BetroffenesDTO ist eine Gruppe, die eine Rücksetzung anfasst.
export interface BetroffenesDTO {
  // was ist eine Marke — der Satz dazu steht im Sprachkatalog.
  was: string

  anzahl: number

  // namen nennen ein paar davon — höchstens fünf.
  namen: string[]
}

export interface HindernisDTO {
  art: string
  stelle: string
}

export interface RuecksetzungsvorschauAntwort extends Umschlag {
  // revision ist die Marke über das, was hier steht (Befund 12).
  revision: string

  art: string
  betroffen: BetroffenesDTO[]
  hindernisse: HindernisDTO[]

  // bereit ist gerechnet und kommt von der Go-Seite — zwei Stellen, die
  // „keine Hindernisse" bilden, laufen auseinander.
  bereit: boolean

  // nameNoetig sagt, dass der Name des Ziels als Bestätigung dazugehört —
  // nicht dasselbe wie sicherungEmpfohlen: die ganze Anwendung löscht am
  // meisten und hat keinen Namen, den man abtippen könnte.
  nameNoetig: boolean

  nachweisNoetig: boolean
  sicherungEmpfohlen: boolean
}

export interface RuecksetzungsbilanzAntwort extends Umschlag {
  art: string
  zurueckgesetzt: number

  // beiseite ist der Ort des alten Bestands — nur beim Neuanfang.
  beiseite: string

  // bestandsepoche ist die Epoche des Bestands danach (J05, #2269).
  bestandsepoche?: number
}

export interface Ruecksetzungsapi {
  Ruecksetzungsarten: () => Promise<string[]>
  Ruecksetzungsvorschau: (
    art: string,
    profilID: string,
    nutzerID: string,
  ) => Promise<RuecksetzungsvorschauAntwort>
  Zuruecksetzen: (
    art: string,
    profilID: string,
    nutzerID: string,
    name: string,
    nachweis: string,

    // revision ist die Marke aus der Vorschau — sie bindet die Ausführung an
    // das, was jemand gesehen hat (Befund 12 des Zweitblicks über Phase 5.2).
    revision: string,
  ) => Promise<RuecksetzungsbilanzAntwort>

  // **Die Freigabe für den Neuanfang** (G12, #2282). Ohne sie ließ sich die
  // ganze Anwendung nirgends zurücksetzen: die Seite fragte nach einer
  // Kennung, die keine Bindung ausstellte. Es gibt sie nur am Fenster.
  RuecksetzungFreigabeErteilen: () => Promise<NachweisAntwort>
}
