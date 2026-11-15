package domain

import (
	"cmp"
	"strings"
)

// ArtrangUebrige ist der Rang jeder Art, die keine der geführten ist.
//
//	TYPE_OTHER,
//
// `src/gui/package_update_list.hpp`, Zeile 66, Fassung f9d9356 vom 2021-07-06
// — der letzte Wert des Enums, und `package_type` gibt ihn für jeden
// unbekannten Namen zurück (`package_update_list.cpp`, Zeile 94).
const ArtrangUebrige = 100

// artraenge ist MSEs `PackageType`, auf die Arten dieses Projekts beschränkt.
//
//	enum PackageType {
//	  TYPE_PROG, TYPE_LOCALE, TYPE_GAME, TYPE_STYLESHEET,
//	  TYPE_EXPORT_TEMPLATE, TYPE_SYMBOL_FONT, TYPE_INCLUDE, TYPE_FONT,
//	  TYPE_OTHER,
//	}
//
// `src/gui/package_update_list.hpp`, Zeilen 57–67, Fassung f9d9356 vom
// 2021-07-06. **Zwei Werte fehlen hier**:
// `TYPE_PROG` ist MSE selbst (dieses Programm verwaltet sich nicht mit, §12),
// `TYPE_FONT` eine lose `.ttf` (Schriften werden über `internal/schriften`
// gefunden, nicht als Paket geführt).
//
// **Die Abstände sind zehn und nicht eins.** Wer eine achte Art dazwischen
// braucht, setzt sie hinein, ohne jede Zahl darunter zu verschieben — und ohne
// dass ein Test, der Zahlen statt Ordnung prüfte, deswegen rot würde.
var artraenge = map[string]int{
	"mse-locale":          10,
	"mse-game":            20,
	"mse-style":           30,
	"mse-export-template": 40,
	"mse-symbol-font":     50,
	"mse-include":         60,
}

// Artrang gibt die Stelle dieser Art in MSEs Reihenfolge.
//
// **Sie ist eine Aussage darüber, was zuerst gebraucht wird**: ohne Spiel kein
// Stil, ohne Stil keine Symbolschrift. Alphabetisch stünde die Ausgabe vor dem
// Spiel, und wer eine Vorlage sucht, fände zuerst den Export.
func (p Paket) Artrang() int {
	if rang, da := artraenge[p.Art]; da {
		return rang
	}

	return ArtrangUebrige
}

// VergleicheNachRang ordnet zwei Pakete, wie MSE2 eine Paketliste ordnet:
// der kleinere `position hint:` zuerst, bei Gleichstand der Name.
//
// **ported** — `PackageList::ComparePackagePosHint`,
// `src/gui/control/package_list.cpp`, Zeilen 52–60, Fassung f9d9356 vom
// 2021-07-06. Nach ihr stehen dort die Spiele und Stile eines neuen Sets
// (`NewSetWindow`), und nach ihr stehen sie hier (UX-05, #1677).
//
// **Der Name nach Codeeinheiten**, wie `wxString` vergleicht: „magic-DayNight"
// steht vor „magic-m15".
func VergleicheNachRang(a, b Paket) int {
	if a.Rang != b.Rang {
		return cmp.Compare(a.Rang, b.Rang)
	}

	return strings.Compare(a.Name, b.Name)
}
