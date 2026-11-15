package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gerechneter Vorgabewert erreicht die Oberfläche** —
// #563.
//
// 142 der 249 Felder von `magic.mse-game` tragen ein `default:`, und alle 142
// sind Skript. Die **gezeichnete** Karte trägt sie — der Kanon steht bei null
// Befunden. Der **Editor** zeigte an derselben Stelle ein leeres Feld: wer die
// Karte ansah, sah eine Farbe; wer das Feld ansah, sah nichts.
//
// # Warum im Plan und nicht in `Kartenfelder`
//
// `Kartenfelder(setID)` spricht über das **Set**. Ein `default:` ist ein
// Skript über `card`, und dieselbe Vorlage gibt für zwei Karten zwei Werte.
//
// # Warum im Plan und nicht in einem zweiten Lauf
//
// Derselbe Grund, den `render.Plan.Feldflaechen` schon nennt: „Wer ihn
// außerhalb noch einmal rechnete, rechnete ein zweites Mal dasselbe und liefe
// beim nächsten Unterschied auseinander." Die Werte stehen beim Planbau fertig
// da.
func vorgabenvorlage() service.Vorlage {
	return service.Vorlage{
		Stil: msepaket.Stil{
			Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "10", Oben: "10", Breite: "300", Hoehe: "20",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			}},
		},
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText},

			// **Ein `default:` als Skript** — so stehen alle 142 im Pack.
			{Name: "card color", Kind: domain.KindText, Default: `"blau"`},

			// **Und eines mit Leerzeichen im Namen**: in der Umgebung heißt es
			// `card_symbol`, im Editor `card symbol` — und der Editor fragt.
			{
				Name: "card symbol", Kind: domain.KindChoice, Initial: "none",
				Choices: []string{"none", "common"},
			},
		},
	}
}

func TestDerPlanNenntDieGerechnetenVorgaben(t *testing.T) {
	t.Parallel()

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorgabenvorlage())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	gefunden := make(map[string]string, len(plan.Vorgaben))
	for _, vorgabe := range plan.Vorgaben {
		gefunden[vorgabe.Feld] = vorgabe.Wert
	}

	if gefunden["card color"] != "blau" {
		t.Errorf("card color = %q — %v", gefunden["card color"], gefunden)
	}

	// **Der Feldname trägt sein Leerzeichen.**
	if gefunden["card symbol"] != "none" {
		t.Errorf("card symbol = %q — %v", gefunden["card symbol"], gefunden)
	}
}

// **Was die Karte selbst führt, steht nicht darunter.**
//
// Eine Vorgabe ist eine **Herleitung**. Ein Feld, das die Karte trägt, ist
// keine — und die Oberfläche machte daraus sonst einen Hinweis „das ist
// gerechnet" über einem getippten Wert.
func TestEinEigenerWertIstKeineVorgabe(t *testing.T) {
	t.Parallel()

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "card color", Text: "rot"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, vorgabenvorlage())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, vorgabe := range plan.Vorgaben {
		if vorgabe.Feld == "card color" {
			t.Errorf("`card color` führt die Karte selbst, steht aber als Vorgabe da")
		}
	}
}

// **Ein leerer Wert ist keine Vorgabe.** Ein Feld, das weder `default:` noch
// `initial:` noch eine erste Wahl hat, bekommt den leeren Text — und ein
// Hinweis „hier steht gerechnet nichts" ist keiner.
func TestEineLeereVorgabeStehtNichtDa(t *testing.T) {
	t.Parallel()

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorgabenvorlage())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, vorgabe := range plan.Vorgaben {
		if vorgabe.Feld == "name" {
			t.Errorf("`name` hat keinen Vorgabewert, steht aber da: %q", vorgabe.Wert)
		}
	}
}

// **Ein geleertes Feld bekommt seine Herleitung zurück** — der dritte
// Abnahmekasten aus #563.
//
// Und er deckt einen Fehler auf, den der erste Anlauf hatte: „führt die Karte"
// hieß dort „hat eine Zeile", und wer ein Feld leert, hat eine Zeile mit
// leerem Text. Die **gezeichnete** Karte trug den Vorgabewert weiter —
// `mitVorgabeskripten` füllt jedes leere Feld —, der Editor zeigte ab da
// nichts mehr. Zwei Antworten auf dieselbe Frage, und die im Editor war die
// falsche.
func TestEinGeleertesFeldBekommtSeineHerleitungZurueck(t *testing.T) {
	t.Parallel()

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "card color", Text: ""},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, vorgabenvorlage())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, vorgabe := range plan.Vorgaben {
		if vorgabe.Feld == "card color" && vorgabe.Wert == "blau" {
			return
		}
	}

	t.Errorf("`card color` ist geleert und bekommt seine Herleitung nicht zurück: %+v",
		plan.Vorgaben)
}
