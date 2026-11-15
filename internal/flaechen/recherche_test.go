package flaechen_test

// **Die Recherche über Kommandozeile, MCP und Dienst** — Werkstück A16 von
// Phase 5.3
// (#1515).
//
// # Was er hält
//
// **Dieselbe Antwort an drei Flächen** — und vor allem: **denselben Grund**,
// wenn nichts kommt. Das Issue verlangt es als eigene Zeile: „Derselbe
// fachliche Aufruf auf vier Flächen liefert dieselben Identitäten, Filter,
// Grenzen und Fehler."
//
// **Und dieselbe Absage.** Ein Recherchefall, den es nicht gibt, ist an jeder
// Fläche eine fachliche Abweisung — nicht an einer ein Fehler und an der
// nächsten eine leere Liste.
//
// # Was er nicht hält
//
// **Das Fenster** — dieser Lauf startet Prozesse.
//
// **Und eine Antwort mit Treffern.** Dafür braucht es einen gebauten Index,
// und der braucht eine laufende Einbettungslaufzeit; beides steht hinter
// `netzlauf` und nicht in der Pipeline. Was hier steht, ist die **Gleichheit
// der Flächen** — die Zusage aus §16.

import (
	"net/http"
	"strings"
	"testing"
)

// derKontextweg ist das Muster, unter dem das Kontextpaket angemeldet ist.
const derKontextweg = "POST /kontextpaket"

func TestDieRechercheGehtUeberDreiFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		aus, klage, code := kommandozeile(t, umgebung, "recherche", "regeln",
			"trampelschaden")

		switch {
		case code != 0:
			t.Fatalf("recherche gab %d (%s)", code, klage)
		case !strings.Contains(aus, "ohne-index"):
			t.Errorf("Ausgabe %q nennt den Grund nicht", aus)
		}

		// **Ein Fall, den es nicht gibt, wird abgewiesen** — und nicht
		// stillschweigend zum Kartenfall gemacht.
		if _, _, abgewiesen := kommandozeile(t, umgebung, "recherche",
			"erfunden", "trampelschaden"); abgewiesen == 0 {
			t.Error("der erfundene Fall kam durch")
		}

		// **Eine leere Frage ebenso.**
		if _, _, abgewiesen := kommandozeile(t, umgebung, "recherche",
			"regeln", ""); abgewiesen == 0 {
			t.Error("die leere Frage kam durch")
		}

		// **Eine Ähnlichkeitsfrage ohne Index liest die Karte gar nicht.**
		//
		// **Und das ist richtig**: wo nichts zu durchsuchen ist, gibt es
		// auch nichts, wozu eine Ausgangskarte gehörte. Der Grund steht in
		// der Antwort — die fehlende Karte fiele erst auf, wenn ein Index
		// dastünde, und dann fällt sie auf.
		aus, klage, code = kommandozeile(t, umgebung, "aehnlich", "4711")

		switch {
		case code != 0:
			t.Fatalf("aehnlich gab %d (%s)", code, klage)
		case !strings.Contains(aus, "ohne-index"):
			t.Errorf("Ausgabe %q nennt den Grund nicht", aus)
		}

		// **Eine Karte, die keine Zahl ist, wird abgewiesen.**
		if _, _, abgewiesen := kommandozeile(t, umgebung, "aehnlich",
			"keinezahl"); abgewiesen == 0 {
			t.Error("die Karte ohne Zahl kam durch")
		}

		// **Das Kontextpaket entsteht auch ohne Index** — und sagt, dass es
		// leer ist. Ein Fehler an dieser Stelle nähme dem Aufrufer den Grund.
		aus, klage, code = kommandozeile(t, umgebung, "kontextpaket",
			"Erkläre Trampelschaden", "regeln", "trampelschaden", "ollama",
			"gemma3", "8192", "1024")

		switch {
		case code != 0:
			t.Fatalf("kontextpaket gab %d (%s)", code, klage)
		case !strings.Contains(aus, "Keine Belege"):
			t.Errorf("Ausgabe %q sagt nicht, dass nichts zusammenkam", aus)
		}

		// **Ohne Aufgabe gibt es kein Paket.**
		if _, _, abgewiesen := kommandozeile(t, umgebung, "kontextpaket", "",
			"regeln", "trampelschaden", "ollama", "gemma3", "8192",
			"1024"); abgewiesen == 0 {
			t.Error("das Paket ohne Aufgabe kam durch")
		}

		// **Und ohne Budget ebenso** — Fenster und Reserve gehören zum
		// Zielmodell, und eine Vorgabe hier gälte für keines.
		if _, _, abgewiesen := kommandozeile(t, umgebung, "kontextpaket", "a",
			"regeln", "b", "ollama", "gemma3", "0", "0"); abgewiesen == 0 {
			t.Error("das Paket ohne Budget kam durch")
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		if antwort := mussMCP(t, umgebung, "lesen.recherche",
			map[string]any{"fall": "regeln", "frage": "trampelschaden"},
		); !strings.Contains(antwort, `"leere": "ohne-index"`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Ohne Fall gilt der Kartenfall** — und nicht der
		// Ähnlichkeitsfall, der eine Ausgangskarte bräuchte, die niemand
		// genannt hat.
		if antwort := mussMCP(t, umgebung, "lesen.recherche",
			map[string]any{"frage": "trampelschaden"},
		); !strings.Contains(antwort, `"modus": "ohne-index"`) {
			t.Errorf("Antwort %q", antwort)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		an := starteDienst(t, deckumgebung(t))

		if antwort := mussREST(t, an, http.MethodGet,
			"/recherche?fall=regeln&frage=trampelschaden", ""); !strings.Contains(
			antwort, `"leere":"ohne-index"`) {
			t.Errorf("Antwort %q", antwort)
		}

		// **Gefahren wird das Muster, unter dem der Weg angemeldet ist**
		// (`internal/rest/betriebswege.go`) — wer Methode und Pfad getrennt
		// hinschriebe, prüfte einen Weg, den so niemand angemeldet hat.
		methode, weg, _ := strings.Cut(derKontextweg, " ")

		if antwort := mussREST(t, an, methode, weg,
			`{"aufgabe":"Erkläre Trampelschaden","fall":"regeln",`+
				`"frage":"trampelschaden","grenze":10,"anbieter":"ollama",`+
				`"modell":"gemma3","kontextgrenze":8192,`+
				`"ausgabereserve":1024}`); !strings.Contains(
			antwort, `"belege":[]`) {
			t.Errorf("Antwort %q", antwort)
		}
	})
}

// **Das Kontextpaket gibt es über MCP nicht — mit Absicht.**
//
// **Die Gegenprobe zum Nein** (§10): der Lauf sieht nach, dass die beiden
// lesenden Werkzeuge **da** sind. Sonst prüfte er eine Werkzeugliste, die er
// gar nicht erreicht, und das Nein wäre keine Aussage.
func TestUeberMCPGibtEsKeinKontextpaket(t *testing.T) {
	t.Parallel()

	werkzeuge := map[string]bool{}
	for _, name := range angemeldeteWerkzeuge(t, deckumgebung(t)) {
		werkzeuge[name] = true
	}

	for _, name := range []string{
		"lesen." + "recherche", "lesen." + "aehnlichekarten",
	} {
		if !werkzeuge[name] {
			t.Errorf("%s fehlt — dann sagt der Lauf nichts über das Nein", name)
		}
	}

	// **Gesucht wird die Familie und nicht das Wort** — zwei Teile und keine
	// Zeichenkette, die wie ein Werkzeugname aussieht (siehe
	// `suchindex_test.go`).
	for name := range werkzeuge {
		if strings.Contains(name, "kontext"+"paket") {
			t.Errorf("%s stellt über MCP ein Kontextpaket zusammen", name)
		}
	}
}
