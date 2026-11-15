package mcp_test

// **Das Werkzeug für das eigene Konto** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Was hier geprüft wird
//
//   - Es kommt durch, mit seinen Argumenten.
//   - **Ohne Naht gibt es kein Werkzeug**: `nil` lässt die Liste schrumpfen
//     statt einen Ruf anzubieten, der abstürzt.
//   - **Es ist lesend und idempotent** — zweimal nachsehen ändert nichts.
//   - **Und es gibt kein zweites**: kein Werkzeug dieser Fläche legt ein
//     Angebot an, kauft oder versendet.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// kontoprobe ist eine Naht, die mitschreibt.
type kontoprobe struct {
	quelle string
	seite  int
	menge  int

	fehler error
}

func (k *kontoprobe) Marktkonto(_ context.Context, quelle string, seite, menge int) (
	mcp.Kontostand, error,
) {
	k.quelle, k.seite, k.menge = quelle, seite, menge

	if k.fehler != nil {
		return mcp.Kontostand{}, k.fehler
	}

	return mcp.Kontostand{
		Name: "Test App", NutzerID: 51242,
		Angebote: []mcp.Eigenangebot{{ID: 1, Name: "Tranquil Cove", Menge: 2}},
		Bestellungen: []mcp.Marktbestellung{{
			ID: 9, Zustand: "hub_pending", UeberNullkonto: true,
			BewegtBestand: true,
		}},
		Befunde: []string{},
	}, nil
}

// rufeKonto sucht das Werkzeug und ruft es.
func rufeKonto(t *testing.T, naht mcp.Kontoauskunft, argumente string) (
	string, error,
) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Konto: naht}) {
		if werkzeug.Name == "lesen.marktkonto" {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug lesen.marktkonto steht nicht in der Liste")

	return "", nil
}

func TestDasMarktkontowerkzeugKommtDurch(t *testing.T) {
	naht := &kontoprobe{}

	aus, err := rufeKonto(t, naht,
		`{"quelle":"cardtrader","seite":2,"menge":50}`)
	if err != nil {
		t.Fatalf("rufen: %v", err)
	}

	if naht.quelle != "cardtrader" || naht.seite != 2 || naht.menge != 50 {
		t.Errorf("gefragt wurde %q / %d / %d", naht.quelle, naht.seite, naht.menge)
	}

	for _, erwartet := range []string{
		`"nutzerID": 51242`, `"bewegtBestand": true`, `"zustand": "hub_pending"`,
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}
}

// **Ein Fehler der Naht kommt an, ein kaputtes Argument endet davor.**
func TestDasMarktkontowerkzeugReichtSeinenFehlerDurch(t *testing.T) {
	naht := &kontoprobe{fehler: errors.New("die Kontoschau ist aus")}

	if _, err := rufeKonto(t, naht, `{"quelle":"cardtrader"}`); err == nil {
		t.Error("der Fehler wurde verschluckt")
	}

	naht = &kontoprobe{}

	if _, err := rufeKonto(t, naht, `kein JSON`); err == nil {
		t.Error("das kaputte Argument kam durch")
	}

	if naht.quelle != "" {
		t.Errorf("die Naht wurde trotzdem gerufen: %q", naht.quelle)
	}
}

// **Ohne Naht gibt es das Werkzeug nicht.**
func TestOhneKontonahtGibtEsDasWerkzeugNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.marktkonto" {
			t.Error("lesen.marktkonto steht da, obwohl keine Naht eingerichtet ist")
		}
	}
}

// **Es ist lesend und idempotent — und es gibt kein schreibendes daneben.**
//
// **Die zweite Hälfte ist die wichtigere**: ein Modell sieht nach und kauft
// nicht. Ein Werkzeug, das ein Angebot anlegte oder einen Warenkorb
// abschickte, gäbe es hier nicht — und dieser Lauf hält es fest, statt es zu
// behaupten (§10).
func TestEsGibtKeinSchreibendesMarktkontowerkzeug(t *testing.T) {
	gesehen := false

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Konto: &kontoprobe{}}) {
		if strings.Contains(werkzeug.Name, "markt") &&
			werkzeug.Art == mcp.Schreibend {
			t.Errorf("%s schreibt bei einem Marktplatz", werkzeug.Name)
		}

		if werkzeug.Name != "lesen.marktkonto" {
			continue
		}

		gesehen = true

		if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
			t.Errorf("es ist als %v geführt, idempotent=%v",
				werkzeug.Art, werkzeug.Idempotent)
		}
	}

	if !gesehen {
		t.Fatal("das Werkzeug steht nicht in der Liste")
	}
}
