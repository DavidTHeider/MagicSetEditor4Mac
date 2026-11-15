package sicherung_test

// **Das gedeckelte Entpacken** — SEC-03
// (#2138).
// Die Grenzen selbst prüft `internal/austausch`, der einzige Aufrufer; hier
// stehen die Wege, die jeder Aufrufer nimmt: eine Datei, die kein Archiv ist,
// ein Teil, das fehlt, und der gewöhnliche Fall.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

const gedeckeltesManifest = "aussen.json"

func gedeckeltesArchiv(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "probe.zip")
	if fehler := sicherung.PackenAls(pfad, gedeckeltesManifest, []sicherung.Teil{
		{Name: "klein.json", Inhalt: []byte("{}")},
		{Name: "anderes.json", Inhalt: []byte("[]")},
	}, sicherung.Angaben{}); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

func TestGedeckeltWirdNurGenommenWasGewolltIst(t *testing.T) {
	t.Parallel()

	ziel := t.TempDir()
	if _, fehler := sicherung.EntpackenGedeckelt(gedeckeltesArchiv(t), ziel,
		gedeckeltesManifest, 4096, map[string]int64{"klein.json": 16}); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "klein.json")); fehler != nil {
		t.Errorf("das gewollte Teil fehlt: %v", fehler)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "anderes.json")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("ein nicht gewolltes Teil liegt da: %v", fehler)
	}
}

func TestGedeckeltWeistKeinArchivUndEinFehlendesTeilAb(t *testing.T) {
	t.Parallel()

	keinArchiv := filepath.Join(t.TempDir(), "kein.zip")
	if fehler := os.WriteFile(keinArchiv, []byte("kein Archiv"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := sicherung.EntpackenGedeckelt(keinArchiv, t.TempDir(),
		gedeckeltesManifest, 4096, map[string]int64{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("kein Archiv: %v", fehler)
	}

	ohneTeil := gedeckeltesArchiv(t)
	entfernen(t, ohneTeil, "klein.json")

	ziel := t.TempDir()
	if _, fehler := sicherung.EntpackenGedeckelt(ohneTeil, ziel,
		gedeckeltesManifest, 4096, map[string]int64{"klein.json": 16, "anderes.json": 16}); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein fehlendes Teil: %v", fehler)
	}

	// **Und nichts davor geschrieben** — auch das andere, vorhandene nicht.
	if eintraege, lesefehler := os.ReadDir(ziel); lesefehler != nil || len(eintraege) != 0 {
		t.Errorf("geschrieben wurde %v (%v)", eintraege, lesefehler)
	}
}

func TestGedeckeltWeistEinTeilUeberSeinerGrenzeAb(t *testing.T) {
	t.Parallel()

	if _, fehler := sicherung.EntpackenGedeckelt(gedeckeltesArchiv(t), t.TempDir(),
		gedeckeltesManifest, 4096, map[string]int64{"klein.json": 1}); !errors.Is(fehler, sicherung.ErrUeberDemDeckel) {
		t.Errorf("zwei Byte über einer Grenze von einem: %v", fehler)
	}
}
