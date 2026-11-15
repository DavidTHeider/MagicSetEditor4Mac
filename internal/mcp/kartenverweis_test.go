// **Die Verweiswerkzeuge** — Werkstück 2 von Phase 3.6
// (#1167).
package mcp_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// verweisprobe hält Verweise im Arbeitsspeicher.
type verweisprobe struct {
	verweise map[int64]mcp.Verweisangabe
	fehler   error
	gesetzt  string
}

func neueVerweisprobe() *verweisprobe {
	return &verweisprobe{verweise: map[int64]mcp.Verweisangabe{}}
}

func (v *verweisprobe) VerweisSetzen(_ context.Context, karteID int64, printing string) (
	mcp.Verweisangabe, error,
) {
	v.gesetzt = printing

	if v.fehler != nil {
		return mcp.Verweisangabe{}, v.fehler
	}

	angabe := mcp.Verweisangabe{
		KarteID: "7", Printing: printing, Name: "Lightning Bolt",
		Setcode: "lea", Nummer: "161",
	}
	v.verweise[karteID] = angabe

	return angabe, nil
}

func (v *verweisprobe) Verweis(_ context.Context, karteID int64) (
	*mcp.Verweisangabe, error,
) {
	if v.fehler != nil {
		return nil, v.fehler
	}

	angabe, da := v.verweise[karteID]
	if !da {
		return nil, nil //nolint:nilnil // „keiner" ist kein Fehler
	}

	return &angabe, nil
}

func (v *verweisprobe) VerweisLoesen(_ context.Context, karteID int64) error {
	if v.fehler != nil {
		return v.fehler
	}

	delete(v.verweise, karteID)

	return nil
}

// rufeVerweis sucht ein Verweiswerkzeug und ruft es.
func rufeVerweis(
	t *testing.T, verweise mcp.Fremdverweise, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand:  neueBestandsprobeWeit(),
		Verweise: verweise,
	}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %s gibt es nicht", name)

	return "", nil
}

// **Setzen gibt den aufgelösten Verweis.**
func TestDasVerweiswerkzeugSetztUndGibtZurueck(t *testing.T) {
	t.Parallel()

	probe := neueVerweisprobe()

	antwort, fehler := rufeVerweis(t, probe, "schreiben.verweis",
		`{"karteID":7,"printing":"516e137a"}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, "Lightning Bolt") {
		t.Errorf("Antwort: %s", antwort)
	}

	if probe.gesetzt != "516e137a" {
		t.Errorf("gesetzt wurde %q", probe.gesetzt)
	}
}

// **Lesen ohne Verweis sagt es, statt zu scheitern.**
//
// Ein Sprachmodell soll „diese Karte hat keinen Verweis" von „der Bestand ist
// kaputt" unterscheiden können.
func TestDasVerweiswerkzeugSagtWennKeinerDaIst(t *testing.T) {
	t.Parallel()

	antwort, fehler := rufeVerweis(t, neueVerweisprobe(), "lesen.verweis", `{"karteID":7}`)
	if fehler != nil {
		t.Fatalf("Ruf: %v", fehler)
	}

	if !strings.Contains(antwort, "null") && !strings.Contains(antwort, "kein") {
		t.Errorf("Antwort: %s", antwort)
	}
}

// **Lösen geht durch.**
func TestDasVerweiswerkzeugLoest(t *testing.T) {
	t.Parallel()

	probe := neueVerweisprobe()

	if _, fehler := rufeVerweis(t, probe, "schreiben.verweis",
		`{"karteID":7,"printing":"abc"}`); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if _, fehler := rufeVerweis(t, probe, "schreiben.verweisloesen",
		`{"karteID":7}`); fehler != nil {
		t.Fatalf("lösen: %v", fehler)
	}

	if len(probe.verweise) != 0 {
		t.Errorf("es blieb %v zurück", probe.verweise)
	}
}

// **Eine fehlende Angabe wird abgewiesen, bevor etwas geschieht.**
func TestDieVerweiswerkzeugeWeisenLeereAngabenAb(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ werkzeug, argumente string }{
		"setzen ohne Karte": {"schreiben.verweis", `{"printing":"abc"}`},
		"setzen ohne Druck": {"schreiben.verweis", `{"karteID":7}`},
		"lesen ohne Karte":  {"lesen.verweis", `{}`},
		"lösen ohne Karte":  {"schreiben.verweisloesen", `{}`},
		"setzen mit Unsinn": {"schreiben.verweis", `{"karteID":"sieben"}`},

		// **Auch die beiden Wege über `verweiskarte`** — ein unlesbares
		// Argument geht dort denselben Zweig, und ein Zweig, den kein Lauf
		// betritt, ist eine Behauptung (§3).
		"lesen mit Unsinn": {"lesen.verweis", `{"karteID":"sieben"}`},
		"lösen mit Unsinn": {"schreiben.verweisloesen", `{"karteID":"sieben"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			probe := neueVerweisprobe()

			if _, fehler := rufeVerweis(t, probe, fall.werkzeug, fall.argumente); fehler == nil {
				t.Fatal("die leere Angabe kam als Erfolg zurück")
			}

			if probe.gesetzt != "" {
				t.Errorf("es ging %q hinein", probe.gesetzt)
			}
		})
	}
}

// **Ein Fehler der Naht kommt durch.**
func TestEinFehlerDerVerweisnahtKommtDurch(t *testing.T) {
	t.Parallel()

	probe := neueVerweisprobe()
	probe.fehler = domain.ErrNotFound

	for name, fall := range map[string]struct{ werkzeug, argumente string }{
		"setzen": {"schreiben.verweis", `{"karteID":7,"printing":"abc"}`},
		"lesen":  {"lesen.verweis", `{"karteID":7}`},
		"lösen":  {"schreiben.verweisloesen", `{"karteID":7}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := rufeVerweis(t, probe, fall.werkzeug, fall.argumente); fehler == nil {
				t.Fatal("der Fehler kam als Erfolg zurück")
			}
		})
	}
}

// **Ohne Naht gibt es die Werkzeuge nicht.**
//
// Ein Werkzeug, das dasteht und bei jedem Ruf scheitert, ist schlimmer als
// keines: ein Sprachmodell versucht es dreimal.
func TestOhneVerweisnahtGibtEsDieWerkzeugeNicht(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "verweis") {
			t.Errorf("ohne Naht steht %s in der Liste", werkzeug.Name)
		}
	}
}
