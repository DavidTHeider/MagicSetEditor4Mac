package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Das Skript einer Maßangabe sieht `value`** — den bisherigen Wert **dieser
// Angabe** (#555).
//
// # Wo MSE ihn bindet
//
//	inline bool update(Context& ctx) {
//	  return script.invokeOn(ctx, value);
//	}
//
//	template <typename T> bool invokeOn(Context& ctx, T& value) const {
//	  if (script) {
//	    T new_value;
//	    ctx.setVariable(SCRIPT_VAR_value, to_script(value));
//
// `src/script/scriptable.hpp`, Zeilen 134–136 und 58–66, gerufen aus
// `Style::update`, `src/data/field.cpp`, Zeilen 141–149. Fassung f9d9356 vom
// 2021-07-06.
//
// # Und womit er vorbelegt ist
//
//	Style::Style(const FieldP& field)
//	  : … left (1000000), top   (1000000)
//	  , width(0),       height(0)
//	  , right(1000000), bottom(1000000)
//
// `src/data/field.cpp`, Zeilen 97–104. **Nicht null für alle vier**: die
// Million ist zugleich MSEs Zeichen für „nicht gesetzt" (`is_set`, Zeile 132),
// und ein Skript, das sie abfragt, bekäme bei null eine andere Antwort.
//
// # Was das im Pack betrifft
//
// Über alle 376 Stilpakete steht `value` als **freier Name** in genau **einer**
// Maßangabe — `magic-m15-mainframe-dfc.mse-style/style`, Zeile 4059:
//
//	top: {if match(value, match:"^<sym>[^<]+</sym>$") then 472 else 471}
//
// Die vier weiteren Fundstellen heißen `value:` und sind **Argumentnamen** von
// `font_field(…)`, kein freier Name.
//
// **Der Rest ist Zahl.** `match(1000000, …)` trifft das Muster nicht, also 471
// — und weil MSE die Angabe danach auf 471 setzt, trifft es auch im zweiten
// Durchlauf nicht. MSE zeichnet dort **immer** 471; das Skript liest die
// Angabe, wo sein Verfasser das Feld gemeint hat.
//
// Ohne die Bindung meldete dieses Programm über den Kanon **70-mal** „der Name
// `value` ist hier nicht bekannt" — sichtbar geworden, seit die kartenweisen
// Stiloptionen den Zweig betretbar machen
// (#547).
func TestEineMassangabeSiehtIhrenBisherigenWert(t *testing.T) {
	faelle := []struct {
		name        string
		stil        msepaket.Feldstil
		links, oben float64
		breite      float64
	}{
		{
			name: "die Kante ist mit einer Million vorbelegt",
			stil: msepaket.Feldstil{
				Links: "0", Breite: "40",
				Oben: "{ if value == 1000000 then 5 else 90 }", Hoehe: "20",
			},
			links: 0, oben: 5, breite: 40,
		},
		{
			name: "die Länge ist mit null vorbelegt",
			stil: msepaket.Feldstil{
				Links: "0", Oben: "5", Hoehe: "20",
				Breite: "{ if value == 0 then 40 else 90 }",
			},
			links: 0, oben: 5, breite: 40,
		},
		{
			name: "die Gegenkante ebenfalls mit einer Million",
			stil: msepaket.Feldstil{
				Oben: "5", Hoehe: "20", Breite: "40",
				Rechts: "{ if value == 1000000 then 40 else 900 }",
			},
			links: 0, oben: 5, breite: 40,
		},
		{
			name: "und die eine Zeile aus dem Pack ergibt 471",
			stil: msepaket.Feldstil{
				Links: "0", Breite: "40", Hoehe: "20",
				Oben: `{if match(value, match:"^<sym>[^<]+</sym>$") then 472 else 471}`,
			},
			links: 0, oben: 471, breite: 40,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stil := msepaket.Stil{Breite: 375, Hoehe: 523, Felder: []msepaket.Feldstil{{
				Feld: "text", Links: fall.stil.Links, Oben: fall.stil.Oben,
				Rechts: fall.stil.Rechts, Breite: fall.stil.Breite, Hoehe: fall.stil.Hoehe,
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			}}}

			karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
				{Field: "text", Text: "abc"},
			}}

			plan, befunde, err := service.NeuerPlanbauer().
				MitSchriften(schriftprobe{breite: 0.5}).
				Stilplan(t.Context(), karte, service.Vorlage{Stil: stil})
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}

			for _, befund := range befunde {
				t.Errorf("Befund: %s — %s", befund.Ort, befund.Meldung)
			}

			for stelle := range plan.Ebenen {
				if plan.Ebenen[stelle].Art != render.ArtText {
					continue
				}

				rahmen := plan.Ebenen[stelle].Rahmen
				if rahmen.Links != fall.links || rahmen.Oben != fall.oben ||
					rahmen.Breite != fall.breite {
					t.Errorf("Rahmen (%g,%g) %g breit, erwartet (%g,%g) %g breit",
						rahmen.Links, rahmen.Oben, rahmen.Breite,
						fall.links, fall.oben, fall.breite)
				}

				return
			}

			t.Fatal("der Plan führt keine Textebene")
		})
	}
}
