package service_test

// **Zwei Schreiber, ein Handel, im Dienst** — TRADE-03.AK-01
// (#1717).
//
// **Zwischen Lesen und Schreiben stimmt Bert zu**, und Annas Schritt vom
// alten Stand wird abgewiesen — seine Stimme bleibt.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinSchrittVomAltenStandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, ablage := eineHandelsleitung()

	angebot := einHandelsangebot()
	if _, fehler := leitung.Eroeffnen(t.Context(), angebot); fehler != nil {
		t.Fatalf("eröffnen: %v", fehler)
	}

	if _, fehler := leitung.Schreiten(t.Context(), angebot.Vorgang,
		domain.SchrittAnbieten, angebot.Initiator); fehler != nil {
		t.Fatalf("anbieten: %v", fehler)
	}

	andere := angebot.Parteien[1]

	ablage.dazwischen = func(staende map[string]domain.Handelsangebot) {
		stand := staende[angebot.Vorgang]
		stand.Zustimmungen = []string{andere}
		staende[angebot.Vorgang] = stand
	}

	_, fehler := leitung.Schreiten(t.Context(), angebot.Vorgang,
		domain.SchrittAbbrechen, angebot.Initiator)
	if !errors.Is(fehler, domain.ErrUeberholt) {
		t.Fatalf("Schritt vom alten Stand: %v", fehler)
	}

	if stand := ablage.staende[angebot.Vorgang]; stand.Zustand != domain.HandelAngeboten ||
		len(stand.Zustimmungen) != 1 {
		t.Fatalf("die Stimme ist fort: %+v", stand)
	}
}

func TestDasProtokollKommtAusDerAblage(t *testing.T) {
	t.Parallel()

	leitung, _ := eineHandelsleitung()

	protokoll, fehler := leitung.Protokoll(t.Context(), "h1")
	if fehler != nil || len(protokoll) != 1 || protokoll[0].Nach != domain.HandelEntwurf {
		t.Fatalf("Protokoll %+v, %v", protokoll, fehler)
	}
}
