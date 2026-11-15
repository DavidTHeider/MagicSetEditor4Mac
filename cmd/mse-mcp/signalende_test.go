package main

// **Ein Signal beendet die Bedienung über stdio regulär** — BK-07c
// (#2117).
//
// Über HTTP endet `mse-mcp` auf SIGINT und SIGTERM mit Status 0; über stdio
// endete es bis zum 2026-10-06 gar nicht, solange die Eingabe offen war — und
// seit es endet, soll es dasselbe sagen: ein gewolltes Ende ist kein Fehler.

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinSignalBeendetDieBedienungUeberStdioRegulaer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ein, schreiben := io.Pipe()
	t.Cleanup(func() {
		if fehler := schreiben.Close(); fehler != nil {
			t.Error(fehler)
		}
	})

	ctx, abbrechen := context.WithCancel(storage.MitProfil(t.Context(), 1))
	antworten, aus := io.Pipe()
	ende := make(chan error, 1)

	go func() { ende <- lauf(ctx, nil, ein, aus) }()

	// **Erst steht der Server** — seine Antwort auf `ping` ist der Beleg —,
	// dann kommt das Signal.
	if _, fehler := io.WriteString(schreiben, `{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n"); fehler != nil {
		t.Fatal(fehler)
	}

	if zeile, fehler := bufio.NewReader(antworten).ReadString('\n'); fehler != nil || !strings.Contains(zeile, `"id":1`) {
		t.Fatalf("keine Antwort auf ping: %q, %v", zeile, fehler)
	}

	abbrechen()

	select {
	case fehler := <-ende:
		if fehler != nil {
			t.Errorf("ein Signal endete mit %v, erwartet ein reguläres Ende", fehler)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("die Bedienung endet nach dem Abbruch nicht")
	}
}
