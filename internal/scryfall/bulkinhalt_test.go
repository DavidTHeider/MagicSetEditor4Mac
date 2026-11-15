package scryfall_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// **Eine förmlich geglückte Antwort mit Unsinn darin ersetzt den letzten
// guten Stand nicht** —
// #1027.
//
// # Der Zustand, den es zu verhindern gilt
//
// `BulkHolen` prüfte den Status und die Größengrenze, nicht den Inhalt. Eine
// Antwort mit 200 und 24 MB Unsinn — die Fehlerseite eines Zwischenspeichers,
// eine abgeschnittene Datei mit gültigem gzip-Kopf — ersetzte die vorhandene
// Sammeldatei.
//
// **Der Zweischritt schützt davor nicht.** Er schließt das Fenster einer halb
// **geschriebenen** Datei; eine vollständig übertragene, inhaltlich falsche
// geht durch ihn hindurch.
//
// Was das kostet, ist die Umkehrung von §16, Regel 5: der Rückfall ohne Netz
// zeigt danach nichts mehr, und der letzte gute Stand ist fort.
func TestEineUnsinnigeAntwortErsetztDenStandNicht(t *testing.T) {
	t.Parallel()

	gut := gezipptesJSONL(t,
		`{"object":"card","id":"a","name":"Lightning Bolt"}`)

	faelle := []struct {
		name    string
		antwort []byte
		eintrag scryfall.Bulkeintrag
		meldung string
	}{
		{
			name:    "eine Fehlerseite statt der Datei",
			antwort: []byte("<html><body>502 Bad Gateway</body></html>"),
			meldung: "kein gzip",
		},
		{
			name:    "ein gültiges gzip ohne JSON darin",
			antwort: gezipptesJSONL(t, "Service Unavailable"),
			meldung: "kein JSON",
		},
		{
			name: "ein JSON, das keine Karte ist",
			antwort: gezipptesJSONL(t,
				`{"object":"error","code":"not_found","details":"weg"}`),
			meldung: `"error"`,
		},
		{
			// **Über der Zeilengrenze** (4 MiB). Ein Kartenobjekt wiegt
			// wenige Kilobyte; eine 5-MB-Zeile ist keine Sammeldatei.
			name: "eine erste Zeile, die niemand liest",
			antwort: gezipptesJSONL(t,
				`{"object":"card","name":"`+strings.Repeat("x", 5<<20)+`"}`),
			meldung: "ließ sich nicht lesen",
		},
		{
			name:    "ein JSON-Objekt ohne `object`",
			antwort: gezipptesJSONL(t, `{"name":"Lightning Bolt"}`),
			meldung: "kein `object`",
		},
		{
			name:    "eine leere Datei",
			antwort: gezipptesJSONL(t),
			meldung: "keine Zeile",
		},
		{
			name:    "abgeschnitten — viel kleiner als angekündigt",
			antwort: gut,
			eintrag: scryfall.Bulkeintrag{Groesse: 24_531_600},
			meldung: "angekündigt",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")
			if fehler := os.WriteFile(ziel, gut, 0o600); fehler != nil {
				t.Fatalf("den guten Stand ablegen: %v", fehler)
			}

			server := httptest.NewServer(http.HandlerFunc(
				func(schreiben http.ResponseWriter, _ *http.Request) {
					if _, fehler := schreiben.Write(fall.antwort); fehler != nil {
						t.Errorf("schreiben: %v", fehler)
					}
				}))
			t.Cleanup(server.Close)

			klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))

			eintrag := fall.eintrag
			eintrag.Art, eintrag.Adresse = "oracle_cards", server.URL+"/oracle.jsonl.gz"

			fehler := klient.BulkHolen(t.Context(), eintrag, ziel)
			if fehler == nil {
				t.Fatal("die unsinnige Antwort kam durch")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v — kein Validierungsbefund", fehler)
			}

			if !bytes.Contains([]byte(fehler.Error()), []byte(fall.meldung)) {
				t.Errorf("Fehler %v — erwartet war %q", fehler, fall.meldung)
			}

			// **Und der letzte gute Stand liegt noch da.** Das ist der Punkt:
			// eine abgewiesene Antwort darf nichts kosten.
			abgelegt, lesefehler := os.ReadFile(ziel) //nolint:gosec // Pfad aus t.TempDir()
			if lesefehler != nil {
				t.Fatalf("der gute Stand ist fort: %v", lesefehler)
			}

			if !bytes.Equal(abgelegt, gut) {
				t.Error("der gute Stand wurde überschrieben")
			}

			// **Und es liegt keine Nebendatei herum**, die beim nächsten Lauf
			// jemanden verwirrte.
			if _, fehler := os.Stat(ziel + ".teil"); fehler == nil {
				t.Error("die Nebendatei blieb liegen")
			}
		})
	}
}

// **Eine gültige Datei, die größer ist als angekündigt, kommt durch.**
//
// # Warum die Grenze nur nach unten prüft
//
// `compressed_size` stammt aus dem Verzeichnis, und das darf bis zu einem Tag
// alt sein (der Zwischenspeicher von `Bulkverzeichnis`). Scryfall
// veröffentlicht die Sammeldateien täglich neu; die Datei ist zwischen
// Verzeichnis und Abruf also **legitim** gewachsen. Eine Prüfung nach oben
// wiese genau diesen Normalfall ab.
func TestEineGewachseneDateiKommtDurch(t *testing.T) {
	t.Parallel()

	// **Mit einer Leerzeile davor.** Sie ist kein Befund: gesucht ist die
	// erste **Zeile mit Inhalt**, und eine Datei, die mit einem Zeilenumbruch
	// beginnt, ist deshalb noch keine kaputte.
	inhalt := gezipptesJSONL(t, ``, `{"object":"card","id":"a","name":"Lightning Bolt"}`)

	server := httptest.NewServer(http.HandlerFunc(
		func(schreiben http.ResponseWriter, _ *http.Request) {
			if _, fehler := schreiben.Write(inhalt); fehler != nil {
				t.Errorf("schreiben: %v", fehler)
			}
		}))
	t.Cleanup(server.Close)

	klient := klientFuer(t, server.URL, netz.KlientUeber(server.Client()))
	ziel := filepath.Join(t.TempDir(), "oracle.jsonl.gz")

	if fehler := klient.BulkHolen(t.Context(), scryfall.Bulkeintrag{
		Art: "oracle_cards", Adresse: server.URL + "/x", Groesse: 10,
	}, ziel); fehler != nil {
		t.Fatalf("BulkHolen: %v", fehler)
	}
}
