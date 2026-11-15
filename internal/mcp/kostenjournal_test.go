package mcp_test

// **Das Kostenjournal über MCP** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type kostenprobe struct {
	antwort mcp.Haushaltsstand
	fehler  error

	konto, zeitraum string
}

func (p *kostenprobe) Kostenstand(_ context.Context, konto, zeitraum string) (
	mcp.Haushaltsstand, error,
) {
	p.konto, p.zeitraum = konto, zeitraum

	return p.antwort, p.fehler
}

func TestOhneNahtGibtEsKeinKostenwerkzeug(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if werkzeug.Name == "lesen.kostenstand" {
			t.Fatal("das Werkzeug stand ohne Naht da")
		}
	}
}

// **Die vier Zahlen kommen, und die Angaben reisen bis zur Naht.**
func TestDasKostenwerkzeugNenntDieVierZahlen(t *testing.T) {
	t.Parallel()

	probe := &kostenprobe{antwort: mcp.Haushaltsstand{
		Konto: "anbieter", Zeitraum: "monat",
		Grenze:     mcp.Betrag{Wert: 50_000, Waehrung: "EUR"},
		Unklar:     mcp.Betrag{Wert: 500, Waehrung: "EUR"},
		Verfuegbar: mcp.Betrag{Wert: 34_500, Waehrung: "EUR"},
	}}

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{Kosten: probe}, "lesen.kostenstand")

	antwort, err := werkzeug.Ruf(t.Context(),
		[]byte(`{"konto":"anbieter","zeitraum":"monat"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.konto != "anbieter" || probe.zeitraum != "monat" {
		t.Errorf("die Naht sah %q über %q", probe.konto, probe.zeitraum)
	}

	for _, teil := range []string{
		`"verfuegbar"`, `"wert": 34500`, `"waehrung": "EUR"`, `"unklar"`,
	} {
		if !strings.Contains(antwort, teil) {
			t.Errorf("Antwort %q nennt %q nicht", antwort, teil)
		}
	}

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("die Marken sind %q und %v", werkzeug.Art, werkzeug.Idempotent)
	}
}

// **Ein Fehler kommt durch** — eine leere Antwort sähe aus wie ein Konto ohne
// Grenze, und das ist die Auskunft, die niemals falsch sein darf.
func TestDasKostenwerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Kosten: &kostenprobe{fehler: errors.New("das Journal ist zu")},
	}, "lesen.kostenstand")

	if _, err := werkzeug.Ruf(t.Context(), []byte(`{"konto":"anbieter"}`)); err == nil {
		t.Error("der Fehler ging verloren")
	}

	if _, err := werkzeug.Ruf(t.Context(), []byte(`{"konto":`)); err == nil {
		t.Error("die krumme Eingabe kam durch")
	}
}
