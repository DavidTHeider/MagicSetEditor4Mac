// **Der Schnellfilter am Fenster und an den Flächen** — Werkstück U5 von
// Phase 5.1 (#1416).
package app_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerFilterGibtKennungenAlsZeichenketten(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{gefiltert: []int64{7, 9}}
	anwendung := mitListenbau(quelle)

	antwort := anwendung.KartenFiltern(context.Background(), "3", "type:creature")
	if antwort.Fehler != nil {
		t.Fatalf("filtern: %+v", antwort.Fehler)
	}

	if !slices.Equal(antwort.KartenIDs, []string{"7", "9"}) {
		t.Errorf("KartenIDs = %v, erwartet [7 9]", antwort.KartenIDs)
	}

	if quelle.gefragtesSet != 3 || quelle.gefragteAnfrage != "type:creature" {
		t.Errorf("gefragt wurde Set %d nach %q", quelle.gefragtesSet, quelle.gefragteAnfrage)
	}
}

func TestDerFilterAntwortetOhneTrefferMitLeererListeStattNull(t *testing.T) {
	t.Parallel()

	antwort := mitListenbau(&listenquelle{}).KartenFiltern(context.Background(), "3", "x")
	if antwort.Fehler != nil || antwort.KartenIDs == nil || len(antwort.KartenIDs) != 0 {
		t.Errorf("Antwort %+v, erwartet eine leere Liste ohne Fehler", antwort)
	}
}

func TestDerFilterWeistAbWasKeineSetkennungIst(t *testing.T) {
	t.Parallel()

	antwort := mitListenbau(&listenquelle{}).KartenFiltern(context.Background(), "drei", "x")
	if antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
		t.Errorf("Antwort %+v, erwartet einen Befund am Feld setID", antwort)
	}
}

func TestDerFilterOhneListenbauSagtEs(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KartenFiltern(context.Background(), "3", "x")
	if antwort.Fehler == nil {
		t.Fatal("ohne Listenbau kam eine Antwort ohne Fehler")
	}
}

func TestDerFilterReichtDenFehlerDerQuelleDurch(t *testing.T) {
	t.Parallel()

	antwort := mitListenbau(&listenquelle{fehler: domain.ErrNotFound}).
		KartenFiltern(context.Background(), "3", "x")
	if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
		t.Errorf("Antwort %+v, erwartet nichtGefunden", antwort)
	}
}

func TestDieFlaechennahtFiltertUeberDieselbeQuelle(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdliste{}).MitListenquelle(&listenquelle{gefiltert: []int64{1}})

	kennungen, fehler := naht.Filtern(context.Background(), 3, "x")
	if fehler != nil || !slices.Equal(kennungen, []int64{1}) {
		t.Errorf("Filtern gab %v, %v", kennungen, fehler)
	}

	kaputt := (&app.Fremdliste{}).MitListenquelle(&listenquelle{fehler: domain.ErrBelegt})
	if _, fehler := kaputt.Filtern(context.Background(), 3, "x"); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Fehler ist %v, erwartet ErrBelegt", fehler)
	}
}
