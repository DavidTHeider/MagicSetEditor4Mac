package app

// **Die Ränder der Protokollweiche** — OPS-04
// (#1721):
// eine Datei, die sich nicht öffnen lässt, eine alte, die sich nicht
// schließen lässt, die Stufe »fehler« und gebundene Gruppen und Merkmale
// hinter der Weiche und dem Filter.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Weiche aus `main`** schreibt neben den Bestand und ins Systemprotokoll.
func TestDieWeicheAusMainSchreibtNebenDenBestand(t *testing.T) {
	ordner := t.TempDir()
	weiche := NeueProtokollweiche(ordner)

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	stand := weiche.Stand()
	if stand.Datei != filepath.Join(ordner, Protokollname) || !kanal(stand, KanalSystem).Aktiv ||
		kanal(stand, KanalAusgabe).Aktiv {
		t.Errorf("%+v", stand)
	}
}

// **Eine Datei, die sich nicht öffnen lässt**, ist ein fehlender Kanal mit
// Grund — hier liegt an ihrer Stelle ein Ordner.
func TestEineUnoeffnbareDateiIstEinFehlenderKanal(t *testing.T) {
	ordner := t.TempDir()
	if fehler := os.Mkdir(filepath.Join(ordner, Protokollname), 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	weiche, system, _ := weicheMitPuffern(t, ordner)
	datei := kanal(weiche.Stand(), KanalDatei)

	weiche.Logger().Info("trotzdem")

	if datei.Aktiv || datei.Grund == "" || !strings.Contains(system.String(), "trotzdem") {
		t.Errorf("%+v, %q", datei, system.String())
	}
}

// **Eine alte Datei, die sich nicht schließen lässt, hält den Wechsel nicht
// auf** — ihre Zeilen sind geschrieben, und die neue gilt.
func TestEinSchliessfehlerHaeltDenWechselNichtAuf(t *testing.T) {
	weiche, system, _ := weicheMitPuffern(t, t.TempDir())
	weiche.zu = func() error { return errors.New("schon zu") }

	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollStufe: domain.ProtokollstufeFehler}))

	weiche.Logger().Warn("eine Warnung")
	weiche.Logger().Error("ein Fehler")

	if strings.Contains(system.String(), "eine Warnung") || !strings.Contains(system.String(), "ein Fehler") {
		t.Errorf("%q", system.String())
	}
}

// **Gruppen und Merkmale gehen durch Weiche und Filter** — auch eine Gruppe,
// die vor dem Wechsel gebunden wurde. **Ihre Werte bleiben geschwärzt**, denn
// `a`, `b` und `c` sind keine angemeldeten Merkmale; hier zählt der Name.
//
// **Gelesen wird die Datei**: die ausführliche Zeile geht seit dem Nachlauf
// über Phase 6.4 (N5) nicht mehr ins Systemprotokoll.
func TestGruppenUndMerkmaleGehenDurchWeicheUndFilter(t *testing.T) {
	ordner := t.TempDir()
	weiche, _, _ := weicheMitPuffern(t, ordner)
	gebunden := weiche.Logger().WithGroup("vorher")

	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{
		domain.ProtokollStufe: domain.ProtokollstufeAusfuehrlich, domain.ProtokollKomponenten: "app",
	}))

	gebunden.Info("eins", "a", 1)
	weiche.Logger().With("b", 2).WithGroup("gruppe").Debug("zwei", "c", 3)

	datei, fehler := os.ReadFile(filepath.Join(ordner, Protokollname)) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	for _, erwartet := range []string{"vorher.a=", " b=", "gruppe.c="} {
		if !strings.Contains(string(datei), erwartet) {
			t.Errorf("%q fehlt in %q", erwartet, datei)
		}
	}
}
