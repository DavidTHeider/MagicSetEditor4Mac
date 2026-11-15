package mcp_test

// **Kein Pfad kommt unter einem anderen Namen am Vertrag vorbei** —
// #2073.
// Dieselbe Lücke wie beim Dienst: der Dekodierer legt `"Ziel"` in das Feld
// `json:"ziel"`, und `quelle` beim Installieren eines Pakets ist ein
// Verzeichnis, obwohl der Name nach keinem klingt.

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// probeserver hat ein Werkzeug mit den genannten Namensfeldern und
// Pfadfeldern; der Vertrag merkt sich, wonach er gefragt wird, und weist
// alles ab.
func probeserver(namen, pfade []string) (server *mcp.Server, gefragt *[]string, gerufen *int) {
	gefragt, gerufen = &[]string{}, new(int)

	server = mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "zählt",
		Schema: map[string]any{"type": "object"}, Namensfelder: namen, Pfadfelder: pfade,
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			*gerufen++

			return "gut", nil
		},
	}}).MitPfadvertrag(func(_ context.Context, pfad string) error {
		*gefragt = append(*gefragt, pfad)

		return errors.New("liegt außerhalb dessen, was der Dienst anfassen darf")
	})

	return server, gefragt, gerufen
}

func probeRuf(t *testing.T, server *mcp.Server, argumente string) string {
	t.Helper()

	antwort := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"schreiben.probe","arguments":`+argumente+`}}`)
	if len(antwort) != 1 {
		t.Fatalf("%v", antwort)
	}

	return antwort[0]
}

func TestEinPfadInAndererSchreibweiseGehtDurchDenVertrag(t *testing.T) {
	for _, argumente := range []string{`{"Ziel":"/fremd"}`, `{"PFAD":"/fremd"}`, `{"bildPfad":"/fremd"}`} {
		server, gefragt, gerufen := probeserver(nil, nil)

		if antwort := probeRuf(t, server, argumente); !strings.Contains(antwort, "außerhalb") ||
			*gerufen != 0 || !slices.Equal(*gefragt, []string{"/fremd"}) {
			t.Errorf("%s: gerufen %d, gefragt %v — %s", argumente, *gerufen, *gefragt, antwort)
		}
	}
}

// **Ein Namensfeld ist es in jeder Schreibweise** — der Dekodierer legt
// `"Datei"` in dasselbe Feld wie `"datei"`, und dort steht ein Name im
// Paket, kein Ort.
func TestEinNamensfeldGiltInJederSchreibweise(t *testing.T) {
	server, gefragt, gerufen := probeserver([]string{"datei"}, nil)

	if antwort := probeRuf(t, server, `{"Datei":"style"}`); !strings.Contains(antwort, "gut") || *gerufen != 1 || len(*gefragt) != 0 {
		t.Errorf("gerufen %d, gefragt %v — %s", *gerufen, *gefragt, antwort)
	}
}

// **Ein Pfadfeld eines Werkzeugs** geht durch den Vertrag, auch wenn sein
// Name nach keinem Pfad klingt — und nur bei dem Werkzeug, das es nennt.
func TestEinPfadfeldDesWerkzeugsGehtDurchDenVertrag(t *testing.T) {
	mit, gefragt, _ := probeserver(nil, []string{"quelle"})

	probeRuf(t, mit, `{"Quelle":"/fremd"}`)

	if !slices.Equal(*gefragt, []string{"/fremd"}) {
		t.Errorf("mit Pfadfeld gefragt: %v", *gefragt)
	}

	ohne, gefragt, gerufen := probeserver(nil, nil)

	probeRuf(t, ohne, `{"quelle":"/fremd"}`)

	if len(*gefragt) != 0 || *gerufen != 1 {
		t.Errorf("ohne Pfadfeld gefragt: %v, gerufen %d", *gefragt, *gerufen)
	}
}

// **Die Werkzeuge, die einen Pfad unter `quelle` oder `ort` nehmen, sagen
// es** — die vier Paketwege des Dienstes; einen Indexaufbau, den fünften,
// hat der Werkzeugserver nicht.
func TestDiePaketwerkzeugeNennenIhrePfadfelder(t *testing.T) {
	erwartet := map[string]string{
		"schreiben.paket.installieren": "quelle",
		"schreiben.paket.ersetzen":     "quelle",
		"lesen.paket.plan":             "quelle",
		"schreiben.paket.uebernehmen":  "ort",
	}

	for _, werkzeug := range mcp.Werkzeuge(probenaehte(&bestandsprobe{})) {
		if feld, gibt := erwartet[werkzeug.Name]; gibt {
			if !slices.Equal(werkzeug.Pfadfelder, []string{feld}) {
				t.Errorf("%s nennt %v, erwartet [%s]", werkzeug.Name, werkzeug.Pfadfelder, feld)
			}

			delete(erwartet, werkzeug.Name)
		} else if len(werkzeug.Pfadfelder) > 0 {
			t.Errorf("%s nennt %v und steht nicht in der Liste", werkzeug.Name, werkzeug.Pfadfelder)
		}
	}

	for name := range erwartet {
		t.Errorf("%s fehlt in der Werkzeugliste", name)
	}
}
