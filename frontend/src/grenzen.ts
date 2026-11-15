// **Die Grenzen dieses Prozesses** — Werkstück O5 von Phase 5.2
// (#1429).
// Die Gestalt kommt aus `internal/app/grenzen.go`.
import type { Umschlag } from './umschlag'

export interface KontingentDTO {
  burst: number
  nachfuellen: number
  fensterSekunden: number
}

// **Das Auftragsbuch des Dienstes** — #2003.
export interface AuftragsgrenzenDTO {
  laufend: number
  behalten: number
  aufbewahrungSekunden: number
}

// **Eine Zeile der Kostentafel** — #2003, `domain.Kostenauskunft`: je teurer
// Fähigkeit ihre Klasse und ihre Wege an jeder Fläche.
export interface KostenauskunftDTO {
  faehigkeit: string
  klasse: string
  kosten: number
  rest: string[]
  mcp: string[]
  cli: string[]
}

export interface GrenzenDTO {
  parallel: Record<string, number>
  antwortBytes: number
  principal: KontingentDTO
  global: KontingentDTO
  reserve: number
  kosten: Record<string, number>
  // **Ob die Guthaben gelten** — QA-08.AK-03 (#1739), `grenzen.kontingente`.
  kontingente: 'an' | 'aus'
  // **Verbindungen, Aufträge und Kostentafel** — #2003: dieselbe Auskunft wie
  // `mse grenzen`, `lesen.grenzen` und `GET /grenzen`.
  verbindungen: number
  auftraege: AuftragsgrenzenDTO
  kostentafel: KostenauskunftDTO[]
}

export interface GrenzenAntwort extends Umschlag {
  grenzen: GrenzenDTO
}
