package app

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Spiel eines Sets, ohne eine Karte zu brauchen** —
// #1032.
//
// Das Anlegen aus einem Scryfall-Namen braucht die Zuordnung, und die hängt am
// Spiel. [Kartenstand] verlangt eine Karte und wäre hier eine Vorbedingung,
// die es nicht gibt — die Karte entsteht ja erst.
type setquelle struct {
	koepfe []domain.SetHead
	fehler error
}

func (s setquelle) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return s.koepfe, s.fehler
}

func (setquelle) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return nil, errors.New("wird hier nicht gebraucht")
}

func TestSetspielFindetDasSpiel(t *testing.T) {
	quelle := setquelle{koepfe: []domain.SetHead{{ID: 3, Game: "magic"}}}

	spiel, fehler := setspielUeber(probekontext(t), quelle, 3)
	if fehler != nil {
		t.Fatalf("setspielUeber: %v", fehler)
	}

	if spiel != "magic" {
		t.Errorf("Spiel = %q", spiel)
	}
}

// **Ein Set, das es nicht gibt, ist ein Befund und keine leere Antwort.**
//
// Ein leeres Spiel führte auf eine Zuordnung, die es nicht gibt — und die
// Meldung nennte dann das Spiel statt das Set.
func TestSetspielWeistEinUnbekanntesSetAb(t *testing.T) {
	quelle := setquelle{koepfe: []domain.SetHead{{ID: 3, Game: "magic"}}}

	if _, fehler := setspielUeber(probekontext(t), quelle, 9); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", fehler)
	}
}

// **Ein Bestand, der nicht antwortet, wird durchgereicht.**
func TestSetspielReichtDenLesefehlerDurch(t *testing.T) {
	quelle := setquelle{fehler: errors.New("der Bestand ist nicht erreichbar")}

	if _, fehler := setspielUeber(probekontext(t), quelle, 3); fehler == nil {
		t.Fatal("kein Fehler")
	}
}
