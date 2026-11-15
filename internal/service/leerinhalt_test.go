package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein leeres Feld meldet seinen leeren Satz, nicht seine Kastenmaße.**
//
// MSE2 nimmt `content_width`, `content_height` und `content_lines` aus dem
// gesetzten Text (`reflect_layout`, `src/data/field/text.cpp`, Zeilen
// 114–121; `update_size`, `src/render/text/viewer.cpp`, Zeilen 350–353;
// Fassung f9d9356 vom 2021-07-06). **Gefragt am 2026-10-11**, mit einem
// Orakelstil, der die drei Maße des Felds `illustrator` als Lage auf die
// Karte schreibt: leer gibt Breite 0, eine Zeile und dieselbe Höhe wie
// „Probe Zeichner". Hier stand für ein leeres Feld die deklarierte Breite —
// bei `magic-nouveau` rückte das Pinselzeichen
// (`left: { 150 + card_style.illustrator.content_width * -0.5 }`) dadurch
// von der Mitte an den Rand.
//
// **Eine leere Wahl ohne Bildanteil meldet Nullmaße**: `ChoiceStyle` beginnt
// bei `content_width(0.0), content_height(0.0)` und bekommt erst ein Maß, wenn
// ein Bild vorbereitet ist (`src/data/field/choice.cpp`, Zeile 168;
// `src/render/value/choice.cpp`, Zeilen 35–58) — mit Bild siehe
// [TestEineWahlMeldetDasMassIhresBilds].
func TestEinLeeresFeldMeldetSeinenLeerenSatz(t *testing.T) {
	maße := func(t *testing.T, art domain.FieldKind, wert string) (breite, hoehe, zeilen float64) {
		t.Helper()

		stil := msepaket.Stil{
			Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{
				{
					Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "200",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
				{
					Feld: "breite", Oben: "40", Hoehe: "5", Breite: "5",
					Links:   "{ 10 + card_style.name.content_width }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
				{
					Feld: "hoehe", Oben: "50", Hoehe: "5", Breite: "5",
					Links:   "{ 10 + card_style.name.content_height }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
				{
					Feld: "zeilen", Oben: "60", Hoehe: "5", Breite: "5",
					Links:   "{ 10 + card_style.name.content_lines }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
			},
		}

		karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
			{Field: "name", Text: wert},
			{Field: "breite", Text: "b"},
			{Field: "hoehe", Text: "h"},
			{Field: "zeilen", Text: "z"},
		}}

		vorlage := service.Vorlage{
			Stil: stil,
			Kartenfelder: []domain.Field{
				{Name: "name", Kind: art, Choices: []string{"eins"}},
				{Name: "breite", Kind: domain.KindText},
				{Name: "hoehe", Kind: domain.KindText},
				{Name: "zeilen", Kind: domain.KindText},
			},
		}

		plan, _, err := service.NeuerPlanbauer().
			MitSchriften(schriftprobe{breite: 0.5}).
			Stilplan(t.Context(), karte, vorlage)
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		lage := map[string]float64{}
		for _, ebene := range plan.Ebenen {
			lage[ebene.Klartext()] = ebene.Rahmen.Links - 10
		}

		return lage["b"], lage["h"], lage["z"]
	}

	t.Run("ein leerer Text hat keine Breite, aber eine Zeile", func(t *testing.T) {
		breite, hoehe, zeilen := maße(t, domain.KindText, "")
		_, gesetzt, _ := maße(t, domain.KindText, "abcd")

		if breite != 0 || zeilen != 1 || hoehe != gesetzt || hoehe <= 0 {
			t.Errorf("Breite %.1f, Höhe %.1f, Zeilen %.0f — erwartet 0, %.1f, 1", breite, hoehe, zeilen, gesetzt)
		}
	})

	t.Run("eine leere Wahl hat Nullmaße", func(t *testing.T) {
		breite, hoehe, _ := maße(t, domain.KindChoice, "")

		if breite != 0 || hoehe != 0 {
			t.Errorf("Breite %.1f, Höhe %.1f — erwartet 0 und 0", breite, hoehe)
		}
	})
}

// **Ein leerer Symbolzug misst sein Ersatzsymbol** — die Manakosten eines
// Lands. MSE2 nimmt für die Höhe der leeren Zeile dann nicht die Schrift,
// sondern `defaultSymbolSize`: das erste Muster, das „0" trifft, samt
// Abstand (`TextViewer::prepareLines`, `src/render/text/viewer.cpp`, Zeilen
// 397–401; `SymbolFont::defaultSymbol` und `defaultSymbolSize`,
// `src/data/symbol_font.cpp`, Zeilen 263–268 und 403–410; Fassung f9d9356
// vom 2021-07-06).
func TestEinLeererSymbolzugMisstSeinErsatzsymbol(t *testing.T) {
	messe := func(t *testing.T, symbole []msepaket.Symbol, wert string) (breite, hoehe float64) {
		t.Helper()

		stil := msepaket.Stil{
			Breite: 375, Hoehe: 523,
			Felder: []msepaket.Feldstil{
				{
					Feld: "casting cost", Links: "300", Oben: "29", Breite: "40", Hoehe: "20",
					Symbolschrift: msepaket.Symbolangabe{Name: "probe-symbole", Groesse: "15"},
					ImmerSymbole:  true,
				},
				{
					Feld: "breite", Oben: "40", Hoehe: "5", Breite: "5",
					Links:   "{ 10 + card_style.casting_cost.content_width }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
				{
					Feld: "hoehe", Oben: "50", Hoehe: "5", Breite: "5",
					Links:   "{ 10 + card_style.casting_cost.content_height }",
					Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
				},
			},
		}

		plan, _, err := service.NeuerPlanbauer().
			MitSchriften(schriftprobe{breite: 0.5}).
			Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "casting cost", Text: wert},
				{Field: "breite", Text: "b"},
				{Field: "hoehe", Text: "h"},
			}}, service.Vorlage{
				Stil: stil,
				Kartenfelder: []domain.Field{
					{Name: "casting cost", Kind: domain.KindText},
					{Name: "breite", Kind: domain.KindText},
					{Name: "hoehe", Kind: domain.KindText},
				},
				Symbolschriften: map[string]msepaket.Symbolschrift{
					"probe-symbole": {Name: "probe-symbole", Symbole: symbole, AbstandSenkrecht: 2},
				},
			})
		if err != nil {
			t.Fatalf("Plan: %v", err)
		}

		lage := map[string]float64{}
		for _, ebene := range plan.Ebenen {
			lage[ebene.Klartext()] = ebene.Rahmen.Links - 10
		}

		return lage["b"], lage["h"]
	}

	mitZahl := []msepaket.Symbol{
		{Code: "W", Bild: "w.png"},
		{Code: "[0-9]+", Bild: "zahl.png", Ausdruck: true},
	}

	t.Run("mit einem Muster für Zahlen", func(t *testing.T) {
		breite, hoehe := messe(t, mitZahl, "")
		_, null := messe(t, mitZahl, "0")

		if breite != 0 || hoehe != null || hoehe <= 0 {
			t.Errorf("Breite %.1f, Höhe %.1f — erwartet 0 und %.1f", breite, hoehe, null)
		}
	})

	// `add_diagonal(RealSize(1,1), spacing)`.
	t.Run("ohne ein solches Muster ein Bildpunkt samt Abstand", func(t *testing.T) {
		breite, hoehe := messe(t, []msepaket.Symbol{{Code: "W", Bild: "w.png"}}, "")

		if breite != 0 || hoehe != 3 {
			t.Errorf("Breite %.1f, Höhe %.1f — erwartet 0 und 3", breite, hoehe)
		}
	})
}

// **Wo sich der leere Symbolzug nicht messen lässt, bleibt es beim Kasten**,
// wie bei einem gefüllten ([TestOhneAusrechenbareSymbolgroesseWirdNichtGemessen]);
// ein unlesbares Muster wird übersprungen wie beim Zerlegen.
func TestEinLeererSymbolzugOhneMassBleibtBeimKasten(t *testing.T) {
	for _, fall := range []struct {
		name    string
		schrift string
		groesse string
		symbole []msepaket.Symbol
		breite  float64
	}{
		{name: "die Symbolschrift fehlt", schrift: "gibtsnicht", groesse: "15", breite: 40},
		{name: "die Kantenlänge lässt sich nicht ausrechnen", schrift: "probe-symbole", groesse: "{ gibtsnicht }", breite: 40},
		{
			name: "ein unlesbares Muster wird übersprungen", schrift: "probe-symbole", groesse: "15",
			symbole: []msepaket.Symbol{
				{Code: "(", Bild: "kaputt.png", Ausdruck: true},
				{Code: "[0-9]+", Bild: "zahl.png", Ausdruck: true},
			},
			breite: 0,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523,
				Felder: []msepaket.Feldstil{
					{
						Feld: "casting cost", Links: "300", Oben: "29", Breite: "40", Hoehe: "{ gibtsnicht }",
						Symbolschrift: msepaket.Symbolangabe{Name: fall.schrift, Groesse: fall.groesse},
						ImmerSymbole:  true,
					},
					{
						Feld: "breite", Oben: "40", Hoehe: "5", Breite: "5",
						Links:   "{ 10 + card_style.casting_cost.content_width }",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			plan, _, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "breite", Text: "b"},
				}}, service.Vorlage{
					Stil: stil,
					Kartenfelder: []domain.Field{
						{Name: "casting cost", Kind: domain.KindText},
						{Name: "breite", Kind: domain.KindText},
					},
					Symbolschriften: map[string]msepaket.Symbolschrift{
						"probe-symbole": {Name: "probe-symbole", Symbole: fall.symbole},
					},
				})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Klartext() == "b" {
					if breite := ebene.Rahmen.Links - 10; breite != fall.breite {
						t.Errorf("content_width %.1f, erwartet %.1f", breite, fall.breite)
					}

					return
				}
			}

			t.Fatal("der Plan führt das Messfeld nicht")
		})
	}
}

// **Auch die Zusatzstile melden ihren gesetzten Inhalt** — Befund 7 des
// Zweitblicks über E/F. Hier bekam `extra_card_style` die Kastenmaße, wo
// `card_style` den Satz misst; MSE2 setzt beide mit demselben `TextViewer`.
func TestDieZusatzstileMeldenIhrenInhalt(t *testing.T) {
	for _, fall := range []struct {
		name, wert, pfad string
	}{
		{name: "ein gesetzter Text", wert: `"abcd"`, pfad: "w20.png"},
		{name: "ein leerer Text", wert: `""`, pfad: "w0.png"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			vorlage := service.Vorlage{
				Stil: msepaket.Stil{
					Spiel: "probe", Breite: 375, Hoehe: 523,
					Felder: []msepaket.Feldstil{{
						Feld: "probe", Links: "0", Oben: "100", Breite: "30", Hoehe: "12",
						Darstellung: "image", Bild: `{ "w" + to_string(extra_card_style.label.content_width) + ".png" }`,
					}},
					Zusatzfelder: []domain.Field{{Name: "label", Kind: domain.KindText, Skript: fall.wert}},
					Zusatzstile: []msepaket.Feldstil{{
						Feld: "label", Links: "0", Oben: "0", Breite: "100", Hoehe: "20",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					}},
				},
				Kartenfelder: []domain.Field{{Name: "probe", Kind: domain.KindChoice}},
			}

			plan, befunde, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "probe", Text: "x"}}}, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Feld == "probe" {
					if !strings.HasSuffix(ebene.Bild.Pfad, "/"+fall.pfad) {
						t.Errorf("das Bild ist %q, erwartet %s (Befunde %v)", ebene.Bild.Pfad, fall.pfad, befunde)
					}

					return
				}
			}

			t.Fatalf("keine Ebene für probe, Befunde %v", befunde)
		})
	}
}

// **Eine leere Wahl mit Stilbild meldet ihren Kasten** — Befund 8 des
// Zweitblicks über E/F, und seit #2303 gefragt statt erschlossen.
//
// MSE2 bereitet das Stilbild auch für eine leere Eingabe vor und übernimmt
// das Maß des erzeugten Bilds (`prepare_choice_viewer`,
// `src/render/value/choice.cpp`, Zeilen 35–58). Ein Skript, das `""` gibt,
// ergibt ein **leeres Bild in der gefragten Größe** (`BlankImage`,
// `src/script/value.cpp`, Zeilen 369–375) — ohne `alignment:` gestreckt auf
// den Kasten. **Hier stand für diesen Fall null**, erschlossen aus „für die
// leere kommt keines zustande". Nur ein Bild, das sich nicht erzeugen lässt,
// lässt das Maß bei null. Fassung f9d9356 vom 2021-07-06.
func TestEineLeereWahlMitFestemBildMeldetIhrenKasten(t *testing.T) {
	for _, fall := range []struct {
		name, bild string
		breite     float64
	}{
		{name: "ein festes Stilbild", bild: `{ "fest.png" }`, breite: 200},
		{name: "ein Bild aus der Wahl", bild: `{ if input == "" then "" else input + ".png" }`, breite: 200},
		{name: "ein Bild, das fehlt", bild: `{ "fehlt.png" }`, breite: 0},
	} {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{
				Breite: 375, Hoehe: 523, Paket: "probe.mse-style",
				Felder: []msepaket.Feldstil{
					{
						Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "200",
						Darstellung: "image", Bild: fall.bild,
					},
					{
						Feld: "breite", Oben: "40", Hoehe: "5", Breite: "5",
						Links:   "{ 10 + card_style.name.content_width }",
						Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
					},
				},
			}

			plan, _, err := service.NeuerPlanbauer().
				MitBildern(massquelle{"/probe.mse-style/fest.png": {X: 50, Y: 50}}).
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
					{Field: "name", Text: ""}, {Field: "breite", Text: "b"},
				}}, service.Vorlage{Stil: stil, Kartenfelder: []domain.Field{
					{Name: "name", Kind: domain.KindChoice, Choices: []string{"eins"}},
					{Name: "breite", Kind: domain.KindText},
				}})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, ebene := range plan.Ebenen {
				if ebene.Klartext() == "b" {
					if breite := ebene.Rahmen.Links - 10; breite != fall.breite {
						t.Errorf("content_width %.1f, erwartet %.1f", breite, fall.breite)
					}

					return
				}
			}

			t.Fatal("der Plan führt das Messfeld nicht")
		})
	}
}
