package service_test

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einAbzug gibt einen Data-URL mit gültiger PNG-Signatur.
func einAbzug() string {
	rohes := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 'x'}

	return domain.Abzugsvorsatz + base64.StdEncoding.EncodeToString(rohes)
}

func TestTischbildausfuhrSchreibt(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.png")

	if err := service.NeueTischbildausfuhr().Schreibe(einAbzug(), ziel); err != nil {
		t.Fatalf("unerwarteter Fehler: %v", err)
	}

	gelesen, err := os.ReadFile(ziel) // #nosec G304 -- ein Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	if len(gelesen) != 9 || gelesen[1] != 'P' {
		t.Errorf("geschrieben wurde %q", gelesen)
	}
}

// **Ein leerer Pfad ist keine Datei.** Er entstünde, wo der Aufrufer den
// Abbruch des Dialogs nicht abgefangen hat — und dann soll er es hören und
// nicht ins Arbeitsverzeichnis schreiben.
func TestTischbildausfuhrOhnePfad(t *testing.T) {
	t.Parallel()

	fehler := service.NeueTischbildausfuhr().Schreibe(einAbzug(), "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Ein Abzug, der keiner ist, kommt nicht an die Platte** — die Prüfung
// steht in [domain.DeuteAbzug], und dieser Lauf hält fest, dass der Dienst
// sie wirklich ruft.
func TestTischbildausfuhrWeistFremdesAb(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "tisch.png")

	fehler := service.NeueTischbildausfuhr().Schreibe("data:text/plain,hallo", ziel)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("die Datei steht da, obwohl der Abzug abgewiesen wurde")
	}
}

// **Ein Verzeichnis, das es nicht gibt, ist ein Fehler und kein halbes
// Bild.**
func TestTischbildausfuhrOhneVerzeichnis(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "gibtsnicht", "tisch.png")

	if fehler := service.NeueTischbildausfuhr().Schreibe(einAbzug(), ziel); fehler == nil {
		t.Fatal("erwartet einen Fehler")
	}
}
