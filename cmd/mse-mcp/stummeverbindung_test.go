package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

// **Eine stumme Verbindung hält das Ende nicht auf** — dieselbe Klasse wie im
// Dienst (`cmd/mse-rest/stummeverbindung_test.go`): `http.Server.Shutdown`
// hält eine Verbindung ohne Anfrage erst nach fünf Sekunden für ruhend.
func TestEineStummeVerbindungHaeltDasEndeNichtAuf(t *testing.T) {
	t.Parallel()

	roh, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	ctx, ende := context.WithCancel(context.Background())
	beendet := make(chan error, 1)

	go func() {
		beendet <- bediene(ctx, roh, http.NotFoundHandler(), 2)
	}()

	stumm, fehler := (&net.Dialer{}).DialContext(t.Context(), "tcp", roh.Addr().String())
	if fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	t.Cleanup(func() { _ = stumm.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	time.Sleep(200 * time.Millisecond)

	angefangen := time.Now()
	ende()

	select {
	case fehler := <-beendet:
		if fehler != nil {
			t.Errorf("endete mit %v", fehler)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("lief 15 s nach dem Ende noch")
	}

	if dauer := time.Since(angefangen); dauer > 2*time.Second {
		t.Errorf("das Ende dauerte %v — eine stumme Verbindung hielt es auf", dauer)
	}
}
