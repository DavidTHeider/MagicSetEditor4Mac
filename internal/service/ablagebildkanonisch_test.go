package service_test

// **Ein Bildfeld wird auch in MSEs Schreibweise erkannt** —
// #1980.
//
// MSE schreibt Feldnamen kanonisch (`image 2` wird `image_2:`, so wie die
// kanonischen Sets `rule_text:` tragen), und das Einfügen schlug den
// Schlüssel roh unter den Feldnamen nach. Ein Verweis ohne Datei ging bei
// einem solchen Feld still durch. Die eigene Ablage schreibt den Namen, wie
// er erklärt ist — der Fall entsteht an Text, den MSE geschrieben hat.

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweitbilddoppel führt ein Bildfeld, dessen Name ein Leerzeichen trägt.
type zweitbilddoppel struct{}

func (zweitbilddoppel) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return []domain.Field{
		{Name: "name", Gespeichert: true},
		{Name: "image 2", Kind: domain.KindImage, Gespeichert: true},
	}, nil
}

func TestEinVerweisOhneBildInMSEsSchreibweiseWirdBenannt(t *testing.T) {
	t.Parallel()

	vor := neueBildvorrichtung(t)
	vor.ablage = vor.ablage.MitFeldern(zweitbilddoppel{})

	einfuegung, fehler := vor.ablage.Einfuegen(t.Context(), 2,
		"game_name: probe\ncard:\n\timage_2: image0\n")
	if fehler != nil {
		t.Fatalf("Einfuegen: %v", fehler)
	}

	if len(einfuegung.Befunde) == 0 || !strings.Contains(einfuegung.Befunde[0].Meldung, "image0") {
		t.Errorf("der Verweis ging still durch — Befunde %v", einfuegung.Befunde)
	}
}
