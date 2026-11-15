package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **`alignment:` darf ein Skript sein.**
//
// MSE führt sie als `Scriptable<Alignment>` und rechnet sie mit jedem Wechsel
// neu (`TextStyle::update`, `src/data/field/text.cpp`, Zeile 95, Fassung
// f9d9356 vom 2021-07-06). `magic-m15` entscheidet damit, ob ein kurzer
// Regeltext mittig steht:
//
//	alignment:
//		script:
//			if (… and card_style.text.content_lines <= 2) …
//			then "middle center" else "middle left"
//
// **216 der 1.597 Ausrichtungen im Pack stehen als Skript.** Wer den rohen
// Text als Wortliste läse, fände darin weder `middle` noch `left` verlässlich
// — und bekäme `links oben`, die Vorgabe. Genau daran stand der Regeltext an
// der Oberkante seines Kastens
// (#345).
func TestEineAusrichtungDarfEinSkriptSein(t *testing.T) {
	faelle := []struct {
		name     string
		angabe   string
		erwartet render.Ausrichtung
	}{
		{"schlicht", "middle center", render.MitteMitte},
		{"als Skript", `{ "middle center" }`, render.MitteMitte},
		{"ein Skript mit Bedingung", `{ if 1 < 2 then "bottom right" else "top left" }`, render.RechtsUnten},

		// **Ein Skript, das scheitert, ergibt die Vorgabe** — links oben, so
		// wie eine leere Angabe. Eine erfundene Ausrichtung säße zuverlässig
		// falsch.
		{"ein Skript, das scheitert", `{ gibtsnicht() }`, render.LinksOben},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{{
					Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
					Ausrichtung: fall.angabe,
					Schrift:     msepaket.Schrift{Name: "Probe", Groesse: "10"},
				}},
			}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "abc"},
			}}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for stelle := range plan.Ebenen {
				if plan.Ebenen[stelle].Art != render.ArtText {
					continue
				}

				if ist := plan.Ebenen[stelle].Satz.Ausrichtung; ist != fall.erwartet {
					t.Errorf("die Ausrichtung ist %q, erwartet %q", ist, fall.erwartet)
				}

				return
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}

// **Und `shrink-overflow` wird im Ergebnis des Skripts gelesen.**
//
// `magic-m15` schreibt `if styling.center_name then "center middle
// shrink-overflow" else "left middle shrink-overflow"` — die Ansage steckt im
// Ergebnis, nicht im Quelltext.
func TestDasStauchenStehtImErgebnisDesSkripts(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
			Ausrichtung: `{ "center middle shrink-overflow" }`,
			Schrift:     msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "abc"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for stelle := range plan.Ebenen {
		if plan.Ebenen[stelle].Art != render.ArtText {
			continue
		}

		satz := plan.Ebenen[stelle].Satz
		if !satz.Stauchen || !satz.NurBeiUeberlauf || satz.Ausrichtung != render.MitteMitte {
			t.Errorf("der Satz ist %+v", satz)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}

// **Und `alignment:` mit einem `script:`-Block darunter — die Form des Packs.**
//
// 216 der 1.597 Ausrichtungen stehen so. Der **Wert** des Knotens ist dann
// leer; wer nur ihn läse, bekäme die Vorgabe „links oben", und der Regeltext
// stünde an der Oberkante seines Kastens.
func TestEineAusrichtungAlsSkriptblockWirdGelesen(t *testing.T) {
	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{{
			Feld: "text", Links: "0", Oben: "0", Breite: "300", Hoehe: "100",
			AusrichtungSkript: `if 1 < 2 then "middle center" else "top left"`,
			Schrift:           msepaket.Schrift{Name: "Probe", Groesse: "10"},
		}},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "text", Text: "abc"},
	}}

	plan, _, err := service.NeuerPlanbauer().
		MitSchriften(schriftprobe{breite: 0.5}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for stelle := range plan.Ebenen {
		if plan.Ebenen[stelle].Art != render.ArtText {
			continue
		}

		if ist := plan.Ebenen[stelle].Satz.Ausrichtung; ist != render.MitteMitte {
			t.Errorf("die Ausrichtung ist %q, erwartet %q", ist, render.MitteMitte)
		}

		return
	}

	t.Fatal("der Plan führt keine Textebene")
}
