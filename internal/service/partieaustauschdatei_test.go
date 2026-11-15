package service_test

// **Eine Partie als Datei auf der Platte** — PLAY-03.AK-05
// (#1711).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinePartiedateiGehtHinUndZurueck(t *testing.T) {
	t.Parallel()

	datei := service.NeuePartieaustauschdatei()
	pfad := filepath.Join(t.TempDir(), "partie.json")

	if fehler := datei.Schreibe(pfad, []byte("hallo\n")); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := datei.Lies(pfad)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if string(gelesen) != "hallo\n" {
		t.Errorf("gelesen wurde %q", gelesen)
	}
}

// **Ein Ort, der fehlt, ist ein Befund** — und zwar bei beiden Wegen.
func TestDiePartiedateiBrauchtEinenOrt(t *testing.T) {
	t.Parallel()

	datei := service.NeuePartieaustauschdatei()

	if fehler := datei.Schreibe("", nil); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("das Schreiben meldete %v", fehler)
	}

	if _, fehler := datei.Lies(""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("das Lesen meldete %v", fehler)
	}
}

// **Eine Datei, die es nicht gibt, ist etwas anderes als eine kaputte.**
func TestDiePartiedateiMeldetDieFehlende(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuePartieaustauschdatei().
		Lies(filepath.Join(t.TempDir(), "niemals.json"))

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Ein Verzeichnis ist keine Datei** — der zweite Lesefehlerweg, und er ist
// ein anderer als „gibt es nicht" (§10).
func TestDiePartiedateiMeldetDasVerzeichnis(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	_, fehler := service.NeuePartieaustauschdatei().Lies(ordner)
	if fehler == nil || !strings.Contains(fehler.Error(), "die Partie lesen") {
		t.Fatalf("erwartet den Lesefehler, bekam %v", fehler)
	}
}

// **Ein Ort, an den nicht geschrieben werden kann, klagt.**
func TestDiePartiedateiMeldetDenSchreibfehler(t *testing.T) {
	t.Parallel()

	fehler := service.NeuePartieaustauschdatei().
		Schreibe(filepath.Join(t.TempDir(), "gibtsnicht", "partie.json"), nil)

	if fehler == nil ||
		!strings.Contains(fehler.Error(), "die Partie schreiben") {
		t.Fatalf("erwartet den Schreibfehler, bekam %v", fehler)
	}
}

// **Mehr als die Grenze wird nicht gelesen** — und wer genau ein Byte zu
// viel anbietet, bekommt genau ein Byte zu viel zurück, damit die Domäne es
// ablehnen kann.
func TestDiePartiedateiLiestNichtMehrAlsErlaubt(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "gross.json")

	if fehler := os.WriteFile(pfad,
		make([]byte, domain.PartieaustauschHoechstens+100), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := service.NeuePartieaustauschdatei().Lies(pfad)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != domain.PartieaustauschHoechstens+1 {
		t.Errorf("gelesen wurden %d Bytes", len(gelesen))
	}

	if _, _, fehler := domain.ZerlegePartie(gelesen); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("die Domäne nahm sie an: %v", fehler)
	}
}

// **Eine Datei, die dazuliegt und sich nicht öffnen lässt** — der dritte
// Lesefehlerweg, und er ist ein anderer als „gibt es nicht" und als „ist ein
// Verzeichnis" (§10).
//
// **Über den Entzug des Leserechts**, weil es der einzige Weg ist, der ohne
// besondere Hardware auskommt — dieselbe Form wie in `atomarschreiben_test.go`.
func TestDiePartiedateiMeldetDieVerschlossene(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "verschlossen.json")

	if fehler := os.WriteFile(pfad, []byte("{}"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := os.Chmod(pfad, 0o000); fehler != nil {
		t.Fatalf("Rechte entziehen: %v", fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(pfad, 0o600) }) //nolint:errcheck,gosec // nur fürs Aufräumen

	_, fehler := service.NeuePartieaustauschdatei().Lies(pfad)
	if fehler == nil || errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet der Lesefehler, bekam %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "die Partie lesen") {
		t.Errorf("die Meldung lautet %q", fehler)
	}
}

// **Dieselbe Datei für ein Turnier** — DOM-04.AK-03 (#1699): mit der Grenze
// des Turnierformats und mit seinem Namen in der Meldung.
func TestDieTurnierdateiLiestBisZuIhrerGrenze(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	pfad := filepath.Join(ordner, "gross.json")

	if fehler := os.WriteFile(pfad,
		make([]byte, domain.TurnieraustauschHoechstens+100), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := service.NeueTurnieraustauschdatei().Lies(pfad)
	if fehler != nil || len(gelesen) != domain.TurnieraustauschHoechstens+1 {
		t.Errorf("gelesen wurden %d Bytes (%v)", len(gelesen), fehler)
	}

	fehler = service.NeueTurnieraustauschdatei().Schreibe(filepath.Join(ordner, "fehlt", "t.json"), nil)
	if fehler == nil || !strings.Contains(fehler.Error(), "das Turnier schreiben") {
		t.Errorf("die Meldung lautet %v", fehler)
	}
}
