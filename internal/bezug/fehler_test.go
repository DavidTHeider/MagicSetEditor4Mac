package bezug_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
)

// Was beim Beziehen schiefgehen kann, wird gemeldet — mit der Stelle.
func TestWasSchiefgehtWirdGemeldet(t *testing.T) {
	gutesArchiv := archiv(t, "wurzel", probeeintraege())

	// Ein Dienst, der nichts kann: die Adresse zeigt ins Leere.
	tot := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(tot.Close)

	// Ein Dienst, der die Marke kennt und dann Unsinn liefert.
	kaputt := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/marke/") {
			return
		}

		if _, err := w.Write([]byte("das ist kein gzip")); err != nil {
			t.Errorf("ausliefern: %v", err)
		}
	}))
	t.Cleanup(kaputt.Close)

	faelle := map[string]struct {
		quelle   bezug.Repositorium
		enthaelt string
	}{
		"kein Stand": {
			quelle:   bezug.Repositorium{Adresse: tot.URL, Marken: tot.URL + "/marke/"},
			enthaelt: "Stand",
		},
		"eine Adresse, die es nicht gibt": {
			quelle: bezug.Repositorium{
				Adresse: tot.URL + "/archiv",
				Marken:  kaputt.URL + "/marke/",
				Stand:   "3.1.0",
			},
			enthaelt: "404",
		},
		"kein Archiv, sondern Text": {
			quelle: bezug.Repositorium{
				Adresse: kaputt.URL + "/archiv",
				Marken:  kaputt.URL + "/marke/",
				Stand:   "3.1.0",
			},
			enthaelt: "auspacken",
		},
		"eine Gegenstelle, die es nicht gibt": {
			quelle: bezug.Repositorium{
				Adresse: "http://127.0.0.1:1/archiv",
				Marken:  "http://127.0.0.1:1/marke/",
				Stand:   "3.1.0",
			},
			enthaelt: "nachschlagen",
		},
		"ein Unterverzeichnis, das das Archiv nicht führt": {
			quelle: bezug.Repositorium{
				Adresse:          kaputt.URL + "/archiv",
				Marken:           kaputt.URL + "/marke/",
				Stand:            "3.1.0",
				Unterverzeichnis: "gibtsnicht",
			},
			enthaelt: "auspacken",
		},
	}

	_ = gutesArchiv

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			_, err := bezug.Hole(t.Context(), fall.quelle, t.TempDir(), "MagicSetEditor4Mac/probe")
			if err == nil {
				t.Fatal("kein Fehler")
			}

			if !strings.Contains(err.Error(), fall.enthaelt) {
				t.Errorf("Fehler: %q — %q fehlt darin", err, fall.enthaelt)
			}
		})
	}
}

// Ein Ziel, in das sich nicht schreiben lässt, wird gemeldet.
func TestEinUnschreibbaresZielWirdGemeldet(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root ist jedes Verzeichnis beschreibbar")
	}

	dienst := server(t, archiv(t, "wurzel", probeeintraege()), "3.1.0")

	ziel := filepath.Join(t.TempDir(), "zu")
	if err := os.Mkdir(ziel, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(ziel, 0o700); err != nil { // #nosec G302 -- Verzeichnis
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	if _, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), ziel, "MagicSetEditor4Mac/probe"); err == nil {
		t.Error("kein Fehler")
	}
}

// Verzeichnisse und andere Einträge werden übergangen, nicht bemängelt.
//
// Ein `tar` führt Verzeichniseinträge; sie sind keine Dateien und kein Befund.
func TestEinVerzeichniseintragIstKeinBefund(t *testing.T) {
	dienst := server(t, archiv(t, "wurzel", []eintrag{
		{name: "data/probe.mse-game/game", inhalt: "short name: Probe\n"},
	}), "3.1.0")

	ergebnis, err := bezug.Hole(t.Context(), quelle(dienst, "3.1.0"), t.TempDir(), "MagicSetEditor4Mac/probe")
	if err != nil {
		t.Fatalf("Hole: %v", err)
	}

	if len(ergebnis.Uebergangen) != 0 {
		t.Errorf("Übergangen: %v, erwartet keine", ergebnis.Uebergangen)
	}
}

// Die Herkunft kommt auch mit einer Adresse zurecht, die kein Repositorium ist.
//
// **Weil Nutzende tippen, was sie wollen.** Eine Auskunft, die bei einer
// ungewöhnlichen Adresse abstürzt, verhindert genau das, was vor dem Bezug
// stehen soll.
func TestDieHerkunftTraegtJedeAdresse(t *testing.T) {
	faelle := map[string]string{
		"ohne Schema":     "example.org/pakete/probe",
		"nur ein Rechner": "https://example.org",
		"leer":            "",
	}

	for name, adresse := range faelle {
		t.Run(name, func(t *testing.T) {
			woher := bezug.Woher(bezug.Repositorium{Adresse: adresse, Stand: "1.0"})
			if woher.Stand != "1.0" {
				t.Errorf("Stand: %q", woher.Stand)
			}
		})
	}
}
