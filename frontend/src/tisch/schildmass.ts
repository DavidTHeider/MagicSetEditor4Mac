// **Wie groß ein Schild auf dem Tisch steht — oder ob überhaupt** — die
// Einstellung `tisch.beschriftung`, seit dem 2026-09-24 (`internal/domain/
// beschriftung.go`).
//
// # Warum eine eigene Datei
//
// **Weil es eine Entscheidung ist, und der Zeichner wird nicht gemessen**
// (`pixibeiwerk.ts` steht in `.covignore`). Hier steht, welche Stufe welche
// Größe gibt und welches Schild bei „aus" bleibt; dort wird nur geschrieben.
import type { Zonenschild } from './zonenschilder'

// Stufenfaktor ist das Verhältnis einer Stufe zu „mittel".
//
// **Ein Fünftel je Stufe** — kleiner, und man sähe den Unterschied nicht;
// größer, und „groß" liefe über sein Band hinaus (`masse.ts`).
const Stufenfaktor: Record<string, number> = {
  klein: 0.8,
  normal: 1,
  gross: 1.2,
}

// schriftgroesse gibt die Höhe der Schrift eines Schildes in Weltpunkten —
// oder nichts, wenn es bei dieser Stufe nicht dasteht.
//
// **Ein unbekannter Wert liest sich wie „mittel"**: eine spätere Fassung kann
// eine Stufe mehr haben, und ein Tisch ohne Schrift wäre die falsche Antwort
// darauf.
export function schriftgroesse(
  schild: Zonenschild,
  beschriftung: string,
): number | undefined {
  if (beschriftung === 'aus') {
    return schild.art === 'zahl' ? schild.groesse : undefined
  }

  return schild.groesse * (Stufenfaktor[beschriftung] ?? 1)
}
