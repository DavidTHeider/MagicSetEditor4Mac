package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

type katalogprobe struct {
	eintraege []mcp.Dienstfaehigkeit
	fehler    error
}

func (k katalogprobe) Dienstkatalog(context.Context) ([]mcp.Dienstfaehigkeit, error) {
	return k.eintraege, k.fehler
}

func katalogwerkzeug(t *testing.T, naht mcp.Fremdkatalog) (mcp.Werkzeug, bool) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Katalog: naht}) {
		if werkzeug.Name == "lesen.dienstkatalog" {
			return werkzeug, true
		}
	}

	return mcp.Werkzeug{}, false
}

func TestDasKatalogwerkzeugNenntJedeFaehigkeit(t *testing.T) {
	t.Parallel()

	werkzeug, gibts := katalogwerkzeug(t, katalogprobe{eintraege: []mcp.Dienstfaehigkeit{{
		Dienst: "scryfall", Frage: "suche", Status: "bereit", HatWahl: true, Gewaehlt: true,
		Wirkung:  "externSenden",
		Vertrag:  mcp.Dienstvertrag{Basisadresse: "https://api.scryfall.com", Abgerufen: "2026-08-23"},
		Nachweis: &mcp.Dienstnachweis{Lauf: "internal/scryfall/netzlauf_test.go", Datum: "2026-09-01"},
	}, {
		Dienst: "scrydex", Frage: "suche", Status: "zugangFehlt", Grund: "team fehlt",
	}}})
	if !gibts {
		t.Fatal("das Werkzeug lesen.dienstkatalog gibt es nicht")
	}

	if werkzeug.Art != mcp.Lesend || !werkzeug.Idempotent {
		t.Errorf("das Werkzeug ist %v und idempotent=%v — es liest nur", werkzeug.Art, werkzeug.Idempotent)
	}

	text, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	for _, erwartet := range []string{
		`"status": "bereit"`, `"giltFuerDieseFassung": false`, `"team fehlt"`,
		`"nachweis": null`, "https://api.scryfall.com",
	} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, text)
		}
	}
}

func TestDasKatalogwerkzeugGibtEineLeereListeAlsListe(t *testing.T) {
	t.Parallel()

	werkzeug, _ := katalogwerkzeug(t, katalogprobe{})

	text, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if !strings.Contains(text, `"eintraege": []`) {
		t.Errorf("die leere Liste ist keine: %s", text)
	}
}

func TestDasKatalogwerkzeugReichtDenFehlerDurch(t *testing.T) {
	t.Parallel()

	werkzeug, _ := katalogwerkzeug(t, katalogprobe{fehler: errors.New("die Einstellungen sind nicht lesbar")})

	if _, err := werkzeug.Ruf(t.Context(), json.RawMessage(`{}`)); err == nil ||
		!strings.Contains(err.Error(), "nicht lesbar") {
		t.Errorf("der Fehler ist %v", err)
	}
}

func TestOhneNahtGibtEsDasKatalogwerkzeugNicht(t *testing.T) {
	t.Parallel()

	if _, gibts := katalogwerkzeug(t, nil); gibts {
		t.Error("ohne Naht steht das Werkzeug in der Liste")
	}
}
