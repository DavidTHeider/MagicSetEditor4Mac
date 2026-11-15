// **Die Listeneinfuhr über MCP** — Werkstück 2 von Phase 4.5
// (#1292).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// einfuhrprobe merkt sich, was gerufen wurde.
type einfuhrprobe struct {
	gedeutet   string
	eingelesen []mcp.Listeneinfuhrzeile
	fehler     error

	// teilerfolg und teilerhoeht sind, was **vor** dem Fehler geschrieben
	// wurde (Befund 3): der Dienst gibt beides zusammen zurück.
	teilerfolg  []int64
	teilerhoeht int

	// zusammenfuehrung und vorgeschaut halten fest, was DOM-03.AK-01 (#1698)
	// durchreicht.
	zusammenfuehrung string
	vorgeschaut      []mcp.Listeneinfuhrzeile
}

func (e *einfuhrprobe) Vorschauen(_ context.Context, zeilen []mcp.Listeneinfuhrzeile) (
	mcp.Sammelvorschau, error,
) {
	e.vorgeschaut = zeilen

	if e.fehler != nil {
		return mcp.Sammelvorschau{}, e.fehler
	}

	return mcp.Sammelvorschau{
		Posten:  []mcp.Vorschauposten{{Zeile: 1, Art: "zusammen", Vorher: 2, Nachher: 4}},
		Befunde: []string{},
	}, nil
}

func (e *einfuhrprobe) Deuten(_ context.Context, text string) (mcp.Listendeutung, error) {
	e.gedeutet = text

	if e.fehler != nil {
		return mcp.Listendeutung{}, e.fehler
	}

	return mcp.Listendeutung{
		Zeilen: []mcp.Listendeutungszeile{{
			Roh: "2 Erste Probe (probe) 1", Zeilennummer: 1, Menge: 2,
			Name: "Erste Probe", Setcode: "probe", Nummer: "1",
			Treffer: []mcp.Listentreffer{{
				Printing: "probe-druck", Name: "Erste Probe",
				Setcode: "probe", Setname: "Probesatz", Nummer: "1",
			}},
			Drucke: 1, Eindeutig: true,
		}},
		Befunde: []string{},
	}, nil
}

func (e *einfuhrprobe) Einlesen(
	_ context.Context, zeilen []mcp.Listeneinfuhrzeile, zusammenfuehrung string,
) (mcp.Einfuhrbericht, error) {
	e.eingelesen, e.zusammenfuehrung = zeilen, zusammenfuehrung

	if e.fehler != nil {
		return mcp.Einfuhrbericht{
			Angelegt: e.teilerfolg, Erhoeht: e.teilerhoeht,
			Befunde: []string{},
		}, e.fehler
	}

	return mcp.Einfuhrbericht{
		Angelegt: []int64{1}, Erhoeht: 0, Stueck: 2, Befunde: []string{},
	}, nil
}

// mitEinfuhr sucht ein Werkzeug in der Liste mit Listeneinfuhr.
func mitEinfuhr(t *testing.T, probe *einfuhrprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Listeneinfuhr: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Beide stehen in der Liste und tun, was sie sagen.**
func TestDieListeneinfuhrStehtUeberMCP(t *testing.T) {
	t.Run("deuten", func(t *testing.T) {
		probe := &einfuhrprobe{}

		antwort := rufeRoh(t, mitEinfuhr(t, probe, "lesen.liste.deuten"),
			`{"text":"2 Erste Probe (probe) 1"}`)

		if probe.gedeutet != "2 Erste Probe (probe) 1" {
			t.Errorf("gedeutet wurde %q", probe.gedeutet)
		}

		if !strings.Contains(antwort, `"eindeutig": true`) &&
			!strings.Contains(antwort, `"eindeutig":true`) {
			t.Errorf("die Antwort nennt die Eindeutigkeit nicht: %s", antwort)
		}
	})

	t.Run("einlesen", func(t *testing.T) {
		probe := &einfuhrprobe{}

		antwort := rufeRoh(t, mitEinfuhr(t, probe, "schreiben.liste.einlesen"),
			`{"zeilen":[{"printing":"probe-druck","menge":2,"zustand":"nearMint",`+
				`"sprache":"en","name":"Erste Probe","setcode":"probe","nummer":"1"}]}`)

		if len(probe.eingelesen) != 1 || probe.eingelesen[0].Printing != "probe-druck" {
			t.Fatalf("eingelesen wurde %+v", probe.eingelesen)
		}

		if probe.eingelesen[0].Menge != 2 || probe.eingelesen[0].Zustand != "nearMint" {
			t.Errorf("die Zeile kam als %+v an", probe.eingelesen[0])
		}

		if !strings.Contains(antwort, "stueck") {
			t.Errorf("die Antwort nennt die Stückzahl nicht: %s", antwort)
		}
	})
}

// **Ein Fehler geht als Fehler hinaus und nicht als Antwort.**
func TestDieListeneinfuhrReichtFehlerDurch(t *testing.T) {
	gestolpert := errors.New("kein Netz")

	for _, fall := range []struct {
		name      string
		werkzeug  string
		argumente string
	}{
		{
			name: "deuten", werkzeug: "lesen.liste.deuten",
			argumente: `{"text":"1 Probe"}`,
		},
		{
			name: "einlesen", werkzeug: "schreiben.liste.einlesen",
			argumente: `{"zeilen":[{"printing":"probe-druck","menge":1}]}`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			werkzeug := mitEinfuhr(t, &einfuhrprobe{fehler: gestolpert}, fall.werkzeug)

			if _, err := werkzeug.Ruf(t.Context(),
				[]byte(fall.argumente)); !errors.Is(err, gestolpert) {
				t.Errorf("Fehler %v, erwartet %v", err, gestolpert)
			}
		})
	}
}

// **Unbrauchbare Argumente sind ein Fehler und kein Absturz.**
func TestDieListeneinfuhrMitUnbrauchbarenArgumenten(t *testing.T) {
	for _, name := range []string{"lesen.liste.deuten", "schreiben.liste.einlesen"} {
		werkzeug := mitEinfuhr(t, &einfuhrprobe{}, name)

		if _, err := werkzeug.Ruf(t.Context(), []byte(`{`)); err == nil {
			t.Errorf("%s: kein Fehler bei kaputtem JSON", name)
		}
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht** — und keinen Absturz.
//
// Die Gegenprobe zur `nil`-Wache: eine Anwendung ohne Zuhause baut die Naht
// nicht, und dann darf die Werkzeugliste sie nicht führen.
func TestOhneEinfuhrnahtKeineWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		// **`liste.` mit Punkt und nicht `liste`**: `lesen.wortlisten` enthält
		// die vier Buchstaben und hat mit der Einfuhr nichts zu tun. Der erste
		// Anlauf war rot und hatte recht — mit der falschen Begründung.
		if strings.Contains(werkzeug.Name, "liste.") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}

// **Ein Fehler nennt, was schon geschrieben ist** — Befund 3 des Zweitblicks
// über Phase 4.5.
//
// MCP trägt entweder einen Text oder einen Fehler, nicht beides. Wer bei
// Zeile 217 von 300 abbricht, muss erfahren, dass 216 stehen — sonst schreibt
// der zweite Versuch sie noch einmal (§7, Regel 3).
func TestEinFehlerNenntWasSchonGeschriebenIst(t *testing.T) {
	probe := &einfuhrprobe{
		fehler:      errors.New("die Platte klemmt"),
		teilerfolg:  []int64{1, 2},
		teilerhoeht: 3,
	}

	werkzeug := mitEinfuhr(t, probe, "schreiben.liste.einlesen")

	_, err := werkzeug.Ruf(t.Context(),
		[]byte(`{"zeilen":[{"printing":"abc","menge":1}]}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "2 angelegt") ||
		!strings.Contains(err.Error(), "3 geändert") {
		t.Errorf("die Meldung lautet %q", err)
	}
}

// **Und ohne Teilerfolg hängt nichts daran.**
//
// Ein „(0 angelegt, 0 geändert)" hinter jedem Fehler wäre Rauschen — und der
// Satz soll gelesen werden, wenn er dasteht.
func TestOhneTeilerfolgHaengtNichtsAnDenFehler(t *testing.T) {
	werkzeug := mitEinfuhr(t, &einfuhrprobe{
		fehler: errors.New("die Platte klemmt"),
	}, "schreiben.liste.einlesen")

	_, err := werkzeug.Ruf(t.Context(),
		[]byte(`{"zeilen":[{"printing":"abc","menge":1}]}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if strings.Contains(err.Error(), "angelegt") {
		t.Errorf("die Meldung lautet %q", err)
	}
}
