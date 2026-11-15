package app

// **Ein Austausch, der sich nicht packen lässt, wird gemeldet** — Werkstück O3
// von Phase 5.2
// (#1427).
//
// **Ein interner Lauf, weil die Packnaht intern ist** — derselbe Grund wie bei
// `personendatenpacken_intern_test.go` daneben: sie von außen erreichbar zu
// machen hieße, eine Naht für einen Test zu öffnen, die im Betrieb niemand
// braucht.
//
// **Und die Datei bleibt aus.** Eine Meldung, neben der trotzdem eine halbe
// Datei läge, wäre der schlechtere Ausgang (§8).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEinNichtPackbarerAustauschWirdGemeldet(t *testing.T) {
	alt := auskunftPacken

	t.Cleanup(func() { auskunftPacken = alt })

	auskunftPacken = func(any) ([]byte, error) {
		return nil, errors.New("der Austausch ist zu groß für diese Welt")
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

	if antwort := anwendung.EinstellungenAusfuehren(
		t.Context(), nil); antwort.Fehler == nil {
		t.Fatal("der Packfehler ging verloren")
	}

	ziel := filepath.Join(t.TempDir(), "einstellungen.json")

	if antwort := anwendung.EinstellungenSchreiben(
		t.Context(), nil, ziel); antwort.Fehler == nil {
		t.Fatal("der Packfehler ging beim Schreiben verloren")
	}

	if _, err := os.Stat(ziel); !os.IsNotExist(err) {
		t.Errorf("es liegt trotzdem eine Datei: %v", err)
	}

	// **Und die Naht ohne Fenster ebenso.** Sie packt selbst — ein Lauf, der
	// nur den Weg des Fensters prüft, ließe ihren Fehlerzweig unbetreten.
	//
	// **Im Profil**, sonst scheitert schon die Ausfuhr und nicht das Packen:
	// ein Fehler ist dann da, aber der falsche — die Gegenprobe, die nichts
	// ändert (§10). Angestrichen von der Abdeckung am 2026-09-08.
	if _, fehler := EinstellungsnahtUeber(anwendung.db).Ausfuhr(
		anwendung.imProfil(t.Context()), nil); fehler == nil ||
		!strings.Contains(fehler.Error(), "packen") {
		t.Errorf("der Packfehler ging an der Naht verloren: %v", fehler)
	}
}
