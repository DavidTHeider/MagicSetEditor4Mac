package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`set` trägt die Angaben des Sets.**
//
// # Der Befund
//
// `stilumgebung.go` sagte seit Phase 6 ausdrücklich: „`set` bleibt leer, und
// das ist benannt." Das war ehrlich und ist jetzt einlösbar — `set_values`
// liegt im Bestand, und der Leser dafür steht seit Phase 9.
//
// Gezählt über `M15-Magic-Pack`: **420 Zugriffe auf `set.symbol`**, 345 auf
// `set.border_color`, 101 auf `set.mainframe_rarity_name`, dazu ein langer
// Schwanz. Es ist nach `stylesheet.card_width` der meistgelesene Name des
// Packs.
//
// # Was das nicht löst
//
// `set.symbol` ist bei MSE ein **Symbolwert** — eine `.mse-symbol`-Vektordatei
// im Set —, und `symbol_variation` rastert sie
// (`src/script/functions/image.cpp:142`). Hier kommt der rohe Wert an; das
// Rastern ist eigener Umfang
// (#321).
//
// Fassung f9d9356 vom 2021-07-06.
func TestSetTraegtSeineAngaben(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ set.symbol + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Setangaben: []domain.Value{
		{Field: "symbol", Text: "eigenes"},
		{Field: "border color", Text: "rgb(0,0,0)"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/eigenes.png" {
				t.Errorf("gezeichnet wird %q, erwartet eigenes.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein Feldname mit Leerzeichen wird zum Unterstrich** — wie bei der Karte.
//
// Die Angaben stehen im Set als `border color`, die Skripte lesen
// `set.border_color`.
func TestEineSetangabeMitLeerzeichenBekommtEinenUnterstrich(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ set.border_color + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{Stil: stil, Setangaben: []domain.Value{
		{Field: "border color", Text: "schwarz"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/schwarz.png" {
				t.Errorf("gezeichnet wird %q, erwartet schwarz.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein Setfeld ohne Wert trägt die Vorgabe des Spiels — und sonst nichts.**
//
// # Die Klasse hinter dem Fall
//
// Es ist dieselbe wie bei den Kartenfeldern
// (#278): das
// Spiel erklärt seine Felder, das Set füllt aus, was es ausfüllt, und der Rest
// steht mit der Vorgabe da. Ohne die Erklärung gäbe `set.copyright` an einem
// Set, das es nicht ausgefüllt hat, **nichts** statt einer leeren Zeichenkette
// — und ein `set.copyright + " x"` bräche ab, wo MSE weiterrechnet.
//
// Gezählt über `M15-Magic-Pack`: 89 Setfelder, 1.400 Zugriffe in 167 Dateien.
func TestEinSetfeldOhneWertTraegtDieVorgabeDesSpiels(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ set.copyright + set.titel + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{
		Stil: stil,
		Setfelder: []domain.Field{
			// **Ein `default:` ist ein Skript** — hier eine Zeichenkette in
			// Skriptsyntax, genau die Form, die `magic.mse-game/set_fields`
			// führt: `default: "Type custom types here; …"`. Roh
			// durchgereicht stünden die Anführungszeichen im Wert (#298).
			{Name: "copyright", Default: `"vorgabe"`},
			{Name: "titel"},
		},
		// **Nichts ausgefüllt.** Die Vorgabe trägt das eine Feld, das andere
		// bleibt leer — leer und nicht nichts.
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/vorgabe.png" {
				t.Errorf("gezeichnet wird %q, erwartet vorgabe.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Was das Set ausgefüllt hat, schlägt die Vorgabe.**
func TestEinAusgefuelltesSetfeldSchlaegtDieVorgabe(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
			Bild: `{ set.copyright + ".png" }`,
		}},
	}

	vorlage := service.Vorlage{
		Stil:       stil,
		Setfelder:  []domain.Field{{Name: "copyright", Default: "vorgabe"}},
		Setangaben: []domain.Value{{Field: "copyright", Text: "eigenes"}},
	}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/eigenes.png" {
				t.Errorf("gezeichnet wird %q, erwartet eigenes.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}
