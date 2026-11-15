package msepaket_test

// **Die Fehlerwege des Installers** — was geschieht, wenn die Platte voll
// ist, eine Datei unlesbar oder ein Eintrag im Archiv beschädigt.

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"errors"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// brechendePlatte nimmt kein Byte an.
type brechendePlatte struct{}

func (brechendePlatte) Write([]byte) (int, error) { return 0, errors.New("die Platte ist voll") }

// zufall gibt Bytes, die sich nicht packen lassen — sie gehen am Puffer des
// Archivs vorbei auf die Platte.
func zufall(t *testing.T, laenge int) string {
	t.Helper()

	roh := make([]byte, laenge)
	if _, err := rand.Read(roh); err != nil {
		t.Fatal(err)
	}

	return string(roh)
}

func TestEineVollePlatteHaeltDenInstallerAn(t *testing.T) {
	kopf := "mse version: 2.0.0\n"
	faelle := []struct {
		name    string
		dateien map[string]string
	}{
		{"beim Abschluss", map[string]string{"include": kopf}},
		{"beim nächsten Eintrag", map[string]string{"a.bin": zufall(t, 10<<10), "include": kopf}},
		{"bei der Hauptdatei", map[string]string{"include": kopf, "z.bin": zufall(t, 10<<10)}},
		{"mitten in einer Datei", map[string]string{"a.bin": zufall(t, 256<<10), "include": kopf}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			paket := legePaketAn(t, t.TempDir(), "probe.mse-include", fall.dateien)

			if err := msepaket.SchreibeInstaller(brechendePlatte{}, []*msepaket.Paket{paket}); err == nil {
				t.Error("der Installer meldete Erfolg auf einer vollen Platte")
			}
		})
	}
}

func TestEineUnlesbareDateiHaeltDenInstallerAn(t *testing.T) {
	faelle := []struct{ name, sperre string }{
		{"eine Datei", "geheim.png"},
		{"ein Ordner", "bilder"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			ort := t.TempDir()
			paket := legePaketAn(t, ort, "probe.mse-include", map[string]string{
				"include": "mse version: 2.0.0\n", "geheim.png": "x", "bilder/a.png": "x",
			})

			gesperrt := filepath.Join(ort, "probe.mse-include", fall.sperre)
			if err := os.Chmod(gesperrt, 0o000); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // nur für das Aufräumen

			if err := msepaket.SchreibeInstaller(io.Discard, []*msepaket.Paket{paket}); err == nil {
				t.Error("eine unlesbare Datei ging still verloren")
			}
		})
	}
}

// **Ohne vollen Namen der kurze, ohne Fassung nur der Name** —
// `PackageDescription` und `Writer::handle(PackageDependency)`
// (`src/util/io/package.cpp`, Zeilen 529–535, Fassung f9d9356 vom
// 2021-07-06).
func TestOhneVollenNamenUndOhneFassung(t *testing.T) {
	paket := legePaketAn(t, t.TempDir(), "probe.mse-game", map[string]string{
		"game": "mse version: 2.0.0\nshort name: Probe\ndepends on: probe-marken.mse-include\n",
	})

	var puffer bytes.Buffer
	if err := msepaket.SchreibeInstaller(&puffer, []*msepaket.Paket{paket, paket}); err != nil {
		t.Fatal(err)
	}

	archiv, err := zip.NewReader(bytes.NewReader(puffer.Bytes()), int64(puffer.Len()))
	if err != nil {
		t.Fatal(err)
	}

	haupt := eintrag(t, archiv, "installer")
	for _, zeile := range []string{"\tfull_name: Probe\n", "\tdepends_on: probe-marken.mse-include\n"} {
		if !strings.Contains(haupt, zeile) {
			t.Errorf("%q fehlt:\n%s", zeile, haupt)
		}
	}

	if strings.Count(haupt, "package:") != 1 {
		t.Errorf("ein doppelt genanntes Paket steht zweimal da:\n%s", haupt)
	}
}

func TestEineUnlesbareHauptdateiIstKeinInstaller(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "probe.mse-installer")
	if err := os.WriteFile(pfad, archivAus(t, map[string]string{"installer": ": ohne Schlüssel\n"}), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := msepaket.OeffneInstaller(pfad); err == nil {
		t.Error("eine unlesbare Hauptdatei wurde angenommen")
	}
}

// roherEintrag ist ein Eintrag, wie ihn ein beschädigtes Archiv trägt.
type roherEintrag struct {
	kopf   zip.FileHeader
	inhalt string
}

func TestBeschaedigteEintraegeWerdenBenanntOderAbgewiesen(t *testing.T) {
	haupt := roherEintrag{kopf: zip.FileHeader{Name: "installer"}, inhalt: "package:\n\tname: probe.mse-game\n"}
	gut := roherEintrag{kopf: zip.FileHeader{Name: "probe.mse-game/game"}, inhalt: "mse version: 2.0.0\n"}

	verweis := zip.FileHeader{Name: "probe.mse-game/weg"}
	verweis.SetMode(fs.ModeSymlink | 0o777)

	faelle := []struct {
		name      string
		weitere   []roherEintrag
		befunde   int
		scheitert bool
	}{
		{"ein Ordnereintrag wird übergangen", []roherEintrag{{kopf: zip.FileHeader{Name: "probe.mse-game/bilder/"}}}, 0, false},
		{"ein Verweis wird benannt", []roherEintrag{{kopf: verweis, inhalt: "/etc/passwd"}}, 1, false},
		{"ein unbekanntes Packverfahren", []roherEintrag{{kopf: zip.FileHeader{Name: "probe.mse-game/a", Method: 99}, inhalt: "x"}}, 0, true},
		{"eine falsche Prüfsumme", []roherEintrag{{kopf: zip.FileHeader{Name: "probe.mse-game/a", CRC32: 1}, inhalt: "x"}}, 0, true},
		{"derselbe Name zweimal", []roherEintrag{gut}, 0, true},
		{"eine Datei, wo ein Ordner sein müsste", []roherEintrag{
			{kopf: zip.FileHeader{Name: "probe.mse-game/a"}, inhalt: "x"},
			{kopf: zip.FileHeader{Name: "probe.mse-game/a/b"}, inhalt: "x"},
		}, 0, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			installer := oeffneInstaller(t, rohesArchiv(t, append([]roherEintrag{haupt, gut}, fall.weitere...)))

			if len(installer.Befunde) != fall.befunde {
				t.Errorf("Befunde %v, erwartet %d", installer.Befunde, fall.befunde)
			}

			_, err := installer.Auspacken("probe.mse-game", t.TempDir(), 1<<20)
			if (err != nil) != fall.scheitert {
				t.Errorf("Fehler %v, erwartet ein Scheitern: %v", err, fall.scheitert)
			}
		})
	}
}

func TestEinFehlendesZielWirdGemeldet(t *testing.T) {
	installer := installerMit(t, map[string]string{
		"installer":           "package:\n\tname: probe.mse-game\n",
		"probe.mse-game/game": "mse version: 2.0.0\n",
	})

	if _, err := installer.Auspacken("probe.mse-game", filepath.Join(t.TempDir(), "fehlt"), 1<<20); err == nil {
		t.Error("in ein fehlendes Ziel wurde ausgepackt")
	}
}

// rohesArchiv schreibt Einträge ungeprüft — Kopf und Bytes, wie sie sind.
func rohesArchiv(t *testing.T, eintraege []roherEintrag) []byte {
	t.Helper()

	var puffer bytes.Buffer
	archiv := zip.NewWriter(&puffer)

	for _, roh := range eintraege {
		kopf := roh.kopf
		kopf.CompressedSize64 = uint64(len(roh.inhalt))
		kopf.UncompressedSize64 = uint64(len(roh.inhalt))

		if kopf.CRC32 == 0 && kopf.Method == zip.Store {
			kopf.CRC32 = crc(roh.inhalt)
		}

		schreiber, err := archiv.CreateRaw(&kopf)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := schreiber.Write([]byte(roh.inhalt)); err != nil {
			t.Fatal(err)
		}
	}

	if err := archiv.Close(); err != nil {
		t.Fatal(err)
	}

	return puffer.Bytes()
}

func crc(inhalt string) uint32 { return crc32.ChecksumIEEE([]byte(inhalt)) }
