package main

import (
	"bytes"
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
)

// **Eine stumme Verbindung hält das Dienstende nicht auf** — gefunden am
// 2026-10-07 in `make ci`, als `TestEinDienstendeMittenImAuftrag…` einmal von
// drei Läufen rot wurde: der Dienst endete mit „beim Dienstende liefen noch
// Aufträge", obwohl kein Auftrag mehr lief.
//
// `http.Server.Shutdown` hält eine Verbindung, die noch keine Anfrage
// geschickt hat (`StateNew`), erst nach **fünf Sekunden** für ruhend
// (`net/http/server.go`, Zeilen 3310–3314, Go 1.27.1, Issue 22682). Ein
// Klient, der eine Verbindung vorwählt und nicht braucht, hielt damit
// `Shutdown` die ganze Frist fest — und das Auftragsbuch, das erst danach
// schloss, bekam keine Zeit mehr und meldete einen Fehler.
func TestEineStummeVerbindungHaeltDasDienstendeNichtAuf(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	ctx, beenden := context.WithCancel(probekontext(t))
	angesagt := &sofort{ziel: &bytes.Buffer{}}

	fertig := make(chan error, 1)
	go func() { fertig <- lauf(ctx, []string{"-port", "0"}, angesagt) }()

	adresse, _ := warteAufAnsage(t, angesagt)

	// **Verbunden, aber nichts gesagt** — so wie eine vorgewählte Verbindung
	// im Pool eines Klienten.
	stumm, fehler := (&net.Dialer{}).DialContext(t.Context(), "tcp", adresse)
	if fehler != nil {
		t.Fatalf("verbinden: %v", fehler)
	}

	t.Cleanup(func() { _ = stumm.Close() }) //nolint:errcheck // beim Aufräumen ist nichts mehr zu retten

	// **Der Dienst muss sie gesehen haben**, sonst prüfte dieser Lauf ein Ende
	// ohne stumme Verbindung.
	time.Sleep(200 * time.Millisecond)

	angefangen := time.Now()
	beenden()

	select {
	case fehler := <-fertig:
		if fehler != nil {
			t.Errorf("der Dienst endete mit einem Fehler: %v", fehler)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("der Dienst lief 10 s nach dem Ende noch")
	}

	if dauer := time.Since(angefangen); dauer > 2*time.Second {
		t.Errorf("das Ende dauerte %v — eine stumme Verbindung hielt es auf", dauer)
	}
}
