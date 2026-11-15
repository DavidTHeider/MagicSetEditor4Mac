// **Der Probelauf an jeder Fläche** — UXR-F7
// (#2015).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier gefahren wird
//
// Derselbe Ablauf über Kommandozeile, MCP und REST: ein Probelauf, der sagt,
// was der Druckbogen und die Exportvorlage ergäben — und danach der echte Lauf
// über **dieselbe** Fläche, dessen Ergebnis der Probelauf vorhergesagt haben
// muss. Dazu die fachlichen Fehler: der Probelauf gibt sie an jeder Fläche
// ebenso wie der Lauf, und er hinterlässt nichts auf der Platte.
//
// # Was sie nicht halten
//
// Das Fenster fährt er nicht (gemessen in `FLAECHENPARITAET.md`); seinen
// Probelauf prüfen die Läufe in `internal/app` und die Hüllenprobe.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// probeZahlen sind die Zahlen, in denen sich die drei Flächen treffen.
type probeZahlen struct {
	Probelauf bool `json:"probelauf"`
	Karten    int  `json:"karten"`
	Seiten    int  `json:"seiten"`
	JeSeite   int  `json:"jeSeite"`
	DPI       int  `json:"dpi"`
	Bytes     int  `json:"bytes"`
	Dateiname string
}

// zahlenAus liest die Antwort einer Fläche.
func zahlenAus(t *testing.T, wer, roh string) probeZahlen {
	t.Helper()

	var gelesen probeZahlen
	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("%s: Antwort lesen: %v (%s)", wer, err, roh)
	}

	return gelesen
}

// leerGeblieben sieht nach, dass im Ordner nichts liegt.
func leerGeblieben(t *testing.T, wer, ordner string) {
	t.Helper()

	eintraege, err := os.ReadDir(ordner)
	if err != nil || len(eintraege) != 0 {
		t.Errorf("%s hat trotz Probelauf etwas angelegt: %v (%v)", wer, eintraege, err)
	}
}

// **Der Probelauf des Bogens sagt vorher, was der Lauf legt.**
func TestDerBogenprobelaufSagtVorherWasDerLaufLegt(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	t.Run("Kommandozeile", func(t *testing.T) {
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "cli.pdf")

		aus, fehler, code := kommandozeile(t, umgebung, "bogen", setID, pfad, "--probe", "--json")
		if code != 0 {
			t.Fatalf("bogen --probe: %d (%s)", code, fehler)
		}

		probe := zahlenAus(t, "die Kommandozeile", aus)
		leerGeblieben(t, "die Kommandozeile", ordner)

		echt, fehler, code := kommandozeile(t, umgebung, "bogen", setID, pfad, "--json")
		if code != 0 {
			t.Fatalf("bogen: %d (%s)", code, fehler)
		}

		gelegt := zahlenAus(t, "die Kommandozeile", echt)
		if !probe.Probelauf || probe.Karten != gelegt.Karten || probe.Seiten != gelegt.Seiten {
			t.Errorf("Probelauf %+v, Lauf %+v", probe, gelegt)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "mcp.pdf")

		text, istFehler := mcpRuf(t, umgebung, "lesen.set.bogenprobelauf",
			map[string]any{"setID": zahl(t, setID), "ziel": pfad})
		if istFehler {
			t.Fatalf("lesen.set.bogenprobelauf: %s", text)
		}

		probe := zahlenAus(t, "MCP", text)
		leerGeblieben(t, "MCP", ordner)

		echt, istFehler := mcpRuf(t, umgebung, "schreiben.set.bogen",
			map[string]any{"setID": zahl(t, setID), "ziel": pfad})
		if istFehler {
			t.Fatalf("schreiben.set.bogen: %s", echt)
		}

		gelegt := zahlenAus(t, "MCP", echt)
		if !probe.Probelauf || probe.Karten != gelegt.Karten || probe.Seiten != gelegt.Seiten {
			t.Errorf("Probelauf %+v, Lauf %+v", probe, gelegt)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "rest.pdf")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/bogen/probelauf",
			`{"ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusOK {
			t.Fatalf("POST probelauf: %d (%s)", status, rumpf)
		}

		probe := zahlenAus(t, "REST", rumpf)
		leerGeblieben(t, "REST", ordner)

		status, echt := an.ruf(t, http.MethodPost, "/sets/"+setID+"/bogen",
			`{"ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusCreated {
			t.Fatalf("POST bogen: %d (%s)", status, echt)
		}

		gelegt := zahlenAus(t, "REST", echt)
		if !probe.Probelauf || probe.Karten != gelegt.Karten || probe.Seiten != gelegt.Seiten {
			t.Errorf("Probelauf %+v, Lauf %+v", probe, gelegt)
		}
	})
}

// **Der Probelauf der Exportvorlage nennt genau die Größe, die danach auf der
// Platte liegt.**
func TestDerVorlagenprobelaufNenntDieGroesseDerDatei(t *testing.T) {
	const paket = "probe-liste.mse-export-template"

	groesse := func(t *testing.T, pfad string) int {
		t.Helper()

		roh, err := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
		if err != nil {
			t.Fatalf("die Datei fehlt: %v", err)
		}

		return len(roh)
	}

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "cli.txt")

		aus, fehler, code := kommandozeile(t, umgebung,
			"vorlagen", setID, "ausfuehren", paket, pfad, "--probe", "--json")
		if code != 0 {
			t.Fatalf("--probe: %d (%s)", code, fehler)
		}

		probe := zahlenAus(t, "die Kommandozeile", aus)
		leerGeblieben(t, "die Kommandozeile", ordner)

		mussKommandozeile(t, umgebung, "vorlagen", setID, "ausfuehren", paket, pfad)

		if !probe.Probelauf || probe.Karten != 2 || probe.Bytes != groesse(t, pfad) {
			t.Errorf("Probelauf %+v, Datei mit %d Bytes", probe, groesse(t, pfad))
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "mcp.txt")

		text, istFehler := mcpRuf(t, umgebung, "lesen.vorlage.probelauf",
			map[string]any{"setID": zahl(t, setID), "paket": paket, "ziel": pfad})
		if istFehler {
			t.Fatalf("lesen.vorlage.probelauf: %s", text)
		}

		probe := zahlenAus(t, "MCP", text)
		leerGeblieben(t, "MCP", ordner)

		if echt, istFehler := mcpRuf(t, umgebung, "schreiben.vorlage.ausfuehren",
			map[string]any{"setID": zahl(t, setID), "paket": paket, "ziel": pfad}); istFehler {
			t.Fatalf("schreiben.vorlage.ausfuehren: %s", echt)
		}

		if !probe.Probelauf || probe.Karten != 2 || probe.Bytes != groesse(t, pfad) {
			t.Errorf("Probelauf %+v, Datei mit %d Bytes", probe, groesse(t, pfad))
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		an := starteDienst(t, umgebung)
		ordner := t.TempDir()
		pfad := filepath.Join(ordner, "rest.txt")

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/vorlagen/probelauf",
			`{"paket":"`+paket+`","ziel":`+alsJSONText(t, pfad)+`}`)
		if status != http.StatusOK {
			t.Fatalf("POST probelauf: %d (%s)", status, rumpf)
		}

		probe := zahlenAus(t, "REST", rumpf)
		leerGeblieben(t, "REST", ordner)

		if status, echt := an.ruf(t, http.MethodPost, "/sets/"+setID+"/vorlagen",
			`{"paket":"`+paket+`","ziel":`+alsJSONText(t, pfad)+`}`); status != http.StatusCreated {
			t.Fatalf("POST vorlagen: %d (%s)", status, echt)
		}

		if !probe.Probelauf || probe.Karten != 2 || probe.Bytes != groesse(t, pfad) {
			t.Errorf("Probelauf %+v, Datei mit %d Bytes", probe, groesse(t, pfad))
		}
	})
}

// **Dieselben fachlichen Fehler wie der Lauf — an jeder Fläche.**
//
// Ein Set, das es nicht gibt, und eine Vorlage, die es nicht gibt: der Lauf
// sagt Nein, und der Probelauf sagt es ebenso, bevor jemand den Knopf drückt.
func TestDerProbelaufGibtAnJederFlaecheDieFehlerDesLaufs(t *testing.T) {
	ziel := t.TempDir()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		_, _, bogen := kommandozeile(t, umgebung, "bogen", "999", filepath.Join(ziel, "a.pdf"), "--probe")
		_, _, lauf := kommandozeile(t, umgebung, "bogen", "999", filepath.Join(ziel, "a.pdf"))

		if bogen == 0 || lauf == 0 {
			t.Errorf("Bogen: Probelauf %d, Lauf %d", bogen, lauf)
		}

		_, _, probe := kommandozeile(t, umgebung, "vorlagen", setID, "ausfuehren",
			"gibtsnicht.mse-export-template", filepath.Join(ziel, "b.txt"), "--probe")
		_, _, echt := kommandozeile(t, umgebung, "vorlagen", setID, "ausfuehren",
			"gibtsnicht.mse-export-template", filepath.Join(ziel, "b.txt"))

		if probe == 0 || echt == 0 {
			t.Errorf("Vorlage: Probelauf %d, Lauf %d", probe, echt)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)

		if text, istFehler := mcpRuf(t, umgebung, "lesen.set.bogenprobelauf",
			map[string]any{"setID": 999}); !istFehler {
			t.Errorf("Bogen: MCP hat es hingenommen: %s", text)
		}

		if text, istFehler := mcpRuf(t, umgebung, "lesen.vorlage.probelauf",
			map[string]any{
				"setID": zahl(t, setID), "paket": "gibtsnicht.mse-export-template",
			}); !istFehler {
			t.Errorf("Vorlage: MCP hat es hingenommen: %s", text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := mitEigenerSaat(t)
		setID := vorlagenset(t, umgebung)
		an := starteDienst(t, umgebung)

		if status, rumpf := an.ruf(t, http.MethodPost, "/sets/999/bogen/probelauf",
			`{}`); status < 400 {
			t.Errorf("Bogen: REST hat es hingenommen: %d (%s)", status, rumpf)
		}

		if status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/vorlagen/probelauf",
			`{"paket":"gibtsnicht.mse-export-template"}`); status < 400 {
			t.Errorf("Vorlage: REST hat es hingenommen: %d (%s)", status, rumpf)
		}
	})
}

// **Ein Ziel, an dem schon etwas liegt, weisen MCP und REST im Probelauf ab,
// wie im Lauf** — und die Datei dort bleibt, wie sie war.
//
// Die Kommandozeile hat diese Wache nicht und überschreibt auch im Lauf
// (`internal/mcpnaehte/ausfuhr.go`, 0044): der Unterschied ist die Aufsicht.
func TestEinBelegtesZielWirdImProbelaufAbgewiesen(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	setID := setkennungAus(t,
		mussKommandozeile(t, umgebung, "oeffnen", probeset(t), "--json"))

	belegt := filepath.Join(t.TempDir(), "da.pdf")
	if err := os.WriteFile(belegt, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	t.Run("MCP", func(t *testing.T) {
		text, istFehler := mcpRuf(t, umgebung, "lesen.set.bogenprobelauf",
			map[string]any{"setID": zahl(t, setID), "ziel": belegt})
		if !istFehler || !strings.Contains(text, "überschreibt nichts") {
			t.Errorf("MCP: %v, %s", istFehler, text)
		}
	})

	t.Run("REST", func(t *testing.T) {
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/sets/"+setID+"/bogen/probelauf",
			`{"ziel":`+alsJSONText(t, belegt)+`}`)
		if status != http.StatusConflict {
			t.Errorf("REST: %d (%s), erwartet %d", status, rumpf, http.StatusConflict)
		}
	})

	if roh, err := os.ReadFile(belegt); err != nil || string(roh) != "x" { // #nosec G304 -- t.TempDir()
		t.Errorf("der vorhandene Inhalt ist hin: %q, %v", roh, err)
	}
}
