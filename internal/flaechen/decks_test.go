// **Decks und Formate — an jeder Fläche** — Werkstück 7 von Phase 4.6
// (#1293),
// nach Vermerk [0061](../../docs/entscheidungen/0061-was-ein-deck-hier-ist.md).
//
// **Die Erklärung, was diese Läufe halten und was nicht, steht in
// `scryfallsatz_test.go`.**
//
// # Was hier geprüft wird
//
// **Derselbe Arbeitsablauf über jede Fläche**, und dasselbe Ergebnis: ein Deck
// anlegen, eine Karte hineinlegen, ein Format anlegen, das Deck darauf
// stellen, prüfen. Ein Wächter zählt, **dass** eine Fläche eine Fähigkeit hat;
// dass sie dasselbe tut, hält erst dieser Lauf (§16, letzter Absatz).
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Ein Deck entsteht an jeder Fläche und kommt überall gleich zurück.**
//
// **Nicht parallel**: das Fenster bekommt seine Umgebung über `t.Setenv`
// (`fensterAus`), und das verträgt Go nicht in einem parallelen Lauf.
func TestEinDeckEntstehtUeberallGleich(t *testing.T) {
	// **Und am Fenster** — Zweitblick über Phase C, Befund 3; die Antwort
	// serialisiert, wie Wails sie serialisiert.
	t.Run("Fenster", func(t *testing.T) {
		umgebung := deckumgebung(t)
		b := fensterAus(t, umgebung)

		if angelegt := b.DeckAnlegen(t.Context(), "Aggro", "magic", ""); angelegt.Fehler != nil {
			t.Fatalf("DeckAnlegen: %+v", angelegt.Fehler)
		}

		decks := b.Decks(t.Context(), "")
		if decks.Fehler != nil {
			t.Fatalf("Decks: %+v", decks.Fehler)
		}

		roh, fehler := json.Marshal(decks)
		if fehler != nil {
			t.Fatalf("Decks serialisieren: %v", fehler)
		}

		pruefeDeckliste(t, string(roh))
	})

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		mussKommandozeile(t, umgebung, "deck", "neu", "Aggro", "magic")

		pruefeDeckliste(t, mussKommandozeile(t, umgebung, "decks", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		if roh, istFehler := mcpRuf(t, umgebung, "schreiben.deck.anlegen",
			map[string]any{"name": "Aggro", "spiel": "magic"}); istFehler {
			t.Fatalf("schreiben.deck.anlegen: %s", roh)
		}

		roh, istFehler := mcpRuf(t, umgebung, "lesen.decks", map[string]any{})
		if istFehler {
			t.Fatalf("lesen.decks: %s", roh)
		}

		pruefeDeckliste(t, `{"decks":`+roh+`}`)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/decks",
			`{"name":"Aggro","spiel":"magic","notiz":""}`)
		if status != http.StatusCreated {
			t.Fatalf("POST /decks: %d (%s)", status, rumpf)
		}

		status, rumpf = an.ruf(t, http.MethodGet, "/decks", "")
		if status != http.StatusOK {
			t.Fatalf("GET /decks: %d (%s)", status, rumpf)
		}

		pruefeDeckliste(t, `{"decks":`+rumpf+`}`)
	})
}

// **Ein Format entsteht an jeder Fläche mit denselben nachgelesenen Zahlen.**
func TestEinFormatEntstehtUeberallGleich(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		mussKommandozeile(t, umgebung,
			"format", "neu", "commander", "Commander", "magic", "commander")

		pruefeFormatliste(t, mussKommandozeile(t, umgebung, "formate", "--json"))
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		if roh, istFehler := mcpRuf(t, umgebung, "schreiben.format.ausbauart",
			map[string]any{
				"bauart": "commander", "name": "Commander",
				"spiel": "magic", "scryfallname": "commander",
			}); istFehler {
			t.Fatalf("schreiben.format.ausbauart: %s", roh)
		}

		roh, istFehler := mcpRuf(t, umgebung, "lesen.formate", map[string]any{})
		if istFehler {
			t.Fatalf("lesen.formate: %s", roh)
		}

		pruefeFormatliste(t, `{"formate":`+roh+`}`)
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		status, rumpf := an.ruf(t, http.MethodPost, "/formate",
			`{"bauart":"commander","name":"Commander","spiel":"magic","scryfallname":"commander"}`)
		if status != http.StatusCreated {
			t.Fatalf("POST /formate: %d (%s)", status, rumpf)
		}

		status, rumpf = an.ruf(t, http.MethodGet, "/formate", "")
		if status != http.StatusOK {
			t.Fatalf("GET /formate: %d (%s)", status, rumpf)
		}

		pruefeFormatliste(t, `{"formate":`+rumpf+`}`)
	})
}

// **Und ein Format aus einem anderen Spiel wird überall abgewiesen.**
//
// **Derselbe fachliche Fehler an jeder Fläche** — das ist die Zusage aus §16,
// und sie ist mehr als „die Fähigkeit ist da".
func TestEinFremdesFormatWirdUeberallAbgewiesen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", "Aggro", "magic"))
		formatID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"format", "neu", "konstruiert", "Golden Age", "vs"))

		_, _, code := kommandozeile(t, umgebung, "deck", "format", deckID, formatID)
		if code != 1 {
			t.Errorf("Rückgabewert %d, erwartet 1 (fachliches Nein)", code)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		umgebung := deckumgebung(t)

		deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
			map[string]any{"name": "Aggro", "spiel": "magic"})
		formatID := kennungAusMCP(t, umgebung, "schreiben.format.ausbauart",
			map[string]any{"bauart": "konstruiert", "name": "Golden Age", "spiel": "vs"})

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.deck.format",
			map[string]any{"deckID": deckID, "formatID": formatID})
		if !istFehler {
			t.Errorf("das fremde Format kam durch: %s", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		deckID := kennungAusREST(t, an, "/decks",
			`{"name":"Aggro","spiel":"magic","notiz":""}`)
		formatID := kennungAusREST(t, an, "/formate",
			`{"bauart":"konstruiert","name":"Golden Age","spiel":"vs"}`)

		status, rumpf := an.ruf(t, http.MethodPut, "/decks/"+deckID+"/format",
			`{"formatID":`+formatID+`}`)
		if status < 400 {
			t.Errorf("das fremde Format kam durch: %d (%s)", status, rumpf)
		}
	})
}

// pruefeDeckliste sichert zu, was in einer Deckliste steht.
func pruefeDeckliste(t *testing.T, roh string) {
	t.Helper()

	var gelesen struct {
		Decks []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Spiel string `json:"spiel"`
		} `json:"decks"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(gelesen.Decks) != 1 {
		t.Fatalf("%d Decks, erwartet eines (%s)", len(gelesen.Decks), roh)
	}

	if gelesen.Decks[0].Name != "Aggro" || gelesen.Decks[0].Spiel != "magic" {
		t.Errorf("%+v", gelesen.Decks[0])
	}

	if gelesen.Decks[0].ID == "" {
		t.Error("das Deck hat keine Kennung")
	}
}

// pruefeFormatliste sichert zu, dass die Bauart ihre Zahlen mitgebracht hat.
func pruefeFormatliste(t *testing.T, roh string) {
	t.Helper()

	var gelesen struct {
		Formate []struct {
			Name              string `json:"name"`
			Mindestgroesse    int    `json:"mindestgroesse"`
			Hoechstgroesse    int    `json:"hoechstgroesse"`
			Sideboardgroesse  int    `json:"sideboardgroesse"`
			HoechstzahlGleich int    `json:"hoechstzahlGleich"`
			Scryfallname      string `json:"scryfallname"`
		} `json:"formate"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Antwort ist kein gültiges JSON: %v (%s)", err, roh)
	}

	if len(gelesen.Formate) != 1 {
		t.Fatalf("%d Formate, erwartet eines (%s)", len(gelesen.Formate), roh)
	}

	format := gelesen.Formate[0]

	// **CR 903.5a, 903.5b und 903.5e** — genau 100, jede Karte einmal, kein
	// Sideboard.
	if format.Mindestgroesse != 100 || format.Hoechstgroesse != 100 ||
		format.Sideboardgroesse != 0 || format.HoechstzahlGleich != 1 {
		t.Errorf("%+v", format)
	}

	if format.Scryfallname != "commander" {
		t.Errorf("Scryfallname %q", format.Scryfallname)
	}
}

// kennungAusMCP ruft ein Werkzeug und gibt die Kennung aus seiner Antwort.
func kennungAusMCP(
	t *testing.T, umgebung []string, werkzeug string, argumente map[string]any,
) int64 {
	t.Helper()

	roh, istFehler := mcpRuf(t, umgebung, werkzeug, argumente)
	if istFehler {
		t.Fatalf("%s: %s", werkzeug, roh)
	}

	var gelesen struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("%s: die Antwort ist kein gültiges JSON: %v (%s)", werkzeug, err, roh)
	}

	return zahl(t, gelesen.ID)
}

// kennungAusREST legt etwas an und gibt die Kennung aus der Antwort.
func kennungAusREST(t *testing.T, an dienst, weg, rumpf string) string {
	t.Helper()

	status, antwort := an.ruf(t, http.MethodPost, weg, rumpf)
	if status != http.StatusCreated {
		t.Fatalf("POST %s: %d (%s)", weg, status, antwort)
	}

	var gelesen struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(antwort), &gelesen); err != nil {
		t.Fatalf("POST %s: die Antwort ist kein gültiges JSON: %v (%s)", weg, err, antwort)
	}

	return gelesen.ID
}
