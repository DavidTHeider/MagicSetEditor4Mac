package flaechen_test

// **Der Werkzeugserver über HTTP nimmt nicht mehr Verbindungen an als der
// Dienst** — Folgebefund F1 des Lastlaufs QA-08
// (#2003).
//
// Bis zum 2026-10-02 nahm `mse-mcp --adresse` beliebig viele an; `mse-rest`
// sechzehn. Gemessen wird am gebauten Programm, mit der Zahl, die `mse
// grenzen` nennt: so viele Verbindungen, die nichts sagen, halten jeden Platz
// bis zur Lesefrist — die nächste wartet, statt bedient zu werden.

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestDerWerkzeugserverUeberHTTPHaeltDieVerbindungsgrenze(t *testing.T) {
	umgebung := mitEigenerSaat(t)

	var grenzen struct {
		Verbindungen int `json:"verbindungen"`
	}

	if fehler := json.Unmarshal([]byte(mussKommandozeile(t, umgebung, "grenzen", "--json")), &grenzen); fehler != nil ||
		grenzen.Verbindungen <= 0 {
		t.Fatalf("mse grenzen: %+v, %v", grenzen, fehler)
	}

	angesagt := starteMCPAn(t, umgebung, "127.0.0.1")

	ziel, fehler := url.Parse(angesagt.adresse)
	if fehler != nil {
		t.Fatalf("die Adresse %q: %v", angesagt.adresse, fehler)
	}

	// **Erst der Beleg, dass er überhaupt antwortet** — sonst sagte das
	// Schweigen darunter nichts.
	liste := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`

	if status, rumpf := sendeMCP(t, angesagt.adresse, angesagt.marke, liste); status != http.StatusOK {
		t.Fatalf("ohne Gedränge: %d %s", status, rumpf)
	}

	gewaehlt := &net.Dialer{}
	stumme := make([]net.Conn, 0, grenzen.Verbindungen)

	t.Cleanup(func() {
		for _, verbindung := range stumme {
			_ = verbindung.Close() //nolint:errcheck // aufgeräumt wird ohnehin
		}
	})

	for range grenzen.Verbindungen {
		verbindung, wahlfehler := gewaehlt.DialContext(t.Context(), "tcp", ziel.Host)
		if wahlfehler != nil {
			t.Fatalf("wählen: %v", wahlfehler)
		}

		stumme = append(stumme, verbindung)
	}

	// **Die nächste wird nicht bedient** — sie wartet auf einen Platz.
	ungeduldig := &http.Client{Timeout: 1500 * time.Millisecond, Transport: &http.Transport{DisableKeepAlives: true}}
	if status, _ := rufMitKlient(t, ungeduldig, http.MethodPost, angesagt.adresse, angesagt.marke, liste); status != 0 {
		t.Errorf("die Verbindung über der Grenze wurde bedient: %d", status)
	}

	// **Und sie kommt durch, sobald ein Platz frei wird.**
	_ = stumme[0].Close() //nolint:errcheck // geschlossen ist geschlossen

	if status, rumpf := sendeMCP(t, angesagt.adresse, angesagt.marke, liste); status != http.StatusOK {
		t.Errorf("nach einem freien Platz: %d %s", status, rumpf)
	}
}
