package mcp_test

// **Ein Ende beantwortet den laufenden Ruf noch** — M02
// (#2346).
//
// Die Fassungswache bricht den Kontext ab, wenn eine neuere Fassung den
// Bestand gewandert hat, und lehnt den Ruf ab, der sie fragte. Über stdio
// kehrte die Bedienung dann sofort zurück — die Ablehnung ging nie hinaus,
// und das Gegenüber sah nur das Ende (gemessen mit einem echten `mse-mcp` in
// `internal/flaechen`). Am Dienst wartet `http.Server.Shutdown` die laufende
// Anfrage ab; hier jetzt auch.

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestEinEndeBeantwortetDenLaufendenRufNoch(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.probe", Art: mcp.Schreibend, Erklaerung: "schreibt",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			t.Error("das Werkzeug lief trotz der Wache")

			return "", nil
		},
	}}).MitSchreibwache(func(context.Context) error {
		abbrechen()

		return &domain.ProzessVeraltet{Fassung: &domain.NeuereFassung{Unbekannt: []string{"0099_kuenftig.sql"}}}
	})

	einLesen, einSchreiben := io.Pipe()
	ausLesen, ausSchreiben := io.Pipe()

	t.Cleanup(func() { _ = errors.Join(einSchreiben.Close(), ausLesen.Close()) }) //nolint:errcheck // Aufräumen

	ende := make(chan error, 1)

	go func() {
		ende <- server.Bediene(ctx, einLesen, ausSchreiben)
		_ = ausSchreiben.Close() //nolint:errcheck // das Ende der Ausgabe
	}()

	gelesen := make(chan string, 8)

	go func() {
		antworten := bufio.NewScanner(ausLesen)
		for antworten.Scan() {
			gelesen <- antworten.Text()
		}

		close(gelesen)
	}()

	for _, zeile := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"schreiben.probe","arguments":{}}}`,
	} {
		if _, fehler := io.WriteString(einSchreiben, zeile+"\n"); fehler != nil {
			t.Fatal(fehler)
		}
	}

	var alle []string

	// **Mit Frist**: ein Server, der nicht endet, hielte das Lesen sonst
	// für immer offen.
	frist := time.After(5 * time.Second)

lesen:
	for {
		select {
		case zeile, offen := <-gelesen:
			if !offen {
				break lesen
			}

			alle = append(alle, zeile)
		case <-frist:
			break lesen
		}
	}

	if len(alle) != 2 || !strings.Contains(alle[1], `"id":2`) || !strings.Contains(alle[1], "älter als der Bestand") {
		t.Errorf("vor dem Ende kam an: %q — erwartet die Ablehnung des laufenden Rufs", alle)
	}

	select {
	case fehler := <-ende:
		if !errors.Is(fehler, context.Canceled) {
			t.Errorf("die Bedienung endete mit %v", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("die Bedienung endete nicht")
	}
}

// **Und ein Ruf, der in der Nachfrist nicht fertig wird, hält das Ende nicht
// auf** — sein Kontext ist abgebrochen, und wer ihn ignoriert, bleibt ohne
// Antwort. Nicht parallel: die Frist ist eine Paketvariable.
func TestEinEndeWartetHoechstensDieNachfrist(t *testing.T) { //nolint:paralleltest // siehe darüber
	alteFrist := mcp.Nachfrist
	mcp.Nachfrist = 50 * time.Millisecond

	t.Cleanup(func() { mcp.Nachfrist = alteFrist })

	ctx, abbrechen := context.WithCancel(t.Context())
	los := make(chan struct{})

	t.Cleanup(func() { close(los) })

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "schreiben.zaeh", Art: mcp.Schreibend, Erklaerung: "hängt",
		Schema: map[string]any{"type": "object"},
		Ruf: func(context.Context, json.RawMessage) (string, error) {
			abbrechen()
			<-los

			return "", nil
		},
	}})

	einLesen, einSchreiben := io.Pipe()
	t.Cleanup(func() { _ = einSchreiben.Close() }) //nolint:errcheck // Aufräumen

	ende := make(chan error, 1)

	go func() { ende <- server.Bediene(ctx, einLesen, io.Discard) }()

	for _, zeile := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"probe","version":"1"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"schreiben.zaeh","arguments":{}}}`,
	} {
		if _, fehler := io.WriteString(einSchreiben, zeile+"\n"); fehler != nil {
			t.Fatal(fehler)
		}
	}

	select {
	case fehler := <-ende:
		if !errors.Is(fehler, context.Canceled) {
			t.Errorf("die Bedienung endete mit %v", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ein zäher Ruf hielt das Ende auf")
	}
}
