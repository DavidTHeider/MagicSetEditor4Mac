// **Eine Partieverwaltung, die nichts tut** — für Läufe, die sie nicht
// prüfen (PLAY-03.AK-05,
// #1711).
//
// **Eine Vorrichtung und keine zweite Fassung** (§3): wer die Verwaltung
// selbst prüft, baut sich ihre Hand mit `vi.fn()` und sieht nach, was
// ankam. Diese hier steht in den Läufen, die die **Liste** prüfen und die
// Verwaltung nur brauchen, weil sie darin steht.
import type { Verwaltungsstand } from '../haken/usePartieverwaltung'

export function stilleVerwaltung(
  ueberschreibung: Partial<Verwaltungsstand> = {},
): Verwaltungsstand {
  return {
    laeuft: () => false,
    meldung: undefined,
    loeschen: () => undefined,
    kopieren: () => undefined,
    ausfuehren: () => undefined,
    einlesen: () => undefined,
    ...ueberschreibung,
  }
}
