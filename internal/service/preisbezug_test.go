package service_test

// **Preise aus einem Bulkbezug** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum ein eigener Anwendungsfall neben `Preispflege.Holen`
//
// `Holen` fragt **einen** Druck bei einer Quelle, die je Druck antwortet.
// MTGJSON antwortet nicht je Druck: es schickt **eine** Datei mit allem, und
// was daraus gebraucht wird, sagt der Bestand. Das ist ein anderer Ablauf,
// eine andere Grenze und eine andere Kostenklasse — und ein `Holen`, das
// beides könnte, wäre eine Naht mit zwei Bedeutungen.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bulkdoppel antwortet mit festen Ständen.
type bulkdoppel struct {
	staende  []domain.Preisstand
	befunde  []domain.Befund
	fehler   error
	gefragt  []string
	umfaenge []string
}

func (b *bulkdoppel) Holen(_ context.Context, umfang string, crosswalk map[string]string) (
	[]domain.Preisstand, []domain.Befund, error,
) {
	b.umfaenge = append(b.umfaenge, umfang)

	for kennung := range crosswalk {
		b.gefragt = append(b.gefragt, kennung)
	}

	return b.staende, b.befunde, b.fehler
}

func (b *bulkdoppel) Name() domain.Preisquelle { return "mtgjson" }

// krosswalkdoppel gibt eine feste Karte.
type krosswalkdoppel struct {
	karte  map[string]string
	fehler error
}

func (k krosswalkdoppel) Krosswalk(context.Context) (map[string]string, error) {
	return k.karte, k.fehler
}

// bulkvorrichtung baut den Bezug samt seiner Doppel.
func bulkvorrichtung(staende []domain.Preisstand) (
	*service.Preisbezug, *preisdoppel, *bulkdoppel,
) {
	ablage := &preisdoppel{}
	quelle := &bulkdoppel{staende: staende}
	jetzt := time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC)

	return service.NeuerPreisbezug(ablage, quelle,
		krosswalkdoppel{karte: map[string]string{"mtg-1": "scry-1"}},
		func() time.Time { return jetzt }), ablage, quelle
}

// **Was der Bezug bringt, steht danach im Bestand — mit dem Abrufzeitpunkt.**
func TestEinBulkbezugLegtSeineStaendeAb(t *testing.T) {
	t.Parallel()

	gebracht := domain.Preisstand{
		Printing: "scry-1", Quelle: "mtgjson:cardkingdom",
		Markt: domain.MarktPapier, Art: domain.PreisartVerkauf,
		Waehrung: "USD", Finish: domain.FinishNonfoil,
		Lage: domain.PreisGenannt, Wert: 417,
		Beobachtet: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
	}

	bezug, ablage, quelle := bulkvorrichtung([]domain.Preisstand{gebracht})

	bilanz, fehler := bezug.Holen(t.Context(), "heute")
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if bilanz.Staende != 1 || bilanz.Drucke != 1 {
		t.Errorf("die Bilanz ist %+v", bilanz)
	}

	if len(quelle.gefragt) != 1 || quelle.gefragt[0] != "mtg-1" {
		t.Errorf("gefragt wurde nach %v", quelle.gefragt)
	}

	if len(ablage.staende) != 1 {
		t.Fatalf("%d Stände im Bestand", len(ablage.staende))
	}

	abgelegt := ablage.staende[0]

	if abgelegt.Erhoben.IsZero() || abgelegt.Erhoben.Equal(abgelegt.Beobachtet) {
		t.Errorf("der Abrufzeitpunkt fehlt oder ist der Beobachtungstag: %+v", abgelegt)
	}

	// **Der Beobachtungstag bleibt, wie die Quelle ihn nennt.**
	if !abgelegt.Beobachtet.Equal(gebracht.Beobachtet) {
		t.Errorf("der Beobachtungstag ist %v", abgelegt.Beobachtet)
	}
}

// **Ein leerer Bestand fragt gar nicht erst.**
//
// **Die Datei wiegt ein Gigabyte** (§16, Satz 4): wer keine Karte besitzt, die
// MTGJSON kennt, hat nichts davon — und ein Abruf wäre Last für den Anbieter
// und Wartezeit für niemanden.
func TestOhneKrosswalkWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}
	quelle := &bulkdoppel{}

	bezug := service.NeuerPreisbezug(ablage, quelle,
		krosswalkdoppel{karte: map[string]string{}},
		time.Now)

	bilanz, fehler := bezug.Holen(t.Context(), "heute")
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if len(quelle.umfaenge) != 0 {
		t.Error("es wurde trotzdem gefragt")
	}

	if bilanz.Staende != 0 || bilanz.Drucke != 0 {
		t.Errorf("die Bilanz ist %+v", bilanz)
	}
}

// **Die Befunde der Quelle reisen mit** (§8).
func TestDieBefundeDesBulkbezugsKommenAn(t *testing.T) {
	t.Parallel()

	ablage := &preisdoppel{}
	quelle := &bulkdoppel{befunde: []domain.Befund{{
		Datei: "mtg-1", Ort: "mond", Art: domain.ArtUmgebung,
		Meldung: "einen Markt kennt diese Anwendung nicht",
	}}}

	bezug := service.NeuerPreisbezug(ablage, quelle,
		krosswalkdoppel{karte: map[string]string{"mtg-1": "scry-1"}},
		time.Now)

	bilanz, fehler := bezug.Holen(t.Context(), "heute")
	if fehler != nil {
		t.Fatalf("Holen: %v", fehler)
	}

	if len(bilanz.Befunde) != 1 {
		t.Errorf("die Bilanz nennt %d Befunde", len(bilanz.Befunde))
	}
}

// **Jeder Fehlerweg kommt heraus — und schreibt nichts.**
func TestJederFehlerwegDesBulkbezugsKommtHeraus(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("planmäßig gescheitert")

	faelle := map[string]func() *service.Preisbezug{
		"der Krosswalk klemmt": func() *service.Preisbezug {
			return service.NeuerPreisbezug(&preisdoppel{}, &bulkdoppel{},
				krosswalkdoppel{fehler: gestolpert}, time.Now)
		},
		"die Quelle klemmt": func() *service.Preisbezug {
			return service.NeuerPreisbezug(&preisdoppel{},
				&bulkdoppel{fehler: gestolpert},
				krosswalkdoppel{karte: map[string]string{"mtg-1": "scry-1"}},
				time.Now)
		},
		"das Ablegen klemmt": func() *service.Preisbezug {
			return service.NeuerPreisbezug(&preisdoppel{fehler: gestolpert},
				&bulkdoppel{staende: []domain.Preisstand{{Printing: "scry-1"}}},
				krosswalkdoppel{karte: map[string]string{"mtg-1": "scry-1"}},
				time.Now)
		},
	}

	for name, bauen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := bauen().Holen(t.Context(), "heute"); fehler == nil {
				t.Error("es kam durch")
			}
		})
	}
}
