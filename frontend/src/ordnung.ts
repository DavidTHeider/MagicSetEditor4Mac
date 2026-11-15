// **Zwei Arten zu sortieren** — LOC-03.AK-01
// (#1689).
//
// # Warum es zwei sind und nicht eine
//
// **Weil `localeCompare` an zwei Stellen dasselbe tut und nur an einer
// richtig ist.** Der Name eines Pakets wird von einem Menschen gelesen, und
// dann gehört er in die Ordnung seiner Sprache. Die Kennung eines Objekts
// liest niemand; sie wird sortiert, damit die Reihenfolge **stabil** ist —
// und eine Ordnung, die sich mit der Systemsprache ändert, ist dort ein
// Fehler: dieselbe Liste sähe auf zwei Rechnern verschieden aus.
//
// # Was das Inventar dazu sagt
//
// [`docs/LOKALISIERUNG.md`](../../docs/LOKALISIERUNG.md) führt jede Stelle
// mit ihrer Klasse. Ein Wächter hält die Liste vollständig; **welche Klasse
// die richtige ist, sieht nur ein Mensch** — das steht dort auch.

// schluesselvergleich ordnet technische Schlüssel — **ohne Sprache**.
//
// **Zeichenweise nach ihren Codepunkten.** Das ist dieselbe Ordnung auf
// jedem Rechner und in jeder Sprache, und genau das ist hier verlangt.
export function schluesselvergleich(links: string, rechts: string): number {
  if (links === rechts) {
    return 0
  }

  return links < rechts ? -1 : 1
}
