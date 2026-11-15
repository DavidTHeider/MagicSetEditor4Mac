package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein `choice images:`-Block wählt das Bild nach dem Wert des Feldes.**
//
// # Der Mechanismus, nachgelesen in MSE2
//
// `src/data/field/choice.cpp:171` baut aus
//
//	choice images:
//	     a: {uvw}
//	     b: {xyz}
//
// das Skript
//
//	[a: {uvw}, b: {xyz}][input]() or else nil
//
// — eine Abbildung von Wahlname auf Bildskript, indiziert mit dem **Wert des
// Feldes**, gerufen, und `or else nil`.
//
// **Jeder Name steht zweimal darin:** einmal wie geschrieben und einmal in der
// „unkanonischen" Form, in der Unterstriche zu Leerzeichen werden
// (`util/string.cpp:127`).
//
// # Warum das 342 Befunde sind
//
// Über die 190 Stile von `M15-Magic-Pack` melden 342 Feldstile „dieses Feld
// zeichnet ein Bild aus den Wahlbildern"; es sind `rarity`, `type symbol` und
// die Verwandten. Sie sind der zweitgrößte Posten nach den Bildbäumen
// (#286).
//
// **Und sie stehen im Stil, nicht im Spiel.** Unsere Meldung sagte bis zum
// 2026-08-16 „aus den Wahlbildern des Spiels" — `magic-m15.mse-style:162` führt
// sie.
//
// Fassung f9d9356 vom 2021-07-06.
func TestEinWahlbildWirdNachDemFeldwertGewaehlt(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Wahlbilder: []msepaket.Wahlbild{
				{Wahl: "common", Bild: "/probe.mse-style/haeufig.png"},
				{Wahl: "rare", Bild: "/probe.mse-style/selten.png"},
			},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "rarity", Text: "rare"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild {
			continue
		}

		if ebene.Bild.Pfad != "/probe.mse-style/selten.png" {
			t.Errorf("gezeichnet wird %q, erwartet selten.png", ebene.Bild.Pfad)
		}

		return
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein Unterstrich im Wahlnamen meint ein Leerzeichen.**
//
// MSE legt jeden Namen zweimal ab — wie geschrieben und mit Unterstrichen zu
// Leerzeichen (`uncanonical_name_form`, `util/string.cpp:127`). Der Stil
// schreibt `diagonal T`, ein Kartenwert könnte `diagonal_T` sein.
func TestEinUnterstrichImWahlnamenMeintEinLeerzeichen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "tap symbol", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Wahlbilder: []msepaket.Wahlbild{
				{Wahl: "diagonal T", Bild: "/probe.mse-style/schraeg.png"},
			},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "tap symbol", Text: "diagonal_T"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/schraeg.png" {
				t.Errorf("gezeichnet wird %q, erwartet schraeg.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein Wert ohne Wahlbild zeichnet nichts — und meldet nichts.**
//
//	[a: {uvw}, b: {xyz}][input]() or else nil
//
// `ChoiceStyle::initImage`, `src/data/field/choice.cpp`, Zeilen 171–198, Fassung
// f9d9356 vom 2021-07-06. **`or else nil`** ist der letzte Schritt: ein Wert
// ohne hinterlegte Wahl ergibt nichts, und die Ebene bleibt leer.
//
// Hier stand bis zum 2026-08-17 „wird benannt", mit der Begründung „ein
// fehlendes Seltenheitssymbol ist etwas, das jemand sehen will". Über die 190
// Stile von `M15-Magic-Pack` waren das **80 Meldungen**, und keine nannte einen
// Fehler: `card symbol: none` hat kein Bild, weil es keines haben soll
// (#270).
func TestEineWahlOhneBildZeichnetNichts(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Wahlbilder:  []msepaket.Wahlbild{{Wahl: "common", Bild: "haeufig.png"}},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "rarity", Text: "mythic rare"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Und keine Bildebene** — gezeichnet wird nichts, nicht etwas Falsches.
	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			t.Errorf("es wurde ein Bild gezeichnet: %+v", ebene.Bild)
		}
	}
}

// **Ein Wahlbild darf ein `script:`-Block sein.**
//
// # Beide Formen stehen im Pack
//
// `magic-m15.mse-style:163` schreibt den Pfad hin; die Seltenheitssymbole in
// `magic-modules.mse-include/rarities/choice_images` schachteln ein Skript:
//
//	choice images:
//	     basic land:
//	          script:
//	               if use_main_rarity() then …
//
// **Der erste Anlauf las nur die erste Form** und ließ genau die 342 Felder
// leer, um die es geht.
func TestEinWahlbildDarfEinSkriptBlockSein(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Wahlbilder: []msepaket.Wahlbild{
				{Wahl: "rare", Bild: `{ "selten" + ".png" }`},
			},
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(),
		&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "rarity", Text: "rare"},
		}},
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			if ebene.Bild.Pfad != "/probe.mse-style/selten.png" {
				t.Errorf("gezeichnet wird %q, erwartet selten.png", ebene.Bild.Pfad)
			}

			return
		}
	}

	t.Fatal("der Plan führt keine Bildebene")
}

// **Ein leeres Feld bekommt kein Wahlbild und keinen Befund.**
//
// MSE gibt dort `nil` und zeichnet nichts — eine Karte ohne Seltenheit hat
// eben kein Seltenheitssymbol. Der erste Anlauf meldete auch das und trieb die
// Befundzahl über die 190 Stile von 1.668 auf 1.797; eine Meldung, die für
// jedes ungefüllte Feld erscheint, liest niemand mehr.
func TestEinLeeresFeldBekommtKeinWahlbildUndKeinenBefund(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rarity", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Darstellung: "image",
			Wahlbilder:  []msepaket.Wahlbild{{Wahl: "common", Bild: "haeufig.png"}},
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(), &domain.CardRow{ID: 1, Number: 1}, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde für ein leeres Feld: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild {
			t.Errorf("ein leeres Feld hat eine Bildebene ergeben: %+v", ebene.Bild)
		}
	}
}

// **Ein Feld ohne `choice images:` geht den Weg gar nicht erst.**
func TestOhneWahlbilderKeinWahlbildweg(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
		Felder: []msepaket.Feldstil{{
			Feld: "rahmen", Links: "0", Oben: "0", Breite: "40", Hoehe: "20",
			Bild: "rahmen.png",
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(
		t.Context(),
		&domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "rahmen", Text: "egal"},
		}},
		service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art == render.ArtBild && ebene.Bild.Pfad != "/probe.mse-style/rahmen.png" {
			t.Errorf("gezeichnet wird %q", ebene.Bild.Pfad)
		}
	}
}
