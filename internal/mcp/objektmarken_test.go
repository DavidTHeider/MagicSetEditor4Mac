package mcp_test

// **Marken an jedem Fachobjekt über MCP** — UX-12
// (#1684).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// objektmarkenprobe merkt sich, womit sie gerufen wurde.
type objektmarkenprobe struct {
	gerufen []string
	fehler  error
}

func (p *objektmarkenprobe) merke(teile ...string) error {
	p.gerufen = append(p.gerufen, strings.Join(teile, " "))

	return p.fehler
}

func (p *objektmarkenprobe) ObjektmarkeSetzen(_ context.Context, art, kennung, marke string) error {
	return p.merke("setzen", art, kennung, marke)
}

func (p *objektmarkenprobe) ObjektmarkeEntfernen(_ context.Context, art, kennung, marke string) error {
	return p.merke("entfernen", art, kennung, marke)
}

func (p *objektmarkenprobe) ObjektmarkenAn(_ context.Context, art, kennung string) ([]string, error) {
	return nil, p.merke("an", art, kennung)
}

func (p *objektmarkenprobe) MarkenJeObjekt(_ context.Context, art string) (map[string][]string, error) {
	return map[string][]string{"11": {"rot"}}, p.merke("je", art)
}

func (p *objektmarkenprobe) MarkeVergeben(_ context.Context, art string, kennungen []string, marke string) error {
	return p.merke(append([]string{"vergeben", art, marke}, kennungen...)...)
}

func (p *objektmarkenprobe) MarkenFiltern(_ context.Context, art string, marken []string, verknuepfung string) ([]string, error) {
	return nil, p.merke(append([]string{"filtern", art, verknuepfung}, marken...)...)
}

func (p *objektmarkenprobe) Markenkatalog(context.Context) ([]domain.Markenposten, error) {
	return []domain.Markenposten{{Marke: "rot", Anzahl: 2}}, p.merke("katalog")
}

func rufeObjektmarke(t *testing.T, probe mcp.Objektmarken, name, argumente string) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Objektmarken: probe}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatalf("das Werkzeug %s gibt es nicht", name)

	return "", nil
}

func TestJedesMarkenwerkzeugErreichtSeineNaht(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		werkzeug, argumente, gerufen, antwort string
	}{
		{"lesen.objektmarken", `{"art":"deck"}`, "je deck", `"11"`},
		{"lesen.objektmarke", `{"art":"deck","kennung":"11"}`, "an deck 11", `"marken": []`},
		{"lesen.markenfilter", `{"art":"deck","marken":["rot"],"verknuepfung":"oder"}`, "filtern deck oder rot", `"kennungen": []`},
		{"lesen.markenkatalog", `{}`, "katalog", `"anzahl": 2`},
		{"schreiben.objektmarke", `{"art":"deck","kennung":"11","marke":" Rot "}`, "setzen deck 11  Rot ", `"marke": "rot"`},
		{"schreiben.objektmarkeentfernen", `{"art":"deck","kennung":"11","marke":"Rot"}`, "entfernen deck 11 Rot", `"entfernt": true`},
		{"schreiben.markevergeben", `{"art":"deck","kennungen":["11","12"],"marke":"Rot"}`, "vergeben deck Rot 11 12", `"marke": "rot"`},
	} {
		t.Run(fall.werkzeug, func(t *testing.T) {
			t.Parallel()

			probe := &objektmarkenprobe{}

			antwort, fehler := rufeObjektmarke(t, probe, fall.werkzeug, fall.argumente)
			if fehler != nil || len(probe.gerufen) != 1 || probe.gerufen[0] != fall.gerufen ||
				!strings.Contains(antwort, fall.antwort) {
				t.Errorf("%s: %v %q %s", fall.werkzeug, fehler, probe.gerufen, antwort)
			}

			// **Und ein Fehler der Naht kommt an.**
			kaputt := &objektmarkenprobe{fehler: errors.New("die Platte klemmt")}
			if _, fehler := rufeObjektmarke(t, kaputt, fall.werkzeug, fall.argumente); fehler == nil {
				t.Errorf("%s: der Fehler ging verloren", fall.werkzeug)
			}

			// **Und kaputtes JSON kommt gar nicht erst an** — außer beim
			// Katalog, der keine Argumente liest.
			if fall.werkzeug == "lesen.markenkatalog" {
				return
			}

			stumm := &objektmarkenprobe{}
			if _, fehler := rufeObjektmarke(t, stumm, fall.werkzeug, `{`); fehler == nil || len(stumm.gerufen) != 0 {
				t.Errorf("%s: kaputtes JSON: %v %q", fall.werkzeug, fehler, stumm.gerufen)
			}
		})
	}
}

// **Umbenennen, Zusammenführen und überall Entfernen gibt es über MCP
// nicht** — sie ändern jedes Objekt, das die Marke trägt, und ein Werkzeug
// schreibt nur, was es nicht überschreibt (0044).
func TestPflegeGibtEsNurAnDenAnderenFlaechen(t *testing.T) {
	t.Parallel()

	// **Nur die Markenwerkzeuge** — hier stand jedes Werkzeug mit
	// „umbenennen" im Namen, und das war breiter als der Satz darüber: der
	// Dateibaum (FMT-02, #1691) benennt **eine** Datei um, mit Stand und
	// Sicherung, und steht unter der Zusage »ändert eine Vorlage«.
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Objektmarken: &objektmarkenprobe{}}) {
		if !strings.Contains(werkzeug.Name, "objektmarke") {
			continue
		}

		if strings.Contains(werkzeug.Name, "umbenennen") || strings.Contains(werkzeug.Name, "tilgen") {
			t.Errorf("%s gibt es über MCP", werkzeug.Name)
		}
	}

	// **Ohne Naht gibt es keines.**
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "objektmarke") {
			t.Errorf("%s ohne Naht", werkzeug.Name)
		}
	}
}
