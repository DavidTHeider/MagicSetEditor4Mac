package flaechen_test

// **Die abgestufte Rücksetzung über Kommandozeile, MCP und Dienst** —
// Werkstück O4 von Phase 5.2
// (#1428).
//
// # Was er hält
//
// **Dieselbe Aussage in drei Gestalten** — und das dritte Nein, das dieselbe
// Klasse hat wie bei der Datenpflege: über MCP wird nicht zurückgesetzt.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse.
//
// **Und eine Löschung, die wirklich etwas wegnimmt.** Gefahren wird die
// **Layoutrücksetzung**: sie ist die einzige der sechs, die nichts wegnimmt,
// was ein zweiter Lauf bräuchte. Dass ein Profil samt Karten verschwindet,
// steht in `internal/service` und `internal/app`; hier steht, dass jede
// Fläche **dieselbe** Antwort gibt — und das ist die Zusage aus §16.

import (
	"net/http"
	"strings"
	"testing"
)

func TestDieRuecksetzungGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "ruecksetzungsarten")

		switch {
		case code != 0:
			t.Fatalf("ruecksetzungsarten gab %d (%s)", code, klage)
		case !strings.Contains(aus, "anwendung"):
			t.Errorf("Ausgabe %q nennt die schwerste Art nicht", aus)
		}

		aus, klage, code = kommandozeile(t, umgebung, "ruecksetzung-vorschau", "layout")

		switch {
		case code != 0:
			t.Fatalf("ruecksetzung-vorschau gab %d (%s)", code, klage)
		case !strings.Contains(aus, "bereit\ttrue"):
			t.Errorf("Ausgabe %q sagt nicht, dass es ginge", aus)
		}

		// **Eine erfundene Art wird abgewiesen.**
		if _, _, code := kommandozeile(t, umgebung,
			"ruecksetzung-vorschau", "alles"); code == 0 {
			t.Error("eine erfundene Art kam durch")
		}

		// **Und die ganze Anwendung ohne Nachweis ebenso** — dieselbe Absage
		// wie an jeder anderen Fläche.
		if _, _, code := kommandozeile(t, umgebung,
			"zuruecksetzen", "anwendung", "alles weg"); code == 0 {
			t.Error("die ganze Anwendung ging ohne Nachweis durch")
		}

		// **Das Layout geht** — und nimmt nichts weg, was ein zweiter Lauf
		// bräuchte.
		if aus, klage, code := kommandozeile(t, umgebung,
			"zuruecksetzen", "layout", ""); code != 0 {
			t.Errorf("zuruecksetzen gab %d (%s / %s)", code, klage, aus)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		if antwort := mussMCP(t, deckumgebung(t), "lesen.ruecksetzungsvorschau",
			map[string]any{"art": "layout"}); !strings.Contains(antwort, `"betroffen"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet, "/ruecksetzung/arten",
			""); !strings.Contains(antwort, `"anwendung"`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodGet, "/ruecksetzung/layout/vorschau",
			""); !strings.Contains(antwort, `"bereit":true`) {
			t.Errorf("Antwort %q", antwort)
		}

		if antwort := mussREST(t, an, http.MethodPost, "/ruecksetzung/layout",
			`{"name":""}`); !strings.Contains(antwort, `"zurueckgesetzt"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}

// **Über MCP wird nicht zurückgesetzt** — mit Absicht.
//
// **Die Gegenprobe zum Nein**: der Lauf sieht nach, dass das lesende Werkzeug
// **da** ist — sonst prüfte er eine Werkzeugliste, die er nicht erreicht (§10).
func TestUeberMCPWirdNichtZurueckgesetzt(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	if !werkzeuge["lesen.ruecksetzungsvorschau"] {
		t.Error("das lesende Werkzeug fehlt — dann sagt der Lauf nichts über das Nein")
	}

	// **Gesucht wird die Familie und nicht das Wort.**
	// `schreiben.stiloption.zuruecksetzen` und
	// `schreiben.autoersetzungen.zuruecksetzen` nehmen eine **Wahl** zurück
	// und löschen nichts; sie heißen nur ähnlich. Wer auf „ruecksetz" prüfte,
	// striche sie mit an und hätte einen Wächter, der eine andere Frage
	// beantwortet als die, die über ihm steht (§10).
	// **Zwei Teile und keine Zeichenkette**, die wie ein Werkzeugname
	// aussieht: der Wächter über die Flächenläufe liest sie sonst als Ruf und
	// verlangt eine Katalogzeile für ein Werkzeug, das es nicht geben soll.
	for name := range werkzeuge {
		if strings.HasPrefix(name, "schreiben.") &&
			strings.Contains(name, "ruecksetzung") {
			t.Errorf("das Werkzeug %q setzt zurück — das gehört an eine Fläche "+
				"mit einem Menschen", name)
		}
	}
}
