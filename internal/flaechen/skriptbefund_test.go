// **Die Skriptprüfung einer Paketdatei über jede Fläche** — Werkstück B von
// Phase 4.9
// (#1324).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Warum diese Fähigkeit über alle Flächen geht
//
// Für die **Einfärbung** steht in `docs/FAEHIGKEITEN.md` „ohne Fenster wird
// nicht getippt" — die Merkmale sind die Darstellung eines Textfelds. Für
// eine **Prüfung** trägt das nicht: sie ist eine Aussage über die Datei. Wer
// eine Vorlage in einem Ablauf ändert, will wissen, ob sie noch zerteilt,
// bevor er sie ausliefert.
package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// schreibePaketdatei setzt den Inhalt der `style`-Datei des Probepakets.
//
// **Von Hand und nicht über die Bindung**: die Vorbereitung eines Laufs soll
// nicht durch dieselbe Naht gehen, die er prüft — sonst prüfte ein grüner
// Lauf zwei Dinge zugleich und sagte über keines etwas.
func schreibePaketdatei(t *testing.T, umgebung []string, paket, inhalt string) {
	t.Helper()

	zuhause := ""

	for _, eintrag := range umgebung {
		if wert, gefunden := strings.CutPrefix(eintrag, "HOME="); gefunden {
			zuhause = wert
		}
	}

	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete", paket, "style")

	if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// **Derselbe Text, dieselbe Stelle — über Kommandozeile, MCP und REST.**
//
// Die Datei trägt einen Fehler in Zeile 2: `script: f(1, 2` schließt seine
// Klammer nicht. Gemeldet wird seit Werkstück B die **öffnende** Klammer, und
// die steht in Spalte 10 — `script:` ist sieben Zeichen, dann `:`, dann ein
// Leerzeichen, dann `f`, dann `(`.
func TestEinenSkriptfehlerFindenUeberJedeFlaeche(t *testing.T) {
	umgebung := eigenesZuhause(t)
	paket := probepaket(t, umgebung)

	kaputt := "mse version: 0.3.8\nscript: f(1, 2\n"

	for name, pruefe := range map[string]func(*testing.T, string) string{
		"Kommandozeile": func(t *testing.T, inhalt string) string {
			t.Helper()

			// **Die Kommandozeile prüft eine installierte Datei.** Sie hat
			// keinen Weg, einen Text entgegenzunehmen — und das ist richtig
			// so: an einer Kommandozeile ist eine Datei das natürliche
			// Argument. Also wird sie zuerst geschrieben.
			schreibePaketdatei(t, umgebung, paket, inhalt)

			aus, fehler, code := kommandozeile(t, umgebung,
				"paketdatei-pruefen", paket, "style")
			if code == 0 {
				t.Fatalf("der Fehler kam nicht durch: %s%s", aus, fehler)
			}

			return aus + fehler
		},
		"MCP": func(t *testing.T, inhalt string) string {
			t.Helper()

			text, istFehler := mcpRuf(t, umgebung, "lesen.skriptbefund",
				map[string]any{"inhalt": inhalt})
			if istFehler {
				t.Fatalf("lesen.skriptbefund: %s", text)
			}

			return text
		},
		"REST": func(t *testing.T, inhalt string) string {
			t.Helper()

			an := starteDienst(t, umgebung)

			status, rumpf := an.ruf(t, http.MethodPost, "/paketdateien/pruefung",
				`{"inhalt":`+jsonText(t, inhalt)+`}`)
			if status != http.StatusOK {
				t.Fatalf("POST: %d (%s)", status, rumpf)
			}

			return rumpf
		},
	} {
		t.Run(name, func(t *testing.T) {
			bericht := pruefe(t, kaputt)

			// **Über den Text und nicht über die Verpackung.** Die drei
			// Flächen antworten in drei Gestalten — eine Zeile, ein JSON mit
			// großen Anfangsbuchstaben, ein JSON mit kleinen. Was hier
			// zählt, ist die fachliche Aussage: Zeile 2, Spalte 10.
			for _, teil := range []string{"2", "10"} {
				if !strings.Contains(bericht, teil) {
					t.Errorf("der Bericht nennt %q nicht: %s", teil, bericht)
				}
			}
		})
	}
}

// **Und eine heile Datei wird überall durchgelassen.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn eine
// Fläche **jede** Datei beanstandete.
func TestEineHeileVorlageWirdUeberallDurchgelassen(t *testing.T) {
	umgebung := eigenesZuhause(t)
	paket := probepaket(t, umgebung)

	heil := "mse version: 0.3.8\nscript: f(1, 2)\n"

	t.Run("Kommandozeile", func(t *testing.T) {
		schreibePaketdatei(t, umgebung, paket, heil)

		aus, fehler, code := kommandozeile(t, umgebung,
			"paketdatei-pruefen", paket, "style")
		if code != 0 {
			t.Fatalf("die heile Datei wurde beanstandet: %s%s", aus, fehler)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.skriptbefund",
			map[string]any{"inhalt": heil})
		if istFehler {
			t.Fatalf("lesen.skriptbefund: %s", text)
		}

		if !strings.Contains(text, "false") {
			t.Errorf("kein »Gefunden: false«: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/paketdateien/pruefung",
			`{"inhalt":`+jsonText(t, heil)+`}`)
		if status != http.StatusOK {
			t.Fatalf("POST: %d (%s)", status, rumpf)
		}

		if !strings.Contains(rumpf, "false") {
			t.Errorf("kein »gefunden: false«: %s", rumpf)
		}
	})
}
