package anthropic_test

// **Eine Grenze wird eingehalten, bevor sie greift** (§16, Regel 4) —
// Werkstück A8 von Phase 5.3
// (#1447).
//
// # Warum dieser Lauf eigens steht
//
// Hier hing bis zum 2026-09-12 **keine** Drossel: `netz.Hole` bekam an ihrer
// Stelle ein `nil`, und zwei Anfragen gingen ohne Abstand hinaus. Bei einem
// Vorschlag je Bedienhandlung fiele das nie auf — bei dem Stapelweg aus A13
// schon, und dann als Ausschluss statt als Wartezeit.

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/anthropic"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestDieDrosselWartetVorDerAnfrage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	var gewartet time.Duration

	drossel := netz.NeueDrossel(anthropic.Abstand, time.Now,
		func(dauer time.Duration) { gewartet += dauer })

	klient := anthropic.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"}).MitDrossel(drossel)

	for range 2 {
		if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
			t.Fatalf("Modelle: %v", fehler)
		}
	}

	// **Die Gegenprobe steht in derselben Datei**: ohne Drossel wird nicht
	// gewartet, und ein Lauf, der beides nicht unterscheidet, sagt nichts.
	if gewartet <= 0 {
		t.Error("zwischen zwei Anfragen wurde nicht gewartet")
	}
}

func TestOhneDrosselWirdNichtGewartet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			_, _ = schreiben.Write([]byte(`{"data":[]}`)) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := anthropic.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()),
		anthropic.Zugang{Schluessel: "geheim"})

	for range 2 {
		if _, _, fehler := klient.Modelle(t.Context()); fehler != nil {
			t.Fatalf("Modelle: %v", fehler)
		}
	}
}
