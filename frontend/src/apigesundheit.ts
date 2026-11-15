// **Der Betriebszustand** — Werkstück O9 von Phase 5.2
// (#1433).
// Die Gestalt kommt aus `internal/app/gesundheit.go`.
import type { Umschlag } from './umschlag'

// FremdzustandDTO ist ein fremder Dienst und seine Erreichbarkeit.
export interface FremdzustandDTO {
  dienst: string
  erreichbar: boolean
  grund: string
}

// GesundheitAntwort sagt, ob dieser Prozess lebt und arbeiten kann.
//
// **`bereit` steht als eigenes Feld da, obwohl es sich rechnen ließe** — eine
// Fläche soll die Regel nicht nachbauen, und vier Nachbauten liefen
// auseinander (§16).
export interface GesundheitAntwort extends Umschlag {
  lebt: boolean
  bereit: boolean

  // grund ist der **eigene** Grund — `bestandZu`, `wanderungLaeuft`,
  // `bestandUnlesbar` oder leer. Ein fremder Dienst steht nie darin.
  grund: string

  fremde: FremdzustandDTO[]
}

export interface Gesundheitsapi {
  Gesundheit: () => Promise<GesundheitAntwort>
}
