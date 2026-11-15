package app

// **Eine Auskunft, die sich nicht packen lässt, wird gemeldet** — Werkstück 8
// von Phase 4.12
// (#1351).
//
// **Ein interner Lauf, weil die Packnaht intern ist** — derselbe Grund wie
// bei `kartenbilder_intern_test.go`: sie von außen erreichbar zu machen hieße,
// eine Naht für einen Test zu öffnen, die im Betrieb niemand braucht.
//
// **Und die Datei bleibt aus.** Eine Meldung, neben der trotzdem eine halbe
// Datei läge, wäre der schlechtere Ausgang (§8).

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEineNichtPackbareAuskunftWirdGemeldet(t *testing.T) {
	alt := auskunftPacken

	t.Cleanup(func() { auskunftPacken = alt })

	auskunftPacken = func(any) ([]byte, error) {
		return nil, errors.New("die Auskunft ist zu groß für diese Welt")
	}

	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	ziel := filepath.Join(t.TempDir(), "meine-daten.json")

	if antwort := anwendung.PersonendatenSchreiben(
		t.Context(), "1", ziel); antwort.Fehler == nil {
		t.Fatal("der Packfehler ging verloren")
	}

	if _, err := os.Stat(ziel); !os.IsNotExist(err) {
		t.Errorf("es liegt trotzdem eine Datei: %v", err)
	}
}
