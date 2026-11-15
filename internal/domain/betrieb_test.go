package domain_test

// **Was Bereitschaft heißt** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Die eine Regel, für die es diesen Typ gibt
//
// **Ein fremder Dienst macht diese Anwendung nicht unbereit.** Das ist die
// Abnahme des Issues — „Scryfall/AI-Ausfall macht lokalen Editor nicht
// unbereit" — und zugleich §16, Satz 5: ohne Netz arbeitet die Anwendung
// weiter. Wer beides in **eine** Zahl rechnete, machte aus einem fremden
// Ausfall einen eigenen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein offener, gewanderter Bestand ist bereit.**
func TestEinOffenerBestandIstBereit(t *testing.T) {
	t.Parallel()

	zustand := domain.Betriebszustand{Lebt: true}

	if !zustand.Bereit() || zustand.Grund != "" {
		t.Errorf("%+v", zustand)
	}
}

// **Und jeder Grund nimmt die Bereitschaft** — aber nur ein eigener.
func TestJederEigeneGrundNimmtDieBereitschaft(t *testing.T) {
	t.Parallel()

	for _, grund := range []domain.Bereitschaftsgrund{
		domain.GrundBestandZu,
		domain.GrundWanderungLaeuft,
		domain.GrundBestandUnlesbar,
	} {
		t.Run(string(grund), func(t *testing.T) {
			t.Parallel()

			zustand := domain.Betriebszustand{Lebt: true, Grund: grund}

			if zustand.Bereit() {
				t.Errorf("%q ließ die Bereitschaft stehen", grund)
			}
		})
	}
}

// **Ein Prozess, der nicht lebt, ist nicht bereit** — die Reihenfolge steht
// fest: wer nicht antwortet, antwortet auch nicht mit „bereit".
func TestOhneLebenKeineBereitschaft(t *testing.T) {
	t.Parallel()

	if (domain.Betriebszustand{}).Bereit() {
		t.Error("ein toter Prozess war bereit")
	}
}

// **Und ein ausgefallener fremder Dienst ändert daran nichts.**
//
// Das ist die Abnahme aus dem Issue, und sie ist der Grund für den ganzen Typ.
func TestEinFremderAusfallNimmtDieBereitschaftNicht(t *testing.T) {
	t.Parallel()

	zustand := domain.Betriebszustand{
		Lebt: true,
		Fremde: []domain.Fremdzustand{
			{Dienst: domain.DienstScryfall, Erreichbar: false, Grund: "kein Netz"},
			{Dienst: domain.DienstMTGJSON, Erreichbar: true},
		},
	}

	if !zustand.Bereit() {
		t.Error("ein fremder Ausfall nahm die Bereitschaft")
	}

	// **Und er steht trotzdem da** — er ist eine Auskunft, kein Grund.
	if len(zustand.Fremde) != 2 || zustand.Fremde[0].Erreichbar {
		t.Errorf("%+v", zustand.Fremde)
	}
}
