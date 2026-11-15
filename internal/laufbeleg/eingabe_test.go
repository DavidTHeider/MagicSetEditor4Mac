package laufbeleg_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/laufbeleg"
)

// schreibe legt eine Datei unter einer Wurzel an.
func schreibe(t *testing.T, wurzel, name, inhalt string) {
	t.Helper()

	pfad := filepath.Join(wurzel, name)
	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(pfad, []byte(inhalt), 0o600); fehler != nil {
		t.Fatal(fehler)
	}
}

// **Eine Pflichteingabe fehlt auf vier Arten**, und jede nennt ihren Grund —
// die vierte ist der Fall aus #1656: das Verzeichnis war da, und es war leer.
func TestEineFehlendeEingabeNenntIhrenGrund(t *testing.T) {
	t.Parallel()

	leer := t.TempDir()
	fremd := t.TempDir()
	schreibe(t, fremd, "liesmich.txt", "kein Set")

	faelle := map[string]struct {
		pfad, muster, grund string
	}{
		"nicht gesetzt":     {"", "", "nicht gesetzt"},
		"liegt nicht vor":   {filepath.Join(leer, "gibtsnicht"), "", "liegt nicht vor"},
		"leer":              {leer, "", "leer"},
		"ohne das Gesuchte": {fremd, "*.mse-set", "kein *.mse-set darin"},
	}

	for name, einer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			eingabe := laufbeleg.Erhebe("MSE_KANON", einer.pfad, einer.muster)
			if !strings.Contains(eingabe.Fehlt, einer.grund) || eingabe.Hash != "" {
				t.Errorf("fehlt %q, Hash %q — erwartet %q und keinen Hash",
					eingabe.Fehlt, eingabe.Hash, einer.grund)
			}

			if eingabe.Name != "MSE_KANON" {
				t.Errorf("Name %q", eingabe.Name)
			}
		})
	}
}

// **Ein Verzeichnis hat einen Hash über Namen und Inhalt** — und er ändert
// sich mit beidem, und mit nichts sonst.
func TestEinVerzeichnisHatEinenInhaltshash(t *testing.T) {
	t.Parallel()

	erheben := func(dateien map[string]string) laufbeleg.Eingabe {
		wurzel := t.TempDir()
		for name, inhalt := range dateien {
			schreibe(t, wurzel, name, inhalt)
		}

		return laufbeleg.Erhebe("MSE_KANON", wurzel, "*.mse-set")
	}

	grund := erheben(map[string]string{"a.mse-set": "eins", "tief/b.mse-set": "zwei", "c.txt": "drei"})
	gleich := erheben(map[string]string{"c.txt": "drei", "tief/b.mse-set": "zwei", "a.mse-set": "eins"})
	anderer := erheben(map[string]string{"a.mse-set": "EINS", "tief/b.mse-set": "zwei", "c.txt": "drei"})
	umbenannt := erheben(map[string]string{"x.mse-set": "eins", "tief/b.mse-set": "zwei", "c.txt": "drei"})

	if grund.Fehlt != "" || grund.Dateien != 3 || grund.Passend != 2 || grund.Bytes != 12 {
		t.Fatalf("erhoben %+v — erwartet drei Dateien, zwei passende, zwölf Bytes", grund)
	}

	if !strings.HasPrefix(grund.Hash, "sha256:") || len(grund.Hash) != len("sha256:")+64 {
		t.Errorf("Hash %q hat nicht die Gestalt sha256:<64 Zeichen>", grund.Hash)
	}

	if grund.Hash != gleich.Hash {
		t.Error("derselbe Inhalt, in anderer Reihenfolge angelegt, gab einen anderen Hash")
	}

	if grund.Hash == anderer.Hash || grund.Hash == umbenannt.Hash {
		t.Error("ein geänderter Inhalt oder Name ließ den Hash gleich")
	}
}

// **Eine Datei ist auch eine Eingabe** — ein einzelnes Set etwa.
func TestEineDateiIstEineEingabe(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	schreibe(t, wurzel, "Archenemy.mse-set", "set")

	passend := laufbeleg.Erhebe("MSE_SETPAKET", filepath.Join(wurzel, "Archenemy.mse-set"), "*.mse-set")
	if passend.Fehlt != "" || passend.Dateien != 1 || passend.Passend != 1 ||
		!strings.HasPrefix(passend.Hash, "sha256:") {
		t.Errorf("erhoben %+v", passend)
	}

	falsch := laufbeleg.Erhebe("MSE_SETPAKET", filepath.Join(wurzel, "Archenemy.mse-set"), "*.png")
	if !strings.Contains(falsch.Fehlt, "kein *.png") {
		t.Errorf("eine Datei, die nicht zum Muster passt, fehlt nicht: %+v", falsch)
	}
}

// **Eine Verknüpfung auf eine Datei zählt mit ihrem Ziel, eine auf ein
// Verzeichnis gar nicht** — sonst liefe ein Kreis ohne Ende.
func TestVerknuepfungenWerdenNichtVerfolgtWennSieKreisenKoennten(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	schreibe(t, wurzel, "echt.mse-set", "set")

	if fehler := os.Symlink(filepath.Join(wurzel, "echt.mse-set"),
		filepath.Join(wurzel, "zeiger.mse-set")); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Symlink(wurzel, filepath.Join(wurzel, "kreis")); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Symlink(filepath.Join(wurzel, "gibtsnicht"),
		filepath.Join(wurzel, "tot")); fehler != nil {
		t.Fatal(fehler)
	}

	eingabe := laufbeleg.Erhebe("MSE_KANON", wurzel, "*.mse-set")
	if eingabe.Fehlt != "" || eingabe.Dateien != 2 || eingabe.Passend != 2 {
		t.Errorf("erhoben %+v — erwartet die Datei und ihren Zeiger", eingabe)
	}
}

// **Eine Datei, die niemand lesen darf, ist keine Eingabe.**
func TestEineUnlesbareDateiFehlt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	schreibe(t, wurzel, "gesperrt.mse-set", "set")

	pfad := filepath.Join(wurzel, "gesperrt.mse-set")
	if fehler := os.Chmod(pfad, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, 0o600) }) //nolint:errcheck // nur zum Aufräumen

	for _, ziel := range []string{wurzel, pfad} {
		if eingabe := laufbeleg.Erhebe("MSE_KANON", ziel, ""); !strings.Contains(eingabe.Fehlt, "nicht lesbar") {
			t.Errorf("%s: erhoben %+v — erwartet »nicht lesbar«", ziel, eingabe)
		}
	}
}

// **Ein Unterverzeichnis, das niemand lesen darf**, macht die Eingabe
// unvollständig — und nicht stillschweigend kleiner.
func TestEinUnlesbaresUnterverzeichnisFehlt(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	schreibe(t, wurzel, "a.mse-set", "set")
	schreibe(t, wurzel, "tief/b.mse-set", "set")

	tief := filepath.Join(wurzel, "tief")
	if fehler := os.Chmod(tief, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(tief, 0o750) }) //nolint:errcheck,gosec // nur zum Aufräumen; ein Verzeichnis braucht das Ausführungsrecht

	if eingabe := laufbeleg.Erhebe("MSE_KANON", wurzel, ""); !strings.Contains(eingabe.Fehlt, "nicht lesbar") {
		t.Errorf("erhoben %+v — erwartet »nicht lesbar«", eingabe)
	}
}

// **Verborgenes zählt nicht** — der Finder legt `.DS_Store` an, wann er will,
// und mit ihm änderte sich der Hash, ohne dass sich an den Daten etwas
// geändert hätte.
func TestVerborgenesZaehltNicht(t *testing.T) {
	t.Parallel()

	ohne := t.TempDir()
	schreibe(t, ohne, "a.mse-set", "set")

	mit := t.TempDir()
	schreibe(t, mit, "a.mse-set", "set")
	schreibe(t, mit, ".DS_Store", "Finder")
	schreibe(t, mit, ".git/HEAD", "ref: refs/heads/main")

	erst, dann := laufbeleg.Erhebe("MSE_KANON", ohne, ""), laufbeleg.Erhebe("MSE_KANON", mit, "")
	if erst.Hash != dann.Hash || dann.Dateien != 1 {
		t.Errorf("ohne %+v\nmit  %+v", erst, dann)
	}
}

// **Eine Wurzel, die eine Verknüpfung ist, wird gelesen** — gefunden am
// 2026-09-22 an echten Daten: MSEs `data` ist eine Verknüpfung auf das
// Paketverzeichnis, und die Eingabe hieß »leer«. Innen wird weiter keiner
// Verknüpfung auf ein Verzeichnis gefolgt; die Wurzel nennt, wer den Lauf
// startet.
func TestEineVerknuepfteWurzelWirdGelesen(t *testing.T) {
	t.Parallel()

	echt := t.TempDir()
	schreibe(t, echt, "magic.mse-style/style", "stil")

	zeiger := filepath.Join(t.TempDir(), "data")
	if fehler := os.Symlink(echt, zeiger); fehler != nil {
		t.Fatal(fehler)
	}

	ueberZeiger, direkt := laufbeleg.Erhebe("MSE2_DATEN", zeiger, "style"), laufbeleg.Erhebe("MSE2_DATEN", echt, "style")
	if ueberZeiger.Fehlt != "" || ueberZeiger.Dateien != 1 || ueberZeiger.Hash != direkt.Hash {
		t.Errorf("über die Verknüpfung %+v\ndirekt %+v", ueberZeiger, direkt)
	}

	// **Der Pfad im Beleg bleibt der genannte** — und nicht das Ziel.
	if ueberZeiger.Pfad != zeiger {
		t.Errorf("Pfad %q, erwartet %q", ueberZeiger.Pfad, zeiger)
	}
}
