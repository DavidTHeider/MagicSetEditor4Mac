package main

// **MCP über HTTP nimmt nicht mehr Verbindungen an als `mse-rest`** —
// Folgebefund F1 des Lastlaufs QA-08
// (#2003).
//
// Bis zum 2026-10-02 nahm der Werkzeugserver über HTTP beliebig viele
// Verbindungen an; der Dienst daneben sechzehn. Jetzt sitzt derselbe
// Lauscher davor (`rest.HoechstensGleichzeitig`), mit derselben Zahl aus
// `domain.Vorgabegrenzen`.

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestUeberHTTPNimmtNichtMehrVerbindungenAlsDieGrenze(t *testing.T) {
	t.Parallel()

	roh, fehler := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if fehler != nil {
		t.Fatalf("lauschen: %v", fehler)
	}

	const grenze = 2

	angekommen := make(chan struct{}, grenze+1)
	frei := make(chan struct{})

	behandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		angekommen <- struct{}{}
		<-frei
	})

	ctx, ende := context.WithCancel(context.Background())
	beendet := make(chan error, 1)

	go func() { beendet <- bediene(ctx, roh, behandler, grenze) }()

	// **Jede Anfrage auf einer eigenen Verbindung** — ein Pool, der eine
	// wiederverwendete, prüfte die Grenze nicht.
	klient := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	adresse := "http://" + roh.Addr().String() + "/mcp"

	for range grenze + 1 {
		go func() {
			anfrage, fehler := http.NewRequestWithContext(ctx, http.MethodPost, adresse, http.NoBody)
			if fehler != nil {
				return
			}

			if antwort, err := klient.Do(anfrage); err == nil {
				_ = antwort.Body.Close() //nolint:errcheck // Aufräumen, die Antwort ist nicht gefragt
			}
		}()
	}

	for range grenze {
		select {
		case <-angekommen:
		case <-time.After(5 * time.Second):
			t.Fatal("eine der erlaubten Verbindungen kam nicht an")
		}
	}

	select {
	case <-angekommen:
		t.Error("die Verbindung über der Grenze wurde angenommen")
	case <-time.After(300 * time.Millisecond):
	}

	// **Und sie wartet, statt abgewiesen zu werden**: wird ein Platz frei,
	// kommt sie an.
	close(frei)

	select {
	case <-angekommen:
	case <-time.After(5 * time.Second):
		t.Error("die wartende Verbindung kam nie an")
	}

	ende()

	if fehler := <-beendet; fehler != nil {
		t.Errorf("bediene: %v", fehler)
	}
}
