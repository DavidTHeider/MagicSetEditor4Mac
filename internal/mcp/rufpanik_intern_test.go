package mcp

// **Ein Werkzeug, das zusammenbricht, reißt den Server nicht mit** —
// BK-05b-MCP (#2113).
//
// Ein Werkzeugruf läuft über stdio in einer eigenen Goroutine
// (`bedienen.go`, `beginne`). Eine Panik dort fängt niemand: nicht
// `net/http`, nicht Wails — der Prozess endet, und mit ihm jede andere
// Bedienung. Über HTTP fängt `net/http` sie zwar, aber ohne Antwort an das
// Gegenüber und ohne den Ruf aus der Liste der laufenden zu nehmen.

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// panikwerkzeug bricht zusammen, wenn es `{"panik":true}` bekommt.
func panikwerkzeug() Werkzeug {
	return Werkzeug{
		Name: "lesen.probe", Art: Lesend, Erklaerung: "bricht auf Wunsch zusammen",
		Schema: map[string]any{"type": "object"},
		Ruf: func(_ context.Context, argumente json.RawMessage) (string, error) {
			var wunsch struct {
				Panik bool `json:"panik"`
			}

			_ = json.Unmarshal(argumente, &wunsch) //nolint:errcheck // leer heißt: keine Panik

			if wunsch.Panik {
				panic("das Werkzeug bricht zusammen")
			}

			return "heil", nil
		},
	}
}

func TestEinZusammenbrechendesWerkzeugBekommtGenauEineFehlerantwort(t *testing.T) {
	t.Parallel()

	server := NeuerServer([]Werkzeug{panikwerkzeug()})
	einLesen, einSchreiben := io.Pipe()
	ausLesen, ausSchreiben := io.Pipe()

	ctx, abbrechen := context.WithCancel(t.Context())
	t.Cleanup(func() {
		abbrechen()
		_ = errors.Join(einSchreiben.Close(), ausLesen.Close()) //nolint:errcheck // Aufräumen
	})

	go func() { _ = server.Bediene(ctx, einLesen, ausSchreiben) }() //nolint:errcheck // endet mit dem Test

	gelesen := make(chan string, 8)

	go func() {
		zeilen := bufio.NewScanner(ausLesen)
		for zeilen.Scan() {
			gelesen <- zeilen.Text()
		}

		close(gelesen)
	}()

	ruf := func(id int, argumente string) {
		zeile := `{"jsonrpc":"2.0","id":` + string(rune('0'+id)) +
			`,"method":"tools/call","params":{"name":"lesen.probe","arguments":` + argumente + `}}` + "\n"
		if _, fehler := io.WriteString(einSchreiben, zeile); fehler != nil {
			t.Fatal(fehler)
		}
	}

	antwort := func(was string) Antwort {
		t.Helper()

		select {
		case zeile, offen := <-gelesen:
			if !offen {
				t.Fatalf("%s: die Bedienung endete ohne Antwort", was)
			}

			var gelesene Antwort
			if fehler := json.Unmarshal([]byte(zeile), &gelesene); fehler != nil {
				t.Fatalf("%s: %v in %s", was, fehler, zeile)
			}

			return gelesene
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: keine Antwort nach fünf Sekunden", was)
		}

		return Antwort{}
	}

	// **Erst die Wirkung über denselben Weg** (§3).
	ruf(1, `{}`)
	if heil := antwort("der heile Ruf"); heil.Fehler != nil {
		t.Fatalf("der heile Ruf: %+v", heil.Fehler)
	}

	ruf(2, `{"panik":true}`)

	zusammengebrochen := antwort("der zusammengebrochene Ruf")
	if zusammengebrochen.Fehler == nil || zusammengebrochen.Fehler.Code != codeInnererFehler ||
		string(zusammengebrochen.ID) != "2" {
		t.Errorf("der zusammengebrochene Ruf: %+v (Kennung %s)", zusammengebrochen.Fehler, zusammengebrochen.ID)
	}

	if zusammengebrochen.Fehler != nil && strings.Contains(zusammengebrochen.Fehler.Meldung, "bricht zusammen") {
		t.Errorf("die Meldung trägt den Grund der Panik nach außen: %q", zusammengebrochen.Fehler.Meldung)
	}

	// **Und danach geht derselbe Weg wieder** — ohne Leiche in der Liste.
	ruf(3, `{}`)
	if danach := antwort("der Ruf danach"); danach.Fehler != nil {
		t.Errorf("der Ruf danach: %+v", danach.Fehler)
	}

	server.laufsperre.Lock()
	uebrig := len(server.laufende)
	server.laufsperre.Unlock()

	if uebrig != 0 {
		t.Errorf("%d Rufe stehen noch als laufend", uebrig)
	}
}

// **Über HTTP dasselbe** — eine Antwort statt einer abgerissenen Leitung, und
// kein Ruf, der als laufend stehen bleibt.
func TestEinZusammenbrechendesWerkzeugAntwortetAuchUeberHTTP(t *testing.T) {
	t.Parallel()

	server := NeuerServer([]Werkzeug{panikwerkzeug()})
	behandler := UeberHTTPMit(server, "probemarke")

	rufe := func(argumente string) Antwort {
		t.Helper()

		anfrage := httptest.NewRequestWithContext(t.Context(), http.MethodPost, Endpunkt,
			strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"lesen.probe","arguments":`+argumente+`}}`))
		anfrage.Host = "127.0.0.1:7778" // der eigene Name (SEC-11)
		anfrage.Header.Set("Authorization", "Bearer probemarke")

		geschrieben := httptest.NewRecorder()
		behandler.ServeHTTP(geschrieben, anfrage)

		var antwort Antwort
		if fehler := json.Unmarshal(geschrieben.Body.Bytes(), &antwort); fehler != nil {
			t.Fatalf("%d %q: %v", geschrieben.Code, geschrieben.Body.String(), fehler)
		}

		return antwort
	}

	if heil := rufe(`{}`); heil.Fehler != nil {
		t.Fatalf("der heile Ruf: %+v", heil.Fehler)
	}

	if zusammengebrochen := rufe(`{"panik":true}`); zusammengebrochen.Fehler == nil ||
		zusammengebrochen.Fehler.Code != codeInnererFehler {
		t.Errorf("der zusammengebrochene Ruf: %+v", zusammengebrochen.Fehler)
	}

	if danach := rufe(`{}`); danach.Fehler != nil {
		t.Errorf("der Ruf danach: %+v", danach.Fehler)
	}

	server.laufsperre.Lock()
	uebrig := len(server.laufende)
	server.laufsperre.Unlock()

	if uebrig != 0 {
		t.Errorf("%d Rufe stehen noch als laufend", uebrig)
	}
}
