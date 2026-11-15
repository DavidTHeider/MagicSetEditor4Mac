// Die Schlüsselwortbindungen erreichen ihre Anwendung
// (#1062).
//
// **Eine eigene Datei**, weil `bindungen_test.go` mit ihnen über die
// Zeilengrenze ginge (§2).
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ohne eingerichtete Naht meldet jede der vier**, statt abzustürzen.
//
// Vor `MitSchluesselwoertern` gibt es sie nicht — im Betrieb setzt der
// Composition Root sie, in einem Test nicht zwingend. Ein Nilzeiger wäre dort
// ein Absturz statt einer Meldung.
func TestDieSchluesselwortbindungenMeldenDieFehlendeNaht(t *testing.T) {
	bindungen := app.NeueBindungen(app.NeueApp(""))

	// **Eine leere Liste und kein `nil`**: über die Wails-Bindung stünde dort
	// sonst `null`, und das Frontend prüfte auf zwei Dinge statt auf eines.
	liste := bindungen.Schluesselwoerter(t.Context(), "1")
	if liste.Fehler == nil {
		t.Error("das Lesen ohne Naht kam durch")
	}

	if liste.Woerter == nil || liste.Modi == nil {
		t.Error("nil statt leerer Listen")
	}

	if bindungen.SchluesselwortAnlegen(t.Context(), "1").Fehler == nil {
		t.Error("das Anlegen ohne Naht kam durch")
	}

	if bindungen.SchluesselwortSetzen(t.Context(), "1", "1", "wort", "X", nil).Fehler == nil {
		t.Error("das Setzen ohne Naht kam durch")
	}

	if bindungen.SchluesselwortLoeschen(t.Context(), "1", "1").Fehler == nil {
		t.Error("das Löschen ohne Naht kam durch")
	}
}

// **Und die Konsole ebenso** (#1070) — sie ist dieselbe Klasse: eine Naht, die
// im Betrieb der Composition Root setzt, in einem Test aber nicht zwingend.
func TestDieKonsolenbindungMeldetDieFehlendeNaht(t *testing.T) {
	antwort := app.NeueBindungen(app.NeueApp("")).Skript(t.Context(), "1", "1 + 1")

	if antwort.Fehler == nil {
		t.Error("die Auswertung ohne Naht kam durch")
	}

	if antwort.Befunde == nil {
		t.Error("nil statt einer leeren Liste")
	}
}

// **Und das Zufallspack ebenso** (#1072) — dieselbe Klasse.
func TestDiePackbindungenMeldenDieFehlendeNaht(t *testing.T) {
	bindungen := app.NeueBindungen(app.NeueApp(""))

	angebot := bindungen.Packarten(t.Context(), "1")
	if angebot.Fehler == nil {
		t.Error("das Angebot ohne Naht kam durch")
	}

	if angebot.Arten == nil || angebot.Befunde == nil {
		t.Error("nil statt leerer Listen")
	}

	ziehung := bindungen.PackZiehen(t.Context(), "1", "booster", "1")
	if ziehung.Fehler == nil {
		t.Error("das Ziehen ohne Naht kam durch")
	}

	if ziehung.Karten == nil || ziehung.Befunde == nil {
		t.Error("nil statt leerer Listen")
	}
}

// **Und die Symbolbindungen ebenso** (#890) — dieselbe Klasse.
//
// **`Grundform` ist die Ausnahme darin**: sie braucht keine Naht, weil sie
// nichts liest und nichts schreibt — sie rechnet. Sie steht hier trotzdem,
// weil die Durchreiche zu prüfen ist: eine Bindung, die die falsche Methode
// ruft, fiele sonst erst im Fenster auf.
func TestDieSymbolbindungenReichenDurch(t *testing.T) {
	bindungen := app.NeueBindungen(app.NeueApp(""))

	form := bindungen.Grundform(t.Context(), app.SymbolgrundformDTO{Art: "Vieleck", Ecken: 3})
	if form.Fehler != nil {
		t.Fatalf("die Grundform: %+v", form.Fehler)
	}

	if len(form.Teile) != 1 || len(form.Teile[0].Punkte) != 3 {
		t.Errorf("die Form ist %+v", form.Teile)
	}

	gesetzt := bindungen.SymbolSetzen(t.Context(), "1", nil, "")
	if gesetzt.Fehler == nil {
		t.Error("das Ablegen ohne Naht kam durch")
	}

	if gesetzt.Teile == nil {
		t.Error("nil statt einer leeren Liste")
	}

	gelesen := bindungen.Symbol(t.Context(), "1", "symbol1.mse-symbol")
	if gelesen.Fehler == nil {
		t.Error("das Lesen ohne Naht kam durch")
	}

	// **Die Punktwege brauchen keine Naht** — sie rechnen. Geprüft wird die
	// Durchreiche: eine Bindung, die die falsche Methode ruft, fiele sonst
	// erst im Fenster auf.
	teile := []app.SymbolteilDTO{{
		Art: "Form",
		Punkte: []app.SymbolpunktDTO{
			{X: 0, Y: 0, Sperre: "free"},
			{X: 1, Y: 0, Sperre: "free"},
			{X: 1, Y: 1, Sperre: "free"},
		},
	}}

	eingefuegt := bindungen.PunktEinfuegen(t.Context(), teile, app.SymbolpunktweiseDTO{})
	if eingefuegt.Fehler != nil || len(eingefuegt.Teile[0].Punkte) != 4 {
		t.Errorf("das Einfügen gab %+v", eingefuegt)
	}

	geloescht := bindungen.PunktLoeschen(t.Context(), teile, app.SymbolpunktweiseDTO{Punkt: 1})
	if geloescht.Fehler != nil || len(geloescht.Teile[0].Punkte) != 2 {
		t.Errorf("das Löschen gab %+v", geloescht)
	}

	// Der Punkt steht auf (0, 0) und hat keinen Griff; die Spitze auf 0,1 zu
	// setzen gibt einen Griff von 0,1.
	gezogen := bindungen.GriffZiehen(t.Context(), teile, app.SymbolgriffDTO{X: 0.1})
	if gezogen.Fehler != nil || gezogen.Teile[0].Punkte[0].GriffDanachX != 0.1 {
		t.Errorf("das Ziehen gab %+v", gezogen)
	}

	pruefeSymmetriebindungen(t, bindungen, teile)
	pruefeTeillistenbindungen(t, bindungen, teile)
}

// pruefeSymmetriebindungen fährt die vier Wege der Symmetrie
// (#1092).
//
// **Auch sie brauchen keine Naht** — sie rechnen über der Zeichnung, die
// ankommt. Geprüft wird die Durchreiche: eine Bindung, die die falsche
// Methode ruft, fiele sonst erst im Fenster auf.
func pruefeSymmetriebindungen(
	t *testing.T, bindungen *app.Bindungen, teile []app.SymbolteilDTO,
) {
	t.Helper()

	angelegt := bindungen.SymmetrieAnlegen(t.Context(), teile, 0)
	if angelegt.Fehler != nil || angelegt.Teile[0].Art != "Symmetrie" {
		t.Fatalf("das Anlegen gab %+v", angelegt)
	}

	gesetzt := bindungen.SymmetrieSetzen(t.Context(), angelegt.Teile,
		app.SymmetrieweiseDTO{Teil: 0, Kopien: 4})
	if gesetzt.Fehler != nil || gesetzt.Teile[0].Kopien != 4 {
		t.Errorf("das Setzen gab %+v", gesetzt)
	}

	kopien := bindungen.Symmetriekopien(t.Context(), gesetzt.Teile, 0)
	if kopien.Fehler != nil || len(kopien.Kopien) != 4 {
		t.Errorf("die Kopien gaben %+v", kopien)
	}

	aufgeloest := bindungen.SymmetrieAufloesen(t.Context(), gesetzt.Teile, 0)
	if aufgeloest.Fehler != nil || aufgeloest.Teile[0].Art != "Form" {
		t.Errorf("das Auflösen gab %+v", aufgeloest)
	}
}

// pruefeTeillistenbindungen fährt die drei Wege über die Teilliste
// (#1168).
//
// **Dieselbe Frage wie eine Funktion höher**: ruft die Bindung die Methode,
// die auf ihrem Schild steht? Was die drei *tun*, steht in
// `internal/service/symbolteile_test.go`.
func pruefeTeillistenbindungen(
	t *testing.T, bindungen *app.Bindungen, teile []app.SymbolteilDTO,
) {
	t.Helper()

	zwei := append(append([]app.SymbolteilDTO{}, teile...), teile...)

	gruppiert := bindungen.TeileGruppieren(t.Context(), zwei, []int{0, 1})
	if gruppiert.Fehler != nil || gruppiert.Teile[0].Art != "Gruppe" {
		t.Fatalf("das Gruppieren gab %+v", gruppiert)
	}

	aufgeloest := bindungen.GruppeAufloesen(t.Context(), gruppiert.Teile, []int{0})
	if aufgeloest.Fehler != nil || len(aufgeloest.Teile) != 2 {
		t.Errorf("das Auflösen gab %+v", aufgeloest)
	}

	verschoben := bindungen.TeilVerschieben(t.Context(), zwei, 0, 1)
	if verschoben.Fehler != nil || len(verschoben.Teile) != 2 {
		t.Errorf("das Verschieben gab %+v", verschoben)
	}
}
