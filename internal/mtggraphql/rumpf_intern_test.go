package mtggraphql

// **Der Fehlerweg beim Bauen der Anfrage** — ein interner Lauf, weil er
// [rumpfBauen] braucht.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was JSON nicht kennt, wird gemeldet und nicht halb geschickt.**
//
// Der Fall entsteht nicht aus [Klient.Karten] — dort sind die Variablen
// Zeichenketten. Er entsteht, sobald jemand eine zweite Abfrage baut und
// einen Wert übergibt, der sich nicht schreiben lässt. Dann soll er es
// erfahren, statt eine Anfrage ohne Variablen abzuschicken.
func TestWasJSONNichtKenntWirdGemeldet(t *testing.T) {
	t.Parallel()

	_, err := rumpfBauen("{ cards { name } }", map[string]any{
		"kanal": make(chan int),
	})

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "JSON") {
		t.Errorf("die Meldung %q sagt nicht, woran es lag", err)
	}
}

// **Und ein brauchbarer Rumpf kommt durch** — sonst prüfte der Lauf oben nur,
// dass die Funktion immer scheitert.
func TestEinBrauchbarerRumpfKommtDurch(t *testing.T) {
	t.Parallel()

	rumpf, err := rumpfBauen("{ cards { name } }", map[string]any{"name": "X"})
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	for _, teil := range []string{`"query"`, `"variables"`, `"name":"X"`} {
		if !strings.Contains(string(rumpf), teil) {
			t.Errorf("der Rumpf %q enthält %q nicht", rumpf, teil)
		}
	}
}
