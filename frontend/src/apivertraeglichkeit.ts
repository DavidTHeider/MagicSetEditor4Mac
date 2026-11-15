// **Kann die neueste Fassung diesen Bestand öffnen?** — Werkstück O6 von
// Phase 5.2 (#1430).
// Die Gestalt kommt aus `internal/app/vertraeglichkeit.go`.
import type { Umschlag } from './umschlag'

// HindernisDTO ist ein Grund, warum ein Wechsel nicht vorbereitet wird.
export interface HindernisDTO {
  art: string

  // stelle benennt es genauer — der Name einer Wanderung, die fremde
  // Plattform. Leer, wo es nichts zu benennen gibt.
  stelle: string
}

// VertraeglichkeitAntwort sagt, ob und warum.
export interface VertraeglichkeitAntwort extends Umschlag {
  fassung: string

  // urteil ist `vertraeglich`, `unvertraeglich`, `ungeklaert` — oder leer,
  // wenn es keine neuere Fassung gibt.
  urteil: string

  bereit: boolean
  hindernisse: HindernisDTO[]
}

export interface Vertraeglichkeitapi {
  Vertraeglichkeit: () => Promise<VertraeglichkeitAntwort>
  WechselVorbereiten: () => Promise<VertraeglichkeitAntwort>
}
