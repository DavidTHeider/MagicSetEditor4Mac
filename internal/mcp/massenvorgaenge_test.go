// **Massenvorgänge über MCP** — Werkstück 6 von Phase 4.5
// (#1292).
package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// masseprobe merkt sich, was gerufen wurde.
type masseprobe struct {
	aenderung mcp.Massenaenderung
	geloescht []mcp.Massenzeile
	stunden   int
	fehler    error
}

func (m *masseprobe) Aendern(_ context.Context, aenderung mcp.Massenaenderung) (
	mcp.Massenbericht, error,
) {
	m.aenderung = aenderung

	if m.fehler != nil {
		return mcp.Massenbericht{}, m.fehler
	}

	return mcp.Massenbericht{
		Geaendert: len(aenderung.Zeilen), Befunde: []string{},
	}, nil
}

func (m *masseprobe) Loeschen(_ context.Context, zeilen []mcp.Massenzeile) (
	mcp.Massenbericht, error,
) {
	m.geloescht = zeilen

	if m.fehler != nil {
		return mcp.Massenbericht{}, m.fehler
	}

	return mcp.Massenbericht{Geaendert: len(zeilen), Befunde: []string{}}, nil
}

func (m *masseprobe) PreiseInMasse(_ context.Context, stunden int) (
	mcp.Preisschub, error,
) {
	m.stunden = stunden

	if m.fehler != nil {
		return mcp.Preisschub{}, m.fehler
	}

	return mcp.Preisschub{Geholt: 3, Offen: 7, Befunde: []string{}}, nil
}

// mitMasse sucht ein Werkzeug in der Liste mit Massenvorgängen.
func mitMasse(t *testing.T, probe *masseprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
		Massenvorgaenge: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}

// **Alle drei stehen in der Liste und tun, was sie sagen.**
func TestDieMassenvorgaengeStehenUeberMCP(t *testing.T) {
	t.Run("ändern", func(t *testing.T) {
		probe := &masseprobe{}

		antwort := rufeRoh(t, mitMasse(t, probe, "schreiben.sammlung.masse"),
			`{"zeilen":[{"id":1,"fassung":2}],"lagerort":"Kiste 4","menge":3}`)

		// **Die Fassung kommt mit** (§7, Regel 2) — sonst überschriebe der
		// Vorgang fremde Änderungen.
		if len(probe.aenderung.Zeilen) != 1 || probe.aenderung.Zeilen[0].Fassung != 2 {
			t.Errorf("die Zeilen kamen als %+v an", probe.aenderung.Zeilen)
		}

		if probe.aenderung.Lagerort == nil || *probe.aenderung.Lagerort != "Kiste 4" {
			t.Errorf("der Lagerort kam als %v an", probe.aenderung.Lagerort)
		}

		// **Was nicht genannt ist, bleibt `nil`** — „unverändert".
		if probe.aenderung.Zustand != nil || probe.aenderung.Wunsch != nil {
			t.Errorf("nicht genannte Felder kamen gesetzt an: %+v", probe.aenderung)
		}

		if !strings.Contains(antwort, "geaendert") {
			t.Errorf("die Antwort lautet %s", antwort)
		}
	})

	t.Run("löschen", func(t *testing.T) {
		probe := &masseprobe{}

		rufeRoh(t, mitMasse(t, probe, "schreiben.sammlung.masseloeschen"),
			`{"zeilen":[{"id":1,"fassung":2},{"id":3,"fassung":1}]}`)

		if len(probe.geloescht) != 2 {
			t.Errorf("gelöscht wurden %+v", probe.geloescht)
		}
	})

	t.Run("Preisschub", func(t *testing.T) {
		probe := &masseprobe{}

		antwort := rufeRoh(t, mitMasse(t, probe, "schreiben.sammlung.preisschub"),
			`{"hoechstensStunden":48}`)

		if probe.stunden != 48 {
			t.Errorf("die Stundenzahl kam als %d an", probe.stunden)
		}

		// **»offen« sagt, wie viele bleiben** — daran hängt der Fortschritt.
		if !strings.Contains(antwort, "offen") {
			t.Errorf("die Antwort nennt den Rest nicht: %s", antwort)
		}
	})
}

// **Ein Fehler geht auf jedem der drei Wege als Fehler hinaus.**
func TestDieMassenvorgaengeReichenFehlerDurch(t *testing.T) {
	gestolpert := errors.New("die Platte klemmt")

	for _, fall := range []struct {
		name      string
		argumente string
	}{
		{
			name:      "schreiben.sammlung.masse",
			argumente: `{"zeilen":[{"id":1,"fassung":1}],"menge":3}`,
		},
		{
			name:      "schreiben.sammlung.masseloeschen",
			argumente: `{"zeilen":[{"id":1,"fassung":1}]}`,
		},
		{
			name:      "schreiben.sammlung.preisschub",
			argumente: `{"hoechstensStunden":24}`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			werkzeug := mitMasse(t, &masseprobe{fehler: gestolpert}, fall.name)

			if _, err := werkzeug.Ruf(t.Context(),
				[]byte(fall.argumente)); !errors.Is(err, gestolpert) {
				t.Errorf("Fehler %v, erwartet %v", err, gestolpert)
			}
		})
	}
}

// **Unbrauchbare Argumente sind ein Fehler und kein Absturz.**
func TestDieMassenvorgaengeMitUnbrauchbarenArgumenten(t *testing.T) {
	for _, name := range []string{
		"schreiben.sammlung.masse", "schreiben.sammlung.masseloeschen",
		"schreiben.sammlung.preisschub",
	} {
		werkzeug := mitMasse(t, &masseprobe{}, name)

		if _, err := werkzeug.Ruf(t.Context(), []byte(`{`)); err == nil {
			t.Errorf("%s: kein Fehler bei kaputtem JSON", name)
		}
	}
}

// **Ohne Naht gibt es die drei Werkzeuge nicht.**
func TestOhneMassennahtKeineWerkzeuge(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{}, Angaben: neueAngabenprobe(),
		Stile: neueStilprobe(), Bestand: neueBestandsprobeWeit(),
	}) {
		if strings.Contains(werkzeug.Name, "masse") ||
			strings.Contains(werkzeug.Name, "preisschub") {
			t.Errorf("ohne Naht steht %q in der Liste", werkzeug.Name)
		}
	}
}
