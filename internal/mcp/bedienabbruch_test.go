package mcp_test

// **Ein Abbruch beendet die Bedienung, auch wenn niemand auflegt** — BK-07c
// (#2117).
//
// `mse-mcp` hängt `SIGINT` und `SIGTERM` an seinen Kontext. Bis zum
// 2026-10-06 wartete `Bediene` aber nur auf eine Zeile oder das Ende eines
// Rufs: über stdio mit offener Eingabe endete der Server erst, wenn das
// Gegenüber auflegte — gemessen im Audit für 0.7.0 nach Abbruch und 150 ms
// noch blockiert.

import (
	"bufio"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

func TestEinAbbruchBeendetDieBedienungTrotzOffenerEingabe(t *testing.T) {
	t.Parallel()

	server := mcp.NeuerServer(nil)
	einLesen, einSchreiben := io.Pipe()
	ausLesen, ausSchreiben := io.Pipe()
	t.Cleanup(func() {
		if fehler := errors.Join(einSchreiben.Close(), ausLesen.Close()); fehler != nil {
			t.Error(fehler)
		}
	})

	ctx, abbrechen := context.WithCancel(t.Context())
	ende := make(chan error, 1)

	go func() { ende <- server.Bediene(ctx, einLesen, ausSchreiben) }()

	// **Erst der Beleg, dass jemand zuhört** (§3): ein `ping` über denselben
	// Weg bekommt seine Antwort.
	antworten := bufio.NewScanner(ausLesen)
	if _, fehler := io.WriteString(einSchreiben, `{"jsonrpc":"2.0","id":1,"method":"ping"}`+"\n"); fehler != nil {
		t.Fatal(fehler)
	}

	gelesen := make(chan string, 4)

	go func() {
		for antworten.Scan() {
			gelesen <- antworten.Text()
		}

		close(gelesen)
	}()

	select {
	case zeile := <-gelesen:
		t.Logf("vor dem Abbruch: %s", zeile)
	case <-time.After(5 * time.Second):
		t.Fatal("das ping blieb ohne Antwort — niemand hört zu")
	}

	abbrechen()

	select {
	case fehler := <-ende:
		if !errors.Is(fehler, context.Canceled) {
			t.Errorf("die Bedienung endete mit %v, erwartet context.Canceled", fehler)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("die Bedienung läuft nach dem Abbruch weiter, solange die Eingabe offen ist")
	}

	// **Und danach geht keine Antwort mehr hinaus**, auch nicht auf eine
	// Zeile, die noch kommt.
	if _, fehler := io.WriteString(einSchreiben, `{"jsonrpc":"2.0","id":2,"method":"ping"}`+"\n"); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := ausSchreiben.Close(); fehler != nil {
		t.Fatal(fehler)
	}

	for zeile := range gelesen {
		t.Errorf("nach dem Abbruch ging noch hinaus: %s", zeile)
	}
}
