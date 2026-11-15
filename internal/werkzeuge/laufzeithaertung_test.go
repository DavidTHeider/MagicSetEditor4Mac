package werkzeuge

// **Gesiegelt wird mit Hardened Runtime** — SEC03
// (#2356),
// Befund A-17, Vermerk 0195.
//
// # Was er prüft
//
// Jeder `codesign`-Aufruf, der siegelt — `--sign` mit **gleich welcher
// Identität**, in **jeder** Skriptsprache unter `scripts/` samt
// Unterordnern und im `Makefile` —, trägt `--options runtime`, und keiner
// trägt `--entitlements`: gemessen ist die Anwendung **ohne** Ausnahmen, und
// eine, die dazukäme, wäre ungemessen.
//
// **Ausgenommen sind die Messprogramme** in [messprogramme], je mit Grund:
// sie siegeln Proben im Gast, deren Flaggen sie gerade messen. Ein Eintrag,
// der nicht mehr siegelt, ist ein Befund — sonst wüchse die Liste still.
//
// **Bis zum Zweitblick über Phase 5 (Befund 2)** las er nur `bauen.sh` und
// `release.sh`; `make bau` siegelte danach ein zweites Mal ohne Laufzeit,
// und `saat.sh` wie `mitliefern.sh` siegelten zwischendurch ohne. **Bis zum
// Nachlauf (Befund 4)** las er nur `scripts/*.sh` und nur `--sign -`.
//
// # Was er nicht prüft
//
//   - **Ob das gebaute Bündel die Flagge trägt** — das prüft `release.sh`
//     selbst am Siegel (`codesign -dv`) an **jedem** ausführbaren Programm im
//     Bündel, und die Hüllenprobe fährt das Abnahmebündel, das `bauen.sh` so
//     siegelt.
//   - **Die Siegel in den Tests** (`*.test.sh`, `*.test.mjs`): sie siegeln
//     Proben und keine Freigabe.
//   - **Einen Aufruf, der über mehrere Zeilen geht** oder `codesign` aus einer
//     Variablen ruft: gelesen wird je Zeile.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// messprogramme siegeln ohne Laufzeit, und mit Absicht.
var messprogramme = map[string]string{
	"bindunggast.mjs":              "siegelt die fremde Bibliothek der Probe I01 ad hoc, damit der Gast ihr Laden misst",
	"quicklookgast.mjs":            "siegelt Probe-App und Erweiterung mit der Sandbox, die der Gast misst",
	"huellenprobe/gast/bindung.pl": "siegelt Leser mit und ohne Härtung — die Flaggen sind das Messergebnis",
}

// siegelt erkennt einen siegelnden Aufruf: `codesign` und `--sign` in einer
// Zeile, als Wort oder als Zeichenkette (`'--sign'`, `"--sign"`).
var siegelt = regexp.MustCompile(`\bcodesign\b.*[\s'"\[,](?:--sign|-s)(?:[\s'"=,]|$)`)

// mitLaufzeit erkennt `--options runtime` — auch als zwei Zeichenketten.
var mitLaufzeit = regexp.MustCompile(`(?:--options|-o)['"]?[\s,=]+['"]?runtime\b`)

func TestGesiegeltWirdMitHardenedRuntime(t *testing.T) {
	t.Parallel()

	wurzel := filepath.Join("..", "..", "scripts")
	gesehen := 0
	gemessen := map[string]bool{}

	var dateien []string

	fehler := filepath.WalkDir(wurzel, func(pfad string, eintrag fs.DirEntry, fehler error) error {
		if fehler != nil || eintrag.IsDir() {
			return fehler
		}

		name := eintrag.Name()
		if strings.Contains(name, ".test.") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		dateien = append(dateien, pfad)

		return nil
	})
	if fehler != nil || len(dateien) < 20 {
		t.Fatalf("die Skripte: %v (%d)", fehler, len(dateien))
	}

	dateien = append(dateien, filepath.Join("..", "..", "Makefile"))

	for _, pfad := range dateien {
		name, fehler := filepath.Rel(wurzel, pfad)
		if fehler != nil {
			t.Fatal(fehler)
		}

		name = filepath.ToSlash(name)

		roh, fehler := os.ReadFile(pfad) //nolint:gosec // Datei im Repositorium
		if fehler != nil {
			t.Fatal(fehler)
		}

		for nummer, zeile := range strings.Split(string(roh), "\n") {
			code := strings.TrimSpace(zeile)
			if strings.HasPrefix(code, "#") || strings.HasPrefix(code, "//") || !siegelt.MatchString(code) {
				continue
			}

			if _, ausgenommen := messprogramme[name]; ausgenommen {
				gemessen[name] = true

				continue
			}

			gesehen++

			if !mitLaufzeit.MatchString(code) {
				t.Errorf("%s:%d siegelt ohne Hardened Runtime: %s", name, nummer+1, code)
			}

			if strings.Contains(code, "--entitlements") {
				t.Errorf("%s:%d siegelt mit Entitlements — gemessen ist die Anwendung ohne: %s", name, nummer+1, code)
			}
		}
	}

	for name, grund := range messprogramme {
		if !gemessen[name] {
			t.Errorf("%s ist als Messprogramm ausgenommen (%s), siegelt aber nicht mehr — der Eintrag fällt", name, grund)
		}
	}

	if gesehen < 4 {
		t.Errorf("nur %d Siegel gefunden — erwartet bauen.sh, release.sh, saat.sh und mitliefern.sh", gesehen)
	}
}
