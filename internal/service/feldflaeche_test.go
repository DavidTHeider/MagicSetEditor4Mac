package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Plan sagt, wo ein Feld liegt — auch wenn dort nichts steht.**
//
// # Wofür
//
// MSE entscheidet die Bedienbarkeit eines Feldes an seinem Kasten, nicht an
// seinem Inhalt:
//
//	bool ValueViewer::isVisible() const {
//	  return getStyle()->visible
//	    && bounding_box.width > 0 && bounding_box.height > 0 …
//	}
//
// `src/render/value/viewer.cpp`, Zeilen 36–42, gefragt von `is_enabled`
// (`src/gui/control/card_editor.cpp`, Zeilen 155–157), Fassung f9d9356 vom
// 2021-07-06. Der Kasten kommt aus dem **Stil**; ein `ValueViewer` bleibt
// bestehen, ob sein Wert leer ist oder nicht.
//
// # Was ohne diese Auskunft fehlte
//
// Der Planbauer lässt ein leeres Textfeld fallen — richtig, denn es zeichnet
// nichts. Der Editor las daraufhin die **Ebenen** und bot genau die Felder an,
// die etwas zeichnen: eine frisch angelegte Karte besteht nur aus leeren
// Feldern und hatte kein einziges Eingabefeld
// (#562).
func TestDerPlanNenntDieFlaecheEinesLeerenFeldes(t *testing.T) {
	// **`farbe` steht darin, `name` und `text` nicht.** Der Stil der Saat
	// rechnet die Stelle des Textfelds aus `card.farbe`; ohne das Feld wäre
	// der Befund daraus der eigentliche Gegenstand des Tests, und das ist er
	// nicht.
	leer := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "farbe", Text: "blau"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), leer, service.Vorlage{Stil: saatstil(t)})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	// **Die Gegenprobe zuerst**: zeichnete das Feld etwas, prüfte der Test die
	// Ebenen und nicht die Flächen.
	for _, ebene := range plan.Ebenen {
		if ebene.Feld == "name" {
			t.Fatalf("`name` zeichnet auf einer leeren Karte — der Test prüft nichts")
		}
	}

	flaeche, da := flaecheFuer(plan, "name")
	if !da {
		t.Fatalf("keine Fläche für `name`; vorhanden: %v", flaechennamen(plan))
	}

	if flaeche.Breite <= 0 || flaeche.Hoehe <= 0 {
		t.Errorf("die Fläche hat keine Ausdehnung: %+v", flaeche)
	}
}

// **Und ein Feld, das der Stil verbirgt, steht nicht darin.**
//
// `getStyle()->visible` ist MSEs **erste** Bedingung in `isVisible`, vor jeder
// Frage nach dem Kasten. Wer ein verborgenes Feld anböte, machte ein
// Eingabefeld für etwas, das die Karte nicht zeigt — und `magic-m15` verbirgt
// so den Kasten für Stärke und Widerstandskraft an jeder Verzauberung.
//
// **Der Stil steht hier und kommt nicht aus der Saat**: `probe-schlicht`
// verbirgt nichts, und ihr ein Feld dafür anzuhängen änderte die Goldbilder.
func TestEinVerborgenesFeldBekommtKeineFlaeche(t *testing.T) {
	stil := saatstil(t)
	stil.Felder = append(stil.Felder, msepaket.Feldstil{
		Feld:     "verborgen",
		Links:    "10",
		Oben:     "10",
		Breite:   "100",
		Hoehe:    "20",
		Sichtbar: "{ false }",
	})

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), stilkarte("blau"), service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if _, da := flaecheFuer(plan, "verborgen"); da {
		t.Errorf("ein verborgenes Feld steht in den Flächen: %v", flaechennamen(plan))
	}

	// **Die Gegenprobe zum Wächter**: ohne `visible:` steht dasselbe Feld drin.
	stil.Felder[len(stil.Felder)-1].Sichtbar = ""

	offen, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), stilkarte("blau"), service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if _, da := flaecheFuer(offen, "verborgen"); !da {
		t.Errorf("ohne `visible:` fehlt die Fläche auch — der Test prüft nichts")
	}
}

func flaecheFuer(plan *render.Plan, feld string) (render.Rechteck, bool) {
	for _, flaeche := range plan.Feldflaechen {
		if flaeche.Feld == feld {
			return flaeche.Rahmen, true
		}
	}

	return render.Rechteck{}, false
}

func flaechennamen(plan *render.Plan) []string {
	namen := make([]string, 0, len(plan.Feldflaechen))
	for _, flaeche := range plan.Feldflaechen {
		namen = append(namen, flaeche.Feld)
	}

	return namen
}
