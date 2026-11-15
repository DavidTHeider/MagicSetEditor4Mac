package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Stil ist in der Ausfuhr ein Gegenstand** — QA-04-F3
// (#2039).
//
// Gefragt am Orakel (MSE2, Fassung f9d9356 vom 2021-07-06, über
// `Multicolor.mse-set`): `set.stylesheet` ist ein `<stylesheet>`, sein
// `short_name` und `full_name` die des Pakets, `card_width` die Breite.
// `card.stylesheet` ist der **eigene** Stil einer Karte (`REFLECT(stylesheet)`,
// `src/data/card.cpp`, Zeile 77) — ohne einen `nil`, und die Vorlagen fallen
// dann auf `set.stylesheet` zurück.
//
// **`card_regions` kommt aus MSE 2.5** (Vermerk 0165): das Pack ist für 2.5.8
// geschrieben, und 2.1.2 bricht die Lackey-Ausfuhr an diesem Glied ganz ab.
// Hier ist es eine Liste der Regionen, die der Stil erklärt — leer ohne; acht
// Stile des Packs erklären welche (N1 des Nachlaufs über die QA-Arbeit).
func TestDieAusfuhrKenntDenStilAlsGegenstand(t *testing.T) {
	t.Parallel()

	stile := func(paket string) (msepaket.Stil, bool) {
		switch paket {
		case "":
			return msepaket.Stil{
				Paket: "probe-mainframe.mse-style", KurzerName: "Mainframe",
				VollerName: "Probe Mainframe", Breite: 375, Hoehe: 523, DPI: 150,
			}, true
		case "eigen":
			return msepaket.Stil{
				Paket: "probe-eigen.mse-style", KurzerName: "Eigen",
				VollerName: "Eigener Stil", Breite: 523, Hoehe: 375, DPI: 150,
			}, true
		case "zweiseitig":
			return msepaket.Stil{
				Paket: "probe-zweiseitig.mse-style", VollerName: "Zweiseitig",
				Breite: 752, Hoehe: 523, DPI: 150,
				Regionen: []msepaket.Kartenregion{
					{Name: "front", Breite: 375, Hoehe: 523},
					{Name: "back", X: 377.5, Breite: 375, Hoehe: 523},
				},
			}, true
		}

		return msepaket.Stil{}, false
	}

	karten := []*domain.CardRow{
		{Values: []domain.Value{{Field: "name", Text: "Ohne"}}},
		{Values: []domain.Value{{Field: "name", Text: "Mit"}, {Field: "stylesheet", Text: "eigen"}}},
		{Values: []domain.Value{{Field: "name", Text: "Doppelt"}, {Field: "stylesheet", Text: "zweiseitig"}}},
	}

	faelle := []struct {
		name, skript, erwartet string
	}{
		{"der Vollname des Setstils", `set.stylesheet.full_name`, "Probe Mainframe"},
		{"der Kurzname des Setstils", `set.stylesheet.short_name`, "Mainframe"},
		{"die Maße des Setstils", `set.stylesheet.card_width + "x" + set.stylesheet.card_height`, "375x523"},
		{"keine Regionen", `length(set.stylesheet.card_regions)`, "0"},
		{"eine Karte ohne eigenen Stil hat nil", `if cards[0].stylesheet == nil then "nil" else "etwas"`, "nil"},
		{"eine Karte mit eigenem Stil hat ihn", `cards[1].stylesheet.full_name`, "Eigener Stil"},
		{
			"der Rückfall der Lackey-Vorlage nimmt das Einzelbild",
			`to_text(for each card in cards do (
				sts := card.stylesheet
				if sts == nil then sts := set.stylesheet
				if length(sts.card_regions) > 0 then "viele" else "eines"
			))`,
			"eineseinesviele",
		},
		{"die Regionen eines Stils, der sie erklärt", `length(cards[2].stylesheet.card_regions)`, "2"},
		{
			"jede Region mit Namen und Maßen",
			`r := cards[2].stylesheet.card_regions[1]
			r.name + ":" + r.x + "," + r.y + "," + r.width + "x" + r.height`,
			"back:377.5,0,375x523",
		},
		{
			"derselbe Stil ist derselbe Gegenstand",
			`if set.stylesheet == set.stylesheet and cards[1].stylesheet != set.stylesheet then "ja" else "nein"`,
			"ja",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{Stile: stile}).Fuehre(
				t.Context(), msepaket.Exportvorlage{Skript: fall.skript}, karten, nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v — %v", err, befunde)
			}

			if ausgabe != fall.erwartet {
				t.Errorf("%q, erwartet %q", ausgabe, fall.erwartet)
			}
		})
	}
}

// **Kennt die Quelle einen Stil nicht, ist er `nil` — und ein Befund sagt
// es.** Eine Vorlage, die `set.stylesheet` liest, fällt dann auf ihren Zweig
// für „kein Stil"; still ein leerer Gegenstand wäre ein Stil, den es nicht
// gibt. **Ohne Quelle** (ein Lauf ohne Bestand) ist er ebenso `nil`, aber
// ohne Befund: dort fragt niemand nach Stilen.
func TestEinUnbekannterStilIstNilUndEinBefund(t *testing.T) {
	t.Parallel()

	karten := []*domain.CardRow{
		{Values: []domain.Value{{Field: "name", Text: "Fremd"}, {Field: "stylesheet", Text: "fehlt"}}},
	}

	for name, fall := range map[string]struct {
		mittel service.Ausfuhrmittel
		befund bool
	}{
		"ohne Quelle": {service.Ausfuhrmittel{}, false},
		"die Quelle kennt ihn nicht": {
			service.Ausfuhrmittel{Stile: func(string) (msepaket.Stil, bool) { return msepaket.Stil{}, false }},
			true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ausgabe, befunde, err := probelauf(fall.mittel).Fuehre(t.Context(),
				msepaket.Exportvorlage{Skript: `(if set.stylesheet == nil then "a" else "b") + (if cards[0].stylesheet == nil then "a" else "b")`},
				karten, nil, nil)
			if err != nil {
				t.Fatalf("ausführen: %v — %v", err, befunde)
			}

			if ausgabe != "aa" {
				t.Errorf("%q, erwartet aa", ausgabe)
			}

			if (len(befunde) > 0) != fall.befund {
				t.Errorf("Befunde %v, erwartet einen: %v", befunde, fall.befund)
			}
		})
	}
}
