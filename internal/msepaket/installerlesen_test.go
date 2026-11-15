package msepaket_test

// **Einen `.mse-installer` lesen und auspacken** — FMT-02.AK-04
// (#1691).
//
// Ein Installer kommt von fremder Hand — genau dafür gibt es ihn. Was hier
// geprüft wird, ist deshalb zur Hälfte, **was nicht geschieht**: kein Eintrag
// landet außerhalb des Ziels, kein Verweis wird gefolgt, kein Paket, das die
// Hauptdatei nicht nennt, kommt mit.

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinGeschriebenerInstallerLiestSichZurueck(t *testing.T) {
	ort := t.TempDir()
	marken := legePaketAn(t, ort, "probe-marken.mse-include", map[string]string{
		"include":        "mse version: 2.0.0\nfull name: Probe Marken\nversion: 2026-09-01\n",
		"bilder/alt.png": "alt",
	})
	spiel := legePaketAn(t, ort, "probe.mse-game", map[string]string{
		"game": "mse version: 2.0.0\nshort name: Probe\nversion: 2026-09-02\n" +
			"depends on: probe-marken.mse-include 2026-09-01\n",
	})

	installer := installerAus(t, marken, spiel)

	if len(installer.Befunde) != 0 {
		t.Fatalf("Befunde %v", installer.Befunde)
	}

	namen := make([]string, 0, len(installer.Pakete))
	for _, paket := range installer.Pakete {
		namen = append(namen, paket.Name)
	}

	if !slices.Equal(namen, []string{"probe-marken.mse-include", "probe.mse-game"}) {
		t.Fatalf("Pakete %v", namen)
	}

	gelesen := installer.Pakete[1]
	if gelesen.KurzerName != "Probe" || gelesen.Version.String() != "2026-09-02" ||
		len(gelesen.Abhaengigkeiten) != 1 || gelesen.Abhaengigkeiten[0].Paket != "probe-marken.mse-include" {
		t.Errorf("Beschreibung %+v", gelesen)
	}

	ziel := t.TempDir()
	benutzt, err := installer.Auspacken("probe-marken.mse-include", ziel, 1<<20)
	if err != nil {
		t.Fatal(err)
	}

	if benutzt != int64(len("alt")+len(marken.KopfRoh)) {
		t.Errorf("%d Bytes gezählt", benutzt)
	}

	if roh, err := os.ReadFile(filepath.Join(ziel, "bilder", "alt.png")); err != nil || string(roh) != "alt" { //nolint:gosec // aus t.TempDir()
		t.Errorf("bilder/alt.png: %q, %v", roh, err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "game")); !errors.Is(err, fs.ErrNotExist) {
		t.Error("eine Datei des anderen Pakets kam mit")
	}
}

// **Was MSE2 schreibt, liest sich** — BOM, Schlüssel mit Unterstrich,
// `depends_on:` in einer Zeile und eine Fassung `0.0.0` für „keine".
func TestEinInstallerAusMSE2WirdGelesen(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer": "\ufeffmse_version: 0.3.7\nshort_name: \nversion: 0.0.0\n" +
			"package:\n\tname: probe.mse-game\n\tversion: 0.0.0\n\tshort_name: Probe\n" +
			"\tinstaller_group: Probe/Spiel\n\tdepends_on: probe-marken.mse-include 2013-12-14\n",
		"probe.mse-game/game": "mse version: 2.0.0\n",
	})

	if len(installer.Befunde) != 0 || len(installer.Pakete) != 1 {
		t.Fatalf("Pakete %+v, Befunde %v", installer.Pakete, installer.Befunde)
	}

	paket := installer.Pakete[0]
	if !paket.Version.Leer() || paket.Gruppe != "Probe/Spiel" || paket.Abhaengigkeiten[0].Version.String() != "2013-12-14" {
		t.Errorf("gelesen %+v", paket)
	}
}

// **Nur benannte Pakete, und nur unter ihrem Namen** — `actual_install`
// nimmt nur Einträge, die mit dem Paketnamen beginnen
// (`src/util/io/package_manager.cpp`, Zeilen 376–400, Fassung f9d9356 vom
// 2021-07-06). Hier ausdrücklich bis zum Schrägstrich: `probe.mse-game-x/`
// gehört nicht zu `probe.mse-game`.
func TestEinInstallerLegtNurAbWasErNennt(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer":                         "package:\n\tname: probe.mse-game\n",
		"probe.mse-game/game":               "mse version: 2.0.0\n",
		"probe.mse-game-x/game":             "fremd",
		"ungenannt.mse-style/style":         "fremd",
		"probe.mse-game/../../ausbruch.txt": "fremd",
	})

	ziel := filepath.Join(t.TempDir(), "innen")
	if err := os.Mkdir(ziel, 0o750); err != nil {
		t.Fatal(err)
	}

	if _, err := installer.Auspacken("probe.mse-game", ziel, 1<<20); err != nil {
		t.Fatal(err)
	}

	var liegt []string
	_ = filepath.WalkDir(filepath.Dir(ziel), func(pfad string, eintrag fs.DirEntry, _ error) error { //nolint:errcheck // nur gezählt
		if !eintrag.IsDir() {
			liegt = append(liegt, pfad)
		}

		return nil
	})

	if !slices.Equal(liegt, []string{filepath.Join(ziel, "game")}) {
		t.Errorf("ausgepackt %v", liegt)
	}

	if len(installer.Befunde) != 1 {
		t.Errorf("Befunde %v — der Ausbruch gehört benannt", installer.Befunde)
	}
}

func TestEinNameMitWegOderOhneEndungWirdUebergangen(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer": "package:\n\tname: ../weg.mse-game\npackage:\n\tname: ohne-endung\n" +
			"package:\n\tname: probe.mse-set\npackage:\n\tname:\npackage:\n\tname: probe.mse-game\n" +
			"package:\n\tname: probe.mse-game\n",
		"probe.mse-game/game": "mse version: 2.0.0\n",
	})

	if len(installer.Pakete) != 1 || installer.Pakete[0].Name != "probe.mse-game" {
		t.Errorf("Pakete %+v", installer.Pakete)
	}

	if len(installer.Befunde) != 5 {
		t.Errorf("Befunde %v — jeder übergangene Eintrag gehört benannt", installer.Befunde)
	}
}

func TestEineUnlesbareFassungImInstallerIstEinBefund(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer":           "package:\n\tname: probe.mse-game\n\tversion: 1.2.3\n\tdepends_on: a.mse-game kaputt\n",
		"probe.mse-game/game": "mse version: 2.0.0\n",
	})

	if len(installer.Pakete) != 1 || len(installer.Befunde) != 2 {
		t.Errorf("Pakete %+v, Befunde %v", installer.Pakete, installer.Befunde)
	}
}

func TestDerDeckelHaeltDasAuspacken(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer":           "package:\n\tname: probe.mse-game\n",
		"probe.mse-game/game": "mse version: 2.0.0\n",
	})

	if _, err := installer.Auspacken("probe.mse-game", t.TempDir(), 4); !errors.Is(err, msepaket.ErrInstallerZuGross) {
		t.Errorf("Fehler %v — mehr als der Deckel wurde ausgepackt, oder er ist nicht als solcher zu erkennen", err)
	}
}

func TestEinPaketOhneDateienLaesstSichNichtAuspacken(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer": "package:\n\tname: probe.mse-game\n",
	})

	if _, err := installer.Auspacken("probe.mse-game", t.TempDir(), 1<<20); err == nil {
		t.Error("ein leeres Paket wurde ausgepackt")
	}

	if _, err := installer.Auspacken("ungenannt.mse-game", t.TempDir(), 1<<20); err == nil {
		t.Error("ein Paket, das der Installer nicht nennt, wurde ausgepackt")
	}
}

func TestOhneHauptdateiIstEsKeinInstaller(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "probe.mse-installer")
	if err := os.WriteFile(pfad, archivAus(t, map[string]string{"probe.mse-game/game": "x"}), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := msepaket.OeffneInstaller(pfad); err == nil {
		t.Error("ein Archiv ohne `installer` wurde als Installer gelesen")
	}

	if _, err := msepaket.OeffneInstaller(filepath.Join(t.TempDir(), "fehlt.mse-installer")); err == nil {
		t.Error("ein fehlender Installer wurde gelesen")
	}
}

// installerAus schreibt die Pakete in einen Installer und öffnet ihn.
func installerAus(t *testing.T, pakete ...*msepaket.Paket) *msepaket.Installer {
	t.Helper()

	var puffer bytes.Buffer
	if err := msepaket.SchreibeInstaller(&puffer, pakete); err != nil {
		t.Fatal(err)
	}

	return oeffneInstaller(t, puffer.Bytes())
}

// installerMit baut einen Installer aus diesen Einträgen und öffnet ihn.
func installerMit(t *testing.T, eintraege map[string]string) *msepaket.Installer {
	t.Helper()

	return oeffneInstaller(t, archivAus(t, eintraege))
}

func oeffneInstaller(t *testing.T, roh []byte) *msepaket.Installer {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.mse-installer")
	if err := os.WriteFile(pfad, roh, 0o600); err != nil {
		t.Fatal(err)
	}

	installer, err := msepaket.OeffneInstaller(pfad)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = installer.Schliessen() }) //nolint:errcheck // nur gelesen

	return installer
}

// archivAus packt Einträge in ein ZIP, in der Ordnung ihrer Namen.
func archivAus(t *testing.T, eintraege map[string]string) []byte {
	t.Helper()

	var puffer bytes.Buffer
	archiv := zip.NewWriter(&puffer)

	namen := make([]string, 0, len(eintraege))
	for name := range eintraege {
		namen = append(namen, name)
	}

	slices.Sort(namen)

	for _, name := range namen {
		schreiber, err := archiv.Create(name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := schreiber.Write([]byte(eintraege[name])); err != nil {
			t.Fatal(err)
		}
	}

	if err := archiv.Close(); err != nil {
		t.Fatal(err)
	}

	return puffer.Bytes()
}
