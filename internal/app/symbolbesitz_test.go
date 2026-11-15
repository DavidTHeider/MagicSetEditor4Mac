//go:build darwin

// **Das Symbol eines fremden Sets ist auch am Fenster nicht zu erreichen** —
// #1815.
//
// Die Symbolzeichnung liegt als Datei neben dem Set, und `Symbol` wie
// `SymbolSetzen` fassten sie über die Setkennung an, ohne den Bestand zu
// fragen. An den drei Flächen ohne Fenster hält das
// `internal/flaechen/symbolbesitz_test.go`; hier steht das Fenster.

package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDasFensterFasstKeinFremdesSymbolAn(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	ctx := t.Context()

	ich := anwendung.WerArbeitet(ctx)
	bob := anwendung.NutzerAnlegen(ctx, "Audit Bob", "")

	if ich.Fehler != nil || bob.Fehler != nil {
		t.Fatalf("vorbereiten: %+v %+v", ich.Fehler, bob.Fehler)
	}

	if wechsel := anwendung.NutzerWechseln(ctx, bob.Nutzer.ID, ""); wechsel.Fehler != nil {
		t.Fatalf("zu Bob: %+v", wechsel.Fehler)
	}

	set := anwendung.SetOeffnen(ctx, probesetVerzeichnis(t))
	if set.Fehler != nil {
		t.Fatalf("Bobs Set: %+v", set.Fehler)
	}

	teile := app.SymbolGrundform(app.SymbolgrundformDTO{Art: "Rechteck"}).Teile

	// **Die Gegenprobe zuerst**: Bob legt ein Symbol in sein Set — ohne sie
	// hieße ein „nicht gefunden" unten nur, dass es keines gibt.
	bobs := anwendung.SymbolSetzen(ctx, set.SetID, teile, "")
	if bobs.Fehler != nil {
		t.Fatalf("Bob legt sein Symbol nicht ab: %+v", bobs.Fehler)
	}

	if zurueck := anwendung.NutzerWechseln(ctx, ich.NutzerID, ""); zurueck.Fehler != nil {
		t.Fatalf("zurück: %+v", zurueck.Fehler)
	}

	for name, fehler := range map[string]*app.Fehler{
		"lesen":     anwendung.Symbol(ctx, set.SetID, bobs.Name).Fehler,
		"schreiben": anwendung.SymbolSetzen(ctx, set.SetID, teile, bobs.Name).Fehler,
	} {
		if fehler == nil || fehler.Code != app.CodeNichtGefunden {
			t.Errorf("%s: %+v, erwartet nicht gefunden", name, fehler)
		}
	}
}
