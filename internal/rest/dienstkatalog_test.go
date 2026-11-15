package rest_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

type katalogprobe struct {
	eintraege []rest.Dienstfaehigkeit
	fehler    error
}

func (k katalogprobe) Dienstkatalog(context.Context) ([]rest.Dienstfaehigkeit, error) {
	return k.eintraege, k.fehler
}

func dienstMitKatalog(t *testing.T, katalog rest.Fremdkatalog) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Leser: williger{}, Katalog: katalog})
}

func TestDerDienstkatalogNenntJedeFaehigkeit(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMitKatalog(t, katalogprobe{eintraege: []rest.Dienstfaehigkeit{{
		Dienst: "scryfall", Frage: "suche", Status: "bereit", HatWahl: true, Gewaehlt: true,
		Wirkung:  "externSenden",
		Vertrag:  rest.Dienstvertrag{Basisadresse: "https://api.scryfall.com"},
		Nachweis: &rest.Dienstnachweis{Datum: "2026-09-01", Fassung: "0.3.0"},
	}}}), endpunkt{methode: http.MethodGet, weg: "/dienstkatalog"}, gueltig)

	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	for _, erwartet := range []string{
		`"status":"bereit"`, `"giltFuerDieseFassung":false`, `"basisadresse":"https://api.scryfall.com"`,
	} {
		if !strings.Contains(antwort.Body.String(), erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, antwort.Body.String())
		}
	}
}

func TestEinLeererDienstkatalogIstEineListe(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMitKatalog(t, katalogprobe{}),
		endpunkt{methode: http.MethodGet, weg: "/dienstkatalog"}, gueltig)

	if antwort.Code != http.StatusOK || !strings.Contains(antwort.Body.String(), `"eintraege":[]`) {
		t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
	}
}

// **Ein fachlicher Fehler kommt übersetzt an, ein fremder nicht wörtlich** (§8).
func TestDerDienstkatalogMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	fachlich := ruf(t, dienstMitKatalog(t, katalogprobe{
		fehler: fmt.Errorf("%w: die Wahl nennt einen Dienst, den es nicht gibt", domain.ErrValidation),
	}), endpunkt{methode: http.MethodGet, weg: "/dienstkatalog"}, gueltig)

	if fachlich.Code != http.StatusBadRequest || !strings.Contains(fachlich.Body.String(), "nicht gibt") {
		t.Errorf("Status %d: %s", fachlich.Code, fachlich.Body.String())
	}

	fremd := ruf(t, dienstMitKatalog(t, katalogprobe{fehler: fmt.Errorf("sqlite: disk I/O error")}),
		endpunkt{methode: http.MethodGet, weg: "/dienstkatalog"}, gueltig)

	if fremd.Code != http.StatusInternalServerError || strings.Contains(fremd.Body.String(), "sqlite") {
		t.Errorf("Status %d: %s — eine rohe Meldung erreicht die Fläche nie", fremd.Code, fremd.Body.String())
	}
}
