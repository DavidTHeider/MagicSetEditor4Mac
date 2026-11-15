package app

import (
	"os"
	"path/filepath"
	"testing"
)

// **Das Beigabenverzeichnis entsteht, bevor jemand hineinschreibt** —
// #732.
//
//	wxFileName fn(outname);
//	info.directory_relative = fn.GetName() + _("-files");
//	fn.SetFullName(info.directory_relative);
//	info.directory_absolute = fn.GetFullPath();
//	if (!wxDirExists(info.directory_absolute)) wxMkdir(info.directory_absolute);
//
// `src/gui/html_export_window.cpp`, Zeilen 63–71, Fassung f9d9356 vom
// 2021-07-06.
//
// # Warum dieser Test hier steht und nicht am Weg durch die Fassade
//
// **Weil die Gegenprobe dort grün war.** Nimmt man das `MkdirAll` heraus, läuft
// der Weg durch die Fassade trotzdem — der Bildschreiber legt sein
// Zielverzeichnis selbst an, und alles Weitere findet es dann vor. Die Zusage
// hing an der Reihenfolge im Skript: eine Vorlage, die zuerst `write_text_file`
// ruft, scheiterte.
//
// Geprüft wird deshalb die Funktion, die die Zusage gibt (§10).
func TestDerBeigabenortEntsteht(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "spoiler.html")

	ort, name, err := beigabenort(ziel, true)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// **Der Stamm ohne Endung, plus `-files`** — `fn.GetName()` gibt bei
	// wxWidgets den Namen ohne Erweiterung.
	if name != "spoiler-files" {
		t.Errorf("Name %q", name)
	}

	if ort != filepath.Join(filepath.Dir(ziel), "spoiler-files") {
		t.Errorf("Ort %q", ort)
	}

	angaben, err := os.Stat(ort)
	if err != nil {
		t.Fatalf("das Verzeichnis liegt nicht da: %v", err)
	}

	if !angaben.IsDir() {
		t.Errorf("%s ist kein Verzeichnis", ort)
	}
}

// **Ohne `create directory:` entsteht keines.**
//
// Ein leeres Verzeichnis neben jeder Ausfuhr wäre Müll — und die Vorlage darf
// ohnehin nicht schreiben.
func TestOhneAngabeEntstehtKeinBeigabenort(t *testing.T) {
	t.Parallel()

	verzeichnis := t.TempDir()
	ziel := filepath.Join(verzeichnis, "liste.txt")

	ort, name, err := beigabenort(ziel, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if name != "" || ort != verzeichnis {
		t.Errorf("Ort %q, Name %q", ort, name)
	}

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 0 {
		t.Errorf("es liegt etwas da: %v", eintraege)
	}
}

// **Ein Ort, an dem sich nichts anlegen lässt, wird gemeldet.**
//
// Eine Ausfuhr, die ihre Beigaben nicht ablegen kann, ist keine (§8) — anders
// als bei einer Karte, wo eine Ebene weniger brauchbar bleibt.
func TestEinUnanlegbarerBeigabenortWirdGemeldet(t *testing.T) {
	t.Parallel()

	// **Eine Datei an der Stelle des Verzeichnisses.** `MkdirAll` scheitert
	// daran, und das ist der Fall, den ein voller oder schreibgeschützter
	// Datenträger auslöst.
	verzeichnis := t.TempDir()
	sperre := filepath.Join(verzeichnis, "liste-files")

	if err := os.WriteFile(sperre, []byte("keine Mappe"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, _, err := beigabenort(filepath.Join(verzeichnis, "liste.txt"), true); err == nil {
		t.Fatal("das ging durch")
	}
}

// **Ein Schreibfehler bei einer Beigabe kommt durch** (§8).
//
// Eine halbe Ausfuhr ist eine, die jemand für vollständig hält — anders als bei
// einer Karte, wo eine Ebene weniger brauchbar bleibt.
func TestEinSchreibfehlerBeiEinerBeigabe(t *testing.T) {
	t.Parallel()

	// **Ein Verzeichnis, in das niemand schreiben darf.** Das ist der Weg, den
	// `copy_file` und `write_text_file` nehmen — und der Fall, den ein voller
	// oder schreibgeschützter Datenträger auslöst.
	gesperrt := filepath.Join(t.TempDir(), "zu")
	if err := os.Mkdir(gesperrt, 0o500); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	//nolint:gosec,errcheck // ein Verzeichnis braucht das Ausführrecht; nur zum Aufräumen
	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) })

	if _, err := (beigaben{ziel: gesperrt}).Lege("x.txt", []byte("y")); err == nil {
		t.Fatal("das ging durch")
	}

	// **Und ein Name, der ein Pfad ist, kommt gar nicht erst bis zur Platte.**
	// Abgewiesen, nicht bereinigt — der Grund steht bei [harmloserName].
	if _, err := (beigaben{ziel: t.TempDir()}).Lege("../oben.txt", []byte("y")); err == nil {
		t.Fatal("ein Pfad ging als Name durch")
	}
}
