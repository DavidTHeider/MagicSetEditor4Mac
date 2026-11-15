// **Die Zwischenspeicher** — OPS-08
// (#1725).
// Die Gestalt kommt aus `internal/app/zwischenspeicher.go`.
import type { Umschlag } from './umschlag'

// ZwischenspeicherDTO ist ein Speicher mit seinen Zahlen — eine Grenze von
// null heißt: diese Grenze hat er nicht.
export interface ZwischenspeicherDTO {
  name: string
  ort: string
  treffer: number
  fehlgriffe: number
  eintraege: number
  bytes: number
  grenzeEintraege: number
  grenzeBytes: number
}

export interface ZwischenspeicherAntwort extends Umschlag {
  speicher: ZwischenspeicherDTO[]
}

export interface Zwischenspeicherapi {
  Zwischenspeicher: () => Promise<ZwischenspeicherAntwort>

  // **Gibt, was vorher belegt war** — was danach gilt, sagt die nächste
  // Auskunft (§7, Regel 3).
  ZwischenspeicherLeeren: () => Promise<ZwischenspeicherAntwort>
}
