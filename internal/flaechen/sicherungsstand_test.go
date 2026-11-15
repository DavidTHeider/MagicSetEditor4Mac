// **Wann der Bestand zuletzt gesichert wurde, an jeder Fläche** — OPS01
// (#2347).
//
// # Die Zusage
//
// **Eine Sicherung über jede Fläche setzt die Zeit**, und jede Fläche nennt
// danach dasselbe: wann, über welche Fläche, welche Datei. Vorher heißt es
// überall „unbekannt" — nie „heute", ohne dass eine Datei zurückgelesen
// wurde.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

// sicherungsstandUeberAlle liest den Stand über Kommandozeile, MCP, Dienst
// und Fenster, vergleicht die vier und gibt ihn.
func sicherungsstandUeberAlle(t *testing.T, umgebung []string, an dienst) map[string]any {
	t.Helper()

	ausCLI := mussKommandozeile(t, umgebung, "--json", "bestand-sicherungsstand")

	ausMCP, istFehler := mcpRuf(t, umgebung, "lesen.bestand.sicherungsstand", map[string]any{})
	if istFehler {
		t.Fatalf("lesen.bestand.sicherungsstand: %s", ausMCP)
	}

	status, ausREST := an.ruf(t, http.MethodGet, "/bestand/sicherungsstand", "")
	if status != http.StatusOK {
		t.Fatalf("GET bestand/sicherungsstand: %d (%s)", status, ausREST)
	}

	antwort := fensterAus(t, umgebung).Sicherungsstand(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Sicherungsstand am Fenster: %+v", antwort.Fehler)
	}

	ausFenster, err := json.Marshal(antwort.Stand)
	if err != nil {
		t.Fatal(err)
	}

	gleicheAuskunft(t, "der Sicherungsstand — Zeile gegen Modell", ausCLI, ausMCP)
	gleicheAuskunft(t, "der Sicherungsstand — Modell gegen Dienst", ausMCP, ausREST)
	gleicheAuskunft(t, "der Sicherungsstand — Dienst gegen Fenster", ausREST, string(ausFenster))

	var stand map[string]any
	if err := json.Unmarshal([]byte(ausCLI), &stand); err != nil {
		t.Fatal(err)
	}

	return stand
}

func TestJedeFlaecheSetztUndNenntDasSicherungsalter(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	eingefahren(t, umgebung)

	an := starteDienst(t, umgebung)

	if stand := sicherungsstandUeberAlle(t, umgebung, an); stand["stufe"] != "unbekannt" || stand["gesichert"] != "" {
		t.Errorf("vor der ersten Sicherung: %v — unbekannt heißt unbekannt", stand)
	}

	ziel := t.TempDir()

	for flaeche, sichere := range map[string]func(string){
		"cli": func(pfad string) { mussKommandozeile(t, umgebung, "bestand-sichern", pfad) },
		"mcp": func(pfad string) {
			if text, istFehler := mcpRuf(t, umgebung, "schreiben.bestand.sichern", map[string]any{"ziel": pfad}); istFehler {
				t.Fatalf("schreiben.bestand.sichern: %s", text)
			}
		},
		"rest": func(pfad string) {
			if status, rumpf := an.ruf(t, http.MethodPost, "/bestand/sicherung", `{"ziel":"`+pfad+`"}`); status != http.StatusCreated {
				t.Fatalf("POST bestand/sicherung: %d (%s)", status, rumpf)
			}
		},
		"fenster": func(pfad string) {
			if antwort := fensterAus(t, umgebung, pfad).BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
				t.Fatalf("BestandSichern: %+v", antwort.Fehler)
			}
		},
	} {
		datei := flaeche + ".msesicherung"
		sichere(filepath.Join(ziel, datei))

		stand := sicherungsstandUeberAlle(t, umgebung, an)
		gesichert, _ := stand["gesichert"].(string) //nolint:errcheck // kein Text ist rot, eine Zeile tiefer
		if stand["stufe"] != "frisch" || stand["flaeche"] != flaeche || stand["datei"] != datei ||
			stand["anlass"] != "hand" || !strings.HasPrefix(gesichert, "20") {
			t.Errorf("nach der Sicherung über %s: %v", flaeche, stand)
		}
	}
}
