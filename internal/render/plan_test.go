package render_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ein Plan ist von Hand hinschreibbar — das ist die Zusage aus #29.
//
// **Sie trägt die ganze Prüfbarkeit der Phase.** Ein Goldbildvergleich braucht
// eine Eingabe, die jemand lesen und ändern kann, ohne einen Interpreter zu
// starten; eine Vorlage von 1100 Zeilen ist das nicht.
func planM15() *render.Plan {
	return &render.Plan{
		Breite: 375, Hoehe: 523,
		Ebenen: []render.Ebene{
			{
				Art:     render.ArtBild,
				Rahmen:  render.Rechteck{Links: 0, Oben: 0, Breite: 375, Hoehe: 523},
				Bild:    render.Blatt("/magic-m15.mse-style/card.jpg"),
				Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: 29, Oben: 32, Breite: 317, Hoehe: 22},
				Farbe:  render.Farbe{R: 156, G: 130, B: 90, A: 255},
				Verlauf: &render.Verlauf{Winkel: 90, Stufen: []render.Verlaufsstufe{
					{Stelle: 0, Farbe: render.Farbe{R: 156, G: 130, B: 90, A: 255}},
					{Stelle: 1, Farbe: render.Farbe{R: 90, G: 70, B: 40, A: 255}},
				}},
				Maske:   render.Blatt("/magic-m15.mse-style/name-mask.png"),
				Deckung: 1,
			},
			{
				Art:        render.ArtText,
				Rahmen:     render.Rechteck{Links: 32, Oben: 34, Breite: 280, Hoehe: 18},
				Farbe:      render.Farbe{A: 255},
				Abschnitte: render.Aufrecht("Feuerball"),
				Satz: render.Satz{
					Schrift: "MPlantin", Groesse: 14,
					Ausrichtung: render.LinksMitte, Laufweite: 0.2,
				},
				Deckung: 1,
			},
		},
	}
}

func TestEinBrauchbarerPlanGehtDurch(t *testing.T) {
	befunde, err := render.Pruefe(planM15())
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}
}

// Ein Mangel an der Karte selbst ist ein Fehler, kein Befund.
//
// Wer eine Größe von null als Befund behandelte, zeichnete auf eine Fläche der
// Größe null und wunderte sich über ein leeres Bild ohne Meldung.
func TestEinMangelAnDerKarteIstEinFehler(t *testing.T) {
	faelle := []struct {
		name string
		plan *render.Plan
	}{
		{name: "gar kein Plan", plan: nil},
		{name: "Breite null", plan: &render.Plan{Breite: 0, Hoehe: 523}},
		{name: "Höhe null", plan: &render.Plan{Breite: 375, Hoehe: 0}},
		{name: "negative Breite", plan: &render.Plan{Breite: -1, Hoehe: 523}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if _, err := render.Pruefe(fall.plan); !errors.Is(err, render.ErrPlan) {
				t.Errorf("Fehler %v, erwartet war ErrPlan", err)
			}
		})
	}
}

// Ein Mangel an einer Ebene ist ein Befund mit ihrer Nummer.
//
// Ohne die Nummer sucht jemand in einer Vorlage mit 60 Ebenen von Hand.
func TestEinMangelAnEinerEbeneIstEinBefundMitNummer(t *testing.T) {
	faelle := []struct {
		name  string
		ebene render.Ebene
		nennt string
	}{
		{
			name:  "ein Bild ohne Bild",
			ebene: render.Ebene{Art: render.ArtBild},
			nennt: "ohne Bild",
		},
		{
			name:  "ein Symbol ohne Bild",
			ebene: render.Ebene{Art: render.ArtSymbol, Bild: render.Blatt("   ")},
			nennt: "ohne Bild",
		},
		{
			name:  "Text ohne Schriftgröße",
			ebene: render.Ebene{Art: render.ArtText, Abschnitte: render.Aufrecht("a")},
			nennt: "Schriftgröße",
		},
		{
			// **Ein Verlauf mit einer Stufe ist keiner.** Er sähe aus wie eine
			// einfarbige Fläche und wäre einer, den jemand halb geschrieben hat.
			name: "ein Verlauf mit einer Stufe",
			ebene: render.Ebene{Art: render.ArtFlaeche, Verlauf: &render.Verlauf{
				Stufen: []render.Verlaufsstufe{{Stelle: 0}},
			}},
			nennt: "Verlauf mit 1 Stufen",
		},
		{
			name:  "ein Verlauf ohne Stufen",
			ebene: render.Ebene{Art: render.ArtFlaeche, Verlauf: &render.Verlauf{}},
			nennt: "Verlauf mit 0 Stufen",
		},
		{
			name: "Stufen, die nicht aufsteigen",
			ebene: render.Ebene{Art: render.ArtFlaeche, Verlauf: &render.Verlauf{
				Stufen: []render.Verlaufsstufe{{Stelle: 1}, {Stelle: 0}},
			}},
			nennt: "aufsteigend",
		},
		{
			name:  "eine Art, die es nicht gibt",
			ebene: render.Ebene{Art: "Unfug"},
			nennt: "unbekannte Ebenenart",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
				{Art: render.ArtFlaeche, Deckung: 1},
				fall.ebene,
			}}

			befunde, err := render.Pruefe(plan)
			if err != nil {
				t.Fatalf("abgewiesen: %v", err)
			}

			if len(befunde) != 1 {
				t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
			}

			if befunde[0].Ebene != 2 {
				t.Errorf("Ebene %d, erwartet war 2", befunde[0].Ebene)
			}

			if !strings.Contains(befunde[0].Meldung, fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", befunde[0].Meldung, fall.nennt)
			}
		})
	}
}

// Eine Fläche ohne Verlauf ist in Ordnung — sie ist der Normalfall.
func TestEineEinfarbigeFlaecheGehtDurch(t *testing.T) {
	plan := &render.Plan{Breite: 10, Hoehe: 10, Ebenen: []render.Ebene{
		{Art: render.ArtFlaeche, Farbe: render.Farbe{A: 255}, Deckung: 1},
	}}

	befunde, err := render.Pruefe(plan)
	if err != nil || len(befunde) != 0 {
		t.Errorf("%v / %v", befunde, err)
	}
}

// Zeichenbar lässt die beanstandeten Ebenen weg und den Rest stehen.
func TestZeichenbarLaesstNurDieBeanstandetenWeg(t *testing.T) {
	plan := planM15()
	plan.Ebenen = append(plan.Ebenen, render.Ebene{Art: render.ArtBild})

	befunde, err := render.Pruefe(plan)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	gekuerzt, nummern := render.Zeichenbar(plan, befunde)

	if len(gekuerzt.Ebenen) != 3 {
		t.Errorf("%d Ebenen, erwartet waren 3", len(gekuerzt.Ebenen))
	}

	// **Der ursprüngliche Plan bleibt unberührt.** Er ist die Auskunft darüber,
	// was die Vorlage wollte, und die soll ein Befund nicht überschreiben.
	if len(plan.Ebenen) != 4 {
		t.Errorf("der ursprüngliche Plan wurde verändert: %d Ebenen", len(plan.Ebenen))
	}

	if gekuerzt.Breite != plan.Breite || gekuerzt.Hoehe != plan.Hoehe {
		t.Error("die Kartengröße ging beim Kürzen verloren")
	}

	// Und die Nummern sind die des ursprünglichen Plans.
	if len(nummern) != 3 || nummern[0] != 1 || nummern[2] != 3 {
		t.Errorf("Nummern %v, erwartet waren [1 2 3]", nummern)
	}
}

// Ohne Befund kommt derselbe Plan zurück.
func TestZeichenbarOhneBefundeGibtDenselbenPlan(t *testing.T) {
	plan := planM15()

	gleicher, nummern := render.Zeichenbar(plan, nil)
	if gleicher != plan {
		t.Error("ohne Befunde wurde der Plan trotzdem umgebaut")
	}

	if len(nummern) != 3 || nummern[1] != 2 {
		t.Errorf("Nummern %v, erwartet waren [1 2 3]", nummern)
	}
}

// Die Darstellungen sind für Testmeldungen da und müssen lesbar sein.
func TestDieDarstellungen(t *testing.T) {
	if got := (render.Farbe{R: 156, G: 130, B: 90, A: 255}).String(); got != "rgba(156,130,90,255)" {
		t.Errorf("Farbe: %s", got)
	}

	if got := (render.Rechteck{Links: 29, Oben: 32, Breite: 317, Hoehe: 22}).String(); got != "(29,32 317x22)" {
		t.Errorf("Rechteck: %s", got)
	}

	if got := (render.Befund{Ebene: 2, Meldung: "kein Pfad"}).String(); got != "Ebene 2: kein Pfad" {
		t.Errorf("Befund: %s", got)
	}
}
