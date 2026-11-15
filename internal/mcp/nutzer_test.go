// **Nutzer und Profile über MCP** — Werkstück 1 von Phase 4.12
// (#1351).
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die drei Werkzeuge tun, was sie sagen.**
func TestDieNutzerwerkzeugeTunWasSieSagen(t *testing.T) {
	t.Run("lesen", func(t *testing.T) {
		probe := &nutzerprobe{zeilen: []mcp.Nutzerzeile{{
			ID: "1", Name: "Ich",
			Profile: []mcp.Profilzeile{{ID: "10", NutzerID: "1", Name: "Mein Profil"}},
		}}}

		antwort := rufeRoh(t, mitNutzern(t, probe, "lesen.nutzer"), `{}`)

		if !strings.Contains(antwort, `"name": "Ich"`) ||
			!strings.Contains(antwort, `"Mein Profil"`) {
			t.Errorf("die Antwort nennt Nutzer oder Profil nicht: %s", antwort)
		}
	})

	// **Eine leere Liste ist eine Antwort, `null` ist keine** (§3).
	t.Run("lesen: leer", func(t *testing.T) {
		antwort := rufeRoh(t, mitNutzern(t, &nutzerprobe{}, "lesen.nutzer"), `{}`)

		if strings.Contains(antwort, `"nutzer": null`) {
			t.Errorf("die Liste ist null: %s", antwort)
		}
	})

	t.Run("Nutzer anlegen", func(t *testing.T) {
		probe := &nutzerprobe{}

		antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.nutzer.neu"),
			`{"name":"Zweite","bild":"z.png"}`)

		if probe.angelegt != "Zweite/z.png" {
			t.Errorf("angelegt wurde %q", probe.angelegt)
		}

		if !strings.Contains(antwort, `"name": "Zweite"`) {
			t.Errorf("die Antwort nennt den Nutzer nicht: %s", antwort)
		}
	})

	t.Run("Profil anlegen", func(t *testing.T) {
		probe := &nutzerprobe{}

		antwort := rufeRoh(t, mitNutzern(t, probe, "schreiben.profil.neu"),
			`{"nutzerID":"7","name":"Turnier"}`)

		if probe.profil != "7/Turnier" {
			t.Errorf("angelegt wurde %q", probe.profil)
		}

		if !strings.Contains(antwort, `"name": "Turnier"`) {
			t.Errorf("die Antwort nennt das Profil nicht: %s", antwort)
		}
	})
}

// **Und jeder Weg meldet seinen Fehler** — der krumme Rumpf wie die klagende
// Naht.
func TestJedesNutzerwerkzeugMeldetSeinenFehler(t *testing.T) {
	faelle := map[string]struct {
		werkzeug  string
		argumente string
		probe     *nutzerprobe
	}{
		"lesen: die Naht klagt": {
			"lesen.nutzer", `{}`, &nutzerprobe{fehler: errors.New("kaputt")},
		},
		"Nutzer: krummer Rumpf": {"schreiben.nutzer.neu", `{`, &nutzerprobe{}},
		"Profil: krummer Rumpf": {"schreiben.profil.neu", `{`, &nutzerprobe{}},
		"Nutzer: die Naht klagt": {
			"schreiben.nutzer.neu", `{"name":"x"}`, &nutzerprobe{fehler: errors.New("kaputt")},
		},
		"Profil: die Naht klagt": {
			"schreiben.profil.neu", `{"nutzerID":"1","name":"x"}`,
			&nutzerprobe{fehler: errors.New("kaputt")},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			if !rufeMitFehler(
				t, mitNutzern(t, fall.probe, fall.werkzeug), fall.argumente) {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// mitNutzern gibt das Werkzeug dieses Namens über einer Probe.
func mitNutzern(t *testing.T, probe *nutzerprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: &bestandsprobe{}, Schreiber: &bestandsprobe{},
		Auskunft: stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(), Nutzer: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("das Werkzeug %q gibt es nicht", name)

	return mcp.Werkzeug{}
}

// rufeMitFehler ruft und sagt, ob es geklagt hat.
func rufeMitFehler(t *testing.T, werkzeug mcp.Werkzeug, argumente string) bool {
	t.Helper()

	_, err := werkzeug.Ruf(t.Context(), json.RawMessage(argumente))

	return err != nil
}

// nutzerprobe ist eine Nutzerpflege im Arbeitsspeicher.
type nutzerprobe struct {
	zeilen   []mcp.Nutzerzeile
	angelegt string
	profil   string
	fehler   error

	// **Der Wechsel** — Werkstück 3 (#1351).
	gezeigt    mcp.Profilzeile
	gewechselt string

	// nutzergewechselt hält fest, auf wen umgestellt wurde.
	nutzergewechselt string

	// **Die Konten** — Werkstück 4 desselben Issues.
	konten  []mcp.Kontozeile
	geloest string

	// **Die Pflege** — Werkstück 5.
	geaendert string
	geloescht string

	// **Das anonyme Konto** — Werkstück 6.
	anonym bool

	// **Der zweite Faktor** — Werkstück 7.
	faktorzeilen []mcp.Faktorzeile

	// **Das Profilbild** — Werkstück 8. `gebildet` hält den Pfad, den das
	// Werkzeug gereicht hat.
	gebildet string

	// **Die Einstellungen desselben Werkstücks** — `gesetzteWerte` hält, was
	// das Werkzeug gesetzt hat.
	gesetzteWerte map[string]string
}

// **Der zweite Faktor** — Werkstück 7 (#1351).
func (p *nutzerprobe) Faktoren(context.Context, string) ([]mcp.Faktorzeile, error) {
	return p.faktorzeilen, p.fehler
}

func (p *nutzerprobe) FaktorEntfernen(_ context.Context, _, art string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geloescht = art

	return nil
}

// KontoAnonym legt ein Konto ohne jede Angabe an — Werkstück 6 (#1351).
func (p *nutzerprobe) KontoAnonym(context.Context) (mcp.Nutzerzeile, error) {
	if p.fehler != nil {
		return mcp.Nutzerzeile{}, p.fehler
	}

	p.anonym = true

	return mcp.Nutzerzeile{ID: "1", Name: "anonym-abc", Anonym: true}, nil
}

// **Ändern und Löschen** — Werkstück 5 (#1351).
func (p *nutzerprobe) NutzerAendern(_ context.Context, angaben mcp.Nutzerangaben) (
	mcp.Nutzerzeile, error,
) {
	if p.fehler != nil {
		return mcp.Nutzerzeile{}, p.fehler
	}

	p.geaendert = angaben.Name

	return mcp.Nutzerzeile{
		ID: angaben.NutzerID, Name: angaben.Name, Bild: angaben.Bild,
		Vorname: angaben.Vorname, Fassung: "2",
	}, nil
}

func (p *nutzerprobe) ProfilAendern(_ context.Context, angaben mcp.Profilangaben) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.geaendert = angaben.Name

	return mcp.Profilzeile{
		ID: angaben.ProfilID, NutzerID: "1", Name: angaben.Name,

		Subslogan:         angaben.Subslogan,
		Kurzebeschreibung: angaben.Kurzebeschreibung,
		Langebeschreibung: angaben.Langebeschreibung,
		Badge:             angaben.Badge,
	}, nil
}

func (p *nutzerprobe) NutzerLoeschen(_ context.Context, nutzerID, _ string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geloescht = nutzerID

	return nil
}

func (p *nutzerprobe) ProfilLoeschen(_ context.Context, profilID, _ string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geloescht = profilID

	return nil
}

func (p *nutzerprobe) Anbieterkonten(context.Context, string) ([]mcp.Kontozeile, error) {
	return p.konten, p.fehler
}

func (p *nutzerprobe) KontoLoesen(_ context.Context, _, anbieter string) error {
	if p.fehler != nil {
		return p.fehler
	}

	p.geloest = anbieter

	return nil
}

func (p *nutzerprobe) GezeigtesProfil(context.Context) (mcp.Profilzeile, error) {
	return p.gezeigt, p.fehler
}

// WerArbeitet nennt den Menschen, der arbeitet — Werkstück 17 von Phase 4.12.
func (p *nutzerprobe) WerArbeitet(context.Context) (string, error) {
	if p.fehler != nil {
		return "", p.fehler
	}

	return "7", nil
}

// Arbeitskontext sagt, wer wo als was arbeitet — Werkstück O1 von Phase 5.2.
func (p *nutzerprobe) Arbeitskontext(context.Context) (mcp.Kontextzeile, error) {
	if p.fehler != nil {
		return mcp.Kontextzeile{}, p.fehler
	}

	return mcp.Kontextzeile{
		NutzerID: "7", ProfilID: "10", EignerID: "7", Rolle: "verwalter",
		Modus: "mcp", Revision: 3,
	}, nil
}

// NutzerWechseln stellt um, wer arbeitet.
func (p *nutzerprobe) NutzerWechseln(_ context.Context, nutzerID, _ string) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.nutzergewechselt = nutzerID

	return mcp.Profilzeile{ID: "1", NutzerID: nutzerID, Name: "Erstes"}, nil
}

func (p *nutzerprobe) ProfilWechseln(_ context.Context, profilID, _ string) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.gewechselt = profilID

	return mcp.Profilzeile{ID: profilID, NutzerID: "1", Name: "Gewechselt"}, nil
}

func (p *nutzerprobe) Nutzer(context.Context) ([]mcp.Nutzerzeile, error) {
	return p.zeilen, p.fehler
}

func (p *nutzerprobe) NutzerAnlegen(_ context.Context, name, bild string) (
	mcp.Nutzerzeile, error,
) {
	if p.fehler != nil {
		return mcp.Nutzerzeile{}, p.fehler
	}

	p.angelegt = name + "/" + bild

	return mcp.Nutzerzeile{ID: "2", Name: name, Bild: bild, Profile: []mcp.Profilzeile{}}, nil
}

func (p *nutzerprobe) ProfilAnlegen(_ context.Context, nutzerID, name string) (
	mcp.Profilzeile, error,
) {
	if p.fehler != nil {
		return mcp.Profilzeile{}, p.fehler
	}

	p.profil = nutzerID + "/" + name

	return mcp.Profilzeile{ID: "20", NutzerID: nutzerID, Name: name}, nil
}
