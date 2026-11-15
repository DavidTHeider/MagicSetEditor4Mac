// **Zurück ist nicht eine Handlung, sondern vier** — Werkstück O7 von Phase
// 5.2 (#1431).
// Die Gestalt kommt aus `internal/app/ruecksprung.go`.
import type { Finish, SchreibAntwort } from './api'
import type { Umschlag } from './umschlag'

// RuecksprungartDTO ist eine Art samt dem, was ihr im Weg steht.
export interface RuecksprungartDTO {
  art: string

  // hindernis ist leer, wenn diese Anwendung die Art selbst ausführt.
  hindernis: string
}

export interface RuecksprungartenAntwort extends Umschlag {
  arten: RuecksprungartDTO[]
}

// VergleichszeileDTO ist ein Zählstand nebeneinander.
export interface VergleichszeileDTO {
  gegenstand: string
  inSicherung: number
  jetzt: number

  // verlust ist, was beim Zurückspielen verschwindet.
  verlust: number

  // unbekannt heißt: eine der beiden Seiten sagt nichts darüber.
  unbekannt: boolean
}

// MengeDTO ist eine Stückzahl eines Gegenstands der Sammlung.
export interface MengeDTO {
  schluessel: string
  zustand: string
  sprache: string

  // finish ist die Ausführung: `nonfoil`, `foil` oder `etched` — Werkstück T2
  // von Phase 5.6 (#1467, 0119).
  //
  // **Das abgekündigte `foil` steht hier nicht**, obwohl die Go-Gestalt es
  // mitschickt. Es ist dort für MCP, REST und die Kommandozeile, die diese
  // Antwort unverändert bekommen; das Fenster liest die Naht als **innen**
  // (0033, Regel 5) und braucht den alten Wahrheitswert nicht.
  finish: Finish

  menge: number
  name: string
  setcode: string
  nummer: string
}

// RuecksprunglageAntwort ist der Vergleich vor der Rückkehr.
export interface RuecksprunglageAntwort extends Umschlag {
  datenstand: string
  zeilen: VergleichszeileDTO[]

  // unklar sind die Mengen, die danach wieder dastünden, ohne dass jemand
  // weiß, ob sie noch da sein sollten.
  unklar: MengeDTO[]

  handpruefungNoetig: boolean
}

// RueckkehrDTO ist eine verbuchte Rückkehr.
export interface RueckkehrDTO {
  id: number
  zurueckgespielt: string
  datenstand: string
  quelle: string
  klaerung: string
  geklaertAm: string
  mengen: MengeDTO[]
}

export interface RueckkehrenAntwort extends Umschlag {
  rueckkehren: RueckkehrDTO[]
}

export interface Rueckkehrapi {
  Ruecksprungarten: () => Promise<RuecksprungartenAntwort>
  Ruecksprunglage: (
    pfad: string,
    kennwort: string,
  ) => Promise<RuecksprunglageAntwort>
  Rueckkehren: () => Promise<RueckkehrenAntwort>
  RueckkehrQuittieren: (kennung: number) => Promise<SchreibAntwort>
}
