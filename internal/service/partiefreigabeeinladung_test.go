package service_test

// **Ein eingeladener Sitz geht an den Eingeladenen** — P9b
// (#2377).
// Wer zuerst kommt, bekommt den nächsten freien Sitz — aber nie einen, auf den
// jemand anderes eingeladen ist. Sonst säße ein Fremder dort, und der Gast,
// der sein Deck mitbringt, käme nicht mehr an den Tisch.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinEingeladenerSitzGehtAnDenEingeladenen(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"bo", "cora"}).
		MitEinladungen(map[string]string{"bo": "peer-bo"})

	// **Wer zuerst kommt und nicht eingeladen ist**, bekommt nicht Bos Sitz.
	if sitz, fehler := freigabe.Sitz("peer-x"); fehler != nil || sitz != "cora" {
		t.Fatalf("peer-x bekam %q, %v", sitz, fehler)
	}

	if _, fehler := freigabe.Sitz("peer-y"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("peer-y: es kam %v", fehler)
	}

	for range 2 {
		if sitz, fehler := freigabe.Sitz("peer-bo"); fehler != nil || sitz != "bo" {
			t.Errorf("Bo bekam %q, %v", sitz, fehler)
		}
	}
}

// **Eine Einladung auf einen Sitz, der nicht frei ist, gibt keinen.**
func TestEineEinladungAufEinenNichtFreienSitzGibtKeinen(t *testing.T) {
	t.Parallel()

	freigabe := service.NeuePartiefreigabe("p1", []string{"cora"}).
		MitEinladungen(map[string]string{"wirt": "peer-bo"})

	if sitz, fehler := freigabe.Sitz("peer-bo"); fehler != nil || sitz != "cora" {
		t.Errorf("Bo bekam %q, %v", sitz, fehler)
	}
}

// **Die Einladungen kommen aus der Sicht des Wirts** — nur dort steht, wen er
// eingeladen hat.
func TestDieEinladungenStehenInDerSichtDesWirts(t *testing.T) {
	t.Parallel()

	einladungen := service.Einladungen(domain.Partiesicht{Sitze: []domain.Sichtsitz{
		{Kennung: "anna"}, {Kennung: "bo", Eingeladen: "peer-bo", Wartet: true},
	}})

	if len(einladungen) != 1 || einladungen["bo"] != "peer-bo" {
		t.Errorf("die Einladungen: %v", einladungen)
	}
}
