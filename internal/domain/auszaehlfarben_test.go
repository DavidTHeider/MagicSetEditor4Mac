package domain_test

// **Eine Farbtafel trifft ihre Gruppe in beiden Schreibweisen** —
// #1980.
//
// MSE liest jeden Schlüssel kanonisch (`src/util/io/reader.cpp`, Zeile 217,
// Fassung f9d9356 vom 2021-07-06), die Tafel unter `colors:` also mit
// Unterstrichen — und schlägt in `GraphData::GraphData` mit dem **rohen**
// Gruppennamen nach (`src/gui/control/graph.cpp`, Zeile 138). Dort findet
// `multi color` seine Farbe nie. Hier wird auf beiden Seiten kanonisch
// verglichen, wie es MSE beim Wahlbild und beim Sprachpaket tut
// (`src/data/field/choice.cpp` Zeile 188, `src/data/locale.cpp` Zeile 82):
// eine Abweichung mit Absicht, Marke `adapted`.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineFarbtafelTrifftBeideSchreibweisen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name, schluessel, gruppe string
	}{
		{"Unterstrich in der Tafel, Leerzeichen in der Gruppe", "multi_color", "multi color"},
		{"Leerzeichen in der Tafel, Unterstrich in der Gruppe", "multi color", "multi_color"},
		{"gleich geschrieben", "multi color", "multi color"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			gruppen := domain.Auszaehlen([]string{fall.gruppe}, domain.Auszaehlart{
				Farben: map[string]string{fall.schluessel: "rgb(1,2,3)"},
			})

			if len(gruppen) != 1 || gruppen[0].Farbe != "rgb(1,2,3)" {
				t.Errorf("Gruppen %+v", gruppen)
			}
		})
	}
}
