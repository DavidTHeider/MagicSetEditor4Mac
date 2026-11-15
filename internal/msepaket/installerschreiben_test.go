package msepaket_test

// **Ein Installer, wie MSE2 ihn schreibt** — FMT-02.AK-04
// (#1691).
//
// Die erwartete Hauptdatei ist **gemessen**: MSE2 (Fassung f9d9356 vom
// 2021-07-06) hat am 2026-09-29 mit `--create-installer` aus einem Include
// und einem Spiel einen Installer geschrieben — BOM, Schlüssel mit
// Unterstrich, `mse_version: 0.3.7`, ein `package:`-Block je Paket und
// `depends_on:` in einer Zeile. Die Pakete hier sind selbst geschrieben
// (§11); Gestalt und Regeln sind die aus jenem Lauf.

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestEinInstallerTraegtDiePaketeWieMSE2(t *testing.T) {
	ort := t.TempDir()
	marken := legePaketAn(t, ort, "probe-marken.mse-include", map[string]string{
		"include":        "mse version: 2.0.0\nfull name: Probe Marken\nversion: 2026-09-01\n",
		"marke.png":      "png",
		"bilder/alt.png": "alt",
		// **Was der Dateimanager hinlegt, reist nicht mit** — bei MSE2
		// `Thumbs.db`, hier auch `.DS_Store` (adapted, §15 Satz 9).
		"Thumbs.db":        "x",
		".DS_Store":        "x",
		".svn/entries":     "x",
		"bilder/.DS_Store": "x",
	})
	spiel := legePaketAn(t, ort, "probe.mse-game", map[string]string{
		"game": "mse version: 2.0.0\nshort name: Probe\nfull name: Probespiel\n" +
			"installer group: Probe/Spieldateien\nicon: bild.png\nposition hint: 02\n" +
			"version: 2026-09-02\ndepends on: probe-marken.mse-include 2026-09-01\n",
		"bild.png": "bild",
	})

	var puffer bytes.Buffer
	if err := msepaket.SchreibeInstaller(&puffer, []*msepaket.Paket{marken, spiel}); err != nil {
		t.Fatal(err)
	}

	archiv, err := zip.NewReader(bytes.NewReader(puffer.Bytes()), int64(puffer.Len()))
	if err != nil {
		t.Fatal(err)
	}

	namen := make([]string, 0, len(archiv.File))
	for _, datei := range archiv.File {
		namen = append(namen, datei.Name)
	}

	erwartet := []string{
		"probe-marken.mse-include/bilder/alt.png",
		"probe-marken.mse-include/include",
		"probe-marken.mse-include/marke.png",
		"probe.mse-game/bild.png",
		"probe.mse-game/game",
		"installer",
	}
	if !slices.Equal(namen, erwartet) {
		t.Errorf("Einträge\n%q\nerwartet\n%q", namen, erwartet)
	}

	haupt := "\ufeffmse_version: 0.3.7\nshort_name: \nfull_name: \nicon: \nposition_hint: 100000\n" +
		"installer_group: \nversion: 0.0.0\ncompatible_version: 0.0.0\n" +
		"package:\n\tname: probe-marken.mse-include\n\tversion: 2026-09-01\n" +
		"\tshort_name: probe-marken\n\tfull_name: Probe Marken\n\ticon_url: \n" +
		"\tinstaller_group: unclassified/probe/probe-marken\n\tposition_hint: 100000\n\tdescription: \n" +
		"package:\n\tname: probe.mse-game\n\tversion: 2026-09-02\n" +
		"\tshort_name: Probe\n\tfull_name: Probespiel\n\ticon_url: bild.png\n" +
		"\tinstaller_group: Probe/Spieldateien\n\tposition_hint: 2\n\tdescription: \n" +
		"\tdepends_on: probe-marken.mse-include 2026-09-01\n"
	if gelesen := eintrag(t, archiv, "installer"); gelesen != haupt {
		t.Errorf("Hauptdatei\n%q\nerwartet\n%q", gelesen, haupt)
	}

	if gelesen := eintrag(t, archiv, "probe-marken.mse-include/bilder/alt.png"); gelesen != "alt" {
		t.Errorf("Inhalt %q — die Datei kam verändert an", gelesen)
	}
}

// **Die Gruppe, wenn das Paket keine nennt** — `PackageDescription`
// (`src/data/installer.cpp`, Zeilen 204–219, Fassung f9d9356 vom
// 2021-07-06): der Name bis zum letzten Bindestrich, Bindestriche als
// Schrägstriche, darunter „Game files" für ein Spiel und sonst der Kurzname.
func TestDieGruppeOhneAngabe(t *testing.T) {
	faelle := []struct{ name, paket, kopf, gruppe string }{
		{"Spiel ohne Bindestrich", "probe.mse-game", "game", "unclassified/probe/Game files"},
		{"Stil mit zwei Bindestrichen", "probe-alt-neu.mse-style", "style", "unclassified/probe/alt/Kurz"},
		{"Symbolschrift", "probe-zeichen.mse-symbol-font", "symbol-font", "unclassified/probe/Kurz"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			paket := legePaketAn(t, t.TempDir(), fall.paket, map[string]string{
				fall.kopf: "mse version: 2.0.0\nshort name: Kurz\n",
			})

			var puffer bytes.Buffer
			if err := msepaket.SchreibeInstaller(&puffer, []*msepaket.Paket{paket}); err != nil {
				t.Fatal(err)
			}

			archiv, err := zip.NewReader(bytes.NewReader(puffer.Bytes()), int64(puffer.Len()))
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Contains([]byte(eintrag(t, archiv, "installer")), []byte("\tinstaller_group: "+fall.gruppe+"\n")) {
				t.Errorf("Gruppe nicht %q:\n%s", fall.gruppe, eintrag(t, archiv, "installer"))
			}
		})
	}
}

func TestEinLeererInstallerWirdAbgewiesen(t *testing.T) {
	if err := msepaket.SchreibeInstaller(io.Discard, nil); err == nil {
		t.Error("ein Installer ohne Paket wurde geschrieben — MSE2 verlangt eines")
	}
}

// legePaketAn schreibt ein Verzeichnispaket und öffnet es.
func legePaketAn(t *testing.T, ort, name string, dateien map[string]string) *msepaket.Paket {
	t.Helper()

	wurzel := filepath.Join(ort, name)
	for datei, inhalt := range dateien {
		pfad := filepath.Join(wurzel, filepath.FromSlash(datei))
		if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	paket, err := msepaket.Oeffne(wurzel)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = paket.Schliessen() }) //nolint:errcheck // nur gelesen

	return paket
}

// eintrag liest einen Eintrag des Archivs.
func eintrag(t *testing.T, archiv *zip.Reader, name string) string {
	t.Helper()

	datei, err := archiv.Open(name)
	if err != nil {
		t.Fatalf("%s fehlt: %v", name, err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	roh, err := io.ReadAll(datei)
	if err != nil {
		t.Fatal(err)
	}

	return string(roh)
}
