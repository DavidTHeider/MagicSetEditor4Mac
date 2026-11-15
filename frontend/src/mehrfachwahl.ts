// Der Wert einer Mehrfachwahl, wie MSE ihn schreibt.
//
// **`"wahl, wahl, wahl"`** — Komma und Leerzeichen, in der Reihenfolge, in der
// die Wahlen erklärt sind (`src/data/field/multiple_choice.hpp`, Zeilen 46–48,
// Fassung f9d9356 vom 2021-07-06).
//
// **Zerlegt wird hier, geordnet im Dienst.** Das Fenster muss wissen, welche
// Kästchen angehakt sind; welche Form gespeichert wird, entscheidet
// `service.normalform` — dieselbe Funktion, die auch die Kartenfelder benutzt.

// Wahltrenner steht zwischen zwei Wahlen.
export const Wahltrenner = ', '

// wahlenAus zerlegt einen Mehrfachwahlwert.
//
// **Leerzeichen hinter dem Komma fallen weg**, wie MSEs `get` sie überspringt
// (`multiple_choice.cpp`, Zeilen 63–80). Ein leerer Wert gibt keine Wahl —
// nicht eine leere: sonst stünde ein Kästchen ohne Namen an.
export function wahlenAus(wert: string): string[] {
  return wert
    .split(',')
    .map((teil) => teil.trim())
    .filter((teil) => teil !== '')
}
