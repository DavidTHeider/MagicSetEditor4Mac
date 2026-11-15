package service

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Was hier gesetzt wurde, liegt über dem Block der Datei** — und beide
// über der Vorgabe des Stils (`stilwirkung.go`).
func TestHierGesetztLiegtUeberDemBlockDerDatei(t *testing.T) {
	stil := msepaket.Stil{Paket: "probe-schlicht.mse-style", Stiloptionen: []domain.StyleField{{
		Name: "hintergrund", Kind: domain.KindChoice, Initial: "hell", Choices: []string{"hell", "dunkel"},
	}}}
	block := []domain.Value{{Field: Setstilfeld, Text: "probe-schlicht/hintergrund: dunkel\n"}}

	for _, fall := range []struct {
		name     string
		angaben  []domain.Value
		gesetzt  []domain.Value
		erwartet string
	}{
		{"ohne beides die Vorgabe", nil, nil, "hell"},
		{"der Block der Datei", block, nil, "dunkel"},
		{"darüber, was hier gesetzt wurde", block, []domain.Value{{Field: "hintergrund", Text: "hell"}}, "hell"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			umgebung := msescript.NeueUmgebung(nil)
			belegeVorlage(t.Context(), umgebung, Vorlage{Stil: stil, Setangaben: fall.angaben, Stilwerte: fall.gesetzt}, nil, nil)

			styling, _ := umgebung.Suche("styling")
			for _, eintrag := range styling.Eintraege {
				if eintrag.Name == "hintergrund" && eintrag.Wert.Text != fall.erwartet {
					t.Errorf("styling.hintergrund ist %q, erwartet %q", eintrag.Wert.Text, fall.erwartet)
				}
			}
		})
	}
}

func TestAlsStilwerteGibtEineFesteReihenfolge(t *testing.T) {
	werte := alsStilwerte(map[string]string{"b": "2", "a": "1"})
	if len(werte) != 2 || werte[0].Field != "a" || werte[1].Field != "b" {
		t.Errorf("%v", werte)
	}
}
