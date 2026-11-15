package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Sitzbindung eines Tokens** —
// #1819.

func TestEineSitzbindungIstGanzOderGarNicht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name    string
		bindung domain.Sitzbindung
		feld    string
	}{
		{"keine Bindung", domain.Sitzbindung{}, ""},
		{"Partie und Sitz", domain.Sitzbindung{Partie: "p1", Sitz: "ada"}, ""},
		{"ein Sitz ohne Partie", domain.Sitzbindung{Sitz: "ada"}, "partie"},
		{"eine Partie ohne Sitz", domain.Sitzbindung{Partie: "p1"}, "sitz"},
		{"Leerzeichen am Rand", domain.Sitzbindung{Partie: "p1", Sitz: " ada"}, "sitz"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			fehler := fall.bindung.Validate()

			var feldfehler *domain.FieldError
			switch {
			case fall.feld == "" && fehler != nil:
				t.Errorf("abgewiesen: %v", fehler)
			case fall.feld != "" && (!errors.As(fehler, &feldfehler) || feldfehler.Field != fall.feld):
				t.Errorf("Fehler = %v, erwartet einen am Feld %q", fehler, fall.feld)
			}
		})
	}
}

// **Und ein Token prüft seine Bindung mit** — eine halbe kommt nicht in den
// Bestand, auch wenn der Weg an `Sitzbindung.Validate` vorbeiführt.
func TestEinTokenMitHalberBindungIstUngueltig(t *testing.T) {
	t.Parallel()

	eines := gutesToken()
	eines.Sitzbindung = domain.Sitzbindung{Partie: "p1"}

	var feldfehler *domain.FieldError
	if fehler := eines.Validate(); !errors.As(fehler, &feldfehler) || feldfehler.Field != "sitz" {
		t.Errorf("Fehler = %v, erwartet einen am Feld \"sitz\"", fehler)
	}
}

func TestEineSitzbindungVerengtUndGibtNichtsDazu(t *testing.T) {
	t.Parallel()

	gebunden := domain.Sitzbindung{Partie: "p1", Sitz: "ada"}
	frei := domain.Sitzbindung{}

	for _, fall := range []struct {
		name  string
		pruef error
		nein  bool
	}{
		{"der eigene Sitz", gebunden.Erlaubt("p1", "ada"), false},
		{"ein fremder Sitz", gebunden.Erlaubt("p1", "bo"), true},
		{"eine andere Partie", gebunden.Erlaubt("p2", "ada"), true},
		{"die eigene Partie ohne Sitz", gebunden.ErlaubtPartie("p1"), false},
		{"eine andere Partie ohne Sitz", gebunden.ErlaubtPartie("p2"), true},
		{"der ganze Zustand", gebunden.ErlaubtVollenZustand(), true},
		{"ungebunden: jeder Sitz", frei.Erlaubt("p2", "bo"), false},
		{"ungebunden: jede Partie", frei.ErlaubtPartie("p2"), false},
		{"ungebunden: der ganze Zustand", frei.ErlaubtVollenZustand(), false},
	} {
		if fall.nein != errors.Is(fall.pruef, domain.ErrErlaubnis) {
			t.Errorf("%s: %v", fall.name, fall.pruef)
		}
	}
}
