// **Die Bindungen der Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// # Warum die Wrapper einen eigenen Lauf brauchen
//
// Sie sind eine Zeile lang und tun scheinbar nichts — sie holen den Kontext des
// Fensters und reichen durch. Genau darin liegt der Fehler, den sie machen
// können: einen **falschen** Kontext, eine **falsche** Methode. Ein Wrapper
// ohne Lauf ist eine Bindung, die im Fenster zum ersten Mal ausprobiert wird.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Jede der acht Bindungen führt zu ihrer Methode.**
func TestJedeSammlungsbindungFuehrtDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	anwendung.MitUhr(festeUhr())
	anwendung.MitSammlung(&sammlungsnahtdoppel{exemplare: nil})
	anwendung.MitPreisen(preisnahtdoppel{})

	bindungen := app.NeueBindungen(anwendung)

	if antwort := bindungen.Sammlung(t.Context()); antwort.Fehler != nil {
		t.Errorf("Sammlung: %+v", antwort.Fehler)
	}

	if antwort := bindungen.ExemplarAnlegen(t.Context(), app.ExemplarDTO{
		Printing: "abc", Menge: 1,
	}); antwort.Fehler != nil || antwort.ID != 7 {
		t.Errorf("ExemplarAnlegen: %+v", antwort)
	}

	if fehler := bindungen.ExemplarAendern(t.Context(), app.ExemplarDTO{
		ID: 1, Printing: "abc", Menge: 1,
	}); fehler != nil {
		t.Errorf("ExemplarAendern: %+v", fehler)
	}

	if fehler := bindungen.ExemplarLoeschen(t.Context(), 1, 1); fehler != nil {
		t.Errorf("ExemplarLoeschen: %+v", fehler)
	}

	for name, antwort := range map[string]app.Preisantwort{
		"PreiseHolen":  bindungen.PreiseHolen(t.Context(), "abc"),
		"Preise":       bindungen.Preise(t.Context(), "abc"),
		"Preisverlauf": bindungen.Preisverlauf(t.Context(), probereihe()),
	} {
		if antwort.Fehler != nil {
			t.Errorf("%s: %+v", name, antwort.Fehler)
		}
	}

	if antwort := bindungen.Portfolio(t.Context()); antwort.Fehler != nil {
		t.Errorf("Portfolio: %+v", antwort.Fehler)
	}
}
