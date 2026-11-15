package domain_test

// **Was ein KI-Modell kann, ist eine Aufzählung** — Werkstück A6 von
// Phase 5.3 (#1445).

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Sieben Fähigkeiten, und sie sind verschieden.**
//
// Das Issue verlangt die Unterscheidung ausdrücklich: „Embeddings liefern
// Vektoren, Reranker Kandidatenbewertungen, Generierung Vorschläge — kein
// generischer »KI-Antwort«-Blob."
func TestEsGibtSiebenKIFaehigkeiten(t *testing.T) {
	t.Parallel()

	alle := domain.KIFaehigkeiten()

	if len(alle) != 7 {
		t.Fatalf("es sind %d Fähigkeiten, erwartet 7: %v", len(alle), alle)
	}

	for _, erwartet := range []domain.KIFaehigkeit{
		domain.KITextgenerierung, domain.KIStrukturierteGenerierung,
		domain.KIVision, domain.KIBildgenerierung, domain.KIBildbearbeitung,
		domain.KIEinbettung, domain.KINeuordnung,
	} {
		if !slices.Contains(alle, erwartet) {
			t.Errorf("%q fehlt in der Aufzählung", erwartet)
		}

		if !domain.IstKIFaehigkeit(erwartet) {
			t.Errorf("%q gilt nicht als Fähigkeit", erwartet)
		}
	}

	if domain.IstKIFaehigkeit("hellsehen") {
		t.Error("eine erfundene Fähigkeit kam durch")
	}
}

// **Eine Fähigkeit ohne Beleg gilt als ungeprüft — und nicht als vorhanden.**
//
// Das Issue: „Unbekannt bleibt unbekannt; Modelllisten garantieren keine
// Capability." Ein Modell, das in einer Liste steht, kann trotzdem nichts.
func TestEineFaehigkeitOhneBelegGiltNichtAlsVorhanden(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		geprueft domain.KIGeprueft
		gueltig  bool
	}{
		{
			"mit Datum und Beleg",
			domain.KIGeprueft{
				Faehigkeit: domain.KIEinbettung,
				Datum:      "2026-09-11",
				Beleg:      "eigener Lauf gegen die Laufzeit",
			},
			true,
		},
		{
			"ohne Datum",
			domain.KIGeprueft{
				Faehigkeit: domain.KIEinbettung, Beleg: "irgendwoher",
			},
			false,
		},
		{
			"ohne Beleg",
			domain.KIGeprueft{
				Faehigkeit: domain.KIEinbettung, Datum: "2026-09-11",
			},
			false,
		},
		{
			"mit einem Datum, das keines ist",
			domain.KIGeprueft{
				Faehigkeit: domain.KIEinbettung, Datum: "neulich", Beleg: "x",
			},
			false,
		},
		{
			"mit einer Fähigkeit, die es nicht gibt",
			domain.KIGeprueft{
				Faehigkeit: "hellsehen", Datum: "2026-09-11", Beleg: "x",
			},
			false,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befunde := fall.geprueft.Validate()

			if (len(befunde) == 0) != fall.gueltig {
				t.Errorf("%+v gab %v", fall.geprueft, befunde)
			}
		})
	}
}
