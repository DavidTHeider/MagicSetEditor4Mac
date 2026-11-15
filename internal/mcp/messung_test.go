package mcp_test

// **Jeder Werkzeugruf wird gemessen** — OPS-05.AK-01/02
// (#1722):
// mit dem Namen des Werkzeugs, ob es schreibt, wie lange es dauerte und wie
// es ausging — und mit dem Kontext, den die Ausstattung gerichtet hat, denn
// **dort** steht die Vorgangskennung.

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type kennungsschluessel struct{}

func TestJederWerkzeugrufWirdGemessen(t *testing.T) {
	t.Parallel()

	var (
		sperre    sync.Mutex
		messungen []mcp.Messung
		kennungen []string
	)

	gut := mcp.Werkzeug{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "Gelingt.",
		Ruf: func(context.Context, json.RawMessage) (string, error) { return "{}", nil },
	}
	abgewiesen := mcp.Werkzeug{
		Name: "lesen.falsch", Art: mcp.Lesend, Idempotent: true, Erklaerung: "Weist ab.",
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			return "", domain.NewFieldError("x", domain.SchluesselValidierung, "falsch")
		},
	}

	server := mcp.NeuerServer([]mcp.Werkzeug{gut, abgewiesen, scheiterndesWerkzeug()}).
		MitAusstattung(func(ctx context.Context) (context.Context, error) {
			return context.WithValue(ctx, kennungsschluessel{}, "0a1b2c3d"), nil
		}).
		MitMessung(func(ctx context.Context, messung mcp.Messung) {
			sperre.Lock()
			defer sperre.Unlock()

			messungen = append(messungen, messung)
			kennung, gesetzt := ctx.Value(kennungsschluessel{}).(string)
			if !gesetzt {
				kennung = ""
			}

			kennungen = append(kennungen, kennung)
		})

	var zeilen strings.Builder
	for stelle, name := range []string{"schreiben.probe", "lesen.falsch", "lesen.probe", "gibt.es.nicht"} {
		zeilen.WriteString(`{"jsonrpc":"2.0","id":` + string(rune('1'+stelle)) + `,"method":"tools/call",` +
			`"params":{"name":"` + name + `","arguments":{}}}` + "\n")
	}

	var aus strings.Builder
	if fehler := server.Bediene(t.Context(), strings.NewReader(zeilen.String()), &aus); fehler != nil {
		t.Fatal(fehler)
	}

	erwartet := []mcp.Messung{
		{Weg: "schreiben.probe", Schreibend: true, Ausgang: domain.RufGelungen},
		{Weg: "lesen.falsch", Ausgang: domain.RufAbgewiesen},
		{Weg: "lesen.probe", Ausgang: domain.RufGescheitert},
		{Weg: mcp.UnbekanntesWerkzeug, Ausgang: domain.RufAbgewiesen},
	}

	if len(messungen) != len(erwartet) {
		t.Fatalf("%d Messungen: %+v", len(messungen), messungen)
	}

	for stelle, soll := range erwartet {
		ist := messungen[stelle]
		if ist.Weg != soll.Weg || ist.Schreibend != soll.Schreibend || ist.Ausgang != soll.Ausgang || ist.Dauer < 0 {
			t.Errorf("%d: %+v, erwartet %+v", stelle, ist, soll)
		}
	}

	// **Die ersten drei mit Kennung** — sie liefen durch die Ausstattung;
	// das unbekannte Werkzeug kam nie dort an.
	if kennungen[0] != "0a1b2c3d" || kennungen[2] != "0a1b2c3d" || kennungen[3] != "" {
		t.Errorf("Kennungen: %v", kennungen)
	}
}
