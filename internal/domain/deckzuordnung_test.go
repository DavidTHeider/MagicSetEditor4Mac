package domain_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Bedarf ist nicht Zuordnung** — Werkstück U9 von Phase 5.1
// (#1420):
// eine Deckliste reserviert nichts; erst eine ausdrückliche Zuordnung nimmt
// Exemplare aus dem Freien, und was frei ist, ist Besitz minus Reservierungen
// minus Zuordnungen — nie unter null.

func TestDieVerfuegbarkeitRechnetBesitzMinusReserviertMinusZugeordnet(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		verfuegbar  domain.Verfuegbarkeit
		frei        int
		erlaubtDrei bool
	}{
		{"nichts belegt", domain.Verfuegbarkeit{Besitz: 3}, 3, true},
		{"zwei reserviert", domain.Verfuegbarkeit{Besitz: 3, Reserviert: 2}, 1, false},
		{"eins zugeordnet, eins reserviert", domain.Verfuegbarkeit{Besitz: 3, Reserviert: 1, Zugeordnet: 1}, 1, false},
		{"mehr belegt als da — nie negativ", domain.Verfuegbarkeit{Besitz: 1, Reserviert: 2, Zugeordnet: 2}, 0, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if got := fall.verfuegbar.Frei(); got != fall.frei {
				t.Errorf("Frei = %d, erwartet %d", got, fall.frei)
			}

			if got := fall.verfuegbar.Erlaubt(3); got != fall.erlaubtDrei {
				t.Errorf("Erlaubt(3) = %v, erwartet %v", got, fall.erlaubtDrei)
			}
		})
	}
}

func TestEineZuordnungNenntPlatzExemplarUndMenge(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		zuordnung domain.Deckzuordnung
		feld      string
	}{
		{"vollständig", domain.Deckzuordnung{DeckID: 1, PlatzID: 2, ExemplarID: 3, Menge: 1}, ""},
		{"ohne Platz", domain.Deckzuordnung{DeckID: 1, ExemplarID: 3, Menge: 1}, "platzID"},
		{"ohne Exemplar", domain.Deckzuordnung{DeckID: 1, PlatzID: 2, Menge: 1}, "exemplarID"},
		{"ohne Menge", domain.Deckzuordnung{DeckID: 1, PlatzID: 2, ExemplarID: 3}, "menge"},
		{"ohne Deck", domain.Deckzuordnung{PlatzID: 2, ExemplarID: 3, Menge: 1}, "deckID"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			pruefeBefund(t, fall.zuordnung.Validate(), fall.feld)
		})
	}
}

// **Ein Wunsch ist nicht zuordenbar, und ein anderer Druck ist eine
// ausdrückliche Abweichung** — keine Namensgleichheit als Identität.
func TestZuordenbarUndAbweichend(t *testing.T) {
	t.Parallel()

	besitz := domain.Exemplar{ID: 1, Printing: "abc", DruckID: 1, Menge: 3}
	wunsch := domain.Exemplar{ID: 2, Printing: "abc", DruckID: 1, Menge: 3, Wunsch: true}

	if !besitz.Zuordenbar() || wunsch.Zuordenbar() {
		t.Errorf("Zuordenbar: Besitz %v, Wunsch %v", besitz.Zuordenbar(), wunsch.Zuordenbar())
	}

	platz := domain.Deckplatz{ID: 2, Printing: "abc", DruckID: 1, Menge: 4}
	anderer := domain.Exemplar{ID: 3, Printing: "def", DruckID: 2, Menge: 1}

	if domain.Abweichend(platz, besitz) {
		t.Error("derselbe Druck ist keine Abweichung")
	}

	if !domain.Abweichend(platz, anderer) {
		t.Error("ein anderer Druck ist eine Abweichung — auch bei gleichem Namen")
	}

	// **Eine eigene Karte ohne Verweis hat keinen Druck** — jede Zuordnung
	// wäre eine Abweichung von nichts: nicht zuordenbar.
	eigene := domain.Deckplatz{ID: 3, KarteID: 9, Menge: 1}
	if eigene.Zuordenbar() || !platz.Zuordenbar() {
		t.Errorf("Zuordenbar: eigene %v, Druck %v", eigene.Zuordenbar(), platz.Zuordenbar())
	}
}
