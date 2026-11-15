package fenster

// **Die Naht für eingespeiste Paniken** — F02
// (#2341).
// Die Abnahme verlangt einen Lauf, in dem Start, Finder, Ablegen, Menü und
// Schluss je eine Panik bekommen; `MSE_ABNAHME_PANIK` nennt die Rückrufe mit
// ihrem Namen aus [riegeln], getrennt durch `|`. Gelesen wird sie nur im
// Abnahmebau (`probepanik_abnahme_darwin.go`).

import "testing"

func TestDieNahtNenntGenauDieGenanntenRueckrufe(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		liste, was string
		genannt    bool
	}{
		"einer von zweien":        {"der Start|der Schluss", "der Schluss", true},
		"mit Leerraum":            {" das Ablegen | der Start ", "das Ablegen", true},
		"ein anderer":             {"der Start", "das Ablegen", false},
		"nur ein Teil des Namens": {"Start", "der Start", false},
		"leer":                    {"", "der Start", false},
		"ein leerer Eintrag":      {"der Start||", "", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if genannt := nennt(fall.liste, fall.was); genannt != fall.genannt {
				t.Errorf("nennt(%q, %q) = %v", fall.liste, fall.was, genannt)
			}
		})
	}
}

// **Ohne Schalter stürzt nichts ab** — die Absturznaht von F04 (#2343)
// startet nur mit `MSE_ABNAHME_ABSTURZ` eine Goroutine, und nur im
// Abnahmebau. Lebt der Lauf nach dem Ruf weiter, hat sie nichts getan.
func TestDieAbsturznahtTutOhneSchalterNichts(t *testing.T) {
	t.Setenv("MSE_ABNAHME_ABSTURZ", "")
	t.Setenv("MSE_ABNAHME_SYNCPANIK", "")

	probeabsturz()
	probeSyncpanik()
}
