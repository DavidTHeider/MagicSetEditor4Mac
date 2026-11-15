//go:build fremdpakete

// **Die Packarten eines großen Sets, im Prozess und profilierbar** —
// #2466.
// Derselbe Weg wie `mse packarten <set>`, nur ohne Prozessstart, damit
// `go test -cpuprofile` ihn aufzeichnet:
//
//	MSE_FREMDPAKETE=/Users/david/full-magic-pack/data \
//	MSE_KANON="$HOME/Library/Application Support/MagicSetEditor4Mac/kanonische-sets" \
//	  go test -tags fremdpakete ./cmd/mse -run TestDiePackartenEinesGrossenSets -count=1 -cpuprofile packarten.pprof
//
// **Was er hält**: dass der Ruf gelingt und für *Blue* Karten in den
// Packarten stehen. **Was er nicht hält**: die Zeit — sie steht im Protokoll.
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiePackartenEinesGrossenSets(t *testing.T) {
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

	if code, _, fehler := fahreAn(t, bestand, "oeffnen", filepath.Join(kanon, "Blue.mse-set")); code != CodeGut {
		t.Fatalf("Blue öffnen: %d %s", code, fehler)
	}

	beginn := time.Now()

	code, ausgabe, fehler := fahreAn(t, bestand, "packarten", "1")
	if code != CodeGut {
		t.Fatalf("packarten: %d %s", code, fehler)
	}

	t.Logf("packarten über Blue: %v\n%s", time.Since(beginn), ausgabe)
}
