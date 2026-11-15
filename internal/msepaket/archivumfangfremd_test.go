//go:build fremdpakete

// **Kein vorhandenes Paket fällt unter die Ressourcenregel** —
// #1886,
// Abnahme 3.
//
// Gelesen werden die zehn kanonischen Sets (`MSE_KANON`, Archive) und jedes
// Paket im vollen Pack (`MSE_FREMDPAKETE`). **Die Pakete des Packs sind
// Verzeichnisse** — für sie gilt die Regel nicht, weil kein Archivverzeichnis
// gelesen wird; der Lauf zählt sie mit, damit „keines abgewiesen" nicht
// „keines angesehen" heißt.
//
//	MSE_KANON=… MSE_FREMDPAKETE=~/full-magic-pack/data \
//	  go test -tags fremdpakete ./internal/msepaket -run Ressourcenregel -v
package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

func TestKeinVorhandenesPaketFaelltUnterDieRessourcenregel(t *testing.T) {
	kanon, pack := os.Getenv("MSE_KANON"), os.Getenv("MSE_FREMDPAKETE")
	if kanon == "" || pack == "" {
		t.Skip("MSE_KANON oder MSE_FREMDPAKETE ist nicht gesetzt")
	}

	sets, fehler := filepath.Glob(filepath.Join(kanon, "*.mse-set"))
	if fehler != nil || len(sets) == 0 {
		t.Fatalf("keine kanonischen Sets unter %s: %v", kanon, fehler)
	}

	eintraege, fehler := os.ReadDir(pack)
	if fehler != nil {
		t.Fatalf("den Pack lesen: %v", fehler)
	}

	pakete := append([]string(nil), sets...)
	for _, eintrag := range eintraege {
		if strings.Contains(eintrag.Name(), ".mse-") {
			pakete = append(pakete, filepath.Join(pack, eintrag.Name()))
		}
	}

	archive := 0

	for _, pfad := range pakete {
		paket, oeffnen := msepaket.Oeffne(pfad)
		if errors.Is(oeffnen, msepaket.ErrArchivZuGross) {
			t.Errorf("%s fällt unter die Regel: %v", filepath.Base(pfad), oeffnen)

			continue
		}

		if oeffnen != nil {
			continue
		}

		if angaben, _ := os.Stat(pfad); angaben != nil && !angaben.IsDir() { //nolint:errcheck,gosec // nur gezählt; der Pfad kommt aus dem eigenen Pack
			archive++
		}

		_ = paket.Schliessen() //nolint:errcheck // nur geöffnet
	}

	t.Logf("%d Pakete geöffnet, davon %d Archive", len(pakete), archive)

	if archive < len(sets) {
		t.Errorf("nur %d von %d kanonischen Sets gingen als Archiv auf", archive, len(sets))
	}
}
