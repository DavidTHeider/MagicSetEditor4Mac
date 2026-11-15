// **Das Suchwerkzeug** — Werkstück 1 von Phase 3.6
// (#1167).
package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// suchprobe gibt feste Treffer und merkt sich die Frage.
type suchprobe struct {
	gefragt  string
	ergebnis mcp.Suchergebnis
	fehler   error
}

func (s *suchprobe) Kartensuche(_ context.Context, frage string) (
	mcp.Suchergebnis, error,
) {
	s.gefragt = frage

	return s.ergebnis, s.fehler
}

// rufeSuche sucht das Werkzeug und ruft es.
func rufeSuche(t *testing.T, suche mcp.Fremdkartensuche, argumente string) (
	string, error,
) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
		Suche:   suche,
	}) {
		if werkzeug.Name == "lesen.scryfall.suche" {
			return werkzeug.Ruf(t.Context(), json.RawMessage(argumente))
		}
	}

	t.Fatal("das Werkzeug lesen.scryfall.suche gibt es nicht")

	return "", nil
}

// **Die Frage geht unverändert hinaus, und die Treffer kommen als JSON.**
func TestDasSuchwerkzeugGibtSeineTreffer(t *testing.T) {
	t.Parallel()

	probe := &suchprobe{ergebnis: mcp.Suchergebnis{
		Treffer: []mcp.Suchtrefferang{{Kennung: "abc", Name: "Lightning Bolt"}},
		Weitere: true,
	}}

	text, err := rufeSuche(t, probe, `{"frage":"t:creature c:r"}`)
	if err != nil {
		t.Fatalf("Ruf: %v", err)
	}

	if probe.gefragt != "t:creature c:r" {
		t.Errorf("gefragt wurde %q", probe.gefragt)
	}

	for _, erwartet := range []string{"abc", "Lightning Bolt", `"weitere": true`} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Antwort nennt %q nicht: %s", erwartet, text)
		}
	}
}

// **Eine leere Frage wird abgewiesen.**
func TestDasSuchwerkzeugWeistEineLeereFrageAb(t *testing.T) {
	t.Parallel()

	probe := &suchprobe{}

	_, err := rufeSuche(t, probe, `{"frage":""}`)

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v — will ErrValidation", err)
	}

	if probe.gefragt != "" {
		t.Errorf("es ging etwas hinaus: %q", probe.gefragt)
	}
}

// **Ein kaputtes Argument ebenso.**
func TestDasSuchwerkzeugWeistEinKaputtesArgumentAb(t *testing.T) {
	t.Parallel()

	if _, err := rufeSuche(t, &suchprobe{}, `{"frage":`); err == nil {
		t.Error("das kaputte Argument kam durch")
	}
}

// **Und ein Fehler der Quelle kommt durch.**
func TestEinFehlerDerSuchquelleErreichtDasWerkzeug(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("Scryfall antwortet nicht")

	if _, err := rufeSuche(t, &suchprobe{fehler: gestolpert},
		`{"frage":"bolt"}`); !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Ohne Naht gibt es das Werkzeug gar nicht.**
//
// **Und das ist die Zusage**: eine Fläche, die den Zugang nicht einrichtet,
// bietet ihn nicht an — statt ihn anzubieten und beim Ruf abzustürzen.
func TestOhneSuchnahtGibtEsDasWerkzeugNicht(t *testing.T) {
	t.Parallel()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser: probenbestand(), Schreiber: probenbestand(),
		Ausfuhr: ruhendeAusfuhr{}, Auskunft: stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Angaben:     neueAngabenprobe(), Stile: neueStilprobe(),
		Bestand: neueBestandsprobeWeit(),
	}) {
		if werkzeug.Name == "lesen.scryfall.suche" {
			t.Fatal("das Werkzeug steht da, obwohl es keine Naht gibt")
		}
	}
}
