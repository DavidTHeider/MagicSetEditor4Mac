package main

// **Der Werkzeugserver über HTTP auf mehreren Adressen** — G08
// (#2272).
//
// `--adresse` lässt sich wiederholen; ohne sie bleibt es die
// Standardeingabe — die Lauschwahl aus den Einstellungen gibt das Fenster
// mit, wenn es ihn startet. Jede angesagte Adresse nimmt an und weist eine
// Anfrage **nicht** als fremde Herkunft ab: über `::1` tat sie das seit N6.

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// mitschrift ist gegen gleichzeitiges Lesen gesichert.
type mitschrift struct {
	schloss sync.Mutex
	puffer  bytes.Buffer
}

func (m *mitschrift) Write(roh []byte) (int, error) {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	return m.puffer.Write(roh) //nolint:wrapcheck // ein Puffer im Speicher
}

func (m *mitschrift) text() string {
	m.schloss.Lock()
	defer m.schloss.Unlock()

	return m.puffer.String()
}

func TestDerWerkzeugserverLauschtAufJederGenanntenAdresse(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	ctx, beenden := context.WithCancel(probekontext(t))
	angesagt := &mitschrift{}

	fertig := make(chan error, 1)
	go func() {
		fertig <- lauf(ctx, []string{"-port", "0", "-adresse", "127.0.0.1", "-adresse", "::1"},
			strings.NewReader(""), angesagt)
	}()

	var adressen []string

	frist := time.After(60 * time.Second)

	for marke := ""; marke == ""; {
		select {
		case <-frist:
			t.Fatalf("nichts angesagt: %q", angesagt.text())
		case <-time.After(20 * time.Millisecond):
		}

		adressen = nil

		for _, zeile := range strings.Split(angesagt.text(), "\n") {
			if wert, ja := strings.CutPrefix(zeile, "Dienst auf "); ja {
				adressen = append(adressen, wert)
			}

			if wert, ja := strings.CutPrefix(zeile, "Marke "); ja {
				marke = wert
			}
		}
	}

	if len(adressen) != 2 || !strings.HasPrefix(adressen[0], "http://127.0.0.1:") || !strings.HasPrefix(adressen[1], "http://[::1]:") {
		t.Fatalf("angesagt sind %v", adressen)
	}

	for _, adresse := range adressen {
		anfrage, fehler := http.NewRequestWithContext(ctx, http.MethodPost, adresse, http.NoBody)
		if fehler != nil {
			t.Fatal(fehler)
		}

		antwort, fehler := http.DefaultClient.Do(anfrage)
		if fehler != nil {
			t.Fatalf("über %s: %v", adresse, fehler)
		}

		_ = antwort.Body.Close() //nolint:errcheck // gelesen wird nur der Status

		if antwort.StatusCode == http.StatusForbidden {
			t.Errorf("über %s: als fremde Herkunft abgewiesen", adresse)
		}
	}

	beenden()

	if fehler := <-fertig; fehler != nil {
		t.Errorf("lauf: %v", fehler)
	}
}
