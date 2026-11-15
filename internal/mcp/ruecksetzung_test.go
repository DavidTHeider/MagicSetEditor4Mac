package mcp_test

// **Die abgestufte Rücksetzung über MCP** — Werkstück O4 von Phase 5.2
// (#1428).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type ruecksetzungsprobe struct {
	antwort mcp.Ruecksetzungsvorschau
	fehler  error

	art, profilID, nutzerID string
}

func (p *ruecksetzungsprobe) Ruecksetzungsvorschau(
	_ context.Context, art, profilID, nutzerID string,
) (mcp.Ruecksetzungsvorschau, error) {
	p.art, p.profilID, p.nutzerID = art, profilID, nutzerID

	return p.antwort, p.fehler
}

func TestOhneNahtGibtEsKeinRuecksetzungswerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.ruecksetzungsvorschau" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

// **Die Vorschau kommt mit Zahlen, Namen und Hindernissen** — und die drei
// Angaben des Rufs reisen bis zur Naht.
func TestDasRuecksetzungswerkzeugReichtSeineAngabenDurch(t *testing.T) {
	t.Parallel()

	probe := &ruecksetzungsprobe{antwort: mcp.Ruecksetzungsvorschau{
		Art: "profil",
		Betroffen: []mcp.Betroffenes{
			{Was: "sets", Anzahl: 3, Namen: []string{"Eins"}},
		},
		Hindernisse:        []mcp.Hindernis{{Art: "offeneRueckkehr", Stelle: "Meins"}},
		NameNoetig:         true,
		SicherungEmpfohlen: true,
	}}

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Ruecksetzung: probe},
		"lesen.ruecksetzungsvorschau")

	antwort, err := werkzeug.Ruf(t.Context(),
		[]byte(`{"art":"profil","profilID":"7","nutzerID":"3"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.art != "profil" || probe.profilID != "7" || probe.nutzerID != "3" {
		t.Errorf("die Naht sah %q, %q, %q", probe.art, probe.profilID, probe.nutzerID)
	}

	for _, teil := range []string{
		`"was": "sets"`, `"anzahl": 3`, `"art": "offeneRueckkehr"`,
		`"nameNoetig": true`, `"sicherungEmpfohlen": true`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}

	// **Das Werkzeug ist lesend und idempotent** — es ändert nichts, und das
	// steht in seiner Marke, nicht nur in seinem Text.
	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("die Marken sind %q und %v", werkzeug.Art, werkzeug.Idempotent)
	}
}

// **Ein Fehler kommt durch** — eine leere Vorschau sähe aus, als wäre nichts
// betroffen, und das ist die eine Auskunft, die niemals falsch sein darf.
func TestDasRuecksetzungswerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Ruecksetzung: &ruecksetzungsprobe{fehler: errors.New("die Datenbank ist zu")},
	}, "lesen.ruecksetzungsvorschau")

	if _, err := werkzeug.Ruf(t.Context(), []byte(`{"art":"layout"}`)); err == nil {
		t.Error("der Fehler ging verloren")
	}

	// **Und eine krumme Eingabe ebenso** — sie ist ein Fehler der Anfrage.
	if _, err := werkzeug.Ruf(t.Context(), []byte(`{"art":`)); err == nil {
		t.Error("die krumme Eingabe kam durch")
	}
}
