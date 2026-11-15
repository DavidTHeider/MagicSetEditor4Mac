// **Die Karten in die berechnete Reihenfolge bringen** — Befund 1 des
// Zweitblicks über Phase 3.7
// (#1168).
//
// # Warum es diese Datei gibt
//
// Zwei Antworten treffen sich in der Kartenspalte: `Kartenzeilen` bringt die
// Karten samt ihren Aufschriften — geordnet nach `cards.reihenfolge` —, und
// `Kartenlistenaufbau` bringt die **Ordnung**, die der Dienst aus der Wahl des
// Spiels gerechnet hat. Bis zum 2026-08-27 traf sich das nirgends: die Ordnung
// kam an und wurde weggeworfen.
//
// **Eine reine Funktion und keine zweite Sortierung.** Sortiert wird auf der
// Go-Seite — `domain.Vergleiche` ist MSEs `smart_compare`, und eine zweite
// Fassung in TypeScript liefe mit jener auseinander (§2). Hier wird nur
// **umgestellt**.

import type { KartenzeileDTO } from '../api'

// geordnet gibt die Karten in der Reihenfolge der Kennungsliste.
//
// **Drei Fälle, und alle drei sind die Regel und nicht der Rand:**
//
//   - **Ohne Ordnung bleibt es, wie es kam.** Solange der Aufbau nicht
//     geladen ist, ist die Liste leer — und eine leere Ordnung darf die Liste
//     nicht leeren.
//   - **Was die Ordnung nicht kennt, steht hinten.** Eine Karte, die nach dem
//     letzten Aufbau entstanden ist, verschwindet sonst aus der Liste, bis
//     jemand neu lädt.
//   - **Was es nicht gibt, wird übersprungen.** Eine gelöschte Karte steht
//     noch in der Ordnung von vorhin.
export function geordnet(
  karten: readonly KartenzeileDTO[],
  ordnung: readonly string[],
): readonly KartenzeileDTO[] {
  if (ordnung.length === 0) {
    return karten
  }

  const nachKennung = new Map(karten.map((karte) => [karte.id, karte]))
  const gestellt: KartenzeileDTO[] = []

  for (const kennung of ordnung) {
    const karte = nachKennung.get(kennung)
    if (karte !== undefined) {
      gestellt.push(karte)
      nachKennung.delete(kennung)
    }
  }

  // **Der Rest in seiner eigenen Reihenfolge** — nicht in der der `Map`.
  for (const karte of karten) {
    if (nachKennung.has(karte.id)) {
      gestellt.push(karte)
    }
  }

  return gestellt
}
