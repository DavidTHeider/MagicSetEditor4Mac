package msepaket_test

// **Eine Abhängigkeit in einer Zeile** — FMT-02
// (#1691).
//
// MSE2 liest `depends on:` in zwei Formen: als Block mit `package:` und
// `version:` und als **eine Zeile** `depends on: paket fassung`
// (`Reader::handle(PackageDependency&)`, `src/util/io/package.cpp`, Zeilen
// 515–528, Fassung f9d9356 vom 2021-07-06). Die Fassung liest es mit
// `sscanf` (`Version::fromString`, `src/util/version.cpp`, Zeilen 33–38) —
// Leerzeichen davor zählen nicht.
//
// **Gemessen am 2026-09-29 über `full-magic-pack`**: elf Paketdateien tragen
// die kurze Form, zweiundzwanzig Zeilen, etwa `magic-crystal.mse-style`. Bis
// hierher verlor dieses Programm jede davon und meldete dazu einen Befund,
// der keiner war. Und ein Installer schreibt sie nur so (`depends_on:`).

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestAbhaengigkeitenInBeidenFormen(t *testing.T) {
	faelle := []struct {
		name    string
		zeilen  string
		paket   string
		fassung string
	}{
		{"als Block", "depends on:\n\tpackage: probe.mse-game\n\tversion: 2008-08-08\n", "probe.mse-game", "2008-08-08"},
		{"in einer Zeile", "depends on: probe.mse-game 2008-08-08\n", "probe.mse-game", "2008-08-08"},
		{"mit Abstand vor der Fassung", "depends on: probe.mse-game                     2009-08-10\n", "probe.mse-game", "2009-08-10"},
		{"ohne Fassung", "depends on: probe.mse-game\n", "probe.mse-game", "0000-00-00"},
		{"mit Unterstrich wie in einem Installer", "depends_on: probe.mse-game 2013-12-14\n", "probe.mse-game", "2013-12-14"},
		// **Eine Leerzeile dahinter macht keinen Block** — so stehen sechs der
		// elf Dateien im Pack, etwa `magic-text.mse-export-template`.
		{"in einer Zeile mit Leerzeile dahinter", "depends on: probe.mse-game 2008-08-08\n\ncard width: 375\n", "probe.mse-game", "2008-08-08"},
		{"in einer Zeile mit Kommentar dahinter", "depends on: probe.mse-game 2008-08-08\n# Maße\n", "probe.mse-game", "2008-08-08"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			paket := stilMit(t, fall.zeilen)

			if len(paket.Befunde) != 0 {
				t.Fatalf("Befunde %v — die Form ist gültig", paket.Befunde)
			}

			if len(paket.Abhaengigkeiten) != 1 {
				t.Fatalf("%d Abhängigkeiten, erwartet eine", len(paket.Abhaengigkeiten))
			}

			gelesen := paket.Abhaengigkeiten[0]
			if gelesen.Paket != fall.paket || gelesen.Version.String() != fall.fassung {
				t.Errorf("gelesen %s %s, erwartet %s %s", gelesen.Paket, gelesen.Version, fall.paket, fall.fassung)
			}
		})
	}
}

func TestEineUnlesbareFassungInEinerZeileIstEinBefund(t *testing.T) {
	paket := stilMit(t, "depends on: probe.mse-game gestern\n")

	if len(paket.Befunde) != 1 {
		t.Fatalf("Befunde %v, erwartet einen", paket.Befunde)
	}

	if len(paket.Abhaengigkeiten) != 1 || paket.Abhaengigkeiten[0].Paket != "probe.mse-game" {
		t.Errorf("Abhängigkeiten %v — der Name ist lesbar und bleibt", paket.Abhaengigkeiten)
	}
}

// **Im Block steht der Befund an der Zeile der Fassung** — wie vor der
// zweiten Form; wer die Datei öffnet, springt dorthin.
func TestEineUnlesbareFassungImBlockNenntIhreZeile(t *testing.T) {
	paket := stilMit(t, "depends on:\n\tpackage: probe.mse-game\n\tversion: gestern\n")

	if len(paket.Befunde) != 1 || paket.Befunde[0].Stelle.Zeile != 5 {
		t.Errorf("Befunde %v, erwartet einen in Zeile 5", paket.Befunde)
	}
}

// **Eine leere Zeile ohne Block ist unvollständig** — wie ein Block ohne
// `package:`.
func TestEineLeereAbhaengigkeitIstEinBefund(t *testing.T) {
	paket := stilMit(t, "depends on:\n")

	if len(paket.Befunde) != 1 || len(paket.Abhaengigkeiten) != 0 {
		t.Errorf("Befunde %v, Abhängigkeiten %v — erwartet ein Befund und keine", paket.Befunde, paket.Abhaengigkeiten)
	}
}

// stilMit legt einen Stil mit diesen Kopfzeilen an und öffnet ihn.
func stilMit(t *testing.T, zeilen string) *msepaket.Paket {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe-form.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatal(err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\n" + zeilen
	if err := os.WriteFile(filepath.Join(ort, "style"), []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}

	paket, err := msepaket.Oeffne(ort)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur gelesen

	return paket
}
