package app

import (
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ohne lesbare Wahl gilt, was ohne Wahl gilt** — #2054: eine Anwendung
// ohne Einstellungen und eine, deren Einstellungen klemmen, fragen beide das
// erste Modell in fester Form, statt den Vorrang ganz zu verweigern.
func TestOhneLesbareWahlGiltDieAutomatik(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name      string
		anwendung *App
	}{
		{name: "ohne Einstellungen", anwendung: &App{}},
		{name: "mit klemmenden Einstellungen", anwendung: &App{einstellungen: klagendeEinstellungen{}}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if wahl := fall.anwendung.vorrangmodellwahl(t.Context()); wahl != "" {
				t.Errorf("gelesen %q, erwartet keine Wahl", wahl)
			}
		})
	}
}

// **Ein Anbieter ohne Stand hat kein Modell** — und keinen Fehler: der
// Vorrang überspringt ihn, wie einen, der keines in fester Form hat.
func TestEinAnbieterOhneStandHatKeinVorrangmodell(t *testing.T) {
	t.Parallel()

	alias, fehler := modellFuer(nil, ollamaanbieter, "gemma3:1b")
	if alias != "" || fehler != nil {
		t.Errorf("%q, %v", alias, fehler)
	}
}

// **Ein Satz ohne Inhalt bleibt leer** — statt beim großen Anfangsbuchstaben
// über eine leere Zeichenkette zu stolpern.
func TestEinLeererSatzBleibtLeer(t *testing.T) {
	t.Parallel()

	if satz := alsSatz(fmt.Errorf("%w: ", domain.ErrNotFound)); satz != "" {
		t.Errorf("%q", satz)
	}

	if satz := alsSatz(fmt.Errorf("%w: das fehlt", domain.ErrNotFound)); satz != "Das fehlt." {
		t.Errorf("%q", satz)
	}
}
