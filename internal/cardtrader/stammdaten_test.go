package cardtrader_test

// **Was CardTrader über Spiele und Erweiterungen sagt** — Werkstück A4 von
// Phase 5.3
// (#1443).
//
// Die Gestalt der Antworten ist am 2026-09-11 nachgelesen und steht in
// [0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md);
// die Beispiele darunter sind **selbst geschrieben** (§3).

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// **Die Spiele kommen an, und Magic ist darunter.**
func TestSpieleLesen(t *testing.T) {
	t.Parallel()

	server := gegenstelle(t, map[string]string{"/games": "spiele.json"})

	spiele, fehler := klientFuer(server).Spiele(t.Context())
	if fehler != nil {
		t.Fatalf("Spiele: %v", fehler)
	}

	if len(spiele) != 2 {
		t.Fatalf("%d Spiele, erwartet 2", len(spiele))
	}

	if spiele[0].ID != 1 || spiele[0].Name != "Magic" ||
		spiele[0].Anzeigename != "Magic: the Gathering" {
		t.Errorf("das erste Spiel ist %+v", spiele[0])
	}
}

// **Die Erweiterungen tragen ihr Spiel und ihren Code.**
//
// **Der Code ist der Schlüssel zum Bestand**: ein Druck hier kennt seinen
// Setcode, und `ptkdf` ist die Brücke von dort zu `expansion_id`.
func TestErweiterungenLesen(t *testing.T) {
	t.Parallel()

	server := gegenstelle(t, map[string]string{"/expansions": "erweiterungen.json"})

	erweiterungen, fehler := klientFuer(server).Erweiterungen(t.Context())
	if fehler != nil {
		t.Fatalf("Erweiterungen: %v", fehler)
	}

	if len(erweiterungen) != 3 {
		t.Fatalf("%d Erweiterungen, erwartet 3", len(erweiterungen))
	}

	zweite := erweiterungen[1]
	if zweite.ID != 92 || zweite.SpielID != 1 ||
		zweite.Code != "ptkdf" || zweite.Name != "Tarkir Dragonfury" {
		t.Errorf("die zweite Erweiterung ist %+v", zweite)
	}
}

// **Die Marke reist im Kopf mit** — und ohne sie fragt niemand.
func TestDieMarkeStehtImKopf(t *testing.T) {
	t.Parallel()

	gesehen := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			gesehen <- gestellt.Header.Get("Authorization")
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	if _, fehler := klientFuer(server).Spiele(t.Context()); fehler != nil {
		t.Fatalf("Spiele: %v", fehler)
	}

	if kopf := <-gesehen; kopf != "Bearer geheim" {
		t.Errorf("der Kopf ist %q, erwartet »Bearer geheim«", kopf)
	}
}

// **Ohne Marke geht keine Anfrage hinaus.**
//
// **Die Gegenprobe zum Lauf darüber**: ohne sie sagte er nur, dass die Marke
// mitreist, wenn es eine gibt.
func TestOhneMarkeWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	var gefragt bool

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			gefragt = true
			_, _ = schreiben.Write([]byte("[]")) //nolint:errcheck // die Gegenstelle
		}))
	t.Cleanup(server.Close)

	klient := cardtrader.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), cardtrader.Zugang{})

	if _, fehler := klient.Spiele(t.Context()); fehler == nil {
		t.Fatal("ohne Marke kam eine Antwort")
	}

	if gefragt {
		t.Error("der Server wurde ohne Marke gefragt")
	}
}

// gegenstelle antwortet je Pfad mit einer Datei aus `testdaten`.
func gegenstelle(t *testing.T, dateien map[string]string) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, gestellt *http.Request) {
			name, bekannt := dateien[gestellt.URL.Path]
			if !bekannt {
				schreiben.WriteHeader(http.StatusNotFound)

				return
			}

			roh, fehler := os.ReadFile(filepath.Join("testdaten", name)) //nolint:gosec // ein Name aus der Aufzählung dieses Tests
			if fehler != nil {
				t.Errorf("%s lesen: %v", name, fehler)
				schreiben.WriteHeader(http.StatusInternalServerError)

				return
			}

			_, _ = schreiben.Write(roh) //nolint:errcheck // die Gegenstelle
		}))

	t.Cleanup(server.Close)

	return server
}

// klientFuer baut einen Klienten gegen eine Gegenstelle.
func klientFuer(server *httptest.Server) *cardtrader.Klient {
	return cardtrader.NeuerKlient(server.URL,
		netz.KlientUeber(server.Client()), cardtrader.Zugang{Marke: "geheim"})
}
