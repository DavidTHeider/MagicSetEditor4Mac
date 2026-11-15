package service_test

// **Die Suche über den gewählten Dienst** — Werkstück 8 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// quellprobe gibt einen festen Treffer und merkt sich, ob sie gefragt wurde.
type quellprobe struct {
	name    string
	gefragt int
	fehler  error
}

func (q *quellprobe) Suche(_ context.Context, _ string) (
	*service.Suchauskunft, error,
) {
	q.gefragt++

	if q.fehler != nil {
		return nil, q.fehler
	}

	return &service.Suchauskunft{
		Treffer: []service.Suchtreffer{{Name: q.name}},
	}, nil
}

// wahlprobe sagt, welcher Dienst die Frage beantwortet.
type wahlprobe struct {
	dienst domain.Dienst
	fehler error
}

func (w wahlprobe) FuerFrage(_ context.Context, _ domain.Frage) (
	domain.Dienst, error,
) {
	return w.dienst, w.fehler
}

// **Gefragt wird der gewählte Dienst — und nur er.**
//
// **Und das ist die Zusage aus 0063, Entscheidung 1**: kein stilles
// Ausweichen. Wer „magicthegathering.io" einstellt und eine Antwort von
// Scryfall bekäme, sähe es nicht.
// kartenprobe ist eine feste Quellenkarte.
//
// **Sie zählt mit, wie oft gefragt wurde** — seit die Karte je Anfrage
// entsteht (Befund 1 des Zweitblicks über Phase 4.8), ist das eine Aussage
// über den Weg und nicht nur über das Ergebnis.
type kartenprobe struct {
	bereit  map[domain.Dienst]service.Suchdienst
	grund   map[domain.Dienst]error
	fehler  error
	gefragt int
}

func (k *kartenprobe) Fuer(context.Context) (service.Quellen, error) {
	k.gefragt++

	if k.fehler != nil {
		return service.Quellen{}, k.fehler
	}

	return service.Quellen{Bereit: k.bereit, Grund: k.grund}, nil
}

func TestGefragtWirdDerGewaehlteDienst(t *testing.T) {
	t.Parallel()

	var (
		scryfall = &quellprobe{name: "von Scryfall"}
		mtgapi   = &quellprobe{name: "von der MTG-API"}
	)

	suche := service.NeueDienstsuche(
		&kartenprobe{bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: scryfall,
			domain.DienstMTGAPI:   mtgapi,
		}},
		wahlprobe{dienst: domain.DienstMTGAPI})

	auskunft, err := suche.Suche(t.Context(), "Blitz")
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if auskunft.Treffer[0].Name != "von der MTG-API" {
		t.Errorf("die Antwort kam von %q", auskunft.Treffer[0].Name)
	}

	if scryfall.gefragt != 0 {
		t.Errorf("Scryfall wurde %dmal gefragt, erwartet keinmal", scryfall.gefragt)
	}

	if mtgapi.gefragt != 1 {
		t.Errorf("die MTG-API wurde %dmal gefragt, erwartet einmal", mtgapi.gefragt)
	}
}

// **Ein Dienst, der scheitert, wird nicht durch einen anderen ersetzt.**
//
// **Das ist der Kern von 0063**: ein Ausweichweg ist still. Eine Antwort aus
// einer nicht gewählten Quelle sieht aus wie eine aus der gewählten — bei
// einem Bild harmlos, bei einem Preis eine falsche Zahl mit einem falschen
// Namen daneben.
func TestEinGescheiterterDienstWirdNichtErsetzt(t *testing.T) {
	t.Parallel()

	var (
		scryfall = &quellprobe{name: "von Scryfall"}
		mtgapi   = &quellprobe{fehler: domain.ErrNotFound}
	)

	suche := service.NeueDienstsuche(
		&kartenprobe{bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: scryfall,
			domain.DienstMTGAPI:   mtgapi,
		}},
		wahlprobe{dienst: domain.DienstMTGAPI})

	if _, err := suche.Suche(t.Context(), "Blitz"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Fatalf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	if scryfall.gefragt != 0 {
		t.Errorf("Scryfall sprang ein — %dmal", scryfall.gefragt)
	}
}

// **Ein gewählter Dienst ohne Quelle sagt es.**
//
// Der Fall entsteht, wenn jemand einen Dienst wählt, dessen Zugang fehlt —
// dann gibt es ihn in der Karte nicht. „Nichts gefunden" wäre die falsche
// Antwort: gefunden wurde nichts, weil gar nicht gefragt wurde.
func TestEinGewaehlterDienstOhneQuelleSagtEs(t *testing.T) {
	t.Parallel()

	suche := service.NeueDienstsuche(
		&kartenprobe{bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: &quellprobe{name: "von Scryfall"},
		}},
		wahlprobe{dienst: domain.DienstScrydex})

	_, err := suche.Suche(t.Context(), "Blitz")

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler ist %v, erwartet eine Validierung", err)
	}

	if !strings.Contains(err.Error(), "scrydex") {
		t.Errorf("die Meldung %q nennt den gewählten Dienst nicht", err)
	}
}

// **Ein Fehler beim Lesen der Wahl kommt an.**
func TestEinFehlerBeimLesenDerWahlKommtAn(t *testing.T) {
	t.Parallel()

	suche := service.NeueDienstsuche(
		&kartenprobe{bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: &quellprobe{name: "von Scryfall"},
		}},
		wahlprobe{dienst: domain.DienstScryfall, fehler: domain.ErrNotFound})

	if _, err := suche.Suche(t.Context(), "Blitz"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Der genannte Grund geht vor** — Befund 1 des Zweitblicks über Phase 4.8.
//
// Seit die Karte je Anfrage entsteht, fehlt ein abgeschalteter Dienst schlicht
// darin. Ohne den Grund daneben sagte die Meldung „fehlt sein Zugang, oder ist
// er abgeschaltet?" — eine Frage zurück an den Lesenden statt einer Auskunft.
func TestDerGenannteGrundGehtVor(t *testing.T) {
	t.Parallel()

	karte := &kartenprobe{
		bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: &quellprobe{name: "von Scryfall"},
		},
		grund: map[domain.Dienst]error{
			domain.DienstMTGAPI: domain.DienstIstAus(domain.DienstMTGAPI),
		},
	}

	suche := service.NeueDienstsuche(karte, wahlprobe{dienst: domain.DienstMTGAPI})

	_, err := suche.Suche(t.Context(), "Blitz")

	if !domain.IstDienstAus(err) {
		t.Fatalf("Fehler ist %v, erwartet ErrDienstAus", err)
	}
}

// **Die Karte wird je Anfrage gelesen.**
//
// Das ist die Zusage aus Befund 1 in einem Satz: zwei Suchen, zwei Fragen an
// die Karte. Eine feste Karte wäre der Stand vom Verdrahten.
func TestDieKarteWirdJeAnfrageGelesen(t *testing.T) {
	t.Parallel()

	karte := &kartenprobe{
		bereit: map[domain.Dienst]service.Suchdienst{
			domain.DienstScryfall: &quellprobe{name: "von Scryfall"},
		},
	}

	suche := service.NeueDienstsuche(karte, wahlprobe{dienst: domain.DienstScryfall})

	for range 2 {
		if _, err := suche.Suche(t.Context(), "Blitz"); err != nil {
			t.Fatalf("suchen: %v", err)
		}
	}

	if karte.gefragt != 2 {
		t.Errorf("die Karte wurde %dmal gefragt, erwartet zweimal", karte.gefragt)
	}
}

// **Ein Fehler beim Lesen der Karte hält die Suche an.**
func TestEinFehlerDerKarteHaeltDieSucheAn(t *testing.T) {
	t.Parallel()

	suche := service.NeueDienstsuche(
		&kartenprobe{fehler: domain.ErrConflict},
		wahlprobe{dienst: domain.DienstScryfall})

	if _, err := suche.Suche(t.Context(), "Blitz"); !errors.Is(
		err, domain.ErrConflict) {
		t.Errorf("Fehler ist %v, erwartet ErrConflict", err)
	}
}
