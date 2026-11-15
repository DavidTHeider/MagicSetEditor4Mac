// **Die Rechnung hinter dem Markenfilter** —
// #1208.
//
// **Eine eigene Datei**, weil `Kartenspalte.tsx` mit ihr über die Zeilengrenze
// ginge (§2) — und weil der Schnitt einer Frage folgt: dort steht, **wie** die
// Spalte aussieht, hier, **welche** Karten sie zeigt.
import type { KartenzeileDTO } from '../api'
import { vergleicher } from '../sprache/vergleich'

// imSet gibt die Marken, die im Set vorkommen — jede einmal, alphabetisch.
//
// **Aus dem Aufbau gerechnet und nicht über eine eigene Bindung geholt**: die
// Antwort steht schon da, und ein zweiter Ruf hätte einen zweiten Stand
// (§7, Regel 2) — nach dem Setzen einer Marke wären beide für einen Augenblick
// verschieden.
export function imSet(
  jeKarte: Readonly<Record<string, readonly string[]>>,
  sprache: string,
): readonly string[] {
  return [...new Set(Object.values(jeKarte).flat())].sort(vergleicher(sprache))
}

// wirksam gibt die Marke, nach der **tatsächlich** gefiltert wird.
//
// # Warum es sie braucht
//
// **Weil eine gewählte Marke verschwinden kann.** Wer die letzte Karte
// entmarkiert, nimmt damit die Marke aus dem Set — die Kachelreihe zeichnet
// sie nicht mehr, und die Wahl bliebe im Zustand stehen. Die Liste wäre dann
// **leer und ohne Weg zurück**: der Knopf, mit dem man »Alle« wählte, ist ja
// gerade verschwunden.
//
// **Gesehen im laufenden Fenster** (§3): der erste Anlauf sagte im Kommentar,
// der Filter räume sich von selbst — er tat es nicht. Die Zusage stand da, und
// die Prüfung fehlte.
//
// **Abgeleitet und nicht zurückgesetzt.** Ein Effekt, der den Zustand nachführt,
// liefe eine Zeichnung zu spät — und für einen Augenblick stünde die leere
// Liste doch da.
export function wirksam(
  angeboten: readonly string[],
  gewaehlt: string,
): string {
  return angeboten.includes(gewaehlt) ? gewaehlt : ''
}

// gefiltert gibt die Karten, die die gewählte Marke tragen.
//
// **Leer heißt: alle.**
export function gefiltert(
  karten: readonly KartenzeileDTO[],
  jeKarte: Readonly<Record<string, readonly string[]>>,
  marke: string,
): readonly KartenzeileDTO[] {
  if (marke === '') {
    return karten
  }

  return karten.filter((karte) => (jeKarte[karte.id] ?? []).includes(marke))
}
