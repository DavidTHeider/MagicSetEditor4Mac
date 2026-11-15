package app

// **Ein Kanal, der beim Schreiben scheitert, nimmt die anderen nicht mit** —
// Befunde 3, 6 und 10 des Zweitblicks über Phase 6.4
// (#1665):
// `io.MultiWriter` hört beim ersten Fehler auf, die Drehung am Vorgabeort
// griff nie, und ein eigener Ort wurde auch auf einem abgezogenen Laufwerk
// angelegt.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type scheiternderSchreiber struct{}

func (scheiternderSchreiber) Write([]byte) (int, error) { return 0, errors.New("logger fehlt") }

func TestEinScheiternderKanalNimmtDieAnderenNichtMit(t *testing.T) {
	ausgabe := &bytes.Buffer{}
	ordner := t.TempDir()
	weiche := neueWeicheMit(ordner, scheiternderSchreiber{}, ausgabe, func() (string, error) { return ordner, nil })

	t.Cleanup(func() {
		if fehler := weiche.Schliessen(); fehler != nil {
			t.Error(fehler)
		}
	})

	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollAusgabe: domain.ProtokollAn}))
	weiche.Logger().Info("eine Zeile")

	if !strings.Contains(ausgabe.String(), "eine Zeile") {
		t.Errorf("die Ausgabe schwieg: %q", ausgabe.String())
	}

	if !strings.Contains(dateiinhalt(t, filepath.Join(ordner, Protokollname)), "eine Zeile") {
		t.Error("die Datei schwieg")
	}

	// **Und der Stand sagt, welcher schweigt** (Befund 10) — hier stand das
	// Systemprotokoll immer auf „an".
	system := kanal(weiche.Stand(), KanalSystem)
	if system.Aktiv || !strings.Contains(system.Grund, "logger fehlt") {
		t.Errorf("%+v", system)
	}
}

// **Die Drehung greift auch dort, wo `main` die Datei schon geöffnet hat.**
func TestDieDrehungGreiftAmVorgabeort(t *testing.T) {
	ordner := t.TempDir()
	pfad := filepath.Join(ordner, Protokollname)

	if fehler := os.WriteFile(pfad, make([]byte, 2<<20), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	weiche, _, _ := weicheMitPuffern(t, ordner)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollGrenze: "1"}))

	if _, fehler := os.Stat(vorfassung(pfad, 1)); fehler != nil {
		t.Errorf("keine Vorfassung: %v", fehler)
	}

	if angabe, fehler := os.Stat(pfad); fehler != nil || angabe.Size() >= 1<<20 {
		t.Errorf("die Datei wurde nicht gedreht: %v", fehler)
	}
}

// **Ein eigener Ort, den es nicht gibt, wird nicht angelegt** — auf einem
// abgezogenen Laufwerk entstünde er sonst still auf der eigenen Platte.
func TestEinFehlenderEigenerOrtWirdNichtAngelegt(t *testing.T) {
	ordner := t.TempDir()
	fehlt := filepath.Join(ordner, "Volumes", "Stick", "Protokolle")

	weiche, _, _ := weicheMitPuffern(t, ordner)
	weiche.Einrichten(domain.ProtokollregelAus(map[string]string{domain.ProtokollOrt: fehlt}))

	if datei := kanal(weiche.Stand(), KanalDatei); datei.Aktiv || datei.Grund == "" {
		t.Errorf("%+v", datei)
	}

	if _, fehler := os.Stat(fehlt); !os.IsNotExist(fehler) {
		t.Errorf("der Ort entstand: %v", fehler)
	}
}
