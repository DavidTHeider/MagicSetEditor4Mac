//go:build darwin || linux

package service_test

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerStilspeicherSchliesstGeleseneArchive(t *testing.T) {
	vorherigeGC := debug.SetGCPercent(-1)
	t.Cleanup(func() { debug.SetGCPercent(vorherigeGC) })

	wurzel := t.TempDir()
	legePaketarchiv(t, filepath.Join(wurzel, "probe-stil.mse-style"), "style",
		"mse version: 2.0.0\ngame: probe\ncard width: 375\ncard height: 523\n")
	legePaketarchiv(t, filepath.Join(wurzel, "probe.mse-game"), "game",
		"mse version: 2.0.0\ninit script:\n\t\"bereit\"\n")

	koepfe := make([]domain.SetHead, 80)
	for i := range koepfe {
		koepfe[i] = domain.SetHead{ID: int64(i + 1), Style: "probe-stil", Game: "probe"}
	}

	vorherigeGrenze := dateigrenzeSenken(t)
	t.Cleanup(func() {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &vorherigeGrenze); err != nil {
			t.Errorf("Grenze offener Dateien wiederherstellen: %v", err)
		}
	})

	speicher := service.NeuerStilspeicher(&koepfequelle{koepfe: koepfe}, wurzel)
	for _, kopf := range koepfe {
		if _, befunde, da := speicher.StilFuerSet(t.Context(), kopf.ID); !da {
			t.Fatalf("Stil für Set %d fehlt: %v", kopf.ID, befunde)
		}
	}
}

func dateigrenzeSenken(t *testing.T) syscall.Rlimit {
	t.Helper()

	var grenze syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &grenze); err != nil {
		t.Skipf("Grenze offener Dateien nicht lesbar: %v", err)
	}
	vorherigeGrenze := grenze
	if grenze.Cur > 64 {
		grenze.Cur = 64
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &grenze); err != nil {
		t.Skipf("Grenze offener Dateien nicht absenkbar: %v", err)
	}

	return vorherigeGrenze
}

func legePaketarchiv(t *testing.T, pfad, hauptdatei, inhalt string) {
	t.Helper()

	datei, err := os.Create(pfad) // #nosec G304 -- Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("Paketarchiv anlegen: %v", err)
	}
	archiv := zip.NewWriter(datei)
	eintrag, err := archiv.Create(hauptdatei)
	if err != nil {
		t.Fatalf("Hauptdatei anlegen: %v", err)
	}
	if _, err = eintrag.Write([]byte(inhalt)); err != nil {
		t.Fatalf("Hauptdatei schreiben: %v", err)
	}
	if err = errors.Join(archiv.Close(), datei.Close()); err != nil {
		t.Fatalf("Paketarchiv schließen: %v", err)
	}
}
