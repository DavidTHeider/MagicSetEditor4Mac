package service_test

// **Eine Fähigkeit messen, statt sie zu glauben** — Werkstück A9 von
// Phase 5.3
// (#1448).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// probequelle antwortet, wie ein Lauf es vorgibt.
type probequelle struct {
	fassung  string
	antwort  service.KIProbeantwort
	fehler   error
	gefragt  []string
	schemen  []map[string]any
	fassfehl error
}

func (q *probequelle) Fassung(context.Context) (string, error) {
	return q.fassung, q.fassfehl
}

func (q *probequelle) Vorschlagen(
	_ context.Context, alias, auftrag string, schema map[string]any,
) (service.KIProbeantwort, error) {
	q.gefragt = append(q.gefragt, alias+"|"+auftrag)
	q.schemen = append(q.schemen, schema)

	return q.antwort, q.fehler
}

// probenablage merkt sich, was geschrieben wurde.
type probenablage struct {
	gemerkt []domain.KIModellprobe
	fehler  error
}

func (a *probenablage) Merken(
	_ context.Context, _ int64, probe domain.KIModellprobe,
) error {
	if a.fehler != nil {
		return a.fehler
	}

	a.gemerkt = append(a.gemerkt, probe)

	return nil
}

// eineUhr gibt einen festen Tag.
func eineUhr() time.Time {
	return time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)
}

// **Eine schemakonforme Antwort belegt zwei Fähigkeiten.**
//
// **Zwei und nicht eine**: sie zeigt, dass das Modell Text erzeugt **und**
// eine Form einhält. Wer nur das zweite führte, ließe ein Modell ungeprüft,
// das gerade eben geantwortet hat.
func TestEineSchemakonformeAntwortBelegtZweiFaehigkeiten(t *testing.T) {
	t.Parallel()

	ablage := &probenablage{}
	quelle := &probequelle{
		fassung: "Responses-API /v1/responses",
		antwort: service.KIProbeantwort{
			Inhalt: `{"probe": "bereit"}`, Modell: "gpt-probe-2026-09-01",
		},
	}

	proben, fehler := service.NeuesKIModellpruefwerk(ablage, eineUhr).
		Pruefen(t.Context(), 1, quelle, "openai", "gpt-probe")
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if len(proben) != 2 || len(ablage.gemerkt) != 2 {
		t.Fatalf("%d Messungen, %d gemerkt", len(proben), len(ablage.gemerkt))
	}

	gesehen := map[domain.KIFaehigkeit]bool{}
	for _, probe := range ablage.gemerkt {
		gesehen[probe.Faehigkeit] = true

		if probe.Datum != "2026-09-12" || probe.Modell != "gpt-probe-2026-09-01" {
			t.Errorf("die Messung ist %+v", probe)
		}

		// **Der Beleg sagt, wie gemessen wurde** — „geprüft" ohne das Wie
		// ist eine Behauptung (§10).
		if befunde := probe.Validate(); len(befunde) != 0 {
			t.Errorf("die Messung ist unvollständig: %v", befunde)
		}
	}

	if !gesehen[domain.KITextgenerierung] ||
		!gesehen[domain.KIStrukturierteGenerierung] {
		t.Errorf("belegt sind %v", gesehen)
	}

	// **Und das Schema ging wirklich mit.** Ohne diese Zusicherung sagte der
	// Lauf nur, dass etwas gefragt wurde — nicht, dass eine **Form**
	// verlangt war.
	if len(quelle.schemen) != 1 || quelle.schemen[0]["type"] != "object" {
		t.Errorf("die Schemen sind %+v", quelle.schemen)
	}
}

// **Eine Antwort, die die Form nicht hält, belegt nichts.**
//
// **Und schreibt nichts**: eine Messung, die danebenliegt, wäre schlimmer
// als keine — sie stünde im Katalog und sähe aus wie ein Beleg.
func TestEineAntwortOhneFormBelegtNichts(t *testing.T) {
	t.Parallel()

	for name, antwort := range map[string]service.KIProbeantwort{
		"kein JSON":          {Inhalt: "gerne doch!"},
		"JSON ohne das Feld": {Inhalt: `{"etwas": "anderes"}`},
		"abgeschnitten": {
			Inhalt: `{"probe": "bereit"}`, Abgeschnitten: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := &probenablage{}

			_, fehler := service.NeuesKIModellpruefwerk(ablage, eineUhr).
				Pruefen(t.Context(), 1, &probequelle{
					fassung: "x", antwort: antwort,
				}, "openai", "gpt-probe")

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if len(ablage.gemerkt) != 0 {
				t.Errorf("trotzdem gemerkt: %+v", ablage.gemerkt)
			}
		})
	}
}

// **Ohne Anbieter oder Modell wird gar nicht erst gefragt.**
func TestOhneAnbieterOderModellWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct{ anbieter, alias string }{
		"ohne Anbieter": {"", "gpt-probe"},
		"ohne Modell":   {"openai", "  "},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			quelle := &probequelle{fassung: "x"}

			if _, fehler := service.NeuesKIModellpruefwerk(&probenablage{}, eineUhr).
				Pruefen(t.Context(), 1, quelle, fall.anbieter, fall.alias); !errors.Is(
				fehler, domain.ErrValidation) {
				t.Fatalf("der Fehler ist %v", fehler)
			}

			if len(quelle.gefragt) != 0 {
				t.Errorf("trotzdem gefragt: %v", quelle.gefragt)
			}
		})
	}
}

// **Ein Fehler des Anbieters kommt durch und wird nicht zur Messung.**
func TestEinFehlerDesAnbietersKommtDurch(t *testing.T) {
	t.Parallel()

	for name, quelle := range map[string]*probequelle{
		"die Fassung ist nicht zu lesen": {
			fassfehl: errors.New("nichts da"),
		},
		"die Antwort scheitert": {
			fassung: "x", fehler: errors.New("abgelehnt"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ablage := &probenablage{}

			if _, fehler := service.NeuesKIModellpruefwerk(ablage, eineUhr).
				Pruefen(t.Context(), 1, quelle, "openai",
					"gpt-probe"); fehler == nil {
				t.Fatal("der Fehler kam nicht durch")
			}

			if len(ablage.gemerkt) != 0 {
				t.Errorf("trotzdem gemerkt: %+v", ablage.gemerkt)
			}
		})
	}
}

// **Ein Schreibfehler kommt ebenso durch.**
func TestEinSchreibfehlerBeiEinerMessungKommtDurch(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuesKIModellpruefwerk(
		&probenablage{fehler: errors.New("die Platte ist voll")}, eineUhr).
		Pruefen(t.Context(), 1, &probequelle{
			fassung: "x",
			antwort: service.KIProbeantwort{Inhalt: `{"probe": "bereit"}`},
		}, "openai", "gpt-probe")

	if fehler == nil {
		t.Fatal("der Schreibfehler kam nicht durch")
	}
}

// **Nennt der Anbieter kein Modell, steht der Alias da.**
//
// **Und nicht nichts**: eine Messung ohne Modell fiele an ihrer eigenen
// Prüfung durch, und dann stünde gar nichts fest.
func TestOhneGenanntesModellStehtDerAlias(t *testing.T) {
	t.Parallel()

	ablage := &probenablage{}

	if _, fehler := service.NeuesKIModellpruefwerk(ablage, eineUhr).
		Pruefen(t.Context(), 1, &probequelle{
			fassung: "x",
			antwort: service.KIProbeantwort{Inhalt: `{"probe": "bereit"}`},
		}, "openai", "gpt-probe"); fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if len(ablage.gemerkt) == 0 || ablage.gemerkt[0].Modell != "gpt-probe" {
		t.Fatalf("die Messungen sind %+v", ablage.gemerkt)
	}
}

// **Der Auftrag nennt das Wort, um das er bittet.**
//
// **Sonst prüfte der Lauf eine Form gegen eine Bitte, die sie nicht
// verlangt** — und ein Modell, das etwas anderes schriebe, fiele zu Recht
// durch, ohne dass jemand sagen könnte, woran es lag.
func TestDerAuftragNenntDasWortUmDasErBittet(t *testing.T) {
	t.Parallel()

	auftrag := service.DerProbeauftrag()

	eigenschaften, sind := service.DasProbeschema()["properties"].(map[string]any)
	if !sind {
		t.Fatalf("das Schema ist %+v", service.DasProbeschema())
	}

	if _, hat := eigenschaften["probe"]; !hat {
		t.Fatalf("die Eigenschaften sind %+v", eigenschaften)
	}

	if !strings.Contains(auftrag, "probe") {
		t.Errorf("der Auftrag ist %q", auftrag)
	}
}
