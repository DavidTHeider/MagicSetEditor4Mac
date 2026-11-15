package service

// **Ein gescheiterter Wechsel hinterlässt den ganzen Stand von vorher** —
// Befunde Z1 und Z2 des Zweitblicks über Phase 6.5 (DOM-01,
// #1696).
//
// Nicht nur die aktuelle Fassung: auch die **vorige**, die sich bis dahin
// zurückholen ließ, und der Katalog. Ein Fehler, der nach dem Dateitausch
// kommt, nimmt den Tausch zurück — sonst meldete er einen unveränderten
// Stand, während auf der Platte der neue liegt.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// vorigeLiegtDa sichert zu, dass die vorige Fassung in `version` samt
// Vermerk da ist.
func vorigeLiegtDa(t *testing.T, verwaltet, version string) {
	t.Helper()

	vorige := filepath.Join(verwaltet+"-vorige", "fremd.mse-style")

	roh, err := os.ReadFile(filepath.Join(vorige, "style")) //nolint:gosec // ein Pfad dieses Laufs
	if err != nil || !strings.Contains(string(roh), "version: "+version) {
		t.Errorf("die vorige Fassung steht nicht in %s da: %q (%v)", version, roh, err)
	}

	if _, err := os.Stat(vorige + ".vermerk"); err != nil {
		t.Errorf("der Vermerk der vorigen Fassung fehlt: %v", err)
	}
}

// zweimal bindet Fassung 2 ein — Fassung 1 ist dann die vorige — und legt
// eine dritte bereit.
func zweimal(t *testing.T) (verwalter *Paketverwalter, katalog *innerkatalog, verwaltet, dritte string) {
	t.Helper()

	verwalter, katalog, verwaltet, neu := innerplatz(t)
	if _, err := verwalter.Ersetzen(t.Context(), neu, false); err != nil {
		t.Fatalf("ersetzen: %v", err)
	}

	return verwalter, katalog, verwaltet, innenPaket(t, t.TempDir(), "2026-03-03")
}

func TestEinGescheitertesErsetzenBehaeltDieVorigeFassung(t *testing.T) {
	verwalter, _, verwaltet, dritte := zweimal(t)
	if err := os.WriteFile(filepath.Join(dritte, "unlesbar"), []byte("x"), 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
		t.Error("der Fehler ging verloren")
	}

	liegtDa(t, verwaltet, "2026-02-02")
	vorigeLiegtDa(t, verwaltet, "2026-01-01")
}

// **Und scheitert schon das Beiseitelegen**, bleibt beides, wie es war.
func TestEinGescheitertesBeiseitelegenAendertNichts(t *testing.T) {
	verwalter, _, verwaltet, dritte := zweimal(t)
	verwalter.umbenennen = scheitertBeim(1)

	if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
		t.Error("der Fehler ging verloren")
	}

	liegtDa(t, verwaltet, "2026-02-02")
	vorigeLiegtDa(t, verwaltet, "2026-01-01")
}

func TestEinKlemmenderKatalogNimmtDenTauschZurueck(t *testing.T) {
	t.Run("beim Ersetzen", func(t *testing.T) {
		verwalter, katalog, verwaltet, dritte := zweimal(t)
		katalog.abgleichfehler = errors.New("klemmt")

		if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-02-02")
		vorigeLiegtDa(t, verwaltet, "2026-01-01")
	})

	t.Run("beim Zurückholen", func(t *testing.T) {
		verwalter, katalog, verwaltet, _ := zweimal(t)
		katalog.abgleichfehler = errors.New("klemmt")

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-02-02")
		vorigeLiegtDa(t, verwaltet, "2026-01-01")
	})

	// **Dieselbe Klasse**: eine vorige Fassung, die sich nach dem Tausch
	// nicht lesen lässt, ist ein Fehler nach dem Tausch.
	//
	// **Und der Katalog sieht sie erst, wenn sie lesbar ist** — Befund N1 des
	// Nachlaufs: der Abgleich ist vollständig, ein fehlendes Paket fiele aus
	// ihm heraus, und der Rückweg brächte nur die Dateien zurück.
	t.Run("eine unlesbare vorige Fassung beim Zurückholen", func(t *testing.T) {
		verwalter, katalog, verwaltet, _ := zweimal(t)
		if err := os.Remove(filepath.Join(verwaltet+"-vorige", "fremd.mse-style", "style")); err != nil {
			t.Fatalf("räumen: %v", err)
		}

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler ist %v", err)
		}

		liegtDa(t, verwaltet, "2026-02-02")

		if !slices.ContainsFunc(katalog.pakete, func(p domain.Paket) bool { return p.Name == "fremd" }) {
			t.Error("das Paket fiel aus dem Katalog")
		}
	})
}
