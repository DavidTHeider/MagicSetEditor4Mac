package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Farbfeld trägt eine Farbe, und seine Vorgabe ist die erste Wahl.**
//
//	value( !field->initial.isDefault() ? field->initial()
//	     : !field->choices.empty()     ? field->choices[0]->color
//	     :                               Color()
//	     , true)
//
// `ColorValue::ColorValue`, `src/data/field/color.cpp`, Zeilen 84–88, Fassung
// f9d9356 vom 2021-07-06.
//
// # Was das gekostet hat
//
// `magic.mse-game` schreibt
// `color := styling.copyright_text_color or else set.copyright_text_color`, und
// `set field: copyright text color` ist ein Farbfeld mit den Wahlen `white` und
// `black`. Ohne die Vorgabe stand dort der leere Text, `recolor_image` bekam
// keine Farbe — und **vier Ebenen je Karte** wurden in ihrer Ausgangsfarbe
// gezeichnet (#270).
func TestEinFarbfeldTraegtEineFarbe(t *testing.T) {
	faelle := []struct {
		name     string
		feld     domain.Field
		wert     string
		erwartet string
	}{
		{
			name: "die erste Wahl",
			feld: domain.Field{
				Name: "grund", Kind: domain.KindColor,
				Wahlfarben: []string{"rgb(255,255,255)", "rgb(0,0,0)"},
			},
			erwartet: "rgba(255,255,255,255)",
		},
		{
			name: "`initial:` schlägt die erste Wahl",
			feld: domain.Field{
				Name: "grund", Kind: domain.KindColor, Initial: "rgb(1,2,3)",
				Wahlfarben: []string{"rgb(255,255,255)"},
			},
			erwartet: "rgba(1,2,3,255)",
		},
		{
			// **Ohne Wahl und ohne `initial:` ist es Schwarz** — MSEs
			// `Color()`, und das ist kein Erfundenes: es ist der Nullwert
			// seines Farbtyps.
			name:     "ohne alles Schwarz",
			feld:     domain.Field{Name: "grund", Kind: domain.KindColor},
			erwartet: "rgba(0,0,0,255)",
		},
		{
			// **Der Wert des Sets schlägt die Vorgabe.**
			name: "der ausgefüllte Wert",
			feld: domain.Field{
				Name: "grund", Kind: domain.KindColor,
				Wahlfarben: []string{"rgb(255,255,255)"},
			},
			wert:     "rgb(9,8,7)",
			erwartet: "rgba(9,8,7,255)",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			vorlage := service.Vorlage{
				Stil: msepaket.Stil{
					Spiel: "probe", Breite: 100, Hoehe: 100,
					Felder: []msepaket.Feldstil{{
						Feld: "name", Links: "0", Oben: "0", Breite: "90", Hoehe: "10",
						Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
					}},
				},
				Setfelder: []domain.Field{fall.feld},
				Kartenfelder: []domain.Field{{
					Name: "name", Kind: domain.KindText, Skript: "to_string(set.grund)",
				}},
			}

			if fall.wert != "" {
				vorlage.Setangaben = []domain.Value{{Field: "grund", Text: fall.wert}}
			}

			plan, befunde, err := service.NeuerPlanbauer().
				Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != fall.erwartet {
				t.Errorf("gezeichnet wird %q, erwartet %q",
					plan.Ebenen[len(plan.Ebenen)-1].Klartext(), fall.erwartet)
			}
		})
	}
}

// **Ein Farbfeld zeichnet eine Fläche — beim Kartenrand einen Rahmen.**
//
//	void ColorValueViewer::draw(RotatedDC& dc) {
//	  dc.SetPen(*wxTRANSPARENT_PEN);
//	  dc.SetBrush(value().value());
//	  … const AlphaMask& alpha_mask = getMask(dc);
//	  if (alpha_mask.isLoaded()) {
//	    dc.DrawImage(alpha_mask.colorImage(value().value()), RealPoint(0,0), style().combine);
//	  } else { … dc.DrawRoundedRectangle(dc.getInternalRect(), style().radius); }
//	  drawFieldBorder(dc);
//	}
//
// `src/render/value/color.cpp`, Zeilen 18–66, Fassung f9d9356 vom 2021-07-06.
//
// **Der schwarze Kartenrand ist dieses Feld.** `magic-m15.mse-style` gibt
// `border color` die Breiten 17/17/17/18, einen Radius von 18 und
// `mask: border_mask.png`; bis zu diesem Werkstück wurde es gar nicht
// gezeichnet und stattdessen gemeldet, es nenne „weder eine Schrift noch ein
// Bild" — 134-mal über die 190 Stile des Packs
// (#270).
func TestEinFarbfeldZeichnetEineFlaeche(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "border color", Links: "0", Oben: "0", Breite: "100", Hoehe: "100",
				Ecke: "18",
				Rand: msepaket.Randangabe{Links: "17", Rechts: "17", Oben: "17", Unten: "18"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "border color", Kind: domain.KindColor,
			Wahlfarben: []string{"rgb(200,180,0)"},
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Zwei Ebenen**: der weiße Kartengrund und das Farbfeld darauf.
	if len(plan.Ebenen) != 2 {
		t.Fatalf("%d Ebenen: %+v", len(plan.Ebenen), plan.Ebenen)
	}

	ebene := plan.Ebenen[1]

	if ebene.Art != render.ArtFlaeche {
		t.Errorf("die Ebene ist %q, erwartet eine Fläche", ebene.Art)
	}

	// **Die Farbe der ersten Wahl**, weil die Karte keine trägt.
	if ebene.Farbe != (render.Farbe{R: 200, G: 180, A: 255}) {
		t.Errorf("die Farbe ist %v", ebene.Farbe)
	}

	if ebene.Ecke != 18 {
		t.Errorf("der Radius ist %v, erwartet 18", ebene.Ecke)
	}

	if ebene.Rand == nil || *ebene.Rand !=
		(render.Randbreiten{Links: 17, Rechts: 17, Oben: 17, Unten: 18}) {
		t.Errorf("die Randbreiten sind %+v", ebene.Rand)
	}
}

// **Eine Stiloption vom Typ `color` trägt eine Farbe.**
//
// `magic-m15.mse-style` führt sieben davon — `text_color`, `textbox_color`,
// `illustrator_text_color` —, und die Vorlagen lesen sie, wo eine Farbe
// hingehört: `font_color(styling.text_color)` steht 72-mal im Pack
// (#270).
func TestEineStiloptionTraegtEineFarbe(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Stiloptionen: []domain.StyleField{
				{Name: "text color", Kind: domain.KindColor, Wahlfarben: []string{"rgb(4,5,6)"}},
				{Name: "randlos", Kind: domain.KindBoolean, Initial: "no"},
			},
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "90", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "name", Kind: domain.KindText,
			Skript: `to_string(styling.text_color) + "|" + styling.randlos`,
		}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	// **Und eine Option, die keine Farbe ist, bleibt Text** — `styling.randlos`
	// wird zu `no` und nicht zu Schwarz.
	if haben := plan.Ebenen[1].Klartext(); haben != "rgba(4,5,6,255)|no" {
		t.Errorf("gezeichnet wird %q", haben)
	}
}

// **Eine einzelne Randbreite macht noch keinen Rahmen.**
//
//	ColorStyle::ColorStyle(const ColorFieldP& field)
//	  : … left_width(100000), right_width (100000)
//	  , top_width (100000), bottom_width(100000)
//
// `src/data/field/color.cpp`, Zeilen 58–63, Fassung f9d9356 vom 2021-07-06.
// MSEs `clip` verlangt, dass **alle vier** kleiner als das Feld sind; wer die
// fehlenden mit null vorbelegte, ließe die Füllung verschwinden.
func TestEineEinzelneRandbreiteFuelltGanz(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "border color", Links: "0", Oben: "0", Breite: "100", Hoehe: "100",
				Rand: msepaket.Randangabe{Links: "17"},
			}},
		},
		Kartenfelder: []domain.Field{{Name: "border color", Kind: domain.KindColor}},
	}

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	rand := plan.Ebenen[1].Rand
	if rand == nil || rand.Links != 17 || rand.Rechts != 100000 {
		t.Fatalf("die Randbreiten sind %+v", rand)
	}

	if rand.Rahmt(100, 100) {
		t.Error("eine einzelne Breite ergab einen Rahmen")
	}
}

// **Ohne Breiten wird gefüllt.** Ein Farbfeldstil ohne `left width:` und
// Geschwister rahmt nichts — MSEs Vorgaben sind 100000, und die Ebene trägt
// dann gar keine Breiten.
func TestEinFarbfeldOhneBreitenFuellt(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "card color", Links: "0", Oben: "0", Breite: "100", Hoehe: "100",
			}},
		},
		Kartenfelder: []domain.Field{{
			Name: "card color", Kind: domain.KindColor, Initial: "rgb(1,2,3)",
		}},
	}

	plan, _, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if ebene := plan.Ebenen[1]; ebene.Rand != nil {
		t.Errorf("die Randbreiten sind %+v", ebene.Rand)
	}
}
