package flaechen_test

// **Der lokale Suchindex über Kommandozeile, MCP und Dienst** — Werkstück
// A15 von Phase 5.3
// (#1514).
//
// # Was er hält
//
// **Dieselbe Auskunft an drei Flächen** — und dieselbe Absage, wo kein Index
// steht: „kein Suchindex aktiv" ist eine Antwort und kein Fehler, und alle
// drei geben sie.
//
// **Und dass ein Aufbau ohne Laufzeit scheitert, statt etwas Halbes zu
// hinterlassen.** Kein Ollama hört auf dieser Maschine im Lauf; genau das ist
// der Fall, den ein Nutzer als erstes trifft, und die Antwort muss ihn
// benennen.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse.
//
// **Und einen wirklich gebauten Index.** Dafür braucht es eine laufende
// Einbettungslaufzeit und eine Sammeldatei; beides steht hinter `netzlauf`
// (`internal/ollama`) und nicht in der Pipeline. Was hier steht, ist die
// **Gleichheit der Flächen** — die Zusage aus §16.

import (
	"net/http"
	"strings"
	"testing"
)

// derAbbruchweg ist das Muster, unter dem der Abbruch angemeldet ist.
const derAbbruchweg = "DELETE /index/aufbau"

func TestDerSuchindexGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "indexstand")

		switch {
		case code != 0:
			t.Fatalf("indexstand gab %d (%s)", code, klage)
		case !strings.Contains(aus, "Kein Suchindex aktiv"):
			t.Errorf("Ausgabe %q sagt nicht, dass keiner dasteht", aus)
		}

		// **Ohne Index findet die Suche nichts — und sagt es als Modus.**
		aus, klage, code = kommandozeile(t, umgebung, "indexsuche", "lightning")

		switch {
		case code != 0:
			t.Fatalf("indexsuche gab %d (%s)", code, klage)
		case !strings.Contains(aus, "ohne-index"):
			t.Errorf("Ausgabe %q nennt den Modus nicht", aus)
		}

		// **Eine leere Frage wird abgewiesen** — an jeder Fläche dieselbe
		// fachliche Absage.
		if _, _, code := kommandozeile(t, umgebung, "indexsuche", ""); code == 0 {
			t.Error("die leere Frage kam durch")
		}

		// **Und ein Aufbau ohne erreichbare Laufzeit scheitert.**
		if aus, _, code := kommandozeile(t, umgebung, "indexaufbau",
			"gibtesnicht", "/gibt/es/nicht.jsonl.gz", "2026-09-01"); code == 0 {
			t.Errorf("der Aufbau kam ohne Laufzeit durch: %q", aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		// **Über MCP kommt die Antwort eingerückt** — sie ist für ein
		// Modell gedacht und nicht für eine Leitung.
		if antwort := mussMCP(t, umgebung, "lesen.indexstand",
			map[string]any{}); !strings.Contains(antwort, `"aktiv": false`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussMCP(t, umgebung, "lesen.indexsuche",
			map[string]any{"frage": "lightning"}); !strings.Contains(
			antwort, `"modus": "ohne-index"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/index",
			""); !strings.Contains(antwort, `"aktiv":false`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodGet,
			"/index/suche?frage=lightning", ""); !strings.Contains(
			antwort, `"modus":"ohne-index"`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Der Abbruch sagt, dass nichts lief** — und ist kein Fehler.
		//
		// **Gefahren wird das Muster, unter dem der Weg angemeldet ist.**
		// `http.ServeMux` nimmt „Methode Pfad" als **ein** Muster entgegen
		// (`internal/rest/betriebswege.go`); hier steht dasselbe Wort und
		// wird für den Ruf zerlegt. Wer Methode und Pfad getrennt hinschriebe,
		// prüfte einen Weg, den so niemand angemeldet hat.
		methode, weg, _ := strings.Cut(derAbbruchweg, " ")

		if antwort := mussREST(t, an, methode, weg,
			""); !strings.Contains(antwort, `"abgebrochen":false`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}

// **Über MCP wird nicht aufgebaut** — mit Absicht.
//
// **Die Gegenprobe zum Nein**: der Lauf sieht nach, dass die beiden lesenden
// Werkzeuge **da** sind — sonst prüfte er eine Werkzeugliste, die er nicht
// erreicht (§10).
func TestUeberMCPWirdKeinIndexGebaut(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	for _, name := range []string{"lesen." + "indexstand", "lesen." + "indexsuche"} {
		if !werkzeuge[name] {
			t.Errorf("%s fehlt — dann sagt der Lauf nichts über das Nein", name)
		}
	}

	// **Gesucht wird die Familie und nicht das Wort.** Zwei Teile und keine
	// Zeichenkette, die wie ein Werkzeugname aussieht: der Wächter über die
	// Flächenläufe liest sie sonst als Ruf und verlangt eine Katalogzeile für
	// ein Werkzeug, das es nicht geben soll.
	for name := range werkzeuge {
		if strings.HasPrefix(name, "schreiben."+"index") {
			t.Errorf("%s baut über MCP einen Index", name)
		}
	}
}
