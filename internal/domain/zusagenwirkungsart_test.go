package domain_test

// **Was ein Zustandswechsel an den festgehaltenen Exemplaren bedeutet** —
// Befund 3 des Nachlaufs zum Zweitblick über Phase 5.6
// (#1477).
//
// # Warum das eine eigene Datei ist
//
// **Weil `handelsstandnahme_test.go` an der Zeilengrenze stand** (§2) — und
// weil die Frage eine andere ist: dort geht es darum, **woran** ein fremder
// Stand anschließt, hier darum, was ein Wechsel **bedeutet**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was ein Wechsel an den Zusagen bedeutet** — Befund 3 des Nachlaufs zum
// Zweitblick über Phase 5.6.
//
// **Ein Wert und keine Handlung**: die Entscheidung reist bis dorthin, wo
// die Transaktion ist — sonst bliebe nach einem gescheiterten Schreiben eine
// Zusage als „übergeben" stehen, während der Vorgang noch davor steht.
func TestWasEinWechselAnDenZusagenBedeutet(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		vorher, nachher domain.Handelszustand
		erwartet        domain.Zusagenwirkungsart
	}{
		"kein Wechsel": {
			vorher: domain.HandelAngeboten, nachher: domain.HandelAngeboten,
			erwartet: domain.ZusagenBleiben,
		},
		"die Übergabe steht offen": {
			vorher:   domain.HandelVorbereitet,
			nachher:  domain.HandelUebergabeOffen,
			erwartet: domain.ZusagenUebergeben,
		},
		"abgelehnt": {
			vorher: domain.HandelAngeboten, nachher: domain.HandelAbgelehnt,
			erwartet: domain.ZusagenFreigeben,
		},
		"abgebrochen": {
			vorher: domain.HandelAngeboten, nachher: domain.HandelAbgebrochen,
			erwartet: domain.ZusagenFreigeben,
		},
		"abgelaufen": {
			vorher: domain.HandelAngeboten, nachher: domain.HandelAbgelaufen,
			erwartet: domain.ZusagenFreigeben,
		},
		"angenommen lässt sie in Ruhe": {
			vorher: domain.HandelAngeboten, nachher: domain.HandelAngenommen,
			erwartet: domain.ZusagenBleiben,
		},
		"lokal verbucht auch": {
			vorher:   domain.HandelUebergabeOffen,
			nachher:  domain.HandelLokalVerbucht,
			erwartet: domain.ZusagenBleiben,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if bedeutet := domain.ZusagenwirkungFuer(
				fall.vorher, fall.nachher); bedeutet != fall.erwartet {
				t.Errorf("es bedeutet %q statt %q", bedeutet, fall.erwartet)
			}
		})
	}
}
