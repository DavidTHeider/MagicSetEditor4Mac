// **Das Diagnosepaket** — Werkstück O10 von Phase 5.2
// (#1434).
// Die Gestalt kommt aus `internal/app/diagnosepaket.go`.
import type { Umschlag } from './umschlag'

// PaketteilDTO ist ein Teil in der Vorschau.
export interface PaketteilDTO {
  teil: string
  datei: string
  zeilen: number
  bytes: number

  // grund steht da, wenn der Teil **nicht** mitkommt — leer sonst. Ein
  // ausgelassener Teil wird genannt und nicht weggelassen.
  grund: string
}

// DiagnosepaketAntwort ist die Vorschau: was im Paket steht, wie groß es ist
// und über welche Zeit es Auskunft gibt.
export interface DiagnosepaketAntwort extends Umschlag {
  teile: PaketteilDTO[]
  bytes: number
  grenze: number

  // von und bis sind der Zeitraum der aufgenommenen Protokollzeilen — leer,
  // wenn keine dabei sind.
  von: string
  bis: string

  // ohne sind die Teile, die jemand abgewählt hat — nur in der Antwort auf
  // das Schreiben belegt, in der Vorschau leer (OPS-04.AK-04).
  ohne: string[]
}

export interface Diagnosepaketapi {
  Diagnosepaket: () => Promise<DiagnosepaketAntwort>

  // **Beide geben nur einen Umschlag zurück** — geschrieben ist geschrieben,
  // und was danach gilt, sagt die nächste Vorschau (§7, Regel 3).
  DiagnosepaketSichern: (ohne: string[]) => Promise<Umschlag>
  DiagnosepaketLoeschen: () => Promise<Umschlag>
}
