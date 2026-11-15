package main

// **Der Probelauf an der Kommandozeile** — UXR-F7
// (#2015).
//
// `bogen` und `vorlagen … ausfuehren` nehmen ein freiwilliges `--probe`: derselbe
// Aufruf, derselbe Weg bis vor das Schreiben, und es wird nichts geschrieben.
// Einordnung nach §16: **erweiternd** — ohne das Wort tut jeder Aufruf, was er
// tat.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// vorgangszeile ist die Kennung des Rufs, die jede Klage trägt und jeder Aufruf
// neu würfelt.
var vorgangszeile = regexp.MustCompile(`Vorgang: [0-9a-f]+`)

// ohneVorgang nimmt sie heraus: verglichen wird der Satz, nicht der Ruf.
func ohneVorgang(klage string) string { return vorgangszeile.ReplaceAllString(klage, "Vorgang: -") }

// nichtsDa sieht nach, dass am Pfad nichts liegt.
func nichtsDa(t *testing.T, was, pfad string) {
	t.Helper()

	if _, err := os.Stat(pfad); err == nil {
		t.Errorf("%s hat trotz --probe etwas angelegt: %s", was, pfad)
	}
}

// **`bogen --probe` nennt, was der Lauf legt — und legt nichts.**
func TestBogenProbeNenntDieZahlenDesLaufs(t *testing.T) {
	ort, _ := eingefahren(t)
	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	code, aus, fehler := fahreAn(t, ort, "bogen", "1", ziel, "--probe", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	nichtsDa(t, "bogen", ziel)

	var probe bogenprobeantwort
	if err := json.Unmarshal([]byte(aus), &probe); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	// **Gegen den echten Lauf desselben Aufrufs.**
	_, echt, echtfehler := fahreAn(t, ort, "bogen", "1", ziel, "--json")

	var gelegt bogenantwort
	if err := json.Unmarshal([]byte(echt), &gelegt); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, echtfehler)
	}

	if !probe.Probelauf || probe.Karten != gelegt.Karten || probe.Seiten != gelegt.Seiten ||
		probe.JeSeite != 9 || probe.DPI != 300 {
		t.Errorf("Probelauf %+v, Lauf %+v", probe, gelegt)
	}
}

// **Und als Satz für Menschen.**
func TestBogenProbeSagtEsInWorten(t *testing.T) {
	ort, _ := eingefahren(t)

	code, aus, fehler := fahreAn(t, ort, "bogen", "1", "--probe")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Probelauf") || !strings.Contains(aus, "nichts geschrieben") ||
		!strings.Contains(aus, "300 dpi") {
		t.Errorf("Ausgabe: %q", aus)
	}
}

// **Dieselben Fehler wie der Lauf, mit demselben Code und demselben Satz.**
func TestBogenProbeGibtDieFehlerDesLaufs(t *testing.T) {
	ort, _ := eingefahren(t)
	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	for name, set := range map[string]string{"ein Set, das es nicht gibt": "999"} {
		t.Run(name, func(t *testing.T) {
			probe, _, probefehler := fahreAn(t, ort, "bogen", set, ziel, "--probe")
			lauf, _, laufklage := fahreAn(t, ort, "bogen", set, ziel)

			if probe != lauf || probe != CodeFachlich {
				t.Errorf("Probelauf %d, Lauf %d, erwartet %d", probe, lauf, CodeFachlich)
			}

			if ohneVorgang(probefehler) != ohneVorgang(laufklage) {
				t.Errorf("Probelauf %q, Lauf %q", probefehler, laufklage)
			}

			nichtsDa(t, "bogen", ziel)
		})
	}
}

// **Die Bedienfehler bleiben Bedienfehler.**
func TestBogenProbeBedienfehler(t *testing.T) {
	ort, _ := eingefahren(t)

	faelle := map[string][]string{
		"ohne Setkennung":   {"bogen", "--probe"},
		"krumme Kennung":    {"bogen", "keine Zahl", "--probe"},
		"zu viele Wörter":   {"bogen", "1", "a.pdf", "b.pdf", "--probe"},
		"ohne alles, nackt": {"bogen"},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}

// **`vorlagen … ausfuehren --probe` nennt genau die Größe der Datei, die der
// Lauf danach schreibt — und schreibt vorher nichts.**
func TestVorlagenProbeNenntDieGroesseDerDatei(t *testing.T) {
	ort, setID := vorlagenset(t)
	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	code, aus, fehler := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"probe-liste.mse-export-template", ziel, "--probe", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	nichtsDa(t, "vorlagen", ziel)

	var probe vorlagenprobeantwort
	if err := json.Unmarshal([]byte(aus), &probe); err != nil {
		t.Fatalf("Antwort lesen: %v (%s)", err, aus)
	}

	if code, _, fehler := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"probe-liste.mse-export-template", ziel); code != CodeGut {
		t.Fatalf("Lauf: %d (%s)", code, fehler)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	if !probe.Probelauf || probe.Bytes != len(roh) || probe.Karten != 1 ||
		probe.Dateiname != "ausfuhr.txt" || probe.Befunde == nil {
		t.Errorf("Probelauf %+v, Datei mit %d Bytes", probe, len(roh))
	}
}

// **Ohne Ziel nennt er den Namen, den ein Dialog vorschlüge — als Satz.**
func TestVorlagenProbeOhneZiel(t *testing.T) {
	ort, setID := vorlagenset(t)

	code, aus, fehler := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"probe-liste.mse-export-template", "--probe")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, "Probelauf") || !strings.Contains(aus, "ausfuhr.txt") ||
		!strings.Contains(aus, "nichts geschrieben") {
		t.Errorf("Ausgabe: %q", aus)
	}
}

// **Dieselben Fehler wie der Lauf — samt den Befunden, die den Grund nennen.**
func TestVorlagenProbeGibtDieFehlerDesLaufs(t *testing.T) {
	ort, setID := vorlagenset(t)
	ziel := filepath.Join(t.TempDir(), "raus.txt")

	probe, _, probefehler := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"gibtsnicht.mse-export-template", ziel, "--probe")
	lauf, _, laufklage := fahreAn(t, ort, "vorlagen", setID, "ausfuehren",
		"gibtsnicht.mse-export-template", ziel)

	if probe != lauf || probe != CodeFachlich {
		t.Errorf("Probelauf %d, Lauf %d, erwartet %d", probe, lauf, CodeFachlich)
	}

	if ohneVorgang(probefehler) != ohneVorgang(laufklage) {
		t.Errorf("Probelauf %q, Lauf %q", probefehler, laufklage)
	}

	nichtsDa(t, "vorlagen", ziel)
}

// **Ein krummes Paar bleibt ein Bedienfehler.**
func TestVorlagenProbeBedienfehler(t *testing.T) {
	ort, setID := vorlagenset(t)

	faelle := map[string][]string{
		"ausfuehren ohne Vorlage": {"vorlagen", setID, "ausfuehren", "--probe"},
		"krummes Paar": {
			"vorlagen", setID, "ausfuehren", "probe-liste.mse-export-template",
			"/tmp/a.txt", "ohneGleich", "--probe",
		},
	}

	for name, aufruf := range faelle {
		t.Run(name, func(t *testing.T) {
			if code, _, _ := fahreAn(t, ort, aufruf...); code != CodeBedienung {
				t.Errorf("Rückgabewert %d, erwartet %d", code, CodeBedienung)
			}
		})
	}
}
