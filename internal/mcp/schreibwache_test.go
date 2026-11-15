package mcp_test

// **Die Schreibwache je Ruf** — ID-03.AK-02
// (#1702):
// der Server hält seine Verbindung über seine ganze Laufzeit, aber wer
// schreiben darf, entscheidet sich bei jedem Ruf neu.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Ein schreibendes Werkzeug fragt vor dem Lauf, ein lesendes nicht — und
// eines, das den Zeiger bewegt, auch nicht**: sonst käme niemand aus einem
// Profil wieder heraus, in dem er nur lesen darf.
func TestDieSchreibwacheStehtVorJedemSchreibendenRuf(t *testing.T) {
	gerufen := map[string]int{}

	werkzeug := func(name string, art mcp.Art, zeiger bool) mcp.Werkzeug {
		return mcp.Werkzeug{
			Name: name, Art: art, Erklaerung: "zählt", Zeigerwechsel: zeiger,
			Schema: map[string]any{"type": "object"},
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				gerufen[name]++

				return "gut", nil
			},
		}
	}

	gefragt := 0

	server := mcp.NeuerServer([]mcp.Werkzeug{
		werkzeug("lesen.probe", mcp.Lesend, false),
		werkzeug("schreiben.probe", mcp.Schreibend, false),
		werkzeug("schreiben.profil.probe", mcp.Schreibend, true),
	}).MitSchreibwache(func(context.Context) error {
		gefragt++

		return errors.New("in diesem Profil wird nur gelesen")
	})

	for _, name := range []string{"lesen.probe", "schreiben.probe", "schreiben.profil.probe"} {
		antwort := antworten(t, server,
			`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"`+name+`","arguments":{}}}`)

		verwehrt := len(antwort) == 1 && strings.Contains(antwort[0], "nur gelesen") &&
			strings.Contains(antwort[0], `"isError":true`)

		if erwartet := name == "schreiben.probe"; verwehrt != erwartet {
			t.Errorf("%s: verwehrt %v, erwartet %v — %v", name, verwehrt, erwartet, antwort)
		}
	}

	if gerufen["schreiben.probe"] != 0 {
		t.Error("das verwehrte Werkzeug lief trotzdem")
	}

	if gerufen["lesen.probe"] != 1 || gerufen["schreiben.profil.probe"] != 1 {
		t.Errorf("gerufen %v, erwartet je einmal für das lesende und das zeigerbewegende", gerufen)
	}

	if gefragt != 1 {
		t.Errorf("die Wache wurde %d-mal gefragt, erwartet einmal", gefragt)
	}
}
