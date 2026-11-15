package service_test

// **Die deterministischen Filter** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Der Satz, gegen den hier geprüft wird
//
// **„Unbekannter Wert ist kein erfundener Treffer."** Eine Regelstelle hat
// keine Farbe. Wer nach roten Karten filtert, bekommt sie nicht — auch nicht,
// weil nichts dagegen spricht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einManawert gibt einen Zeiger auf ihn.
func einManawert(wert float64) *float64 { return &wert }

// diePruefkarten sind drei Dokumente mit verschiedenen Merkmalen.
func diePruefkarten() []domain.Retrievaltreffer {
	blitz := einRetrievaltreffer("blitz", 1, domain.QuelleOracle)
	blitz.Dokument.Merkmale = domain.Retrievalmerkmale{
		Farben: []string{"R"}, Farbidentitaet: []string{"R"},
		Typen: []string{"instant"}, Manawert: einManawert(1), Set: "lea",
		Formate:     map[string]string{"modern": "legal", "standard": "not_legal"},
		Formatstand: "2026-09-12",
	}

	wald := einRetrievaltreffer("wald", 2, domain.QuelleOracle)
	wald.Dokument.Merkmale = domain.Retrievalmerkmale{
		Farbidentitaet: []string{"G"},
		Typen:          []string{"land", "basic"}, Manawert: einManawert(0),
		Set: "neo", Formate: map[string]string{"modern": "legal"},
		Formatstand: "2026-09-12",
	}

	regel := einRetrievaltreffer("regel", 3, domain.QuelleRegelwerk)
	regel.Dokument.Regelnummer = "205.4a"
	regel.Dokument.Quelldatum = "2026-07-26"

	return []domain.Retrievaltreffer{blitz, wald, regel}
}

// **Jeder Filter lässt genau durch, was ihn erfüllt.**
func TestJederStrukturierteFilterLaesstGenauDasSeineDurch(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		filter   service.Retrievalfilter
		erwartet []string
	}{
		// **Farbe ist nicht Farbidentität.** Der Wald ist farblos und
		// grün-identisch; ein Farbfilter findet ihn nicht, ein
		// Identitätsfilter schon.
		"nach Farbe rot":           {service.Retrievalfilter{Farben: []string{"R"}}, []string{"blitz"}},
		"nach Farbe grün":          {service.Retrievalfilter{Farben: []string{"G"}}, nil},
		"nach Farbidentität grün":  {service.Retrievalfilter{Farbidentitaet: []string{"G"}}, []string{"wald"}},
		"nach Typ":                 {service.Retrievalfilter{Typen: []string{"land"}}, []string{"wald"}},
		"nach zwei Typen zugleich": {service.Retrievalfilter{Typen: []string{"land", "basic"}}, []string{"wald"}},
		"nach zwei Typen, die keine Karte zugleich hat": {service.Retrievalfilter{Typen: []string{"land", "instant"}}, nil},
		"nach Set":            {service.Retrievalfilter{Sets: []string{"neo"}}, []string{"wald"}},
		"nach Manawert von 1": {service.Retrievalfilter{ManawertVon: einManawert(1)}, []string{"blitz"}},
		"nach Manawert bis 0": {service.Retrievalfilter{ManawertBis: einManawert(0)}, []string{"wald"}},
		"nach Quelle":         {service.Retrievalfilter{Quellen: []domain.Retrievalquelle{domain.QuelleRegelwerk}}, []string{"regel"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pruefeFilter(t, fall.filter, fall.erwartet)
		})
	}
}

// **Eine Regelstelle hat keine Farbe — und wird deshalb nicht gefunden.**
//
// **Das ist die Zeile, für die es diesen Lauf gibt.** Ein Filter, der
// „unbekannt" als „passt schon" läse, gäbe auf jede Farbfrage das ganze
// Regelwerk mit.
func TestEinUnbekanntesMerkmalIstKeinTreffer(t *testing.T) {
	t.Parallel()

	for name, filter := range map[string]service.Retrievalfilter{
		"Farbe":         {Farben: []string{"R"}},
		"Farbidentität": {Farbidentitaet: []string{"R"}},
		"Typ":           {Typen: []string{"instant"}},
		"Set":           {Sets: []string{"lea"}},
		"Manawert":      {ManawertVon: einManawert(0)},
		"Format":        {Format: "modern"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, treffer := range gefiltert(t, filter) {
				if treffer.Dokument.Kennung == "regel" {
					t.Fatalf("die Regelstelle kam durch den Filter nach %s", name)
				}
			}
		})
	}
}

// **Ein Format filtert nur mit seinem Datenstand.**
func TestEinFormatFiltertNurMitSeinemDatenstand(t *testing.T) {
	t.Parallel()

	pruefeFilter(t, service.Retrievalfilter{
		Format: "modern", Formatlage: "legal",
	}, []string{"blitz", "wald"})

	// **Der Blitz ist in Standard nicht legal** — und das ist eine Auskunft
	// und kein fehlender Eintrag.
	pruefeFilter(t, service.Retrievalfilter{
		Format: "standard", Formatlage: "legal",
	}, nil)

	// **Ein anderer Stand ist kein Treffer.** Wer nach dem Bannstand von
	// morgen fragt, bekommt nicht den von heute.
	pruefeFilter(t, service.Retrievalfilter{
		Format: "modern", Formatlage: "legal", Formatstand: "2026-10-01",
	}, nil)

	pruefeFilter(t, service.Retrievalfilter{
		Format: "modern", Formatlage: "legal", Formatstand: "2026-09-12",
	}, []string{"blitz", "wald"})
}

// **Ein Filter, der nichts sagt, wirft nichts weg.**
func TestEinLeererFilterWirftNichtsWeg(t *testing.T) {
	t.Parallel()

	pruefeFilter(t, service.Retrievalfilter{},
		[]string{"blitz", "wald", "regel"})
}

// gefiltert fährt eine Suche mit diesem Filter.
func gefiltert(t *testing.T, filter service.Retrievalfilter) []domain.Retrievaltreffer {
	t.Helper()

	ergebnis, fehler := einRetrievaldienst(&retrievalablagedoppel{
		manifest: einRetrievalmanifest(), lexikalisch: diePruefkarten(),
	}, nil).Suchen(t.Context(), service.Suchwunsch{
		Frage: "text", Grenze: 10, Filter: filter,
	})
	if fehler != nil {
		t.Fatalf("Suchen: %v", fehler)
	}

	return ergebnis.Treffer
}

// pruefeFilter hält die Treffer gegen die erwarteten Kennungen.
func pruefeFilter(t *testing.T, filter service.Retrievalfilter, erwartet []string) {
	t.Helper()

	treffer := gefiltert(t, filter)

	if len(treffer) != len(erwartet) {
		namen := make([]string, 0, len(treffer))
		for _, einzelner := range treffer {
			namen = append(namen, einzelner.Dokument.Kennung)
		}

		t.Fatalf("durch kamen %v, erwartet %v", namen, erwartet)
	}

	for stelle, kennung := range erwartet {
		if treffer[stelle].Dokument.Kennung != kennung {
			t.Fatalf("der %d. Treffer ist %q, erwartet %q", stelle+1,
				treffer[stelle].Dokument.Kennung, kennung)
		}
	}
}
