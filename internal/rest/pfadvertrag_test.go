package rest_test

// **Ein Pfad an den Dienst geht durch den Pfadvertrag** — OPS-02.AK-04
// (#1719):
// „Ein offener Server darf über Pfadangaben keine beliebigen fremden
// Hostdateien offenlegen oder überschreiben."
//
// **Geprüft wird vor dem Weg und nicht in ihm** — an den Namen, unter denen
// ein Pfad reist (`domain.IstPfadfeld`), im Rumpf wie in der Adresse. Ein
// neuer Endpunkt mit einem Pfad bekommt die Prüfung, ohne dass jemand an sie
// denken muss; welche Namen es gibt, hält
// `internal/werkzeuge/pfadfelder_test.go` gegen die Felder des Dienstes.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// nurUnterErlaubt ist ein Vertrag, der nur `/erlaubt/` annimmt.
func nurUnterErlaubt(_ context.Context, pfad string) error {
	if strings.HasPrefix(pfad, "/erlaubt/") {
		return nil
	}

	return domain.NewFieldError("pfad", domain.SchluesselValidierung, "»%s« liegt außerhalb", pfad)
}

func TestEinPfadAusserhalbDesVertragsWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	dienst := dienstMitLeeremSet(t).MitPfadvertrag(nurUnterErlaubt)

	for name, ziel := range map[string]endpunkt{
		"im Rumpf":       {methode: http.MethodPost, weg: "/sets", rumpf: `{"pfad":"/Users/ich/.ssh/id_ed25519"}`},
		"als Ziel":       {methode: http.MethodPost, weg: "/sets/1/sicherung", rumpf: `{"ziel":"/etc/fremd.mse-set"}`},
		"in der Adresse": {methode: http.MethodGet, weg: "/bestand/sicherung?pfad=/etc/passwd"},
	} {
		antwort := ruf(t, dienst, ziel, gueltig)
		if antwort.Code != http.StatusBadRequest || alsBefund(t, antwort).Code != "validierung" {
			t.Errorf("%s: %d %s", name, antwort.Code, antwort.Body.String())
		}
	}
}

// **Was der Vertrag annimmt, kommt unverändert beim Weg an** — samt Rumpf,
// den die Prüfung gelesen hat. Dass jemand zuhört, belegt derselbe Weg mit
// einem fremden Pfad oben.
func TestEinErlaubterPfadKommtUnveraendertAn(t *testing.T) {
	t.Parallel()

	dienst := dienstMitLeeremSet(t).MitPfadvertrag(nurUnterErlaubt)

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/sets", rumpf: `{"pfad":"/erlaubt/probe.mse-set"}`,
	}, gueltig)
	if antwort.Code == http.StatusBadRequest {
		t.Errorf("abgewiesen: %s", antwort.Body.String())
	}

	// **Ohne Pfad ändert sich nichts.**
	if ohne := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/sets"}, gueltig); ohne.Code != http.StatusOK {
		t.Errorf("ohne Pfad: %d", ohne.Code)
	}
}

// **Ein Rumpf, der kein Objekt ist, geht weiter an den Weg** — der sagt dann,
// was an ihm falsch ist; der Vertrag erfindet keine zweite Meldung.
func TestEinRumpfOhneObjektGehtAnDenWeg(t *testing.T) {
	t.Parallel()

	dienst := dienstMitLeeremSet(t).MitPfadvertrag(func(context.Context, string) error {
		t.Error("der Vertrag wurde gefragt")

		return nil
	})

	anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/sets", strings.NewReader(`["kein","objekt"]`))
	gueltig(anfrage)

	antwort := httptest.NewRecorder()
	dienst.Behandler().ServeHTTP(antwort, anfrage)

	if antwort.Code != http.StatusBadRequest || !strings.Contains(antwort.Body.String(), "validierung") {
		t.Errorf("%d %s", antwort.Code, antwort.Body.String())
	}
}

// abbrechender ist ein Rumpf, der beim Lesen scheitert.
type abbrechender struct{}

func (abbrechender) Read([]byte) (int, error) { return 0, errors.New("die Leitung ist weg") }

// **Die Ränder**: ein Adressparameter, der kein Pfad ist, ein Ruf ohne Rumpf
// und ein Rumpf, der beim Lesen abbricht.
func TestDerPfadvertragAnSeinenRaendern(t *testing.T) {
	t.Parallel()

	gefragt := 0
	dienst := dienstMitLeeremSet(t).MitPfadvertrag(func(context.Context, string) error {
		gefragt++

		return nil
	})

	for _, anfrage := range []*http.Request{
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets?seite=2", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/sets", http.NoBody),
	} {
		gueltig(anfrage)

		antwort := httptest.NewRecorder()
		dienst.Behandler().ServeHTTP(antwort, anfrage)

		if antwort.Code != http.StatusOK {
			t.Errorf("%s: %d", anfrage.URL, antwort.Code)
		}
	}

	if gefragt != 0 {
		t.Errorf("ohne Pfad wurde der Vertrag %d-mal gefragt", gefragt)
	}

	abgebrochen := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/sets", abbrechender{})
	gueltig(abgebrochen)

	antwort := httptest.NewRecorder()
	dienst.Behandler().ServeHTTP(antwort, abgebrochen)

	if antwort.Code != http.StatusBadRequest {
		t.Errorf("ein abgebrochener Rumpf: %d %s", antwort.Code, antwort.Body.String())
	}
}
