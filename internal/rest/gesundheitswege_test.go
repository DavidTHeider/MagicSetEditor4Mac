package rest_test

// **Die zwei Gesundheitswege des Dienstes** — Werkstück O9 von Phase 5.2
// (#1433).
//
// **`/bereit` antwortet mit dem Statuscode**, und das ist der Grund für den
// zweiten Weg: eine Sonde liest 200 oder 503, nicht ein Feld in einem Rumpf.

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// unbereiter meldet einen Prozess, der lebt und nicht arbeiten kann.
type unbereiter struct{}

func (unbereiter) Gesundheit(context.Context) (
	zustand any, bereit bool, fehler error,
) {
	return map[string]any{
		"lebt": true, "bereit": false, "grund": "wanderungLaeuft",
	}, false, nil
}

// **Wer nicht arbeiten kann, bekommt 503 — und den Grund im Rumpf.**
//
// **503 und nicht 500**: „ich kann gerade nicht" ist keine Störung, sondern
// ein Zustand, der vergeht.
func TestBereitGibt503WennErNichtArbeitenKann(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)
	dienst := rest.NeuerDienst(probemarke, func() time.Time { return jetzt },
		rest.Naehte{Gesundheit: unbereiter{}})

	antwort := ruf(t, dienst,
		endpunkt{name: "Bereit", methode: http.MethodGet, weg: "/bereit"}, gueltig)

	switch {
	case antwort.Code != http.StatusServiceUnavailable:
		t.Errorf("Status %d, erwartet 503", antwort.Code)
	case !strings.Contains(antwort.Body.String(), "wanderungLaeuft"):
		t.Errorf("Rumpf %q nennt den Grund nicht", antwort.Body.String())
	}
}
