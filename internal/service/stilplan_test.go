package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// saatstil öffnet den Stil der eigenen Saat.
//
// **Aus `saat/` und nicht aus einem bezogenen Paket** (§11): der Prüflauf muss
// auf jeder Maschine laufen, und fremde Vorlagen liegen nicht im Repository.
func saatstil(t *testing.T) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(
		filepath.Join("..", "..", "saat", "probe-schlicht.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	return stil
}

// stilkarte gibt eine Karte mit Namen, Text und Farbe.
func stilkarte(farbe string) *domain.CardRow {
	return &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Feuerball"},
		{Field: "text", Text: "Füge drei Schaden zu."},
		{Field: "farbe", Text: farbe},
	}}
}

// textebeneMit sucht die Textebene mit diesem Klartext.
//
// **Über den Inhalt und nicht über den Index.** Die Tests hier griffen
// `plan.Ebenen[1]` und `[2]`; als die Saat am 2026-08-22 einen gezeichneten
// Grund bekam (#770),
// verschob das jede Zusicherung um eins. Ein Index ist keine Aussage über die
// Ebene, die gemeint ist.
func textebeneMit(t *testing.T, plan *render.Plan, klartext string) render.Ebene {
	t.Helper()

	for i := range plan.Ebenen {
		if plan.Ebenen[i].Art == render.ArtText && plan.Ebenen[i].Klartext() == klartext {
			return plan.Ebenen[i]
		}
	}

	t.Fatalf("keine Textebene mit %q — der Plan führt %d Ebenen", klartext, len(plan.Ebenen))

	return render.Ebene{}
}

// **Ein Plan aus einem echten Stil setzt jedes Feld an seine Stelle.**
//
// # Was das vom Behelfsplan unterscheidet
//
// Der Behelf schreibt jeden Feldwert untereinander — er ist als Behelf benannt
// und sieht auch so aus. Hier kommen Stelle, Größe, Schrift und Schriftgröße
// aus dem **Stil**, und die Karte misst, was der Stil sagt.
func TestEinPlanEntstehtAusEinemEchtenStil(t *testing.T) {
	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), stilkarte("blau"), service.Vorlage{Stil: saatstil(t)})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Errorf("Befunde: %v", befunde)
	}

	if plan.Breite != 375 || plan.Hoehe != 523 {
		t.Errorf("%g × %g, erwartet 375 × 523", plan.Breite, plan.Hoehe)
	}

	// **Der Grund zuerst, dann die Felder nach `z index`** — vier seit #770:
	// die Zutat des Rasterers, der gezeichnete Grund der Saat, Name und Text.
	// Der Rahmen ist bei Vorgabe unsichtbar.
	if len(plan.Ebenen) != 4 {
		t.Fatalf("%d Ebenen, erwartet 4", len(plan.Ebenen))
	}

	name := textebeneMit(t, plan, "Feuerball")

	if name.Rahmen.Links != 24 || name.Rahmen.Oben != 24 ||
		name.Rahmen.Breite != 327 || name.Rahmen.Hoehe != 28 {
		t.Errorf("Rahmen %+v", name.Rahmen)
	}

	if name.Satz.Schrift != "Helvetica" || name.Satz.Groesse != 20 {
		t.Errorf("Satz %+v", name.Satz)
	}
}

// **Ein Feldskript wird ausgewertet, nicht als Text genommen.**
//
// `top: { if card.farbe == "rot" then 300 else 320 }` — die Karte entscheidet,
// wo ihr Text steht. Wer das Skript stehen ließe, bekäme 0 und legte den Text
// an den oberen Rand.
func TestEinFeldskriptWirdAusgewertet(t *testing.T) {
	for farbe, erwartet := range map[string]float64{"rot": 300, "blau": 320} {
		t.Run(farbe, func(t *testing.T) {
			plan, _, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), stilkarte(farbe), service.Vorlage{Stil: saatstil(t)})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			text := textebeneMit(t, plan, "Füge drei Schaden zu.")
			if text.Rahmen.Oben != erwartet {
				t.Errorf("Oben %g, erwartet %g", text.Rahmen.Oben, erwartet)
			}
		})
	}
}

// **Ein Feld, das die Karte nicht führt, wird übergangen — nicht erfunden.**
//
// Ein Stil beschreibt **alle** Felder eines Spiels; eine Karte führt die, die
// jemand gefüllt hat. Eine leere Textebene wäre eine Ebene, die nichts zeichnet
// und trotzdem Zeit kostet.
func TestEinFeldOhneWertKommtNichtInDenPlan(t *testing.T) {
	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "Nur ein Name"},
	}}

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, service.Vorlage{Stil: saatstil(t)})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(plan.Ebenen) != 3 {
		t.Errorf("%d Ebenen, erwartet 3 — die Zutat des Rasterers, der Grund "+
			"der Saat und der Name", len(plan.Ebenen))
	}
}

// **Ein Skript, das nicht aufgeht, ist ein Befund und kein Abbruch** (§8).
//
// Der Stil kommt aus einem fremden Paket; er darf Funktionen rufen, die wir
// nicht kennen. Eine Karte mit elf von zwölf Ebenen ist brauchbar.
func TestEinUnverstandenesSkriptIstEinBefund(t *testing.T) {
	stil := saatstil(t)
	stil.Felder[1].Oben = "{ gibtesnicht() }"

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) == 0 {
		t.Fatal("kein Befund — die Ebene läge stillschweigend bei 0")
	}

	// Und die Ebene ist trotzdem da. **Vier seit #770**: Zutat, Grund, Name,
	// Text.
	if len(plan.Ebenen) != 4 {
		t.Errorf("%d Ebenen", len(plan.Ebenen))
	}
}

// Die Ränder des Stilplans, jeder mit seiner eigenen Ursache.
func TestDieRaenderDesStilplans(t *testing.T) {
	bauer := service.NeuerPlanbauer()

	t.Run("ein abgebrochener Lauf baut keinen Plan", func(t *testing.T) {
		ctx, abbrechen := context.WithCancel(t.Context())
		abbrechen()

		if _, _, err := bauer.Stilplan(ctx, stilkarte("rot"), service.Vorlage{Stil: saatstil(t)}); err == nil {
			t.Error("kein Fehler")
		}
	})

	t.Run("ein Stil ohne Kartengröße", func(t *testing.T) {
		_, _, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{})
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler: %v", err)
		}
	})

	// **Ein Feld ohne Schrift wird benannt**, nicht weggelassen: es wird über
	// ein Bild, eine Maske oder eine Symbolschrift gezeichnet, und die sind
	// noch nicht gedeutet (§8).
	t.Run("ein Feld ohne Schrift", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Schrift.Name = ""

		plan, befunde, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		// **Drei seit #770**: Zutat, Grund und das übrige Textfeld — das
		// schriftlose fällt weg.
		if len(befunde) != 1 || len(plan.Ebenen) != 3 {
			t.Errorf("%d Befunde, %d Ebenen", len(befunde), len(plan.Ebenen))
		}
	})

	// **Ohne `width` und ohne `right` reicht das Feld bis zum Kartenrand.**
	t.Run("ein Feld ohne Breite", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Breite = ""

		plan, _, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if breite := textebeneMit(t, plan, "Feuerball").Rahmen.Breite; breite != 375-24 {
			t.Errorf("Breite %g, erwartet 351", breite)
		}
	})

	// **`right` statt `width`** — der Fall aus `magic-m15.mse-style`.
	t.Run("ein Feld mit Gegenkante", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Breite = ""
		stil.Felder[0].Rechts = "{ 341 - 20 }"

		plan, _, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if breite := textebeneMit(t, plan, "Feuerball").Rahmen.Breite; breite != 321-24 {
			t.Errorf("Breite %g, erwartet 297", breite)
		}
	})

	// Eine Angabe, die einen Text ergibt, der eine Zahl ist, gilt trotzdem.
	t.Run("eine Zahl als Text", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Oben = `{ "42" }`

		plan, befunde, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if name := textebeneMit(t, plan, "Feuerball"); len(befunde) > 0 || name.Rahmen.Oben != 42 {
			t.Errorf("Befunde %v, Oben %g", befunde, name.Rahmen.Oben)
		}
	})

	// **Eine Angabe, die gar nicht dasteht, ist null und kein Befund.**
	// Ein Stil führt nicht jedes Feld vollständig; `left:` fehlt bei einem
	// Feld, das links steht.
	t.Run("eine Angabe, die fehlt", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Links = ""

		plan, befunde, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if name := textebeneMit(t, plan, "Feuerball"); len(befunde) > 0 || name.Rahmen.Links != 0 {
			t.Errorf("Befunde %v, Links %g", befunde, name.Rahmen.Links)
		}
	})

	// Und eine, die keine Zahl ergibt, ist ein Befund.
	t.Run("ein Text, der keine Zahl ist", func(t *testing.T) {
		stil := saatstil(t)
		stil.Felder[0].Oben = `{ "oben" }`

		_, befunde, err := bauer.Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		if len(befunde) != 1 {
			t.Errorf("Befunde: %v", befunde)
		}
	})
}

// **Die Ausrichtung und die Schriftfarbe kommen aus dem Stil.**
func TestAusrichtungUndFarbe(t *testing.T) {
	faelle := map[string]struct {
		angabe   string
		farbe    string
		erwartet render.Ausrichtung
		weiss    bool
	}{
		"unten links":      {"bottom shrink-overflow", "black", render.LinksUnten, false},
		"mitte rechts":     {"middle right", "white", render.RechtsMitte, true},
		"oben mitte":       {"top center", "", render.MitteOben, false},
		"ohne Angabe":      {"", "", render.LinksOben, false},
		"unbekannte Worte": {"schräg", "grün", render.LinksOben, false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			stil := saatstil(t)
			stil.Felder[0].Ausrichtung = fall.angabe
			stil.Felder[0].Schrift.Farbe = fall.farbe

			plan, _, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			ebene := textebeneMit(t, plan, "Feuerball")
			if ebene.Satz.Ausrichtung != fall.erwartet {
				t.Errorf("Ausrichtung %v, erwartet %v", ebene.Satz.Ausrichtung, fall.erwartet)
			}

			if weiss := ebene.Farbe.R == 255 && ebene.Farbe.G == 255; weiss != fall.weiss {
				t.Errorf("Farbe %+v", ebene.Farbe)
			}
		})
	}
}

// Eine rote Schrift ist rot — der dritte Weg durch die Farbdeutung.
func TestEineRoteSchriftIstRot(t *testing.T) {
	stil := saatstil(t)
	stil.Felder[0].Schrift.Farbe = "red"

	plan, _, err := service.NeuerPlanbauer().Stilplan(t.Context(), stilkarte("rot"), service.Vorlage{Stil: stil})
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if farbe := textebeneMit(t, plan, "Feuerball").Farbe; farbe.R != 255 || farbe.G != 0 {
		t.Errorf("Farbe %+v", farbe)
	}
}
