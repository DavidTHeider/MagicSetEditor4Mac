package service_test

// **Ein Schriftname, der sich nicht ausrechnen lässt, wird nicht gesucht.**
//
// Der erste Durchgang von `card_style` läuft vor den Initskripten, wie bei
// MSE (`src/render/card/viewer.cpp`, Zeilen 44–46, Fassung f9d9356 vom
// 2021-07-06) — ein Name wie `{if choose_style() == "A" then "Matrix" else
// "Beleren Bold"}` (`magic-archenemy`) scheitert dort, weil `choose_style`
// erst danach steht. Hier wurde dann der **Rohtext** als Schrift gesucht: das
// findet nie etwas und läuft bei einem frischen Werk einmal über den ganzen
// Paketbaum (F02, gefunden am Profil von E02).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// gefragteSchriften merkt sich jeden Namen, nach dem gefragt wird.
type gefragteSchriften struct{ namen *[]string }

func (q gefragteSchriften) Suche(name string, _ bool) (render.Schrift, error) {
	*q.namen = append(*q.namen, name)

	return gleichbreiteSchrift{breite: 0.5}, nil
}

func TestEinUnausrechenbarerSchriftnameWirdNichtGesucht(t *testing.T) {
	var gefragt []string

	stil := msepaket.Stil{
		Breite: 375, Hoehe: 523,
		Felder: []msepaket.Feldstil{
			{
				Feld: "name", Links: "10", Oben: "10", Hoehe: "23", Breite: "100",
				Schrift: msepaket.Schrift{Name: "{ gibt_es_nicht() }", Groesse: "10"},
			},
			{
				Feld: "type", Links: "10", Oben: "40", Hoehe: "23",
				Rechts:  "{ card_style.name.content_width }",
				Schrift: msepaket.Schrift{Name: "Probe", Groesse: "10"},
			},
		},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "name", Text: "abcd"},
		{Field: "type", Text: "x"},
	}}

	if _, _, err := service.NeuerPlanbauer().MitSchriften(gefragteSchriften{&gefragt}).
		Stilplan(t.Context(), karte, service.Vorlage{Stil: stil}); err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, name := range gefragt {
		if strings.Contains(name, "gibt_es_nicht") {
			t.Fatalf("gesucht wurde der Rohtext %q — gefragt: %q", name, gefragt)
		}
	}
}
