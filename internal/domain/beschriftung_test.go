package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Feld hat eine Beschriftung, und sie ist nicht sein Name.**
//
//	void Field::after_reading(Version ver) {
//	  name = canonical_name_form(name);
//	  if (caption.default_.empty()) caption.default_ = name_to_caption(name);
//	  …
//	}
//
// `src/data/field.cpp`, Zeilen 67–71, Fassung f9d9356 vom 2021-07-06.
// `name_to_caption` schreibt das **erste** Zeichen groß und macht aus jedem
// Unterstrich ein Leerzeichen — führende bleiben stehen
// (`src/util/string.cpp`, Zeilen 133–150).
//
// # Wofür
//
// Der Editor beschriftete jedes Feld mit seinem rohen Namen: `casting cost`,
// `card symbol`, `rule text`. MSE schreibt `Casting cost`. Das ist kein
// Schönheitsfehler — es ist der Unterschied zwischen einem Bezeichner und
// einer Beschriftung, und `caption:` im Paket schlägt beide
// (#560).
func TestDieBeschriftungEinesFeldes(t *testing.T) {
	faelle := []struct {
		name    string
		feld    domain.Field
		sprache string
		text    string
	}{
		{
			name: "der Name wird großgeschrieben",
			feld: domain.Field{Name: "casting cost"},
			text: "Casting cost",
		},
		{
			name: "ein Unterstrich wird ein Leerzeichen",
			feld: domain.Field{Name: "card_symbol"},
			text: "Card symbol",
		},
		{
			// `ret += leading ? c : ' '` — führende Trenner bleiben, wie sie
			// sind, und das erste **Zeichen** ist erst das danach.
			name: "ein führender Unterstrich bleibt stehen",
			feld: domain.Field{Name: "_intern"},
			text: "_Intern",
		},
		{
			name: "`caption:` schlägt den Namen",
			feld: domain.Field{Name: "pt", Beschriftung: "P/T"},
			text: "P/T",
		},
		{
			name: "die Sprache schlägt `caption:`",
			feld: domain.Field{
				Name: "title", Beschriftung: "Title",
				Uebersetzungen: map[string]string{"de": "Titel", "ch-s": "标题"},
			},
			sprache: "de",
			text:    "Titel",
		},
		{
			name: "eine Sprache ohne Eintrag fällt zurück",
			feld: domain.Field{
				Name: "title", Beschriftung: "Title",
				Uebersetzungen: map[string]string{"ch-s": "标题"},
			},
			sprache: "de",
			text:    "Title",
		},
		{
			// **Das Fenster sagt `zh`, das Paket `ch-s`** — MSE fragt mit dem
			// Namen seines Sprachpakets (`the_locale->name()`,
			// `src/data/locale.cpp`, Zeilen 138–141, Fassung f9d9356 vom
			// 2021-07-06). LOC-01, #1687.
			name: "Chinesisch findet den Eintrag des vereinfachten Pakets",
			feld: domain.Field{
				Name: "title", Beschriftung: "Title",
				Uebersetzungen: map[string]string{"ch-s": "标题", "ch-t": "標題"},
			},
			sprache: "zh",
			text:    "标题",
		},
		{
			name: "Japanisch findet den Eintrag des Pakets jp",
			feld: domain.Field{
				Name:                 "type",
				Listenuebersetzungen: map[string]string{"jp": "タイトル"},
			},
			sprache: "ja",
			text:    "タイトル",
		},
		{
			name: "ein leerer Name bleibt leer",
			feld: domain.Field{Name: ""},
			text: "",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			if gab := fall.feld.Beschriftet(fall.sprache); gab != fall.text {
				t.Errorf("%q statt %q", gab, fall.text)
			}
		})
	}
}
