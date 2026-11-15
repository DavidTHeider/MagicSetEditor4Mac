// **Ein Bildwert ist ein Dateiname, bis er geklammert ist** — FMT-02.AK-02
// (#1691).
//
// Gefunden an der Hüllenprobe `fmt02`: die Kopie der Saat meldete in
// `probe-schlicht` einen Fehler, den es nicht gibt — `hell: hell.png` unter
// `choice images:` wurde als Skript gelesen. MSE2 liest einen Bildwert als
// Dateinamen, außer er beginnt mit `script:` oder trägt eine Klammer
// (`Reader::handle(ScriptableImage&)`, `src/script/image.cpp` Zeilen 67–78,
// Fassung f9d9356 vom 2021-07-06); der Zeichenweg dieses Programms tut
// dasselbe (`bildangabe` in `stilbild.go`). Die Prüfung folgt ihm.
package service_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinBildwertIstKeinSkript(t *testing.T) {
	for name, datei := range map[string]string{
		"ein Dateiname mit Schrägstrich":  "card style:\n\tname:\n\t\tmask: /magic-genevensis-00-main.mse-style/pride_mask.png\n",
		"ein Bild mit Punkt":              "card style:\n\tname:\n\t\timage: rahmen.png\n",
		"Wahlbilder mit Dateinamen":       "styling style:\n\thintergrund:\n\t\tchoice images:\n\t\t\thell:   hell.png\n\t\t\tdunkel: dunkel.png\n",
		"ein Bild als script:-Unterblock": "card style:\n\tstempel:\n\t\timage:\n\t\t\tscript: if card.name == \"\" then \"a.png\" else \"b.png\"\n",
		"ein geklammertes Bild":           "card style:\n\tname:\n\t\timage: { \"rahmen\" + \".png\" }\n",
	} {
		if befund := service.PaketSkriptfehler(datei); befund != nil {
			t.Errorf("%s: %v", name, befund)
		}
	}
}

func TestEinKaputtesBildskriptWirdGefunden(t *testing.T) {
	for name, fall := range map[string]struct {
		datei string
		zeile int
	}{
		"ein geklammertes Bild":    {"card style:\n\tname:\n\t\timage: { f(1, }\n", 3},
		"ein script:-Unterblock":   {"card style:\n\tstempel:\n\t\timage:\n\t\t\tscript: f(1,\n", 4},
		"ein Wahlbild als Skript":  {"card style:\n\tseltenheit:\n\t\tchoice images:\n\t\t\tbasic land:\n\t\t\t\tscript: f(1,\n", 5},
		"eine offene Klammer vorn": {"card style:\n\tname:\n\t\tmask: { f(1)\n", 3},
	} {
		befund := service.PaketSkriptfehler(fall.datei)
		if befund == nil || befund.Zeile != fall.zeile {
			t.Errorf("%s: %v, erwartet in Zeile %d", name, befund, fall.zeile)
		}
	}
}

// **Die Saat meldet nichts** — jede Textdatei jedes mitgelieferten Pakets ist
// gültig, und eine Prüfung, die dort etwas findet, findet einen Fehler in sich.
func TestDieSaatMeldetKeinenSkriptfehler(t *testing.T) {
	gezaehlt := 0

	err := filepath.WalkDir(filepath.Join("..", "..", "saat"), func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil || eintrag.IsDir() || !service.Bearbeitbar(eintrag.Name()) {
			return fehler
		}

		inhalt, lesefehler := os.ReadFile(pfad) //nolint:gosec // Pfad aus der Saat des Repositoriums
		if lesefehler != nil {
			return lesefehler
		}

		gezaehlt++

		if befund := service.PaketSkriptfehler(string(inhalt)); befund != nil {
			t.Errorf("%s: %v", pfad, befund)
		}

		return nil
	})
	if err != nil || gezaehlt < 5 {
		t.Fatalf("die Saat: %d Dateien, %v", gezaehlt, err)
	}
}

// **Die erste Stelle zählt** — ein Skriptfehler und eine unbekannte Feldart
// stehen in derselben Prüfung; gemeldet wird, was weiter oben steht.
func TestDiePruefungNenntEineUnbekannteFeldart(t *testing.T) {
	for name, fall := range map[string]struct {
		datei string
		zeile int
	}{
		"nur die Feldart":    {"card field:\n\ttype: audit-ungueltig\n\tname: a\n", 2},
		"die Feldart zuerst": {"card field:\n\ttype: audit-ungueltig\nscript: f(1,\n", 2},
		"das Skript zuerst":  {"script: f(1,\ncard field:\n\ttype: audit-ungueltig\n", 1},
	} {
		befund := service.PaketSkriptfehler(fall.datei)
		if befund == nil || befund.Zeile != fall.zeile {
			t.Errorf("%s: %v, erwartet Zeile %d", name, befund, fall.zeile)
		}
	}
}

// **Die Feldart trägt ihren eigenen Schlüssel** — das Fenster übersetzt ihn
// (§9); Kommandozeile, MCP und der Dienst bekommen den deutschen Satz.
func TestEinFeldartbefundTraegtSchluesselUndWerte(t *testing.T) {
	fehlt := service.PaketSkriptfehler("card field:\n\tname: a\n")
	if fehlt == nil || fehlt.Schluessel != service.SchluesselFeldartFehlt || len(fehlt.Werte) != 0 {
		t.Errorf("fehlende Art: %+v", fehlt)
	}

	fremd := service.PaketSkriptfehler("card field:\n\ttype: bild\n")
	if fremd == nil || fremd.Schluessel != service.SchluesselFeldartUnbekannt ||
		len(fremd.Werte) != 2 || fremd.Werte[0] != "bild" || !strings.HasPrefix(fremd.Werte[1], "boolean, ") {
		t.Errorf("unbekannte Art: %+v", fremd)
	}
}
