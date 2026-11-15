package flaechen_test

// **Die Datenpflege über Kommandozeile, MCP und Dienst** — Werkstück O8 von
// Phase 5.2 (#1432).
//
// # Was er hält
//
// **Dieselbe Aussage in drei Gestalten** — und die zweite Fähigkeit, die
// **eine** Fläche bewusst nicht hat: über MCP wird nicht gepflegt, und der
// Lauf sieht nach, dass kein Werkzeug es kann.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse.
//
// **Und eine Pflege, die wirklich etwas wegräumt.** Ein frischer Bestand hat
// keine Waisen; dass eine Maßnahme wegräumt, sichert und nachsieht, steht in
// `internal/storage` und `internal/service`. Hier steht, dass **jede Fläche
// dieselbe leere Antwort** gibt — und das ist die Zusage aus §16.

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// jetztAlsRFC3339 gibt den Augenblick in der Form, die die Pflege erwartet.
//
// **Die echte Uhr und kein festes Datum**: die Frist der Vorschau gilt
// gegenüber **jetzt**, und ein Datum im Quelltext wäre in einer Viertelstunde
// abgelaufen — der Lauf würde grün, bis er es irgendwann nicht mehr ist.
func jetztAlsRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

func TestDieDatenpflegeGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "pflege-vorschau")

		// **Ein sauberer Bestand gibt 0** — und das ist der Fall, für den es
		// den Rückgabewert gibt.
		switch {
		case code != 0:
			t.Fatalf("pflege-vorschau gab %d (%s)", code, klage)
		case !strings.Contains(aus, "gueltigbis\t"):
			t.Errorf("Ausgabe %q nennt keine Frist", aus)
		}

		// **Eine Maßnahme, die nicht art:revision ist, wird abgewiesen.**
		if _, _, code := kommandozeile(t, umgebung, "pflege",
			"2026-09-10T12:00:00Z", "keineRevision"); code == 0 {
			t.Error("eine krumme Maßnahme kam durch")
		}

		// **Und ein Zeitpunkt, der keiner ist, ebenso.**
		if _, _, code := kommandozeile(t, umgebung, "pflege",
			"irgendwann", "kartenwerteOhneKarte:abc"); code == 0 {
			t.Error("ein krummer Zeitpunkt kam durch")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		if antwort := mussMCP(t, deckumgebung(t), "lesen.pflegevorschau",
			map[string]any{}); !strings.Contains(antwort, `"befunde"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/pflege/vorschau",
			""); !strings.Contains(antwort, `"befunde":[]`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Und die Pflege selbst gibt es dort** — mit einer leeren Wahl
		// geschieht nichts, und das ist eine Antwort.
		if antwort := mussREST(t, an, http.MethodPost, "/pflege",
			`{"erstellt":"`+jetztAlsRFC3339()+`","wahl":[]}`); !strings.Contains(
			antwort, `"zeilen":[]`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}

// **Über MCP wird nicht gepflegt** — mit Absicht.
//
// **Die Gegenprobe zum Nein**: der Lauf sieht nach, dass das lesende Werkzeug
// **da** ist — sonst prüfte er eine Werkzeugliste, die er nicht erreicht (§10).
func TestUeberMCPWirdNichtGepflegt(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	if !werkzeuge["lesen.pflegevorschau"] {
		t.Error("das lesende Werkzeug fehlt — dann sagt der Lauf nichts über das Nein")
	}

	for name := range werkzeuge {
		if strings.Contains(name, "pfleg") && name != "lesen.pflegevorschau" &&
			!strings.Contains(name, "paket") {
			t.Errorf("das Werkzeug %q pflegt — das gehört an eine Fläche mit einem Menschen", name)
		}
	}
}
