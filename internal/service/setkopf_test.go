package service_test

// **Die fünf Angaben des Setkopfs werden gedeutet** —
// #1325.
//
// # Der Befund
//
// Jedes der zehn kanonischen Sets zeigte im Kasten „Nicht gedeutet" genau
// dieselben fünf Zeilen: `mse_version`, `game_version`, `stylesheet_version`,
// `version_control` und `apprentice_code`. Am 2026-09-03 über die Bindung
// nachgezählt — fünf von fünf, in allen zehn.
//
// **Das ist keine unverstandene Zeile, sondern ein Feld, das MSE2 kennt.**
// Nachgelesen in `src/data/set.cpp` (Fassung f9d9356 vom 2021-07-06):
//
//   - `mse_version` schreibt MSE bei **jedem** Speichern (die eigene Fassung,
//     `src/util/io/writer.cpp` Zeile 26) und warnt beim Lesen, wenn die Datei
//     von einer neueren stammt (`reader.cpp` Zeilen 38–45).
//   - `game_version` und `stylesheet_version` sind die Fassungen der Pakete
//     zum Zeitpunkt des Speicherns; MSE warnt, wenn das installierte Paket
//     älter ist (`set.cpp` Zeilen 169–176, gerufen in 190 und 194).
//   - `version_control` ist das Versionskontrollsystem des Sets
//     (`set.hpp` Zeile 62) — in jeder gesehenen Datei `none`.
//   - `apprentice_code` ist ein Textfeld für den Apprentice-Export, nur für
//     Magic (`set.hpp` Zeile 58).
//
// # Was hier geprüft wird
//
// Dass sie im Bestand ankommen. **Dass sie nicht mehr im Rohweg stehen**,
// prüft der Fall darunter: eine Zeile, die gedeutet **und** roh mitreist,
// stünde zweimal da.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// importiertesSet schreibt die Bytes als Set und gibt, was der Import daraus
// macht.
//
// **Als echte Datei und nicht als Zeichenkette**: der Importeur liest ein
// Paket, und ein Paket ist ein Verzeichnis oder eine Datei — §3 verlangt für
// `msepaket` echte Paketdateien statt nachgebauter Zeichenketten.
func importiertesSet(t *testing.T, inhalt []byte) *domain.Set {
	t.Helper()

	set, _ := importiertesSetMitRoh(t, inhalt)

	return set
}

// importiertesSetMitRoh gibt zusätzlich, was roh mitreist.
func importiertesSetMitRoh(t *testing.T, inhalt []byte) (*domain.Set, []domain.RawBlock) {
	t.Helper()

	verzeichnis := t.TempDir()
	pfad := filepath.Join(verzeichnis, "probe.mse-set")

	if err := os.MkdirAll(pfad, 0o750); err != nil {
		t.Fatalf("das Paketverzeichnis anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pfad, "set"), inhalt, 0o600); err != nil {
		t.Fatalf("die Setdatei schreiben: %v", err)
	}

	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), pfad); err != nil {
		t.Fatalf("Import: %v", err)
	}

	return speicher.set, speicher.roh
}

// istZeileVon sagt, ob ein Rohblock mit diesem Schlüssel beginnt.
func istZeileVon(block domain.RawBlock, schluessel string) bool {
	return strings.HasPrefix(strings.TrimSpace(block.Text), schluessel+":")
}

// setMitKopf gibt die Bytes eines Sets mit allen fünf Kopfangaben.
func setMitKopf() []byte {
	return []byte(`mse_version: 2.1.2
game: magic
game_version: 2020-04-25
stylesheet: m15
stylesheet_version: 2020-09-04
version_control:
	type: none
apprentice_code:
set_info:
	title: Probe
card:
	name: Eine Karte
`)
}

func TestDerSetkopfWirdGedeutet(t *testing.T) {
	t.Parallel()

	set := importiertesSet(t, setMitKopf())

	faelle := []struct {
		name     string
		gelesen  string
		erwartet string
	}{
		{"mse_version", set.Kopf.MSEFassung, "2.1.2"},
		{"game_version", set.Kopf.Spielfassung, "2020-04-25"},
		{"stylesheet_version", set.Kopf.Stilfassung, "2020-09-04"},
		{"version_control", set.Kopf.Versionskontrolle, "none"},
		{"apprentice_code", set.Kopf.Apprenticecode, ""},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if fall.gelesen != fall.erwartet {
				t.Errorf("%s ist %q, erwartet %q", fall.name, fall.gelesen, fall.erwartet)
			}
		})
	}
}

// **Und sie stehen nicht mehr im Rohweg.**
//
// Eine Zeile, die gedeutet **und** roh mitreist, stünde zweimal da: im Bestand
// und im Kasten „Nicht gedeutet". Das ist die Zusage aus `importschluessel.go`
// — was in `gespeichert` steht, wird abgelegt und nicht roh geführt.
func TestDerSetkopfStehtNichtMehrImRohweg(t *testing.T) {
	t.Parallel()

	_, roh := importiertesSetMitRoh(t, setMitKopf())

	for _, block := range roh {
		for _, schluessel := range []string{
			"mse_version", "game_version", "stylesheet_version",
			"version_control", "apprentice_code",
		} {
			if istZeileVon(block, schluessel) {
				t.Errorf("%q steht noch im Rohweg: %q", schluessel, block.Text)
			}
		}
	}
}

// **Ein Set ohne diese Angaben bleibt leer** — und wird nicht erfunden.
//
// `magic-set-generator` schreibt sie nicht alle (§11), und ein Set, das in
// dieser Anwendung entsteht, hat keine davon. Ein erfundener Standardwert wäre
// eine Falschangabe über eine fremde Datei.
func TestEinSetOhneKopfangabenBekommtKeine(t *testing.T) {
	t.Parallel()

	set := importiertesSet(t, []byte("game: magic\nset_info:\n\ttitle: Ohne\n"))

	if set.Kopf != (domain.Setkopf{}) {
		t.Errorf("der Kopf ist %+v, erwartet war er leer", set.Kopf)
	}
}
