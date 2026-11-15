//go:build fremdpakete

// **Ein skriptlastiger Export, im Prozess, profilierbar und gegen MSE2
// gehalten** — E04 (#2241).
//
// Derselbe Weg wie im Vergleich gegen MSE2 (`docs/LEISTUNGSVERGLEICH.md`,
// „Export `magic-namecheck`"): `mse vorlagen <set> ausfuehren
// magic-namecheck.mse-export-template <ziel>` — nur ohne Prozessstart, damit
// `go test -cpuprofile` ihn aufzeichnet (`docs/belege/e04-ausfuhr.md`):
//
//	MSE_FREMDPAKETE=/Users/david/full-magic-pack/data \
//	MSE_KANON="$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets" \
//	  go test -tags fremdpakete ./cmd/mse -run TestDerSkriptlastigeExport -count=1 -cpuprofile ausfuhr.pprof
//
// **Was er hält**: die Ausgabe ist Byte für Byte die von MSE2 — gemessen am
// 2026-10-11 mit `magicseteditor --export`. **Was er nicht hält**: die Zeit.
// Sie steht im Protokoll; eine Grenze darüber machte jeden belasteten Lauf
// rot, ohne etwas über den Export zu sagen.
//
// **Eigenes Zuhause**: das Pack hängt als Verweis unter dessen Paketen, die
// Sets werden in einen frischen Bestand eingelesen.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// wieMSE2 sind die Abdrücke der Ausgabe von MSE2 (`f9d9356`) für
// `magic-namecheck` über die kanonischen Sets.
var wieMSE2 = []struct {
	set, sha256 string
}{
	{"Archenemy", "71650f70e66988a5"},
	{"Black", "9fe74db3b3b562d7"},
}

func TestDerSkriptlastigeExport(t *testing.T) {
	pakete, kanon := os.Getenv("MSE_FREMDPAKETE"), os.Getenv("MSE_KANON")
	if pakete == "" || kanon == "" {
		t.Skip("MSE_FREMDPAKETE und MSE_KANON nennen die Pakete und die kanonischen Sets")
	}

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	ort := filepath.Join(zuhause, "Library", "Application Support", "MagicSetEditor4Mac", "pakete")
	if fehler := os.MkdirAll(filepath.Dir(ort), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.Symlink(pakete, ort); fehler != nil {
		t.Fatal(fehler)
	}

	bestand := frischerBestand(t)

	for kennung, fall := range wieMSE2 {
		if code, _, fehler := fahreAn(t, bestand, "oeffnen", filepath.Join(kanon, fall.set+".mse-set")); code != CodeGut {
			t.Fatalf("%s öffnen: %d %s", fall.set, code, fehler)
		}

		ziel := filepath.Join(t.TempDir(), "ausgabe.txt")
		beginn := time.Now()

		code, _, fehler := fahreAn(t, bestand, "vorlagen", strconv.Itoa(kennung+1), "ausfuehren",
			"magic-namecheck.mse-export-template", ziel)
		if code != CodeGut {
			t.Fatalf("%s ausführen: %d %s", fall.set, code, fehler)
		}

		dauer := time.Since(beginn)

		inhalt, lesefehler := os.ReadFile(ziel) //nolint:gosec // das Ziel liegt in t.TempDir
		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		summe := sha256.Sum256(inhalt)
		abdruck := hex.EncodeToString(summe[:])[:16]
		t.Logf("%s: %v, %d Byte, sha256 %s", fall.set, dauer.Round(time.Millisecond), len(inhalt), abdruck)

		if abdruck != fall.sha256 {
			t.Errorf("%s: die Ausgabe weicht von MSE2 ab — %s statt %s", fall.set, abdruck, fall.sha256)
		}
	}
}
