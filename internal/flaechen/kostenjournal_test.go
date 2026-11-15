package flaechen_test

// **Das Kostenjournal über Kommandozeile, MCP und Dienst** — Werkstück O12
// von Phase 5.2
// (#1439).
//
// # Was er hält
//
// **Dieselben vier Zahlen in drei Gestalten** — und das vierte Nein dieser
// Phase: über MCP wird der Stand gelesen und keine Grenze gesetzt.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse.
//
// **Und einen Auftrag, der wirklich Geld kostet.** Es gibt keinen: Scryfall,
// GitHub und die drei Kartendienste sind kostenlos. Was hier läuft, ist die
// Buchführung — dass sie trägt, steht in `internal/storage` gegen eine echte
// Datenbank.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDasKostenjournalGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "kosten", "anbieter")

		switch {
		case code != 0:
			t.Fatalf("kosten gab %d (%s)", code, klage)
		case !strings.Contains(aus, "verfuegbar\t"):
			t.Errorf("Ausgabe %q nennt das Verfügbare nicht", aus)
		}

		if _, _, gesetzt := kommandozeile(t, umgebung,
			"kostenbudget", "anbieter", "tag", "50000", "EUR"); gesetzt != 0 {
			t.Fatalf("kostenbudget gab %d", gesetzt)
		}

		aus, _, code = kommandozeile(t, umgebung, "kosten", "anbieter")
		if code != 0 || !strings.Contains(aus, "verfuegbar\t50000\tEUR") {
			t.Errorf("nach dem Setzen: %d\n%s", code, aus)
		}

		// **Und die Posten kommen leer** — `mse kostenposten` sagt nichts,
		// weil nichts dasteht, und das ist eine Antwort.
		aus, _, code = kommandozeile(t, umgebung, "kostenposten", "anbieter")
		if code != 0 || strings.TrimSpace(aus) != "" {
			t.Errorf("die Posten: %d %q", code, aus)
		}

		// **Ein erfundener Zeitraum wird abgewiesen.**
		if _, _, erfunden := kommandozeile(t, umgebung,
			"kosten", "anbieter", "stunde"); erfunden == 0 {
			t.Error("ein erfundener Zeitraum kam durch")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		if antwort := mussMCP(t, deckumgebung(t), "lesen.kostenstand",
			map[string]any{"konto": "anbieter"}); !strings.Contains(
			antwort, `"verfuegbar"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodPut, "/kosten/anbieter/budget",
			`{"zeitraum":"tag","wert":50000,"waehrung":"EUR"}`); !strings.Contains(
			antwort, `"fehler":null`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodGet, "/kosten/anbieter",
			""); !strings.Contains(antwort, `"verfuegbar":{"wert":50000,"waehrung":"EUR"}`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodGet, "/kosten/anbieter/posten",
			""); !strings.Contains(antwort, `"posten":[]`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}

// **Über MCP wird keine Grenze gesetzt** — mit Absicht.
//
// **Die Gegenprobe zum Nein**: der Lauf sieht nach, dass das lesende Werkzeug
// **da** ist — sonst prüfte er eine Werkzeugliste, die er nicht erreicht (§10).
func TestUeberMCPWirdKeineGrenzeGesetzt(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	if !werkzeuge["lesen.kostenstand"] {
		t.Error("das lesende Werkzeug fehlt — dann sagt der Lauf nichts über das Nein")
	}

	for name := range werkzeuge {
		if strings.HasPrefix(name, "schreiben.") &&
			(strings.Contains(name, "kosten") || strings.Contains(name, "budget")) {
			t.Errorf("das Werkzeug %q fasst Geld an — das gehört an eine Fläche "+
				"mit einem Menschen", name)
		}
	}
}
