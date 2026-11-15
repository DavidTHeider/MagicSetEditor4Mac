package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// **Die Grenzen als Werkzeug, und das Guthaben am Ruf** — Werkstück O5 von
// Phase 5.2 (#1429).

type grenzenprobe struct {
	grenzen mcp.Grenzen
	fehler  error
}

func (g grenzenprobe) Grenzen(context.Context) (mcp.Grenzen, error) {
	return g.grenzen, g.fehler
}

func grenzenwerkzeug(t *testing.T, naht mcp.Grenzenauskunft) (mcp.Werkzeug, bool) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Grenzen: naht}) {
		if werkzeug.Name == "lesen.grenzen" {
			return werkzeug, true
		}
	}

	return mcp.Werkzeug{}, false
}

func TestDasGrenzenwerkzeugNenntDieGrenzen(t *testing.T) {
	t.Parallel()

	werkzeug, gibts := grenzenwerkzeug(t, grenzenprobe{grenzen: mcp.Grenzen{
		Parallel: map[string]int{"rendern": 2}, AntwortBytes: 1024,
		Principal: mcp.Kontingent{Burst: 120, Nachfuellen: 120, FensterSekunden: 60},
		Reserve:   120, Kosten: map[string]int{"lesen": 1}, Kontingente: "aus",
	}})
	if !gibts {
		t.Fatal("das Werkzeug lesen.grenzen gibt es nicht")
	}

	// **Und die Erklärung nennt den Schalter** (QA-08.AK-03, #1739) — ein
	// Modell, das ihn nicht kennt, deutet ein ausbleibendes Nein falsch.
	if !strings.Contains(werkzeug.Erklaerung, "»kontingente«") {
		t.Errorf("die Erklärung nennt die Kontingente nicht: %s", werkzeug.Erklaerung)
	}

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("das Werkzeug ist %v und idempotent=%v — es liest nur", werkzeug.Art, werkzeug.Idempotent)
	}

	text, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	for _, erwartet := range []string{
		`"rendern": 2`, `"reserve": 120`, `"lesen": 1`, `"antwortBytes": 1024`, `"kontingente": "aus"`,
	} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Antwort nennt %s nicht: %s", erwartet, text)
		}
	}

	// Ohne Naht kein Werkzeug; mit einer, die scheitert, ihr Fehler.
	if _, gibts := grenzenwerkzeug(t, nil); gibts {
		t.Error("ohne Naht ein Werkzeug")
	}

	kaputt, _ := grenzenwerkzeug(t, grenzenprobe{fehler: errors.New("kein Haushalt")})
	if _, err := kaputt.Ruf(t.Context(), json.RawMessage(`{}`)); err == nil || !strings.Contains(err.Error(), "kein Haushalt") {
		t.Errorf("der Fehler ging verloren: %v", err)
	}
}

func TestEinRufUeberDemGuthabenBekommtDieWartezeit(t *testing.T) {
	t.Parallel()

	var gefragt []string

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Idempotent: true,
		Ruf: func(context.Context, json.RawMessage) (string, error) { return "geliefert", nil },
	}}).MitKontingent(func(_ context.Context, _ mcp.Art, name string) (bool, time.Duration) {
		gefragt = append(gefragt, name)

		return len(gefragt) == 1, 4500 * time.Millisecond
	})

	zeilen := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if len(zeilen) != 2 || !strings.Contains(zeilen[0], "geliefert") {
		t.Fatalf("Antworten: %v", zeilen)
	}

	// **Aufgerundet auf ganze Sekunden, als Fehler des Werkzeugs** — kein
	// Protokollfehler, denn das Protokoll war in Ordnung.
	if !strings.Contains(zeilen[1], "in 5 s wieder") || !strings.Contains(zeilen[1], `"isError":true`) {
		t.Errorf("über dem Guthaben: %s", zeilen[1])
	}

	if len(gefragt) != 2 || gefragt[0] != "lesen.probe" {
		t.Errorf("gefragt: %v", gefragt)
	}
}

// **Und mit Fehlerform als Form** — Befund 5 des Zweitblicks über Phase 1 von
// 0.9.0 (F03, #2342): die Ablehnung baute ihr Ergebnis von Hand und trug kein
// `structuredContent.fehler`. Die Wurzel bekommt [domain.ErrKontingent] mit
// der Wartezeit — dieselbe Klasse, die der Dienst als `drossel` nennt.
func TestEinRufUeberDemGuthabenTraegtSeineForm(t *testing.T) {
	t.Parallel()

	var gesehen error

	server := mcp.NeuerServer([]mcp.Werkzeug{{
		Name: "lesen.probe", Art: mcp.Lesend, Idempotent: true,
		Ruf: func(context.Context, json.RawMessage) (string, error) { return "geliefert", nil },
	}}).MitKontingent(func(context.Context, mcp.Art, string) (bool, time.Duration) {
		return false, 4500 * time.Millisecond
	}).MitFehlerform(func(_ context.Context, fehler error) any {
		gesehen = fehler

		return map[string]string{"code": "drossel"}
	})

	zeilen := antworten(t, server,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lesen.probe","arguments":{}}}`)

	if len(zeilen) != 1 || !strings.Contains(zeilen[0], `"structuredContent":{"fehler":{"code":"drossel"}}`) ||
		!strings.Contains(zeilen[0], "in 5 s wieder") {
		t.Fatalf("über dem Guthaben: %v", zeilen)
	}

	var grenze *domain.Kontingentgrenze
	if !errors.Is(gesehen, domain.ErrKontingent) || !errors.As(gesehen, &grenze) || grenze.Warten != 4500*time.Millisecond {
		t.Errorf("die Wurzel bekam %v", gesehen)
	}
}
