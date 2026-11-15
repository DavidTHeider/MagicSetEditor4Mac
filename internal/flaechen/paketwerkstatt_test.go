// Die Paketwerkstatt über jede Fläche
// (#1164).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Eine Paketdatei wird über jede Fläche gelesen und zurückgeschrieben.**
//
// Hier **darf** verglichen werden: der Inhalt einer Datei ist die Antwort
// selbst und nicht ihre Verpackung.
func TestEineVorlageAendernUeberJedeFlaeche(t *testing.T) {
	umgebung := eigenesZuhause(t)
	paket := probepaket(t, umgebung)

	for name, fall := range map[string]struct {
		lesen     func(*testing.T) (inhalt, stand string)
		schreiben func(t *testing.T, inhalt, stand string)
	}{
		"Kommandozeile": {
			lesen: func(t *testing.T) (string, string) {
				t.Helper()

				return ausJSON(t, mussKommandozeile(t, umgebung,
					"paketdatei", paket, "style", "--json"), "inhalt", "stand")
			},
			schreiben: func(t *testing.T, inhalt, stand string) {
				t.Helper()

				quelle := filepath.Join(t.TempDir(), "neu")
				if err := os.WriteFile(quelle, []byte(inhalt), 0o600); err != nil {
					t.Fatalf("schreiben: %v", err)
				}

				mussKommandozeile(t, umgebung,
					"paketdatei-setzen", paket, "style", quelle, stand)
			},
		},
		"MCP": {
			lesen: func(t *testing.T) (string, string) {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "lesen.paketdatei",
					map[string]any{"paket": paket, "datei": "style"})
				if istFehler {
					t.Fatalf("lesen.paketdatei: %s", text)
				}

				return ausJSON(t, text, "Inhalt", "Stand")
			},
			schreiben: func(t *testing.T, inhalt, stand string) {
				t.Helper()

				text, istFehler := mcpRuf(t, umgebung, "schreiben.paketdatei",
					map[string]any{
						"paket": paket, "datei": "style",
						"inhalt": inhalt, "stand": stand,
					})
				if istFehler {
					t.Fatalf("schreiben.paketdatei: %s", text)
				}
			},
		},
		"REST": {
			lesen: func(t *testing.T) (string, string) {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodGet,
					"/pakete/"+paket+"/dateien/style", "")
				if status != http.StatusOK {
					t.Fatalf("GET: %d (%s)", status, rumpf)
				}

				return ausJSON(t, rumpf, "inhalt", "stand")
			},
			schreiben: func(t *testing.T, inhalt, stand string) {
				t.Helper()

				an := starteDienst(t, umgebung)

				status, rumpf := an.ruf(t, http.MethodPut,
					"/pakete/"+paket+"/dateien/style",
					`{"inhalt":`+jsonText(t, inhalt)+`,"stand":"`+stand+`"}`)
				if status != http.StatusOK {
					t.Fatalf("PUT: %d (%s)", status, rumpf)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			inhalt, stand := fall.lesen(t)

			if !strings.Contains(inhalt, "mse version") {
				t.Fatalf("gelesen: %q", inhalt)
			}

			fall.schreiben(t, inhalt+"# "+name+"\n", stand)

			nachher, _ := fall.lesen(t)
			if !strings.Contains(nachher, "# "+name) {
				t.Errorf("die Änderung steht nicht da: %q", nachher)
			}
		})
	}
}

// **Und ein veralteter Stand wird überall abgewiesen.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn keine
// Fläche den Stand überhaupt beachtete.
func TestEinVeralteterStandWirdUeberallAbgewiesen(t *testing.T) {
	umgebung := eigenesZuhause(t)
	paket := probepaket(t, umgebung)

	quelle := filepath.Join(t.TempDir(), "neu")
	if err := os.WriteFile(quelle, []byte("x\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		aus, fehler, code := kommandozeile(t, umgebung,
			"paketdatei-setzen", paket, "style", quelle, "veraltet")
		if code == 0 {
			t.Fatalf("der veraltete Stand kam durch: %s", aus)
		}

		if !strings.Contains(aus+fehler, "geändert") {
			t.Errorf("die Meldung nennt den Grund nicht: %s%s", aus, fehler)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "schreiben.paketdatei",
			map[string]any{
				"paket": paket, "datei": "style", "inhalt": "x\n", "stand": "veraltet",
			})
		if !istFehler {
			t.Fatalf("der veraltete Stand kam durch: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPut, "/pakete/"+paket+"/dateien/style",
			`{"inhalt":"x\n","stand":"veraltet"}`)
		if status == http.StatusOK {
			t.Fatalf("der veraltete Stand kam durch: %s", rumpf)
		}
	})
}

// eigenesZuhause gibt eine Umgebung mit eigenem Bestand **und** eigenem
// Zuhause.
//
// # Warum beides
//
// `app.Paketverzeichnis()` hängt an `os.UserConfigDir()` und damit an `HOME` —
// nicht an `MSE_BESTAND`, denn Pakete gehören dem Rechner. Ein Lauf ohne diese
// Zeile legte ein Paket im **echten** Paketverzeichnis dieses Rechners an und
// schriebe darin. §3 nennt genau diesen Fall: „ein Test, der nach draußen
// greift, ohne dass jemand es wollte" — und er wäre grün, weil er wirklich
// funktioniert hat.
//
// **Und der Bestand liegt darin**, wo er ohne Umgebung auch läge: seit
// QA-10.NAT-01 (#1741) folgen die Bilder dem Bestand und die Pakete dem
// Zuhause — so nennen beide denselben Anwendungsort, und ein Test, der das
// eine oder das andere dort sucht, sucht, wo es liegt.
func eigenesZuhause(t *testing.T) []string {
	t.Helper()

	zuhause := t.TempDir()

	return append(eigenerBestand(t), "HOME="+zuhause,
		"MSE_BESTAND="+filepath.Join(zuhause, "Library", "Application Support",
			"MagicSetEditor4Mac", "bestand.db"))
}

// probepaket legt ein Paket im Paketverzeichnis dieser Umgebung an.
func probepaket(t *testing.T, umgebung []string) string {
	t.Helper()

	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	if zuhause == "" {
		t.Fatal("die Umgebung nennt kein Zuhause")
	}

	// **Derselbe Weg wie `app.Paketverzeichnis()`**, von Hand nachgebaut: die
	// Flächen laufen als eigene Prozesse, und dieser Lauf kann ihnen den Ort
	// nicht abfragen.
	name := "probe.mse-style"
	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", name)

	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "style"),
		[]byte("mse version: 2.0.0\ngame: probe\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return name
}

// ausJSON liest zwei Felder aus einer Antwort.
func ausJSON(t *testing.T, roh, erstes, zweites string) (eins, zwei string) {
	t.Helper()

	var felder map[string]string

	if err := json.Unmarshal([]byte(roh), &felder); err != nil {
		t.Fatalf("kein JSON: %v\n%s", err, roh)
	}

	return felder[erstes], felder[zweites]
}

// jsonText gibt eine Zeichenkette als JSON-Text.
func jsonText(t *testing.T, wert string) string {
	t.Helper()

	roh, err := json.Marshal(wert)
	if err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	return string(roh)
}
