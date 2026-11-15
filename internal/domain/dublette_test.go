package domain_test

// **Derselbe Druck zweimal ist eine Entscheidung über die Menge** — DOM-03,
// S04.NAT-01 bis 03 (#1698).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieGleicheZeileIstDieDerEindeutigkeit(t *testing.T) {
	t.Parallel()

	grund := domain.Exemplar{
		Printing: "druck-1", Zustand: domain.ZustandNearMint, Sprache: "en", Finish: domain.FinishNonfoil,
	}

	for name, fall := range map[string]struct {
		anders  func(*domain.Exemplar)
		gleiche bool
	}{
		"dieselbe Angabe":                     {func(*domain.Exemplar) {}, true},
		"eine andere Menge bleibt dieselbe":   {func(e *domain.Exemplar) { e.Menge = 4 }, true},
		"eine andere Notiz bleibt dieselbe":   {func(e *domain.Exemplar) { e.Notiz = "Ordner" }, true},
		"ohne Ausführung ist es nonfoil":      {func(e *domain.Exemplar) { e.Finish = "" }, true},
		"eine andere Sprache ist eine eigene": {func(e *domain.Exemplar) { e.Sprache = "de" }, false},
		"eine andere Ausführung ist eine eigene": {
			func(e *domain.Exemplar) { e.Finish = domain.FinishFoil }, false,
		},
		"ein anderer Zustand ist ein eigener": {
			func(e *domain.Exemplar) { e.Zustand = domain.ZustandPlayed }, false,
		},
		"ein Wunsch ist kein Besitz":        {func(e *domain.Exemplar) { e.Wunsch = true }, false},
		"ein anderer Druck ist ein anderer": {func(e *domain.Exemplar) { e.Printing = "druck-2" }, false},
		"ein vorläufiger Druck ist ein anderer": {
			func(e *domain.Exemplar) { e.Printing, e.DruckID = "", 7 }, false,
		},
	} {
		anderes := grund
		fall.anders(&anderes)

		if gleich := grund.GleicheZeile(anderes); gleich != fall.gleiche {
			t.Errorf("%s: %v", name, gleich)
		}
	}

	vorlaeufig := domain.Exemplar{DruckID: 7, Zustand: domain.ZustandNearMint, Sprache: "en"}
	if !vorlaeufig.GleicheZeile(vorlaeufig) {
		t.Error("ein vorläufiger Druck ist nicht er selbst")
	}
}

func TestEineDubletteTraegtIhreZeile(t *testing.T) {
	t.Parallel()

	treiber := errors.New("UNIQUE constraint failed: sammlung.profil_id, …")
	vorhanden := domain.Exemplar{ID: 17, Fassung: 3, Printing: "druck-1", Menge: 2}

	dublette := domain.NeueDublette(vorhanden, treiber)

	if !errors.Is(dublette, domain.ErrDoppelt) || !errors.Is(dublette, domain.ErrConflict) {
		t.Errorf("die Art ging verloren: %v", dublette)
	}

	var befund *domain.FieldError
	if !errors.As(dublette, &befund) || befund.Key != domain.SchluesselDoppeltesExemplar {
		t.Fatalf("kein Befund mit Schlüssel: %v", dublette)
	}

	// **Die Flächen ohne Fenster nennen die Zeile im Satz** — dort gibt es
	// keinen Knopf, der zu ihr führt.
	if !strings.Contains(dublette.Error(), "Exemplar 17") || !strings.Contains(dublette.Error(), "2 Stück") {
		t.Errorf("der Satz nennt die Zeile nicht: %q", dublette.Error())
	}

	if domain.GrundVon(dublette) == nil {
		t.Error("der Treibertext ging verloren")
	}

	var gefunden *domain.Dublette
	if !errors.As(dublette, &gefunden) || gefunden.Vorhanden.ID != 17 || gefunden.Vorhanden.Fassung != 3 {
		t.Errorf("die Zeile ging verloren: %+v", gefunden)
	}
}
