// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
//
// **Eine eigene Datei**, damit `api.ts` unter der Zeilengrenze aus §2 bleibt.
import type { FehlerDTO } from './formen'

// PlatzpostenDTO sagt, wie viel ein Posten belegt.
export interface PlatzpostenDTO {
  // posten ist der Schlüssel — `bestand`, `bilder`, `pakete`,
  // `vorsicherungen`, `protokoll` oder `sonstiges`. Der Name dazu steht im
  // Katalog.
  //
  // **Sechs seit Werkstück 6.**
  posten: string

  dateien: number
  bytes: number

  // raeumbar sagt, ob sich dieser Posten wegräumen lässt.
  //
  // **Gerechnet in Go**, damit die Antwort nicht an zwei Stellen steht.
  raeumbar: boolean
}

export interface PlatzAntwort {
  posten: PlatzpostenDTO[]

  // nachladefehler heißt: **weggeräumt ist**, nur die Zählung danach war
  // nicht zu machen (§7, Regel 3; #1970). Die Anzeige zählt selbst neu.
  nachladefehler?: FehlerDTO

  fehler: FehlerDTO | null
}

// Platzapi sind die zwei Bindungen.
export interface Platzapi {
  Platz: () => Promise<PlatzAntwort>

  // PlatzFreigeben räumt einen Posten weg und gibt die Zählung danach.
  //
  // **Die Antwort ist die neue Zählung** und kein bloßes „getan": wer
  // wegräumt, will sehen, was es gebracht hat — und ein zweiter Ruf zum
  // Nachladen wäre ein zweiter Weg zu derselben Antwort (§7, Regel 3).
  PlatzFreigeben: (posten: string) => Promise<PlatzAntwort>
}
