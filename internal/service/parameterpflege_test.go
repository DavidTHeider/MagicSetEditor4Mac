package service_test

// **Die Parameterhilfen an einem Set** — Werkstück D von Phase 4.11
// (#1342).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// parametervorlage ist die Spielvorlage samt Parameterarten.
func parametervorlage() service.Vorlage {
	vorlage := spielvorlage()
	vorlage.Parameterarten = probearten()

	return vorlage
}

// **Beide Menüs kommen in einer Antwort.**
func TestDieParameterhilfenGebenBeideMenues(t *testing.T) {
	t.Parallel()

	ablage := &wortablage{woerter: []*domain.Keyword{
		{ID: 7, Word: "Kicker", Match: "Kicker <atom-param>mana</atom-param>"},
	}}

	hilfen, err := service.NeueSchluesselwortpflege(
		ablage, spielwoerter{parametervorlage(), true}).
		Parameterhilfen(t.Context(), 1, 7)
	if err != nil {
		t.Fatalf("Parameterhilfen: %v", err)
	}

	// **Die Arten des Spiels** — beide, unabhängig vom Muster.
	if len(hilfen.Arten) != 2 || hilfen.Arten[0].Name != "zahl" {
		t.Errorf("die Arten lauten %+v", hilfen.Arten)
	}

	// **Und die Verweise aus dem Muster** — `mana` bringt zwei Skripte mit.
	if len(hilfen.Verweise) != 2 || hilfen.Verweise[0].Einfuegen != "{param1}" {
		t.Errorf("die Verweise lauten %+v", hilfen.Verweise)
	}

	if len(hilfen.Befunde) != 0 {
		t.Errorf("die Befunde lauten %+v", hilfen.Befunde)
	}
}

// **Ein festes Schlüsselwort hat keine Hilfen.**
//
// Es gehört dem **Spielpaket**; seine Parameter zu bearbeiten hieße, das Paket
// zu ändern — und das geht über die Paketwerkstatt, nicht über diese Seite.
func TestEinFestesSchluesselwortHatKeineParameterhilfen(t *testing.T) {
	t.Parallel()

	pflege := service.NeueSchluesselwortpflege(
		&wortablage{}, spielwoerter{parametervorlage(), true})

	if _, err := pflege.Parameterhilfen(t.Context(), 1, 0); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("der Fehler lautet %v", err)
	}
}

// **Ein Schlüsselwort, das es nicht gibt, ist ein fachliches Nein.**
func TestEinUnbekanntesSchluesselwortHatKeineParameterhilfen(t *testing.T) {
	t.Parallel()

	pflege := service.NeueSchluesselwortpflege(
		&wortablage{}, spielwoerter{parametervorlage(), true})

	if _, err := pflege.Parameterhilfen(t.Context(), 1, 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("der Fehler lautet %v", err)
	}
}

// **Ein Lesefehler kommt durch.**
func TestEinLesefehlerKommtDurchDieParameterhilfen(t *testing.T) {
	t.Parallel()

	eigener := errors.New("der Bestand antwortet nicht")

	pflege := service.NeueSchluesselwortpflege(
		&wortablage{fehler: eigener}, spielwoerter{parametervorlage(), true})

	if _, err := pflege.Parameterhilfen(t.Context(), 1, 7); !errors.Is(err, eigener) {
		t.Errorf("der Fehler lautet %v", err)
	}
}

// **Ohne Spielpaket gibt es keine Arten — und keinen Fehler.**
//
// §6: ein leerer Zustand erklärt den nächsten Schritt. Ein Set ohne Spielpaket
// hat keine Parameterarten, und ein leeres Menü ist die richtige Antwort.
func TestOhneSpielpaketGibtEsKeineParameterarten(t *testing.T) {
	t.Parallel()

	ablage := &wortablage{woerter: []*domain.Keyword{
		{ID: 7, Word: "Kicker", Match: "Kicker"},
	}}

	hilfen, err := service.NeueSchluesselwortpflege(ablage, spielwoerter{}).
		Parameterhilfen(t.Context(), 1, 7)
	if err != nil {
		t.Fatalf("Parameterhilfen: %v", err)
	}

	if len(hilfen.Arten) != 0 || len(hilfen.Verweise) != 0 {
		t.Errorf("die Hilfen lauten %+v", hilfen)
	}
}
