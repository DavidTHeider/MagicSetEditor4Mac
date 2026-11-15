package flaechen_test

// **Die Rückkehren über Kommandozeile, MCP und Dienst** — Werkstück O7 von
// Phase 5.2 (#1431).
//
// # Was er hält
//
// **Dieselbe Aussage in drei Gestalten** — und die vierte Fähigkeit, die
// **eine** Fläche bewusst nicht hat: `RueckkehrQuittieren` gibt es über MCP
// nicht, und der Lauf sieht nach, dass das Werkzeug wirklich fehlt.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse. Die Bindungen prüft
// `internal/app/ruecksprung_test.go`.
//
// **Und eine Rückkehr, die wirklich stattgefunden hat.** Ein frischer Bestand
// hat keine; dass eine Rückkehr ihre Epoche hinterlässt, steht in
// `internal/service/ruecksprung_intern_test.go` und in `internal/storage`.
// Hier steht, dass **jede Fläche dieselbe leere Antwort** gibt — und das ist
// die Zusage aus §16.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDerRuecksprungGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "ruecksprung-arten")
		if code != 0 {
			t.Fatalf("ruecksprung-arten gab %d (%s)", code, klage)
		}

		// **Vier Arten, und zwei davon mit einem Hindernis.**
		for _, erwartet := range []string{
			"appfassung\tnachProzessende", "schema\tnurVorwaerts",
			"daten\t", "pakete\t",
		} {
			if !strings.Contains(aus, erwartet) {
				t.Errorf("Ausgabe %q nennt %q nicht", aus, erwartet)
			}
		}

		leer, klage, code := kommandozeile(t, umgebung, "rueckkehren")
		switch {
		case code != 0:
			t.Fatalf("rueckkehren gab %d (%s)", code, klage)
		case strings.TrimSpace(leer) != "":
			t.Errorf("ein frischer Bestand nennt Rückkehren: %q", leer)
		}

		// **Und eine Kennung, die es nicht gibt, wird abgewiesen** — statt
		// still zu glücken.
		_, _, code = kommandozeile(t, umgebung, "rueckkehr-quittieren", "4711")
		if code == 0 {
			t.Error("eine unbekannte Rückkehr ließ sich quittieren")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		if antwort := mussMCP(t, umgebung, "lesen.ruecksprungarten",
			map[string]any{}); !strings.Contains(antwort, `"nurVorwaerts"`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussMCP(t, umgebung, "lesen.rueckkehren",
			map[string]any{}); !strings.Contains(antwort, "[]") {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/ruecksprung/arten",
			""); !strings.Contains(antwort, `"nurVorwaerts"`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodGet, "/rueckkehren",
			""); !strings.Contains(antwort, `"rueckkehren":[]`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Und eine Kennung, die es nicht gibt, wird abgewiesen.**
		if status, rumpf := an.ruf(t, http.MethodPost, "/rueckkehren/4711/quittung",
			""); status < 400 {
			t.Errorf("eine unbekannte Rückkehr gab %d (%s)", status, rumpf)
		}
	})
}

// **Und über MCP gibt es die Quittung nicht** — mit Absicht.
//
// Eine Quittung sagt „ich habe nachgesehen", und nachsehen heißt hier: in die
// Kiste greifen und zählen. Ein Modell kann das nicht. **Die Gegenprobe zum
// Nein**: der Lauf sieht nach, dass die drei lesenden Werkzeuge daneben
// **da** sind — sonst prüfte er eine Werkzeugliste, die er nicht erreicht
// (§10).
func TestDieQuittungGibtEsUeberMCPNicht(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	for _, da := range []string{
		"lesen.ruecksprungarten", "lesen.ruecksprunglage", "lesen.rueckkehren",
	} {
		if !werkzeuge[da] {
			t.Errorf("das Werkzeug %q fehlt — dann sagt der Lauf nichts über das Nein", da)
		}
	}

	// **Über die Gestalt und nicht über einen Namen.** Ein Name, den es nicht
	// geben soll, stünde hier als Zeichenkette — und der Wächter über die
	// Flächenläufe hielte ihn für ein Werkzeug, das im Katalog fehlt. Gesucht
	// wird deshalb nach dem Wortstamm: **kein** Werkzeug quittiert.
	for name := range werkzeuge {
		if strings.Contains(name, "quittier") {
			t.Errorf("das Werkzeug %q gibt es — eine Quittung ist die Aussage eines Menschen", name)
		}
	}
}
