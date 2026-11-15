package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Kommandozentrale: der Katalog realer Fähigkeiten** — Werkstück U12
// von Phase 5.1 (#1423).

func probebefehle() []app.Befehl {
	return []app.Befehl{
		{
			Name: "schreiben.karte.loeschen", Schreibend: true, Idempotent: true, Erklaerung: "löscht",
			Schema: map[string]any{"type": "object", "properties": map[string]any{"karte": map[string]any{"type": "string"}}},
			Ruf:    func(context.Context, json.RawMessage) (string, error) { return "gelöscht", nil },
		},
		{
			Name: "lesen.sets", Schreibend: false, Idempotent: true, Erklaerung: "listet",
			Schema: map[string]any{"type": "object"},
			Ruf: func(_ context.Context, argumente json.RawMessage) (string, error) {
				return `{"sets":[],"argumente":` + string(argumente) + `}`, nil
			},
		},
		{
			Name: "schreiben.karte.anlegen", Schreibend: true, Idempotent: false, Erklaerung: "legt an",
			Ruf: func(context.Context, json.RawMessage) (string, error) { return "", errors.New("kein Set") },
		},
		{
			Name: "schreiben.feld", Schreibend: true, Idempotent: true, Erklaerung: "setzt",
			Ruf: func(context.Context, json.RawMessage) (string, error) { return "gesetzt", nil },
		},
		{
			Name: "lesen.kaputt", Erklaerung: "scheitert",
			Ruf: func(context.Context, json.RawMessage) (string, error) { return "", errors.New("die Naht fehlt") },
		},
	}
}

func TestDerKatalogNenntJedenBefehlMitVerweisUndGefahr(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("").MitBefehlen(probebefehle())
	antwort := anwendung.Befehlskatalog(t.Context())

	if antwort.Fehler != nil || len(antwort.Befehle) != 5 {
		t.Fatalf("%+v", antwort)
	}

	// Sortiert nach Namen, mit den Verweisen aus der Fähigkeitentabelle.
	if antwort.Befehle[0].Name != "lesen.kaputt" || antwort.Befehle[1].Name != "lesen.sets" {
		t.Errorf("Reihenfolge: %s, %s", antwort.Befehle[0].Name, antwort.Befehle[1].Name)
	}

	sets := antwort.Befehle[1]
	if sets.CLI != "mse sets" || sets.REST != "GET /sets" || sets.Faehigkeit != "Sets auflisten" ||
		sets.Schreibend || sets.Gefaehrlich {
		t.Errorf("lesen.sets: %+v", sets)
	}

	// **Gefährlich ist, was schreibt und nicht idempotent ist — oder löscht.**
	gefahr := map[string]bool{}
	for _, befehl := range antwort.Befehle {
		gefahr[befehl.Name] = befehl.Gefaehrlich
	}

	if !gefahr["schreiben.karte.loeschen"] || !gefahr["schreiben.karte.anlegen"] || gefahr["schreiben.feld"] || gefahr["lesen.sets"] {
		t.Errorf("Gefahr: %v", gefahr)
	}

	if antwort.Befehle[1].Schema["type"] != "object" {
		t.Errorf("Schema: %v", antwort.Befehle[1].Schema)
	}

	if gebunden := app.NeueBindungen(anwendung).Befehlskatalog(t.Context()); len(gebunden.Befehle) != 5 {
		t.Errorf("Fensterbindung: %+v", gebunden)
	}
}

func TestOhneQuelleIstDerKatalogLeerUndSagtEs(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").Befehlskatalog(t.Context())
	if antwort.Fehler == nil || len(antwort.Befehle) != 0 {
		t.Errorf("ohne Quelle: %+v", antwort)
	}
}

// Die Sichtprüfung vom 2026-09-09 fand die Kommandozentrale **tot**: das
// echte `lesen.sets` trägt kein Schema, die Naht schrieb `null`, und das
// Fenster griff in `schema.properties`. Die Probe oben gab jedem Befehl ein
// Objekt — eine Vorrichtung, die großzügiger war als der Betrieb (§3). Die
// Zusage gehört an die Naht, und geprüft wird die **serialisierte** Form.
func TestEinBefehlOhneSchemaWirdAlsLeeresObjektSerialisiert(t *testing.T) {
	a := app.NeueApp(t.TempDir()).MitBefehlen([]app.Befehl{{
		Name: "lesen.sets", Erklaerung: "listet",
		Ruf: func(context.Context, json.RawMessage) (string, error) { return "[]", nil },
	}})

	antwort := a.Befehlskatalog(t.Context())
	if antwort.Fehler != nil || len(antwort.Befehle) != 1 {
		t.Fatalf("Befehlskatalog: %+v", antwort)
	}

	roh, err := json.Marshal(antwort.Befehle[0])
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var zurueck struct {
		Schema map[string]any `json:"schema"`
	}
	if err := json.Unmarshal(roh, &zurueck); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if zurueck.Schema == nil {
		t.Fatalf("das Schema kam als null an: %s", roh)
	}

	if zurueck.Schema["type"] != "object" {
		t.Fatalf("das leere Schema soll ein Objekt beschreiben, bekam %v", zurueck.Schema)
	}
}
