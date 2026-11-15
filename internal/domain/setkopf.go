// **Die Angaben, die im Setkopf neben `set info:` stehen** —
// #1325.
//
// # Warum sie eine eigene Gestalt haben
//
// Sie sind **keine** Setangaben: `set info:` führt, was das **Spiel**
// definiert — Titel, Urheberrecht, Setkürzel —, und was dort steht,
// unterscheidet sich von Spiel zu Spiel. Die fünf hier gehören zum **Format**
// und stehen in jeder MSE-Datei gleich.
//
// # Woher sie kommen
//
// Aus MSE2, nachgelesen in `src/data/set.cpp` und `src/data/set.hpp`
// (Fassung f9d9356 vom 2021-07-06). Bis zum 2026-09-03 reisten sie roh mit
// und standen in der Anwendung unter „Nicht gedeutet" — in **jedem** der zehn
// kanonischen Sets dieselben fünf.
//
// # Was mit ihnen nicht geschieht
//
// **Sie werden nicht geändert.** MSE2 schreibt `mse_version` bei jedem
// Speichern neu (die eigene Fassung) und die beiden Paketfassungen aus den
// installierten Paketen; diese Anwendung sichert über die mitgeführte
// Quelldatei ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md))
// und lässt sie deshalb stehen. Wer das ändert, ändert eine fremde Datei —
// und muss erklären, warum.

package domain

// Setkopf sind die fünf Formatangaben eines Sets.
type Setkopf struct {
	// MSEFassung ist die Fassung von Magic Set Editor, die die Datei
	// geschrieben hat.
	//
	// MSE2 schreibt sie bei **jedem** Speichern (`src/util/io/writer.cpp`,
	// Zeile 26) und warnt beim Lesen, wenn sie neuer ist als die eigene
	// (`src/util/io/reader.cpp`, Zeilen 38–45).
	MSEFassung string

	// Spielfassung ist die Fassung des Spielpakets zum Zeitpunkt des
	// Speicherns.
	//
	// MSE2 warnt, wenn das **installierte** Paket älter ist als das, mit dem
	// gespeichert wurde (`src/data/set.cpp`, Zeilen 169–176, gerufen in 190).
	Spielfassung string

	// Stilfassung ist dasselbe für das Stilpaket (`set.cpp`, Zeile 194).
	Stilfassung string

	// Versionskontrolle ist die Art des Versionskontrollsystems.
	//
	// MSE2 hält dafür ein Objekt (`VCSP vcs`, `src/data/set.hpp` Zeile 62,
	// Fassung f9d9356 vom 2021-07-06);
	// in jeder gesehenen Datei steht `none`. Hier steht die **Art** und nicht
	// das Objekt: was ein Versionskontrollsystem täte, tut diese Anwendung
	// nicht.
	Versionskontrolle string

	// Apprenticecode ist ein Textfeld für den Apprentice-Export.
	//
	// **Nur für Magic** (`src/data/set.hpp`, Zeile 58, Fassung f9d9356 vom
	// 2021-07-06: „Code to use for apprentice (Magic only)"). In jedem
	// kanonischen Set leer.
	Apprenticecode string
}
