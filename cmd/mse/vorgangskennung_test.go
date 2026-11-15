package main

// **Der Griff ins Protokoll auf der Kommandozeile** — Werkstück O9 von
// Phase 5.2 (#1433).

import (
	"regexp"
	"strings"
	"testing"
)

// **Ein fachlicher Fehlschlag nennt seinen Vorgang.** Wer ihn hat, sucht damit
// in `mse.log` — der Satz allein sagt nicht, was der Unterbau gemeldet hat.
func TestEinFachlicherFehlschlagNenntSeinenVorgang(t *testing.T) {
	code, _, fehler := fahre(t, "karten", "999")

	if code != CodeFachlich {
		t.Fatalf("Rückgabewert %d, erwartet war %d (%s)", code, CodeFachlich, fehler)
	}

	kennung := regexp.MustCompile(`Vorgang: ([0-9a-f]{8})\n`).FindStringSubmatch(fehler)
	if kennung == nil {
		t.Fatalf("keine Vorgangskennung in %q", fehler)
	}

	// **Und zweimal ist zweimal etwas anderes** — sonst wäre sie kein Griff
	// auf **einen** Ruf.
	_, _, zweiter := fahre(t, "karten", "999")
	if strings.Contains(zweiter, kennung[1]) {
		t.Errorf("zwei Läufe teilen die Kennung %q", kennung[1])
	}
}

// **Ein Bedienfehler nennt keinen.** Der Satz darüber sagt schon alles, und
// im Protokoll steht dazu nichts.
func TestEinBedienfehlerNenntKeinenVorgang(t *testing.T) {
	code, _, fehler := fahre(t, "gibtsnicht")

	if code != CodeBedienung {
		t.Fatalf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}

	if strings.Contains(fehler, "Vorgang:") {
		t.Errorf("der Bedienfehler nennt einen Vorgang: %q", fehler)
	}
}

// **Und ein geglückter Lauf auch nicht** — die Zeile wäre Rauschen in jeder
// Ausgabe, die jemand weiterverarbeitet.
func TestEinGeglueckterLaufNenntKeinenVorgang(t *testing.T) {
	code, aus, fehler := fahre(t, "sets")

	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if strings.Contains(aus, "Vorgang:") || strings.Contains(fehler, "Vorgang:") {
		t.Errorf("der geglückte Lauf nennt einen Vorgang: %q / %q", aus, fehler)
	}
}
