package mcp_test

// **Die Rückkehren über MCP** — Werkstück O7 von Phase 5.2
// (#1431).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type ruecksprungprobe struct {
	arten       []mcp.Ruecksprungart
	lage        mcp.Ruecksprunglage
	rueckkehren []mcp.Rueckkehr
	fehler      error

	// gefragt merkt sich den Pfad, den das Werkzeug weitergereicht hat.
	gefragt *string
}

func (r ruecksprungprobe) Ruecksprungarten(context.Context) ([]mcp.Ruecksprungart, error) {
	return r.arten, r.fehler
}

func (r ruecksprungprobe) Ruecksprunglage(_ context.Context, pfad string) (
	mcp.Ruecksprunglage, error,
) {
	if r.gefragt != nil {
		*r.gefragt = pfad
	}

	return r.lage, r.fehler
}

func (r ruecksprungprobe) Rueckkehren(context.Context) ([]mcp.Rueckkehr, error) {
	return r.rueckkehren, r.fehler
}

func TestOhneNahtGibtEsKeineRuecksprungwerkzeuge(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.HasPrefix(werkzeug.Name, "lesen.ruecksprung") ||
			werkzeug.Name == "lesen.rueckkehren" {
			t.Fatalf("das Werkzeug %q stand ohne Naht da", werkzeug.Name)
		}
	}
}

func TestDasArtenwerkzeugNenntDieHindernisse(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Ruecksprung: ruecksprungprobe{arten: []mcp.Ruecksprungart{
			{Art: "schema", Hindernis: "nurVorwaerts"},
		}},
	}, "lesen.ruecksprungarten")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if !strings.Contains(antwort, `"nurVorwaerts"`) {
		t.Errorf("Antwort %q", antwort)
	}
}

// **Der Pfad kommt aus der Eingabe** — ein Werkzeug, das ihn selbst wählte,
// wäre ein Dateidialog ohne Fenster.
func TestDasLagewerkzeugReichtDenPfadDurch(t *testing.T) {
	t.Parallel()

	var gefragt string

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Ruecksprung: ruecksprungprobe{
			gefragt: &gefragt,
			lage: mcp.Ruecksprunglage{
				Datenstand: "2026-09-01T08:30:00Z",
				Unklar:     []mcp.Menge{{Schluessel: "scryfall:abc", Menge: 3}},
			},
		},
	}, "lesen.ruecksprunglage")

	antwort, err := werkzeug.Ruf(t.Context(),
		json.RawMessage(`{"pfad":"/probe.msesicherung"}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if gefragt != "/probe.msesicherung" {
		t.Errorf("gefragt wurde nach %q", gefragt)
	}

	if !strings.Contains(antwort, `"scryfall:abc"`) {
		t.Errorf("Antwort %q", antwort)
	}
}

// **Eine unbrauchbare Eingabe kommt als Fehler zurück**, nicht als leere Lage.
func TestDasLagewerkzeugWeistEineKrummeEingabeAb(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Ruecksprung: ruecksprungprobe{},
	}, "lesen.ruecksprunglage")

	if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{"pfad":42}`)); err == nil {
		t.Error("die krumme Eingabe kam durch")
	}
}

func TestDasRueckkehrwerkzeugNenntDenKlaerungsstand(t *testing.T) {
	t.Parallel()

	werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
		Ruecksprung: ruecksprungprobe{rueckkehren: []mcp.Rueckkehr{
			{ID: 7, Klaerung: "offen", Quelle: "bestand.msesicherung"},
		}},
	}, "lesen.rueckkehren")

	antwort, err := werkzeug.Ruf(t.Context(), nil)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if !strings.Contains(antwort, `"klaerung": "offen"`) {
		t.Errorf("Antwort %q", antwort)
	}
}

// **Jedes der drei meldet seinen Fehler** — eine leere Antwort sähe aus wie
// „nichts da", und bei den Rückkehren ist das die eine Auskunft, die niemals
// falsch sein darf.
func TestJedesRuecksprungwerkzeugMeldetSeinenFehler(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"lesen.ruecksprungarten", "lesen.ruecksprunglage", "lesen.rueckkehren",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			werkzeug := gesundheitswerkzeug(t, mcp.Naehte{
				Ruecksprung: ruecksprungprobe{fehler: errors.New("das Buch ist zu")},
			}, name)

			if _, err := werkzeug.Ruf(t.Context(),
				json.RawMessage(`{"pfad":"/probe.msesicherung"}`)); err == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}
