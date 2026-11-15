// **Ein neues, leeres Set als Werkzeug** —
// #891.
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// anlegeprobe merkt sich, was angelegt werden sollte.
type anlegeprobe struct {
	name   string
	spiel  string
	stil   string
	fehler error
}

func (a *anlegeprobe) SetAnlegen(
	_ context.Context, name, spiel, stil string,
) (int64, error) {
	a.name, a.spiel, a.stil = name, spiel, stil

	return 7, a.fehler
}

// rufeSetneu ruft das Werkzeug über die **echte** Werkzeugliste.
//
// **Über die Liste und nicht über die Funktion**: ein Werkzeug, das gebaut,
// aber nie eingetragen wird, bestünde jeden Einzeltest und wäre für ein
// Sprachmodell nicht da.
func rufeSetneu(t *testing.T, anleger mcp.Setanleger, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     anleger,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "schreiben.set.neu" {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug schreiben.set.neu steht nicht in der Liste")

	return "", nil
}

// **Das Werkzeug reicht durch und gibt die Kennung.**
func TestDasSetneuwerkzeugLegtAn(t *testing.T) {
	t.Parallel()

	probe := &anlegeprobe{}

	aus, err := rufeSetneu(t, probe,
		`{"name": "Mein Set", "spiel": "probe", "stil": "probe-schlicht"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.name != "Mein Set" || probe.spiel != "probe" ||
		probe.stil != "probe-schlicht" {
		t.Errorf("angelegt wurde %+v", probe)
	}

	if !strings.Contains(aus, `"ID": 7`) {
		t.Errorf("die Antwort nennt die Kennung nicht: %s", aus)
	}
}

// **Ein Fehler kommt durch.**
func TestDasSetneuwerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kein Spiel")

	if _, err := rufeSetneu(t, &anlegeprobe{fehler: kaputt},
		`{"name": "x", "spiel": "y"}`); !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v, erwartet war der eigene", err)
	}
}

// **Unlesbare Argumente melden sich, statt zu raten.**
func TestDasSetneuwerkzeugWeistUnlesbaresAb(t *testing.T) {
	t.Parallel()

	if _, err := rufeSetneu(t, &anlegeprobe{}, `{"name": 7}`); err == nil {
		t.Error("eine Zahl als Name kam durch")
	}
}
