package main

// **Mit 0 endet nur, wer den laufenden Bestand nicht kennt** — Nachlauf über
// Phase 2 von 0.9.0, Befund 2 (#2323): `endcode` suchte die neuere Fassung
// in der ganzen Kette, und auch eine **Vorsicherung** aus einer neueren
// Fassung trägt sie. Dann endete ein echter Fehler der Vorbereitung mit 0, und
// launchd startete den Dienst nicht wieder.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestNurEinNeuererBestandEndetMitNull(t *testing.T) {
	t.Parallel()

	neuer := &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}, Stand: "0098_sicherungsvermerke.sql"}

	for name, fall := range map[string]struct {
		fehler error
		code   int
	}{
		"der laufende Bestand ist neuer": {fmt.Errorf("der Bestand der Migrationen: %w", &domain.BestandNeuer{Grund: neuer}), 0},
		"eine Vorsicherung ist neuer": {
			fmt.Errorf("bestand-vor-0.9.0.db taugt nicht als Sicherung: %w",
				fmt.Errorf("%w: %q: %w", domain.ErrSicherung, "bestand-vor-0.9.0.db", neuer)), 1,
		},
		"ein anderer Fehler": {errors.New("kaputt"), 1},
	} {
		if code := endcode(fall.fehler); code != fall.code {
			t.Errorf("%s: %d, erwartet %d", name, code, fall.code)
		}
	}
}
