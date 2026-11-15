package mcp_test

// **Karten aus einem Skript, über MCP** — Werkstück B von Phase 4.11
// (#1337).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// skriptprobe gibt zwei Skripte und merkt sich den Ruf.
type skriptprobe struct {
	gefragt  int64
	gefuehrt string
	fehler   error
}

func (s *skriptprobe) Angebot(_ context.Context, setID int64) (any, error) {
	s.gefragt = setID

	if s.fehler != nil {
		return nil, s.fehler
	}

	return []map[string]any{{"name": "Zwei Farben", "kuerzel": "Z"}}, nil
}

func (s *skriptprobe) Ausfuehren(_ context.Context, setID int64, name string) (any, error) {
	s.gefragt = setID
	s.gefuehrt = name

	if s.fehler != nil {
		return nil, s.fehler
	}

	return map[string]any{"karten": []string{"2", "3"}, "befunde": []string{}}, nil
}

// rufeSkriptwerkzeug ruft ein Werkzeug der Kartenskripte.
func rufeSkriptwerkzeug(
	t *testing.T, skripte mcp.Kartenskripte, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand:       neueBestandsprobeWeit(),
		Kartenskripte: skripte,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return "", nil
}

// **Das lesende Werkzeug nennt die Skripte.**
func TestDasWerkzeugKartenskripteNenntSie(t *testing.T) {
	t.Parallel()

	probe := &skriptprobe{}

	antwort, err := rufeSkriptwerkzeug(t, probe, "lesen.kartenskripte", `{"setID":7}`)
	if err != nil {
		t.Fatalf("lesen.kartenskripte: %v", err)
	}

	if probe.gefragt != 7 {
		t.Errorf("gefragt wurde nach Set %d", probe.gefragt)
	}

	if !strings.Contains(antwort, "Zwei Farben") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Das schreibende führt eines aus — und liest beide Argumente.**
//
// **Der Kern**: `kennung` deutet **alle** Argumente als Zahlen, und `name` ist
// eine Zeichenkette. Der erste Anlauf gab „die Argumente sind kein JSON:
// invalid number literal" — gefunden hat es der Lauf über alle Flächen.
func TestDasWerkzeugKartenskriptFuehrtAus(t *testing.T) {
	t.Parallel()

	probe := &skriptprobe{}

	antwort, err := rufeSkriptwerkzeug(t, probe, "schreiben.kartenskript",
		`{"setID":7,"name":"Zwei Farben"}`)
	if err != nil {
		t.Fatalf("schreiben.kartenskript: %v", err)
	}

	if probe.gefragt != 7 || probe.gefuehrt != "Zwei Farben" {
		t.Errorf("gerufen wurde Set %d mit %q", probe.gefragt, probe.gefuehrt)
	}

	if !strings.Contains(antwort, "karten") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Eine Kennung, die keine ist, wird abgewiesen — an beiden Werkzeugen.**
func TestDieKartenskriptwerkzeugePruefenIhreKennung(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		"lesen.kartenskripte":    `{"setID":0}`,
		"schreiben.kartenskript": `{"setID":0,"name":"x"}`,
	}

	for name, argumente := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSkriptwerkzeug(t, &skriptprobe{}, name, argumente); err == nil {
				t.Error("die Null ging als Kennung durch")
			}
		})
	}
}

// **Kaputte Argumente melden sich.**
func TestDieKartenskriptwerkzeugeMeldenKaputteArgumente(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"lesen.kartenskripte", "schreiben.kartenskript"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := rufeSkriptwerkzeug(t, &skriptprobe{}, name, `kein JSON`); err == nil {
				t.Error("der Unsinn ging durch")
			}
		})
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerKartenskriptnahtKommtDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")

	for name, argumente := range map[string]string{
		"lesen.kartenskripte":    `{"setID":7}`,
		"schreiben.kartenskript": `{"setID":7,"name":"x"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := rufeSkriptwerkzeug(t, &skriptprobe{fehler: eigener}, name, argumente)
			if !errors.Is(err, eigener) {
				t.Errorf("Fehler %v, erwartet war der eigene", err)
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// Eine Fläche, die sie nicht verdrahtet, soll sie nicht anbieten — ein
// Werkzeug, das `nil` ruft, wäre ein Absturz statt einer Auskunft.
func TestOhneNahtGibtEsDieKartenskriptwerkzeugeNicht(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "kartenskript") {
			t.Errorf("das Werkzeug %q steht ohne Naht da", werkzeug.Name)
		}
	}
}
