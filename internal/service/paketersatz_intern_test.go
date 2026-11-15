package service

// **Die Fehlerwege von Ersetzen und Zurück** — DOM-01.AK-02
// (#1696).
//
// **Ein innerer Lauf**, weil er das Umbenennen einsetzt: ein Verschieben,
// das beim n-ten Mal scheitert, lässt sich auf einer echten Platte nicht
// herbeiführen. Geprüft wird jedes Mal **auch**, dass kein halber Stand
// bleibt — das aktuelle Paket liegt danach, wo es lag.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// innerkatalog ist ein Katalog im Arbeitsspeicher, dessen Abgleich einzeln
// scheitern kann.
type innerkatalog struct {
	pakete         []domain.Paket
	abgleichfehler error
}

func (k *innerkatalog) Abgleichen(_ context.Context, pakete []domain.Paket) error {
	if k.abgleichfehler != nil {
		return k.abgleichfehler
	}

	k.pakete = pakete

	return nil
}

func (k *innerkatalog) Pakete(context.Context) ([]domain.Paket, error) { return k.pakete, nil }

func (k *innerkatalog) PaketeNachMuster(context.Context, string) ([]domain.Paket, error) {
	return nil, nil
}

func (k *innerkatalog) PaketEntfernen(context.Context, string, string) error { return nil }

// innenPaket legt ein Stilpaket an.
func innenPaket(t *testing.T, wurzel, version string) string {
	t.Helper()

	ort := filepath.Join(wurzel, "fremd.mse-style")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, "style"),
		[]byte("mse version: 2.0.0\ngame: probe\nversion: "+version+"\n"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// innerplatz bindet Fassung 1 ein und liefert die Quelle für Fassung 2.
func innerplatz(t *testing.T) (verwalter *Paketverwalter, katalog *innerkatalog, verwaltet, neu string) {
	t.Helper()

	verwaltet = filepath.Join(t.TempDir(), "pakete")
	if err := os.MkdirAll(verwaltet, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	katalog = &innerkatalog{}
	verwalter = NeuerPaketverwalter(katalog, verwaltet, verwaltet)

	if _, err := verwalter.Installieren(t.Context(), innenPaket(t, t.TempDir(), "2026-01-01")); err != nil {
		t.Fatalf("installieren: %v", err)
	}

	return verwalter, katalog, verwaltet, innenPaket(t, t.TempDir(), "2026-02-02")
}

// scheitertBeim gibt ein Umbenennen, das beim n-ten Ruf scheitert.
func scheitertBeim(n int) func(string, string) error {
	rufe := 0

	return func(alt, neu string) error {
		rufe++
		if rufe == n {
			return errors.New("die Platte klemmt")
		}

		return os.Rename(alt, neu)
	}
}

// liegtDa sagt, ob das aktuelle Paket in Fassung `version` dasteht.
func liegtDa(t *testing.T, verwaltet, version string) {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join(verwaltet, "fremd.mse-style", "style")) //nolint:gosec // ein Pfad dieses Laufs
	if err != nil || !strings.Contains(string(roh), "version: "+version) {
		t.Errorf("das aktuelle Paket steht nicht in Fassung %s da: %q (%v)", version, roh, err)
	}
}

func TestErsetzenLaesstBeimScheiternDieAlteFassungStehen(t *testing.T) {
	t.Run("das Aufheben scheitert", func(t *testing.T) {
		verwalter, _, verwaltet, neu := innerplatz(t)
		verwalter.umbenennen = scheitertBeim(1)

		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-01-01")
	})

	t.Run("der Ort der vorigen Fassung ist eine Datei", func(t *testing.T) {
		verwalter, _, verwaltet, neu := innerplatz(t)
		if err := os.WriteFile(verwaltet+"-vorige", []byte("im Weg"), 0o600); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-01-01")
	})

	t.Run("der Vermerk lässt sich nicht schreiben", func(t *testing.T) {
		verwalter, _, verwaltet, neu := innerplatz(t)
		if err := os.MkdirAll(filepath.Join(verwaltet+"-vorige", "fremd.mse-style.vermerk"), 0o750); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-01-01")
	})

	t.Run("die neue Fassung lässt sich nicht kopieren", func(t *testing.T) {
		verwalter, _, verwaltet, neu := innerplatz(t)
		zweite := filepath.Join(neu, "unlesbar")
		if err := os.WriteFile(zweite, []byte("x"), 0o000); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err == nil {
			t.Error("der Fehler ging verloren")
		}

		liegtDa(t, verwaltet, "2026-01-01")
	})

	t.Run("der Katalog klemmt danach", func(t *testing.T) {
		verwalter, katalog, _, neu := innerplatz(t)
		katalog.abgleichfehler = errors.New("klemmt")

		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err == nil {
			t.Error("der Fehler ging verloren")
		}
	})
}

func TestZurueckLaesstBeimScheiternDieAktuelleFassungStehen(t *testing.T) {
	// ersetzt bindet Fassung 2 ein, sodass Fassung 1 die vorige ist.
	ersetzt := func(t *testing.T) (*Paketverwalter, *innerkatalog, string) {
		t.Helper()

		verwalter, katalog, verwaltet, neu := innerplatz(t)
		if _, err := verwalter.Ersetzen(t.Context(), neu, false); err != nil {
			t.Fatalf("ersetzen: %v", err)
		}

		return verwalter, katalog, verwaltet
	}

	for n, name := range map[int]string{
		1: "die vorige Fassung lässt sich nicht bereitlegen",
		2: "die aktuelle lässt sich nicht aufheben",
		3: "die vorige lässt sich nicht einlegen",
	} {
		t.Run(name, func(t *testing.T) {
			verwalter, _, verwaltet := ersetzt(t)
			verwalter.umbenennen = scheitertBeim(n)

			if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); err == nil {
				t.Error("der Fehler ging verloren")
			}

			liegtDa(t, verwaltet, "2026-02-02")
		})
	}

	t.Run("der Vermerk ist unlesbar", func(t *testing.T) {
		verwalter, _, verwaltet := ersetzt(t)
		if err := os.WriteFile(filepath.Join(verwaltet+"-vorige", "fremd.mse-style.vermerk"), []byte("{"), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler ist %v", err)
		}
	})

	t.Run("die vorige Fassung ist kein lesbares Paket", func(t *testing.T) {
		verwalter, _, verwaltet := ersetzt(t)
		if err := os.Remove(filepath.Join(verwaltet+"-vorige", "fremd.mse-style", "style")); err != nil {
			t.Fatalf("räumen: %v", err)
		}

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler ist %v", err)
		}
	})

	t.Run("der Katalog klemmt danach", func(t *testing.T) {
		verwalter, katalog, _ := ersetzt(t)
		katalog.abgleichfehler = errors.New("klemmt")

		if _, err := verwalter.Zurueck(t.Context(), "mse-style", "fremd"); err == nil {
			t.Error("der Fehler ging verloren")
		}
	})
}
