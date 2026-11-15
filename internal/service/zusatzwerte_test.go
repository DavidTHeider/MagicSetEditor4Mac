package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Zusatzfeld liest den Wert, den die Karte für ihren Stil trägt** —
// #2298.
//
// MSE2 hält ihn unter `extra_data:` je Stilvorlage
// (`Card::extraDataFor`, `src/data/card.cpp`, Zeilen 63–65) und rechnet ihn
// in `SetScriptManager::updateStyles` mit `Value::update`: das
// `default:`-Skript nur über einem Wert, der noch Vorgabe ist, das
// `script:` immer (`ChoiceValue::update`, `src/data/field/choice.cpp`,
// Zeilen 290–295; `src/script/script_manager.cpp`, Zeilen 208–222; Fassung
// f9d9356 vom 2021-07-06).
func TestEinZusatzfeldLiestDenGespeichertenWert(t *testing.T) {
	for _, fall := range []struct {
		name     string
		feld     domain.Field
		zusatz   string
		erwartet string
	}{
		{
			name:     "der gespeicherte Wert schlägt die Vorgabe",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`},
			zusatz:   "magic-probe:\n\tkeyword_icons: flying",
			erwartet: "flying.png",
		},
		{
			name:     "das Skript läuft über dem gespeicherten Wert",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Skript: `value + "-x"`},
			zusatz:   "magic-probe:\n\tkeyword_icons: flying",
			erwartet: "flying-x.png",
		},
		{
			name:     "ein Wert für eine andere Stilvorlage zählt nicht",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`},
			zusatz:   "magic-anders:\n\tkeyword_icons: flying",
			erwartet: "vorgabe.png",
		},
		{
			name:     "ohne gespeicherten Wert gilt die Vorgabe",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`},
			zusatz:   "magic-probe:\n\tanderes_feld: flying",
			erwartet: "vorgabe.png",
		},
		{
			// `MultipleChoiceValue::update` ruft `normalForm`
			// (`src/data/field/multiple_choice.cpp`, Zeilen 55–61, Fassung
			// f9d9356 vom 2021-07-06).
			name: "eine gespeicherte Mehrfachwahl kommt in Normalform",
			feld: domain.Field{
				Name: "keyword icons", Kind: domain.KindMultipleChoice,
				Choices: []string{"flying", "haste"},
			},
			zusatz:   "magic-probe:\n\tkeyword_icons: haste, flying",
			erwartet: "flying, haste.png",
		},
		{
			// Befund 6 des Zweitblicks über E/F: so steht der Wert in einem
			// Bestand, der vor #2298 eingelesen wurde.
			name:     "die flache Altform eines älteren Bestands",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`},
			zusatz:   "magic-anders:\nkeyword_icons: fremd\nmagic-probe:\nkeyword_icons: flying",
			erwartet: "flying.png",
		},
		{
			// Eine Zeile, die kein Eintrag ist, bleibt, wo sie steht — und
			// beginnt keinen Block.
			name:     "die flache Altform hinter einem Kommentar",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`},
			zusatz:   "# aus einem älteren Bestand\nmagic-probe:\nkeyword_icons: flying",
			erwartet: "flying.png",
		},
		{
			name:     "der Feldname mit Leerzeichen trifft den Schlüssel mit Unterstrich",
			feld:     domain.Field{Name: "keyword icons", Kind: domain.KindChoice},
			zusatz:   "magic-probe:\n\tkeyword icons: haste",
			erwartet: "haste.png",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			vorlage := service.Vorlage{
				Stil: msepaket.Stil{
					Spiel: "probe", Paket: "magic-probe.mse-style", Breite: 100, Hoehe: 100,
					Zusatzfelder: []domain.Field{fall.feld},
					Zusatzstile: []msepaket.Feldstil{{
						Feld: "keyword icons", Links: "0", Oben: "0", Breite: "30", Hoehe: "12",
						Darstellung: "image", Bild: `{ extra_card.keyword_icons + ".png" }`,
					}},
				},
			}

			karte := &domain.CardRow{ID: 1, Values: []domain.Value{
				{Field: "extra data", Text: fall.zusatz},
			}}

			plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if len(plan.Ebenen) != 2 {
				t.Fatalf("%d Ebenen, erwartet zwei", len(plan.Ebenen))
			}

			if pfad := plan.Ebenen[1].Bild.Pfad; !strings.HasSuffix(pfad, "/"+fall.erwartet) && pfad != fall.erwartet {
				t.Errorf("das Bild ist %q, erwartet %s", pfad, fall.erwartet)
			}
		})
	}
}

// **Unlesbare Zusatzdaten werden benannt, und es gelten die Vorgaben.**
//
// Ein Set mit ungültigem UTF-8 öffnet gar nicht erst; über die Kommandozeile
// oder MCP kann ein solcher Wert aber in den Bestand kommen. Dann zeichnet
// die Karte mit den Vorgaben, und der Befund sagt warum (§8) — statt still
// so zu tun, als trüge sie nichts.
func TestUnlesbareZusatzdatenWerdenBenannt(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Paket: "magic-probe.mse-style", Breite: 100, Hoehe: 100,
			Zusatzfelder: []domain.Field{{Name: "keyword icons", Kind: domain.KindChoice, Default: `"vorgabe"`}},
			Zusatzstile: []msepaket.Feldstil{{
				Feld: "keyword icons", Links: "0", Oben: "0", Breite: "30", Hoehe: "12",
				Darstellung: "image", Bild: `{ extra_card.keyword_icons + ".png" }`,
			}},
		},
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{
		{Field: "extra data", Text: "magic-probe:\n\tkeyword_icons: \xff"},
	}}

	plan, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) != 1 || befunde[0].Ort != "extra data" {
		t.Fatalf("Befunde: %v", befunde)
	}

	if pfad := plan.Ebenen[1].Bild.Pfad; !strings.HasSuffix(pfad, "/vorgabe.png") {
		t.Errorf("das Bild ist %q, erwartet die Vorgabe", pfad)
	}
}
