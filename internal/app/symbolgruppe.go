// **Gruppen und Reihenfolge in der Teilliste** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// **Drei Bindungen, die nichts entscheiden** (§2). Die Regeln stehen in
// [service.TeileGruppieren], [service.GruppeAufloesen] und
// [service.TeilVerschieben]; hier steht die Übersetzung der DTOs und der
// Fehler — dieselbe Aufteilung wie bei [App.Grundform] und [App.PunktEinfuegen].

package app

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SymbolTeileGruppieren fasst die genannten Stellen zu einer Gruppe zusammen.
//
// **Die Gruppe steht dort, wo das erste gewählte Teil stand.** Der Grund und
// die Fundstelle stehen bei [service.TeileGruppieren].
func SymbolTeileGruppieren(teile []SymbolteilDTO, stellen []int) SymbolAntwort {
	return teilweise(teile, func(
		umgesetzt []msepaket.Symbolteil,
	) ([]msepaket.Symbolteil, error) {
		return service.TeileGruppieren(umgesetzt, stellen)
	})
}

// SymbolGruppeAufloesen setzt die Kinder der genannten Gruppen an deren Stelle.
//
// **Eine Symmetrie ist hier keine Gruppe** — sie hat ihre eigene Bedienung in
// [App.SymmetrieAufloesen]. Auch das kommt von MSE; die Fundstelle steht bei
// [service.GruppeAufloesen].
func SymbolGruppeAufloesen(teile []SymbolteilDTO, stellen []int) SymbolAntwort {
	return teilweise(teile, func(
		umgesetzt []msepaket.Symbolteil,
	) ([]msepaket.Symbolteil, error) {
		return service.GruppeAufloesen(umgesetzt, stellen)
	})
}

// SymbolTeilVerschieben nimmt ein Teil heraus und setzt es an anderer Stelle ein.
//
// **Die Zielstelle zählt ohne das bewegte Teil** — auch das ist MSEs Rechnung,
// und sie steht bei [service.TeilVerschieben].
func SymbolTeilVerschieben(teile []SymbolteilDTO, von, nach int) SymbolAntwort {
	return teilweise(teile, func(
		umgesetzt []msepaket.Symbolteil,
	) ([]msepaket.Symbolteil, error) {
		return service.TeilVerschieben(umgesetzt, von, nach)
	})
}

// teilweise führt eine Umformung über die ganze Liste aus.
//
// **Bei einem Fehler kommt die Liste von vorher zurück**, nicht eine leere.
// Sie ist der Stand, mit dem die Oberfläche weiterarbeitet; eine leere Antwort
// auf einen Bedienfehler löschte die Zeichnung. Dieselbe Zusage wie bei
// [punktweise] eine Datei weiter.
func teilweise(
	teile []SymbolteilDTO,
	umformen func([]msepaket.Symbolteil) ([]msepaket.Symbolteil, error),
) SymbolAntwort {
	gebaut, err := umformen(teileAusDTO(teile))
	if err != nil {
		return SymbolAntwort{Teile: nichtNil(teile), Fehler: Nutzerfehler(err)}
	}

	return SymbolAntwort{Teile: teileAlsDTO(gebaut)}
}

// TeileGruppieren fasst die genannten Teile zu einer Gruppe zusammen.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolTeileGruppieren]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) TeileGruppieren(teile []SymbolteilDTO, stellen []int) SymbolAntwort {
	return SymbolTeileGruppieren(teile, stellen)
}

// GruppeAufloesen macht aus einer Gruppe ihre Teile.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolGruppeAufloesen]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) GruppeAufloesen(teile []SymbolteilDTO, stellen []int) SymbolAntwort {
	return SymbolGruppeAufloesen(teile, stellen)
}

// TeilVerschieben ändert die Zeichenreihenfolge.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolTeilVerschieben]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) TeilVerschieben(teile []SymbolteilDTO, von, nach int) SymbolAntwort {
	return SymbolTeilVerschieben(teile, von, nach)
}
