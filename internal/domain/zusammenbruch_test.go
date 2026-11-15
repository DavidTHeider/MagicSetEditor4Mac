package domain_test

// **Ein Nebenlauf, der zusammenbricht, wird zum Fehler** — BK-05
// (#2111,
// #2112, #2113, #2114).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestAuffangenMachtAusEinerPanikEinenFehler(t *testing.T) {
	t.Parallel()

	var gemeldet error

	func() {
		defer domain.Auffangen(func(fehler error) { gemeldet = fehler })

		panic("zerbrochen")
	}()

	if !errors.Is(gemeldet, domain.ErrZusammengebrochen) {
		t.Fatalf("gemeldet: %v", gemeldet)
	}

	// **Der Grund steht im Fehler** — fürs Protokoll; was eine Fläche
	// daraus zeigt, entscheidet sie (§8, `domain.MitGrund`).
	if !strings.Contains(gemeldet.Error(), "zerbrochen") {
		t.Errorf("ohne Grund: %v", gemeldet)
	}
}

func TestAuffangenSchweigtOhnePanik(t *testing.T) {
	t.Parallel()

	gerufen := false

	func() {
		defer domain.Auffangen(func(error) { gerufen = true })
	}()

	if gerufen {
		t.Error("ohne Panik wurde trotzdem gemeldet")
	}
}

// **Ein Fehler als Panik bleibt erkennbar** — `errors.Is` geht durch, und
// eine fachliche Panik wie `torfehler` verliert ihre Gestalt nicht.
func TestAuffangenBehaeltEinenGeworfenenFehler(t *testing.T) {
	t.Parallel()

	var gemeldet error

	func() {
		defer domain.Auffangen(func(fehler error) { gemeldet = fehler })

		panic(domain.ErrConflict)
	}()

	if !errors.Is(gemeldet, domain.ErrZusammengebrochen) || !errors.Is(gemeldet, domain.ErrConflict) {
		t.Errorf("gemeldet: %v", gemeldet)
	}
}

// **Ein Zusammenbruch beim Melden nimmt den Prozess nicht mit.**
func TestAuffangenUeberstehtEinenZusammenbruchBeimMelden(t *testing.T) {
	t.Parallel()

	gemeldet := false

	func() {
		defer domain.Auffangen(func(error) {
			gemeldet = true

			panic("auch das Aufräumen bricht zusammen")
		})

		panic("zerbrochen")
	}()

	if !gemeldet {
		t.Error("die Meldung lief nicht")
	}
}
