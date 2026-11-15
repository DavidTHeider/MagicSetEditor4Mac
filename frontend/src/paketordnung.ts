// **Pakete in MSE2s Ordnung** — der kleinere `position hint:` zuerst, bei
// Gleichstand der Name (UX-05,
// #1677).
//
// **ported** — `PackageList::ComparePackagePosHint`,
// `src/gui/control/package_list.cpp`, Zeilen 52–60, Fassung f9d9356 vom
// 2021-07-06. Nach ihr stehen dort Spiele und Stile eines neuen Sets, und der
// erste ist hier der vorgewählte. Dieselbe Regel steht in Go als
// `domain.VergleicheNachRang` und ordnet die Stilseite eines offenen Sets.
//
// **Der Name ist ein technischer Schlüssel** (`docs/LOKALISIERUNG.md`): er
// wird nach Codeeinheiten verglichen, wie `wxString` es tut — nicht nach der
// Sprache des Fensters.
import { schluesselvergleich } from './ordnung'

// Einordnung ist, was die Ordnung von einem Paket braucht.
interface Einordnung {
  name: string

  // rang ist `position hint:`; ohne Angabe eine Million, also hinten.
  rang: number
}

// nachRang vergleicht zwei Pakete für `sort`.
export function nachRang(a: Einordnung, b: Einordnung): number {
  if (a.rang !== b.rang) {
    return a.rang - b.rang
  }

  return schluesselvergleich(a.name, b.name)
}
