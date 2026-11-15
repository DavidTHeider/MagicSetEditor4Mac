// **Eine Feldbeschreibung für Tests** — an einer Stelle.
//
// # Warum
//
// Sie stand am 2026-08-27 in **vierzehn** Dateien, jede mit ihrer eigenen
// Vorgabe. Als `FeldbeschreibungDTO` um `reihe` und `stelle` wuchs
// (#1167),
// waren das vierzehn Stellen, an denen dasselbe zwei Felder nachzutragen war —
// und vierzehn Gelegenheiten, eine davon zu vergessen. Dieselbe Überlegung wie
// bei `kartenproben.ts`: eine Vorrichtung, die eine Gestalt nachbaut, gehört
// dorthin, wo sie einmal steht.
import type { FeldbeschreibungDTO } from '../api'

// feldform gibt eine Beschreibung mit brauchbaren Vorgaben.
//
// **Die Vorgaben sind die eines gewöhnlichen Textfeldes** — editierbar,
// gespeichert, einzeilig, ohne Reihe. Wer etwas anderes prüft, sagt es im
// Aufruf; wer nichts sagt, bekommt den Fall, der in einem Paket am häufigsten
// steht.
export function feldform(
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name: 'name',
    aufschrift: 'Name',
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    teilfelder: [],
    editierbar: true,
    gespeichert: true,

    // **Ohne Reihe und ohne Stelle.** Ein Feld ohne numerierte Geschwister ist
    // der Normalfall: von 249 Kartenfeldern in `magic.mse-game` gehören 106 zu
    // einer der 43 Reihen, die übrigen zu keiner.
    reihe: '',
    stelle: 0,

    ...teile,
  }
}
