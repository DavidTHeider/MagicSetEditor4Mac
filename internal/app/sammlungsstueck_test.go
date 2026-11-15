package app_test

// **Die Stückzahl der Sammlung, ohne die Liste zu übertragen** — U15
// (#2160).
//
// Die Startseite holte die **ganze** Sammlung, nur um die Mengen zu
// summieren. Das Fenster nennt die Summe jetzt neben `Insgesamt` — wie diese
// **ohne** Sicht, denn sie zählt die Sammlung und nicht den Filter.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDasFensterNenntDieStueckzahlOhneSicht(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		exemplare   []*domain.Exemplar
		ausfuehrung string
		stueck      int
	}{
		"leer":                      {exemplare: []*domain.Exemplar{}, stueck: 0},
		"zehn mit Mengen über eins": {exemplare: zehnExemplare(), stueck: 55},
		"gefiltert zählt dennoch alle": {
			exemplare: zehnExemplare(), ausfuehrung: "foil", stueck: 55,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			antwort := fensterapp(&sammlungsnahtdoppel{exemplare: fall.exemplare}).
				Sammlungsfenster(t.Context(), "", fall.ausfuehrung, "", 0, 1)
			if antwort.Fehler != nil {
				t.Fatalf("%+v", antwort.Fehler)
			}

			if antwort.Stueck != fall.stueck || len(antwort.Exemplare) > 1 {
				t.Errorf("Stück %d bei %d Zeilen, erwartet %d bei höchstens einer",
					antwort.Stueck, len(antwort.Exemplare), fall.stueck)
			}
		})
	}
}
