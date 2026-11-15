package msepaket_test

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Paket kommt von außen, und was darin steht, wird nicht geglaubt** —
// FMT-06.AK-01 (#1695).
//
// # Eine Datei, die mehr wird, als sie war
//
// Ein Eintrag aus Nullen schrumpft beim Packen auf ein Tausendstel. Gelesen
// wurde jede Datei eines Archivs ganz, gleich wie groß sie ausgepackt wird —
// ein Set von ein paar hundert Kilobyte füllte so den Arbeitsspeicher.
func TestEineAufgeblaseneDateiImArchivWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bombe.mse-set")
	archivMitNullen(t, pfad, "set", msepaket.GrenzePaketdatei+1)

	paket, err := msepaket.Oeffne(pfad)
	if err == nil {
		_ = paket.Schliessen() //nolint:errcheck // nur aufgeräumt
		t.Fatal("geöffnet — die Hauptdatei ist größer als die Grenze einer Paketdatei")
	}

	if !errors.Is(err, msepaket.ErrZuGross) {
		t.Errorf("Fehler %v, erwartet ErrZuGross", err)
	}
}

// **Ein Paketverzeichnis führt über keinen Symlink hinaus.**
//
// `os.DirFS` hält einen Symlink nach draußen nicht auf — so steht es in seiner
// eigenen Dokumentation (`os/file.go`, Go 1.27.1: „DirFS does not stop the
// access any more than using os.Open does"). Ein Vorlagenverzeichnis, dessen
// `style` auf eine Datei daneben zeigt, gab deren Inhalt als Paketkopf aus.
func TestEinPaketverzeichnisFolgtKeinemSymlinkHinaus(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	geheim := filepath.Join(ort, "geheim.txt")
	schreibe(t, geheim, "mse version: 2.0.0\nfull name: GEHEIM\n")

	paketordner := filepath.Join(ort, "magic-probe.mse-style")
	if err := os.Mkdir(paketordner, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(geheim, filepath.Join(paketordner, "style")); err != nil {
		t.Fatal(err)
	}

	paket, err := msepaket.Oeffne(paketordner)
	if err == nil {
		defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur aufgeräumt

		t.Fatalf("geöffnet, voller Name %q — der Kopf kam von außerhalb des Pakets", paket.VollerName)
	}

	if strings.Contains(err.Error(), "GEHEIM") {
		t.Errorf("der Fehler nennt den Inhalt der fremden Datei: %v", err)
	}
}

// **Ein Symlink, der im Paket bleibt, bleibt erlaubt.** Sonst wiese die
// Grenze auch Pakete ab, die sich eine Datei teilen.
func TestEinSymlinkImPaketBleibtLesbar(t *testing.T) {
	t.Parallel()

	paketordner := filepath.Join(t.TempDir(), "magic-probe.mse-style")
	if err := os.MkdirAll(filepath.Join(paketordner, "echt"), 0o700); err != nil {
		t.Fatal(err)
	}

	schreibe(t, filepath.Join(paketordner, "echt", "style"), "mse version: 2.0.0\nfull name: Probe\n")

	if err := os.Symlink(filepath.Join("echt", "style"), filepath.Join(paketordner, "style")); err != nil {
		t.Fatal(err)
	}

	paket, err := msepaket.Oeffne(paketordner)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur aufgeräumt

	if paket.VollerName != "Probe" {
		t.Errorf("voller Name %q, erwartet Probe", paket.VollerName)
	}
}

// archivMitNullen legt ein Archiv an, dessen einziger Eintrag aus `laenge`
// Nullbytes besteht — gepackt, also klein auf der Platte.
func archivMitNullen(t *testing.T, pfad, name string, laenge int64) {
	t.Helper()

	datei, err := os.Create(pfad) //nolint:gosec // im Temporärverzeichnis des Tests
	if err != nil {
		t.Fatal(err)
	}

	schreiber := zip.NewWriter(datei)

	eintrag, err := schreiber.Create(name)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(eintrag, io.LimitReader(nullquelle{}, laenge)); err != nil {
		t.Fatal(err)
	}

	if err := schreiber.Close(); err != nil {
		t.Fatal(err)
	}

	if err := datei.Close(); err != nil {
		t.Fatal(err)
	}
}

// nullquelle liefert Nullbytes ohne Ende.
type nullquelle struct{}

func (nullquelle) Read(puffer []byte) (int, error) {
	clear(puffer)

	return len(puffer), nil
}

// schreibe legt eine Datei mit Inhalt an.
func schreibe(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatal(err)
	}
}

// **Ein Archivpfad führt nicht aus dem Paket** — weder `../` noch ein
// führender Schrägstrich. `archive/zip` lässt solche Namen standardmäßig zu
// (`zipinsecurepath`), **bereinigt sie aber in seiner Dateisicht**: aus
// `../../ausbruch` wird `ausbruch`, gemessen an diesem Test. Was
// [msepaket.Paket.Dateien] aufzählt, ist damit ein gültiger Pfad **im**
// Paket — wer die Namen auf die Platte legt, legt sie hinein und nicht
// daneben —, und der rohe Name wird von [msepaket.Paket.Datei] abgewiesen.
func TestEinArchivpfadFuehrtNichtHinaus(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "pfade.mse-set")

	datei, err := os.Create(pfad) //nolint:gosec // im Temporärverzeichnis des Tests
	if err != nil {
		t.Fatal(err)
	}

	schreiber := zip.NewWriter(datei)

	for name, inhalt := range map[string]string{
		"set":            "mse version: 2.0.0\ngame: magic\n",
		"../../ausbruch": "geheim",
		"/absolut":       "geheim",
	} {
		eintrag, fehler := schreiber.Create(name)
		if fehler != nil {
			t.Fatal(fehler)
		}

		if _, fehler := io.WriteString(eintrag, inhalt); fehler != nil {
			t.Fatal(fehler)
		}
	}

	if fehler := schreiber.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := datei.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = paket.Schliessen() }() //nolint:errcheck // nur aufgeräumt

	namen, err := paket.Dateien()
	if err != nil {
		t.Fatalf("aufzählen: %v", err)
	}

	if len(namen) != 3 {
		t.Errorf("%d Namen aufgezählt: %v", len(namen), namen)
	}

	for _, name := range namen {
		if !fs.ValidPath(name) {
			t.Errorf("aufgezählt, und kein Pfad im Paket: %q", name)
		}
	}

	for _, name := range []string{"../../ausbruch", "/absolut"} {
		if _, err := paket.Datei(name); !errors.Is(err, msepaket.ErrAusbruch) {
			t.Errorf("%s: Fehler %v, erwartet ErrAusbruch", name, err)
		}
	}
}
