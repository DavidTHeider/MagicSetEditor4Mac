package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// **Der Verlauf über REST** — Werkstück U6 von Phase 5.1 (#1417):
// `GET /sets/{id}/verlauf`, der jüngste Schritt zuerst.
func TestDerVerlaufAntwortetMitSchritten(t *testing.T) {
	t.Parallel()

	antwort := ruf(t, dienstMit(t, &zaehlender{}), endpunkt{
		methode: http.MethodGet, weg: "/sets/1/verlauf",
	}, gueltig)
	if antwort.Code != http.StatusOK {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}

	var gelesen struct {
		Schritte []struct {
			Vorgang string `json:"vorgang"`
			Umfang  int    `json:"umfang"`
			KarteID string `json:"karteID"`
			Nachher string `json:"nachher"`
		} `json:"schritte"`
	}
	if err := json.Unmarshal(antwort.Body.Bytes(), &gelesen); err != nil {
		t.Fatalf("Rumpf %q: %v", antwort.Body.String(), err)
	}

	if len(gelesen.Schritte) != 1 || gelesen.Schritte[0].Vorgang != "g1" ||
		gelesen.Schritte[0].Umfang != 2 || gelesen.Schritte[0].KarteID != "8" {
		t.Errorf("Schritte: %+v", gelesen.Schritte)
	}
}

// **Ein Bestand, der nicht antwortet, gibt keinen 200** — und keinen leeren
// Verlauf.
func TestDerVerlaufMeldetEinenUnwilligenBestand(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Verlauf: unwilliger{fehler: errors.New("die Naht schweigt")}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/verlauf",
	}, gueltig)

	if antwort.Code < 400 {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}
}

// ohneVerlauf antwortet mit einer nil-Liste — der Fall des jungen Sets.
type ohneVerlauf struct{ williger }

func (ohneVerlauf) Verlauf(context.Context, int64) ([]rest.Verlaufsschritt, error) {
	return nil, nil
}

// **Ein Set ohne Verlauf antwortet mit einer leeren Liste, nicht mit `null`.**
func TestEinLeererVerlaufIstEineLeereListe(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Verlauf: ohneVerlauf{}})

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodGet, weg: "/sets/1/verlauf",
	}, gueltig)

	if antwort.Code != http.StatusOK || !strings.Contains(antwort.Body.String(), `"schritte":[]`) {
		t.Fatalf("Status %d: %s", antwort.Code, antwort.Body.String())
	}
}
