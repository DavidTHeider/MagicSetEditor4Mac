package service

// **Was ein Wechsel neben der vorigen Fassung liegen lässt** — Befunde N2
// bis N4 des Nachlaufs über Phase 6.5 (DOM-01,
// #1696).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// unkopierbar lässt das Einlegen einer Fassung scheitern — nach dem
// Beiseitelegen und dem Aufheben.
func unkopierbar(t *testing.T, paket string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(paket, "unlesbar"), []byte("x"), 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}
}

// liegenlassen stellt den Rest eines abgebrochenen Wechsels her: die vorige
// Fassung liegt unter `rest` neben ihrem Ort, auf Wunsch samt Vermerk.
func liegenlassen(t *testing.T, verwaltet, rest string, mitVermerk bool) {
	t.Helper()

	vorige := filepath.Join(verwaltet+"-vorige", "fremd.mse-style")
	if err := os.Rename(vorige, vorige+rest); err != nil {
		t.Fatalf("liegen lassen: %v", err)
	}

	if mitVermerk {
		if err := os.Rename(vorige+".vermerk", vorige+rest+".vermerk"); err != nil {
			t.Fatalf("liegen lassen: %v", err)
		}
	}
}

// N2: **ohne ihren Vermerk ist die vorige Fassung keine** — scheitert sein
// Beiseitelegen, kommt sie zurück, und nichts ändert sich.
func TestEinGescheiterterVermerkLaesstDieVorigeFassungStehen(t *testing.T) {
	verwalter, _, verwaltet, dritte := zweimal(t)
	verwalter.umbenennen = scheitertBeim(2)

	if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
		t.Error("der Fehler ging verloren")
	}

	liegtDa(t, verwaltet, "2026-02-02")
	vorigeLiegtDa(t, verwaltet, "2026-01-01")
}

// **Eine vorige Fassung ohne Vermerk** — etwa nach einem Abbruch zwischen
// Aufheben und Vermerk — hält den Wechsel nicht auf.
func TestEineVorigeFassungOhneVermerkWirdGehalten(t *testing.T) {
	verwalter, _, verwaltet, dritte := zweimal(t)
	if err := os.Remove(filepath.Join(verwaltet+"-vorige", "fremd.mse-style.vermerk")); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	unkopierbar(t, dritte)

	if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
		t.Error("der Fehler ging verloren")
	}

	liegtDa(t, verwaltet, "2026-02-02")

	if _, err := os.Stat(filepath.Join(verwaltet+"-vorige", "fremd.mse-style", "style")); err != nil {
		t.Errorf("die vorige Fassung kam nicht zurück: %v", err)
	}
}

// N3: **ein Rest eines abgebrochenen Wechsels wird geholt, nicht geräumt** —
// er ist die einzige erreichbare vorige Fassung. Liegt am Ort schon eine,
// hält der Wechsel an und nennt den Rest.
func TestEinRestEinesAbgebrochenenWechsels(t *testing.T) {
	t.Run("beim Ersetzen kommt er zurück", func(t *testing.T) {
		verwalter, _, verwaltet, dritte := zweimal(t)
		liegenlassen(t, verwaltet, ".beiseite", true)
		unkopierbar(t, dritte)

		if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-02-02")
		vorigeLiegtDa(t, verwaltet, "2026-01-01")
	})

	// Beim Zurückholen wandert nur das Verzeichnis, der Vermerk bleibt.
	t.Run("beim Zurückholen kommt er zurück", func(t *testing.T) {
		verwalter, _, verwaltet, _ := zweimal(t)
		liegenlassen(t, verwaltet, ".tausch", false)

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); err != nil {
			t.Fatalf("zurück: %v", err)
		}

		liegtDa(t, verwaltet, "2026-01-01")
		vorigeLiegtDa(t, verwaltet, "2026-02-02")
	})

	t.Run("liegt schon eine da, hält der Wechsel an", func(t *testing.T) {
		for _, fall := range []struct {
			rest    string
			wechsel func(v *Paketverwalter, dritte string) error
		}{
			{".beiseite", func(v *Paketverwalter, dritte string) error {
				_, err := v.Ersetzen(t.Context(), dritte, false)

				return err
			}},
			{".tausch", func(v *Paketverwalter, _ string) error {
				_, err := v.Zurueck(t.Context(), "mse-style", "fremd")

				return err
			}},
		} {
			t.Run(fall.rest, func(t *testing.T) {
				verwalter, _, verwaltet, dritte := zweimal(t)
				rest := filepath.Join(verwaltet+"-vorige", "fremd.mse-style"+fall.rest)
				if err := os.MkdirAll(rest, 0o750); err != nil {
					t.Fatalf("anlegen: %v", err)
				}

				if err := fall.wechsel(verwalter, dritte); !errors.Is(err, domain.ErrConflict) ||
					!strings.Contains(err.Error(), rest) {
					t.Errorf("Fehler ist %v", err)
				}

				liegtDa(t, verwaltet, "2026-02-02")
				vorigeLiegtDa(t, verwaltet, "2026-01-01")
			})
		}
	})

	t.Run("lässt er sich nicht zurücklegen, sagt der Wechsel es", func(t *testing.T) {
		for _, fall := range []struct {
			name string
			rufe int
			rest string
		}{
			{"das Verzeichnis", 1, ".beiseite"},
			{"sein Vermerk", 2, ".beiseite"},
		} {
			t.Run(fall.name, func(t *testing.T) {
				verwalter, _, verwaltet, dritte := zweimal(t)
				liegenlassen(t, verwaltet, fall.rest, true)
				verwalter.umbenennen = scheitertBeim(fall.rufe)

				if _, err := verwalter.Ersetzen(t.Context(), dritte, false); err == nil ||
					!strings.Contains(err.Error(), fall.rest) {
					t.Errorf("Fehler ist %v", err)
				}

				liegtDa(t, verwaltet, "2026-02-02")
			})
		}
	})
}

// N4: **ein Rückweg, der selbst scheitert, bleibt nicht still** — der Fehler
// nennt ihn und den Ort, an dem nachzusehen ist.
func TestEinScheiternderRueckwegWirdGemeldet(t *testing.T) {
	verwalter, _, verwaltet, dritte := zweimal(t)
	unkopierbar(t, dritte)

	// Beiseitelegen, Vermerk, Aufheben — der vierte Zug ist der Rückweg.
	verwalter.umbenennen = scheitertBeim(4)

	_, err := verwalter.Ersetzen(t.Context(), dritte, false)
	if err == nil || !strings.Contains(err.Error(), "Rückweg") ||
		!strings.Contains(err.Error(), filepath.Join(verwaltet, "fremd.mse-style")) {
		t.Errorf("Fehler ist %v", err)
	}
}
