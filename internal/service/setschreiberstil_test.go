package service_test

// **Die setweiten Stiloptionen kommen in die Datei** — F05
// (#2247).
//
// Gefunden im Setdatei-Rundlauf gegen MSE2: `mse stiloption <set> setzen
// "chop top" 7`, dann `mse sichern` — und im geschriebenen Set stand kein
// `chop`. Weder MSE2 noch diese Anwendung sah den Wert nach dem Wiederöffnen.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestSichernSchreibtDieSetweitenStiloptionen(t *testing.T) {
	bestand := mitQuelle()
	bestand.koepfe[0].Style = "probe-stil"
	bestand.stilwerte = []domain.Value{{Field: "chop top", Text: "7"}}

	// **Ohne Umbruch am Ende** — wie die letzte Zeile einer Datei oft endet;
	// die neue darf nicht an ihr kleben.
	bestand.quelle = &domain.Quelle{Datei: "set", Inhalt: []byte(
		"game: probe\nstylesheet: probe-stil\nset info:\n\ttitle: Mein Probeset\ncard:\n\tname: Feuerball")}

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, fehler := service.NeuerSetschreiber(bestand, &ortsschreiber{}).Schreibe(t.Context(), 7, ort); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	gelesen, fehler := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if soll := "\tname: Feuerball\nstyling:\n\tprobe-probe-stil:\n\t\tchop_top: 7\n"; !strings.HasSuffix(string(gelesen), soll) {
		t.Errorf("die Stiloption steht nicht im Block styling:\n%q", string(gelesen))
	}
}

func TestEinFehlerBeimLesenDerStiloptionenHaeltDasSichernAuf(t *testing.T) {
	bestand := mitQuelle()
	bestand.stilwertefehler = errors.New("Bestand weg")

	_, fehler := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, filepath.Join(t.TempDir(), "wieder.mse-set"))
	if !errors.Is(fehler, bestand.stilwertefehler) {
		t.Errorf("Fehler %v, erwartet war der des Bestands", fehler)
	}
}

func TestEinFehlerBeimLesenDerBestandszeileHaeltDenStilblockAuf(t *testing.T) {
	bestand := mitQuelle()
	bestand.stilwerte = []domain.Value{{Field: "chop top", Text: "7"}}
	bestand.zeilenfehler = errors.New("Zeile weg")

	_, fehler := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, filepath.Join(t.TempDir(), "wieder.mse-set"))
	if !errors.Is(fehler, bestand.zeilenfehler) {
		t.Errorf("Fehler %v, erwartet war der der Bestandszeile", fehler)
	}
}
