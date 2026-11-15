package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die vier Verhaltensweisen bei einem Namenskonflikt** —
// #491.
//
//	case CONFLICT_KEEP_OLD:  return !fn.FileExists();
//	case CONFLICT_OVERWRITE: return true;
//	case CONFLICT_NUMBER: {
//	  int i = 0; String ext = fn.GetExt();
//	  while (fn.FileExists()) fn.SetExt(String() << ++i << _(".") << ext);
//	  return true; }
//	case CONFLICT_NUMBER_OVERWRITE: {
//	  int i = 0; String ext = fn.GetExt();
//	  while (used.find(fn.GetFullPath()) != used.end())
//	    fn.SetExt(String() << ++i << _(".") << ext);
//	  return true; }
//
// `src/util/file_utils.cpp`, Zeilen 74–100, Fassung f9d9356 vom 2021-07-06.
//
// # `SetExt` und nicht „Name (1)"
//
// **Das ist die Stelle, an der man raten würde.** `SetExt` **ersetzt** die
// Endung: aus `Feuerball.png` wird `Feuerball.1.png`, nicht `Feuerball (1).png`
// und nicht `Feuerball1.png`. Und `ext` wird **vor** der Schleife gelesen,
// bleibt also die ursprüngliche — sonst hieße der zweite `Feuerball.1.2.png`.
func TestEinFreierNameBleibtWieErIst(t *testing.T) {
	t.Parallel()

	verzeichnis := t.TempDir()

	for _, verhalten := range []string{
		"altbehalten", "ueberschreiben", "nummerieren", "nummerierenneu",
	} {
		t.Run(verhalten, func(t *testing.T) {
			t.Parallel()

			name, schreiben := service.Namenskonflikt(
				verzeichnis, "frei-"+verhalten+".png", verhalten, nil)

			if !schreiben || name != "frei-"+verhalten+".png" {
				t.Errorf("%q / %v", name, schreiben)
			}
		})
	}
}

// **`altbehalten` schreibt gar nicht.** MSE gibt `false` zurück, und der
// Aufrufer macht `continue` — die Karte wird übersprungen, nicht mit einem
// anderen Namen abgelegt.
func TestAltbehaltenSchreibtNicht(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png")

	if _, schreiben := service.Namenskonflikt(verzeichnis, "da.png", "altbehalten", nil); schreiben {
		t.Error("es würde geschrieben")
	}
}

// **`ueberschreiben` schreibt über die vorhandene Datei** — unter demselben
// Namen.
func TestUeberschreibenNimmtDenselbenNamen(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png")

	name, schreiben := service.Namenskonflikt(verzeichnis, "da.png", "ueberschreiben", nil)
	if !schreiben || name != "da.png" {
		t.Errorf("%q / %v", name, schreiben)
	}
}

// **`nummerieren` zählt an der Platte hoch** — `da.png` ist belegt, also
// `da.1.png`.
func TestNummerierenZaehltAnDerPlatteHoch(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png", "da.1.png")

	name, schreiben := service.Namenskonflikt(verzeichnis, "da.png", "nummerieren", nil)
	if !schreiben || name != "da.2.png" {
		t.Errorf("%q / %v", name, schreiben)
	}
}

// **Und die Endung bleibt die ursprüngliche.** MSE liest `ext` **vor** der
// Schleife; wer sie in jedem Durchgang neu läse, bekäme `da.1.2.png`.
func TestDieEndungBleibtDieUrspruengliche(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png", "da.1.png", "da.2.png")

	name, _ := service.Namenskonflikt(verzeichnis, "da.png", "nummerieren", nil)
	if name != "da.3.png" {
		t.Errorf("%q", name)
	}
}

// **`nummerierenneu` sieht nur den laufenden Durchgang** — nicht die Platte.
//
// Damit überschreibt es, was ein früherer Lauf hinterlassen hat, und stößt
// innerhalb eines Laufs nie zusammen. Das ist MSEs Vorgabe.
func TestNummerierenNeuSiehtNurDenLauf(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png")

	// **Die Platte zählt nicht**: `da.png` liegt da und wird trotzdem genommen.
	name, schreiben := service.Namenskonflikt(verzeichnis, "da.png", "nummerierenneu", nil)
	if !schreiben || name != "da.png" {
		t.Errorf("%q / %v", name, schreiben)
	}

	// **Der Lauf zählt**: derselbe Name ein zweites Mal wird nummeriert.
	name, _ = service.Namenskonflikt(
		verzeichnis, "da.png", "nummerierenneu", map[string]bool{"da.png": true})
	if name != "da.1.png" {
		t.Errorf("%q", name)
	}
}

// **Ein unbekanntes Verhalten schreibt** — und zwar unter dem gegebenen Namen.
//
// Der Fall entsteht nur aus einem Bestand, den jemand von Hand geändert hat;
// `domain.PruefeEinstellung` lässt ihn nicht durch. Die Antwort ist trotzdem
// benannt: nicht zu schreiben hieße, eine Ausgabe stillschweigend ausfallen zu
// lassen.
func TestEinUnbekanntesVerhaltenSchreibt(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "da.png")

	name, schreiben := service.Namenskonflikt(verzeichnis, "da.png", "erfunden", nil)
	if !schreiben || name != "da.png" {
		t.Errorf("%q / %v", name, schreiben)
	}
}

// **Ein Name ohne Endung wird trotzdem nummeriert.**
func TestEinNameOhneEndungWirdNummeriert(t *testing.T) {
	t.Parallel()

	verzeichnis := legeBilder(t, "ohne")

	name, _ := service.Namenskonflikt(verzeichnis, "ohne", "nummerieren", nil)
	if name != "ohne.1" {
		t.Errorf("%q", name)
	}
}

// lege legt die genannten Dateien an und gibt das Verzeichnis.
func legeBilder(t *testing.T, namen ...string) string {
	t.Helper()

	verzeichnis := t.TempDir()

	for _, name := range namen {
		if err := os.WriteFile(filepath.Join(verzeichnis, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}
	}

	return verzeichnis
}

// **Und der Weg durch `Schreibe`** — `altbehalten` schreibt gar nicht.
//
// MSE macht an dieser Stelle `continue`: die Karte bleibt ungeschrieben, und
// die vorhandene Datei bleibt stehen. Ein Fehler ist das nicht — genau das
// hat sich jemand gewünscht.
func TestSchreibeUeberspringtEineVorhandeneDatei(t *testing.T) {
	verzeichnis := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(verzeichnis, "da.png"), []byte("alt"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ergebnis, err := bildausfuhr(t, planDouble{}).Schreibe(
		t.Context(), 1, verzeichnis, "da.png", service.FormatPNG, 0,
		domain.BildkonfliktAltbehalten, render.GueteMittel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if !ergebnis.Uebersprungen {
		t.Error("es wurde nicht übersprungen")
	}

	// **Und die alte Datei steht noch da**, unverändert.
	inhalt, err := os.ReadFile(filepath.Join(verzeichnis, "da.png")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if string(inhalt) != "alt" {
		t.Errorf("die Datei wurde angefasst: %q", inhalt)
	}
}

// **`nummerieren` weicht der vorhandenen Datei aus** — und schreibt daneben.
func TestSchreibeWeichtAus(t *testing.T) {
	verzeichnis := t.TempDir()

	if err := os.WriteFile(
		filepath.Join(verzeichnis, "da.png"), []byte("alt"), 0o600); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ergebnis, err := bildausfuhr(t, planDouble{}).Schreibe(
		t.Context(), 1, verzeichnis, "da.png", service.FormatPNG, 0,
		domain.BildkonfliktNummerieren, render.GueteMittel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if ergebnis.Uebersprungen {
		t.Fatal("es wurde übersprungen")
	}

	if filepath.Base(ergebnis.Pfad) != "da.1.png" {
		t.Errorf("%q", ergebnis.Pfad)
	}
}
