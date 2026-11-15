// **Die Kommandozentrale** — Werkstück U12 von Phase 5.1
// (#1423).
// Die Gestalt kommt aus `internal/app/befehlskatalog.go` und
// `befehlausfuehren.go`: die Werkzeuge des MCP-Servers, im Fenster gerufen.
import type { Umschlag } from './umschlag'

// Schema ist das JSON-Schema der Argumente, wie das Werkzeug es nennt.
export type Schema = Record<string, unknown>

export interface BefehlDTO {
  name: string
  faehigkeit: string
  cli: string
  rest: string
  erklaerung: string
  schreibend: boolean
  idempotent: boolean
  gefaehrlich: boolean
  schema: Schema
}

export interface BefehlskatalogAntwort extends Umschlag {
  befehle: BefehlDTO[]
}

export interface BefehlsergebnisAntwort extends Umschlag {
  name: string
  text: string
  dauerMs: number
}

export interface Befehlsapi {
  Befehlskatalog: () => Promise<BefehlskatalogAntwort>
  BefehlAusfuehren: (
    name: string,
    argumente: string,
    bestaetigt: boolean,
  ) => Promise<BefehlsergebnisAntwort>
  BefehlAbbrechen: () => Promise<Umschlag>
}
