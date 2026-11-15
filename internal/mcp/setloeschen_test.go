// **Ein Set löschen als Werkzeug** —
// #1218.
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// loeschprobe merkt sich, was gelöscht werden sollte.
type loeschprobe struct {
	setID   int64
	name    string
	befunde []string
	fehler  error
}

func (l *loeschprobe) SetLoeschen(_ context.Context, setID int64, name string) (
	[]string, error,
) {
	l.setID, l.name = setID, name

	return l.befunde, l.fehler
}

// rufeSetloeschen ruft das Werkzeug über die **echte** Werkzeugliste.
//
// **Über die Liste und nicht über die Funktion**: ein Werkzeug, das gebaut,
// aber nie eingetragen wird, bestünde jeden Einzeltest und wäre für ein
// Sprachmodell nicht da.
func rufeSetloeschen(t *testing.T, loescher mcp.Setloescher, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:       probenbestand(),
		Schreiber:   probenbestand(),
		Ausfuhr:     ruhendeAusfuhr{},
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
		Loescher:    loescher,
	}) {
		if werkzeug.Name == "schreiben.set.loeschen" {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatal("das Werkzeug schreiben.set.loeschen steht nicht in der Liste")

	return "", nil
}

// **Das Werkzeug reicht Kennung und Namen durch.**
func TestDasSetloeschwerkzeugReichtDurch(t *testing.T) {
	t.Parallel()

	probe := &loeschprobe{}

	aus, err := rufeSetloeschen(t, probe, `{"setID": 7, "name": "Mein Set"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.setID != 7 || probe.name != "Mein Set" {
		t.Errorf("gelöscht werden sollte %+v", probe)
	}

	if !strings.Contains(aus, `"geloescht": true`) {
		t.Errorf("die Antwort sagt nicht, dass gelöscht wurde: %s", aus)
	}

	// **Und `befunde` steht als leere Liste da, nicht als `null`.** Ein
	// Gegenüber, das über die Länge geht, stolperte sonst über den
	// Unterschied zwischen „keine Befunde" und „das Feld gibt es nicht".
	if !strings.Contains(aus, `"befunde": []`) {
		t.Errorf("die Antwort führt keine leere Befundliste: %s", aus)
	}
}

// **Die Befunde eines geglückten Löschens kommen mit** — §7, Regel 3.
func TestDasSetloeschwerkzeugGibtDieBefunde(t *testing.T) {
	t.Parallel()

	aus, err := rufeSetloeschen(t,
		&loeschprobe{befunde: []string{"Die Bilder blieben liegen."}},
		`{"setID": 7, "name": "Mein Set"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if !strings.Contains(aus, "Die Bilder blieben liegen.") {
		t.Errorf("der Befund fehlt: %s", aus)
	}
}

// **Ein Fehler kommt durch.**
func TestDasSetloeschwerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("heißt anders")

	if _, err := rufeSetloeschen(t, &loeschprobe{fehler: kaputt},
		`{"setID": 7, "name": "Falsch"}`); !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v, erwartet war der eigene", err)
	}
}

// **Unlesbare Argumente melden sich, statt zu raten.**
func TestDasSetloeschwerkzeugWeistUnlesbaresAb(t *testing.T) {
	t.Parallel()

	probe := &loeschprobe{}

	if _, err := rufeSetloeschen(t, probe, `{"setID": "sieben"}`); err == nil {
		t.Error("die unlesbare Kennung kam durch")
	}

	if probe.name != "" {
		t.Error("die Löschung wurde trotzdem gerufen")
	}
}
