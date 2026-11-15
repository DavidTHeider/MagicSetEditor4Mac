package rest_test

// **Bereiche und Funktionen abschalten, am Dienst** — OPS-10
// (#1908).

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

func (williger) Schalter(context.Context) ([]rest.Schalter, error) {
	return []rest.Schalter{{Kennung: "spielen", Art: "bereich", Reife: "stabil", Vorgabe: "an", An: true}}, nil
}

func (z *zaehlender) Schalter(context.Context) ([]rest.Schalter, error) {
	z.zaehle()

	return []rest.Schalter{}, nil
}

func (u unwilliger) Schalter(context.Context) ([]rest.Schalter, error) {
	return nil, u.fehler
}

// dienstMitSchaltern baut einen Dienst, dessen Schalternaht sich merkt, was
// sie gefragt wurde, und abweist, was in `aus` steht.
func dienstMitSchaltern(t *testing.T, aus map[string]bool, gefragt *[]string) *rest.Dienst {
	t.Helper()

	jetzt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)

	return rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Leser: williger{}, Schalter: williger{}}).
		MitSchaltern(func(_ context.Context, weg string) error {
			*gefragt = append(*gefragt, weg)
			if aus[weg] {
				return &domain.Abschaltung{Schalter: "spielen"}
			}

			return nil
		})
}

// **Ein abgeschalteter Endpunkt weist mit 403 und eigenem Code ab** — und
// gefragt wird mit dem Muster ohne Platzhalternamen.
func TestEinAbgeschalteterEndpunktWeistAb(t *testing.T) {
	t.Parallel()

	var gefragt []string

	dienst := dienstMitSchaltern(t, map[string]bool{"GET /schalter": true}, &gefragt)

	antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: "/schalter"}, gueltig)
	if antwort.Code != http.StatusForbidden || !strings.Contains(antwort.Body.String(), `"code":"abgeschaltet"`) ||
		!strings.Contains(antwort.Body.String(), "spielen") {
		t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	if len(gefragt) != 1 || gefragt[0] != "GET /schalter" {
		t.Errorf("gefragt: %v", gefragt)
	}
}

// **Ohne Marke wird nicht gefragt** — ein Fremder erfährt nicht, was
// abgeschaltet ist.
func TestOhneMarkeFragtDerDienstKeinenSchalter(t *testing.T) {
	t.Parallel()

	var gefragt []string

	antwort := ruf(t, dienstMitSchaltern(t, map[string]bool{}, &gefragt),
		endpunkt{methode: http.MethodGet, weg: "/schalter"}, ohneMarke)
	if antwort.Code != http.StatusUnauthorized || len(gefragt) != 0 {
		t.Errorf("Status %d, gefragt %v", antwort.Code, gefragt)
	}
}

// **Ein offener Endpunkt antwortet** — mit dem Verzeichnis.
func TestDieSchalterSindUeberDenDienstLesbar(t *testing.T) {
	t.Parallel()

	var gefragt []string

	antwort := ruf(t, dienstMitSchaltern(t, map[string]bool{}, &gefragt),
		endpunkt{methode: http.MethodGet, weg: "/schalter"}, gueltig)
	if antwort.Code != http.StatusOK || !strings.Contains(antwort.Body.String(), `"kennung":"spielen"`) {
		t.Errorf("Status %d: %s", antwort.Code, antwort.Body.String())
	}
}
