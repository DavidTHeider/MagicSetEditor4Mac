package service_test

// **Eine Wahrheit für die setweiten Stiloptionen** — F05
// (#2247).
// Die Liste und die Karte lesen dieselbe Reihenfolge: Vorgabe, Block der
// Datei, hier gesetzt (`stilwirkung.go`).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stilwertquelle gibt die hier gesetzten Stiloptionen, die ein Test vorgibt.
type stilwertquelle struct {
	werte  []domain.Value
	fehler error
}

func (s stilwertquelle) Stilwerte(context.Context, int64) ([]domain.Value, error) {
	return s.werte, s.fehler
}

func TestDieListeNenntDenBlockUndDarueberDasGesetzte(t *testing.T) {
	koepfe := &koepfequelle{
		koepfe: []domain.SetHead{{ID: 1, Game: "probe", Style: "probe-schlicht"}},
		werte: []domain.Value{{Field: service.Setstilfeld, Text: "probe-schlicht/hintergrund: dunkel\n" +
			"probe-schlicht/rahmen_zeigen: yes\nprobe-kunst/hintergrund: hell\n"}},
	}
	ablage := &stilablage{werte: []domain.Value{{Field: "rahmen zeigen", Text: "no"}}}

	optionen, err := service.NeuerStilwaehler(koepfe, saatkatalog(), ablage, &feldliste{}, festeSaatwurzel(t)).
		Optionen(t.Context(), 1)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	erwartet := map[string]struct {
		wert    string
		gesetzt bool
	}{"hintergrund": {"dunkel", false}, "rahmen zeigen": {"no", true}}

	for _, option := range optionen {
		if soll, da := erwartet[option.Name]; da && (option.Wert != soll.wert || option.Gesetzt != soll.gesetzt) {
			t.Errorf("%s: %q (gesetzt %v), erwartet %q (gesetzt %v)",
				option.Name, option.Wert, option.Gesetzt, soll.wert, soll.gesetzt)
		}
	}
}

func TestEinFehlerDerSetangabenHaeltDieListeAuf(t *testing.T) {
	kaputt := errors.New("Setangaben weg")
	koepfe := &koepfequelle{
		koepfe:      []domain.SetHead{{ID: 1, Game: "probe", Style: "probe-schlicht"}},
		wertefehler: kaputt,
	}

	if _, err := service.NeuerStilwaehler(koepfe, saatkatalog(), &stilablage{}, &feldliste{}, festeSaatwurzel(t)).
		Optionen(t.Context(), 1); !errors.Is(err, kaputt) {
		t.Errorf("Optionen: %v", err)
	}
}

func TestDerStilspeicherBringtDieGesetztenStilwerteMit(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}}}
	gesetzt := []domain.Value{{Field: "hintergrund", Text: "dunkel"}}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		MitStilwerten(stilwertquelle{werte: gesetzt}).StilFuerSet(t.Context(), 7)
	if !da || len(vorlage.Stilwerte) != 1 || vorlage.Stilwerte[0] != gesetzt[0] {
		t.Errorf("Stilwerte %v, Befunde %v", vorlage.Stilwerte, befunde)
	}

	// **Ein Fehler ist ein Befund und hält die Karte nicht auf** (§8).
	_, befunde, da = service.NeuerStilspeicher(quelle, saatwurzel()).
		MitStilwerten(stilwertquelle{fehler: errors.New("Ablage weg")}).StilFuerSet(t.Context(), 7)
	if !da || len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Stiloptionen lesen") {
		t.Errorf("da %v, Befunde %v", da, befunde)
	}
}
