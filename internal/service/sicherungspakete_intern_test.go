package service

// **Eigene Pakete reisen mit, bezogene als Verweis** — OPS-06.AK-01
// (#1723).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

type festerKatalog []domain.Paket

func (k festerKatalog) Pakete(context.Context) ([]domain.Paket, error) { return k, nil }

func paketAnlegen(t *testing.T, ordner, name string) string {
	t.Helper()

	pfad := filepath.Join(ordner, name)
	if fehler := os.MkdirAll(pfad, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(pfad, "game"), []byte("name: "+name), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad
}

func TestEigenePaketeReisenMitUndBezogeneAlsVerweis(t *testing.T) {
	bau := neuerWegbau(t)
	ordner := filepath.Join(bau.ort, "pakete")

	eigen := paketAnlegen(t, ordner, "eigen.mse-game")
	bezogen := paketAnlegen(t, ordner, "bezogen.mse-style")
	fremd := paketAnlegen(t, t.TempDir(), "mse2.mse-game")

	bau.weg.MitPaketen(festerKatalog{
		{Art: "mse-game", Name: "eigen", Pfad: eigen},
		{Art: "mse-style", Name: "bezogen", Pfad: bezogen, Herkunft: "https://example.org/b", Pruefsumme: "sha256:ab"},
		{Art: "mse-game", Name: "mse2", Pfad: fremd},
	}, func() (string, error) { return ordner, nil })

	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
	if fehler := bau.weg.Sichern(t.Context(), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	schau, fehler := bau.weg.Beschauen(t.Context(), ziel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(schau.PaketeEnthalten, []string{"eigen.mse-game"}) || len(schau.PaketeFehlen) != 0 {
		t.Errorf("enthalten %v, fehlen %v", schau.PaketeEnthalten, schau.PaketeFehlen)
	}

	// **Auf einem leeren Rechner**: kein Paket liegt da. Das eigene kommt aus
	// der Sicherung zurück, das bezogene wird als fehlend genannt.
	if raeumfehler := os.RemoveAll(ordner); raeumfehler != nil {
		t.Fatal(raeumfehler)
	}

	schau, fehler = bau.weg.Beschauen(t.Context(), ziel)
	if fehler != nil || !slices.Equal(schau.PaketeFehlen, []string{"bezogen.mse-style"}) {
		t.Fatalf("fehlen %v, %v", schau.PaketeFehlen, fehler)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler != nil {
		t.Fatalf("zurückspielen: %v", fehler)
	}

	if inhalt, lesefehler := os.ReadFile(filepath.Join(ordner, "eigen.mse-game", "game")); lesefehler != nil || //nolint:gosec // aus t.TempDir()
		string(inhalt) != "name: eigen.mse-game" {
		t.Errorf("das eigene Paket kam nicht zurück: %q, %v", inhalt, lesefehler)
	}

	if _, statfehler := os.Stat(filepath.Join(ordner, "bezogen.mse-style")); statfehler == nil {
		t.Error("das bezogene Paket reiste mit")
	}
}

// **Was schon da ist, bleibt** — ein Zurückspielen überschreibt kein Paket,
// das hier inzwischen jemand bearbeitet hat.
func TestEinVorhandenesPaketWirdNichtUeberschrieben(t *testing.T) {
	bau := neuerWegbau(t)
	ordner := filepath.Join(bau.ort, "pakete")
	eigen := paketAnlegen(t, ordner, "eigen.mse-game")

	bau.weg.MitPaketen(festerKatalog{{Art: "mse-game", Name: "eigen", Pfad: eigen}},
		func() (string, error) { return ordner, nil })

	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
	if fehler := bau.weg.Sichern(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := os.WriteFile(filepath.Join(eigen, "game"), []byte("bearbeitet"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), ziel); fehler != nil {
		t.Fatal(fehler)
	}

	if inhalt, _ := os.ReadFile(filepath.Join(eigen, "game")); string(inhalt) != "bearbeitet" { //nolint:errcheck,gosec // der Vergleich sagt es, aus t.TempDir()
		t.Errorf("überschrieben: %q", inhalt)
	}
}

type kaputterKatalog struct{}

func (kaputterKatalog) Pakete(context.Context) ([]domain.Paket, error) {
	return nil, errors.New("der Katalog ist fort")
}

// **Ohne Ordner, ohne Katalog oder mit einem unlesbaren Paket sichert es
// nicht** — eine Sicherung, der still ein eigenes Paket fehlt, wäre die
// Lücke, die OPS-06 schließt.
func TestEinePaketlueckeHaeltDieSicherungAn(t *testing.T) {
	ordner := t.TempDir()
	unlesbar := paketAnlegen(t, ordner, "zu.mse-game")

	if fehler := os.Chmod(unlesbar, 0o000); fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(unlesbar, 0o700) }) //nolint:errcheck,gosec // Aufräumen

	for name, fall := range map[string]struct {
		quelle Paketquelle
		ordner func() (string, error)
	}{
		"ohne Ordner":      {festerKatalog{}, func() (string, error) { return "", errors.New("kein Zuhause") }},
		"ohne Katalog":     {kaputterKatalog{}, func() (string, error) { return ordner, nil }},
		"unlesbares Paket": {festerKatalog{{Name: "zu", Pfad: unlesbar}}, func() (string, error) { return ordner, nil }},
	} {
		t.Run(name, func(t *testing.T) {
			bau := neuerWegbau(t)
			bau.weg.MitPaketen(fall.quelle, fall.ordner)

			ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
			if fehler := bau.weg.Sichern(t.Context(), ziel); fehler == nil {
				t.Fatal("gesichert, obwohl ein Paket nicht zu lesen war")
			}

			if _, statfehler := os.Stat(ziel); statfehler == nil {
				t.Error("es liegt eine halbe Sicherung da")
			}
		})
	}
}

// **Was sich nicht einsetzen lässt, steht im Protokoll** und hält nichts auf
// — der Bestand ist dann schon getauscht.
func TestEinNichtEinsetzbaresPaketHaeltNichtsAuf(t *testing.T) {
	platz := t.TempDir()
	paketAnlegen(t, platz, "eigen.mse-game")

	versperrt := filepath.Join(t.TempDir(), "eine-datei")
	if fehler := os.WriteFile(versperrt, nil, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	weg := &Bestandsweg{}
	weg.paketeEinsetzen(t.Context(), platz)

	for _, ordner := range []func() (string, error){
		func() (string, error) { return "", errors.New("kein Zuhause") },
		func() (string, error) { return filepath.Join(versperrt, "pakete"), nil },
	} {
		weg.MitPaketen(festerKatalog{}, ordner)
		weg.paketeEinsetzen(t.Context(), platz)
	}

	if _, fehler := os.Stat(filepath.Join(platz, "eigen.mse-game")); fehler != nil {
		t.Errorf("das Paket ist fort: %v", fehler)
	}
}

// **Die Fläche fragt, bevor sie etwas zeigt**, ob eine Datei ein Kennwort
// braucht.
func TestEineFlaecheErkenntEineVerschluesselteSicherung(t *testing.T) {
	if SicherungVerschluesselt(filepath.Join(t.TempDir(), "fehlt")) {
		t.Error("eine fehlende Datei gilt als verschlüsselt")
	}
}

// **Ein bezogenes Paket, das hier jemand bearbeitet hat, reist mit** —
// FMT-02.AK-04 (#1691):
// „angepasste Pakete sind Bestandteil einer geeigneten Sicherung". Als
// Verweis käme es beim Zurückspielen in der bezogenen Fassung wieder, und die
// Bearbeitung wäre fort. Ein unverändertes bleibt ein Verweis.
func TestEinBearbeitetesBezogenesPaketReistMit(t *testing.T) {
	bau := neuerWegbau(t)
	ordner := filepath.Join(bau.ort, "pakete")

	bearbeitet := paketAnlegen(t, ordner, "bearbeitet.mse-style")
	unveraendert := paketAnlegen(t, ordner, "unveraendert.mse-style")

	abdruck := func(pfad string) string {
		wert, fehler := abdruckVon(pfad)
		if fehler != nil {
			t.Fatal(fehler)
		}

		return wert
	}

	vorher := abdruck(bearbeitet)
	if fehler := os.WriteFile(filepath.Join(bearbeitet, "game"), []byte("meins"), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	bau.weg.MitPaketen(festerKatalog{
		{Art: "mse-style", Name: "bearbeitet", Pfad: bearbeitet, Herkunft: "https://example.org/a", Pruefsumme: "sha256:aa", Abdruck: vorher},
		{Art: "mse-style", Name: "unveraendert", Pfad: unveraendert, Herkunft: "https://example.org/b", Pruefsumme: "sha256:bb", Abdruck: abdruck(unveraendert)},
	}, func() (string, error) { return ordner, nil })

	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
	if fehler := bau.weg.Sichern(t.Context(), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	schau, fehler := bau.weg.Beschauen(t.Context(), ziel)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if !slices.Equal(schau.PaketeEnthalten, []string{"bearbeitet.mse-style"}) {
		t.Errorf("enthalten %v", schau.PaketeEnthalten)
	}
}
