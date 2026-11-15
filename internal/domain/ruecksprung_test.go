package domain_test

// **Die vier Rückkehrarten und was sie kosten** — Werkstück O7 von Phase 5.2
// (#1431).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Vier Arten, und sie sind nicht dieselbe Handlung.** Heute gibt es einen
// Knopf; das Issue verlangt die Trennung, weil die Kosten und die Rückwege
// verschieden sind.
func TestEsGibtVierRuecksprungarten(t *testing.T) {
	t.Parallel()

	arten := domain.Ruecksprungarten()
	if len(arten) != 4 {
		t.Fatalf("es sind %d Arten: %v", len(arten), arten)
	}

	gesehen := map[domain.Ruecksprungart]bool{}
	for _, art := range arten {
		if gesehen[art] {
			t.Errorf("%q steht zweimal in der Reihe", art)
		}

		gesehen[art] = true
	}

	for _, erwartet := range []domain.Ruecksprungart{
		domain.RuecksprungAppfassung, domain.RuecksprungSchema,
		domain.RuecksprungDaten, domain.RuecksprungPakete,
	} {
		if !gesehen[erwartet] {
			t.Errorf("%q fehlt in der Reihe", erwartet)
		}
	}
}

// **Zwei der vier kann diese Anwendung nicht selbst** — und das ist kein
// Mangel, sondern eine Aussage, die dasteht. Ein Knopf, der sicher nein sagt,
// wäre schlimmer als ein Satz, der erklärt, wer es tut.
func TestZweiArtenGehenHierNichtUndSagenWarum(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		art       domain.Ruecksprungart
		hindernis string
	}{
		{domain.RuecksprungAppfassung, domain.RuecksprungNachProzessende},
		{domain.RuecksprungSchema, domain.RuecksprungNurVorwaerts},
		{domain.RuecksprungDaten, ""},
		{domain.RuecksprungPakete, ""},
	} {
		t.Run(string(fall.art), func(t *testing.T) {
			t.Parallel()

			if hindernis := domain.RuecksprungHindernis(fall.art); hindernis != fall.hindernis {
				t.Errorf("%q gibt %q, erwartet war %q", fall.art, hindernis, fall.hindernis)
			}
		})
	}
}

// **Eine unbekannte Art ist keine.** Sie darf nicht als „geht hier" durchgehen
// — das wäre ein leerer Hinweis mit der Bedeutung „ja".
func TestEineUnbekannteArtGiltNichtAlsMachbar(t *testing.T) {
	t.Parallel()

	if hindernis := domain.RuecksprungHindernis("zeitreise"); hindernis == "" {
		t.Error("eine erfundene Art kam als machbar durch")
	}
}
