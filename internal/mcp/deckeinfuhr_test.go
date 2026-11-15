package mcp_test

// **Die zwei Decklistenwerkzeuge** — Werkstück `T8` von Phase 5.6
// (#1473).
//
// # Was hier geprüft wird
//
//   - Beide kommen durch, mit ihren Argumenten.
//   - **Es gibt sie nicht ohne Naht**: `nil` heißt kein Werkzeug.
//   - Ein Fehler der Naht kommt an, ein kaputtes Argument endet davor.
//   - **Das schreibende sagt keine Idempotenz zu**, und das ist ehrlich: der
//     Stand aus der Vorschau gilt nach dem ersten Ruf nicht mehr.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// klagendeDeckeinfuhr ist eine Naht, die immer nein sagt.
type klagendeDeckeinfuhr struct{ fehler error }

func (k klagendeDeckeinfuhr) Vorschau(_ context.Context, _ int64, _ string) (
	mcp.Deckvorschau, error,
) {
	return mcp.Deckvorschau{}, k.fehler
}

func (k klagendeDeckeinfuhr) Uebernehmen(
	_ context.Context, _ int64, _ string, _ []mcp.Deckeinfuhrzeile,
) (mcp.Deckeinfuhrbericht, error) {
	return mcp.Deckeinfuhrbericht{}, k.fehler
}

// rufeDeckliste sucht ein Werkzeug und ruft es.
func rufeDeckliste(
	t *testing.T, naht mcp.Deckeinfuhr, name, argumente string,
) (string, error) {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{Deckeinfuhr: naht}) {
		if werkzeug.Name == name {
			return werkzeug.Ruf(t.Context(), []byte(argumente))
		}
	}

	t.Fatalf("das Werkzeug %q steht nicht in der Liste", name)

	return "", nil
}

func TestDieDecklistenwerkzeugeKommenDurch(t *testing.T) {
	naht := deckeinfuhrprobe{}

	aus, err := rufeDeckliste(t, naht, "lesen.deckliste",
		`{"deckID":7,"text":"4 Flammenball"}`)
	if err != nil {
		t.Fatalf("deuten: %v", err)
	}

	for _, erwartet := range []string{
		`"deckID": 7`, `"stand": "abdruck"`, `"uebernehmbar": true`,
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}

	aus, err = rufeDeckliste(t, naht, "schreiben.deckliste",
		`{"deckID":7,"stand":"abdruck","zeilen":[`+
			`{"printing":"a1b2c3d4-0000-0000-0000-00000000000a",`+
			`"menge":4,"bereich":"hauptdeck"}]}`)
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	// **Der neue Stand geht mit** — wer eine zweite Liste einfügt, braucht
	// ihn, und ein zweiter Ruf dafür wäre einer zu viel.
	for _, erwartet := range []string{
		`"gelegt": 1`, `"stueck": 4`, `"stand": "abdruck-danach"`,
	} {
		if !strings.Contains(aus, erwartet) {
			t.Errorf("%s fehlt in %s", erwartet, aus)
		}
	}
}

// **Ein Fehler der Naht kommt an, ein kaputtes Argument endet davor.**
func TestDieDecklistenwerkzeugeReichenIhreFehlerDurch(t *testing.T) {
	naht := klagendeDeckeinfuhr{fehler: errors.New("das Deck gibt es nicht")}

	for name, argumente := range map[string]string{
		"lesen.deckliste":     `{"deckID":7,"text":"4 Flammenball"}`,
		"schreiben.deckliste": `{"deckID":7,"stand":"a","zeilen":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := rufeDeckliste(t, naht, name, argumente); err == nil {
				t.Error("der Fehler wurde verschluckt")
			}

			if _, err := rufeDeckliste(t, naht, name, `kein JSON`); err == nil {
				t.Error("das kaputte Argument kam durch")
			}
		})
	}
}

// **Ohne Naht gibt es die zwei Werkzeuge nicht.**
func TestOhneDeckeinfuhrnahtGibtEsDieWerkzeugeNicht(t *testing.T) {
	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{}) {
		if strings.Contains(werkzeug.Name, "deckliste") {
			t.Errorf("%s steht da, obwohl keine Naht eingerichtet ist",
				werkzeug.Name)
		}
	}
}

// **Das lesende ist lesend, das schreibende schreibend — und keines von
// beiden sagt Idempotenz zu.**
//
// **Beim lesenden ist das eine Entscheidung und keine Auslassung**: es fragt
// Scryfall, und eine Antwort aus dem Zwischenspeicher ist eine andere als
// eine frische. Beim schreibenden ist es die Wirkung selbst — der Stand aus
// der Vorschau gilt nach dem ersten Ruf nicht mehr.
func TestDieDecklistenwerkzeugeTragenIhreArt(t *testing.T) {
	erwartet := map[string]mcp.Art{
		"lesen.deckliste":     mcp.Lesend,
		"schreiben.deckliste": mcp.Schreibend,
	}

	gesehen := map[string]mcp.Art{}

	for _, werkzeug := range mcp.Werkzeuge(
		mcp.Naehte{Deckeinfuhr: deckeinfuhrprobe{}}) {
		if !strings.Contains(werkzeug.Name, "deckliste") {
			continue
		}

		gesehen[werkzeug.Name] = werkzeug.Art

		if werkzeug.Idempotent {
			t.Errorf("%s sagt Idempotenz zu", werkzeug.Name)
		}
	}

	for name, art := range erwartet {
		if gesehen[name] != art {
			t.Errorf("%s ist als %v geführt, erwartet %v",
				name, gesehen[name], art)
		}
	}
}
