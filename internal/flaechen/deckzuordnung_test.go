// **Der ganze Zuordnungsweg — an jeder Fläche** — Werkstück U9 von Phase 5.1
// (#1420).
//
// # Der Ablauf
//
// Exemplar aufnehmen → Deck anlegen → Platz legen → zuordnen → nachsehen
// (zugeordnet, frei) → Rückweg vom Exemplar → Fehlkarten → lösen.
//
// **Und derselbe fachliche Fehler an jeder Fläche**: eine zweite Zuordnung
// über den Besitz hinaus ist ein Konflikt, kein Erfolg.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// zuordnungsstand ist die Antwort von `lesen.deck.zuordnungen`.
type zuordnungsstand struct {
	Zeilen []struct {
		Gebraucht   int `json:"gebraucht"`
		Vorhanden   int `json:"vorhanden"`
		Frei        int `json:"frei"`
		Zugeordnet  int `json:"zugeordnet"`
		Zuordnungen []struct {
			ID         string `json:"id"`
			Menge      int    `json:"menge"`
			Abweichend bool   `json:"abweichend"`
		} `json:"zuordnungen"`
	} `json:"zeilen"`
	Fehlend int `json:"fehlend"`
}

// pruefeZuordnung liest den Stand und hält ihn gegen die vier Zahlen.
//
//nolint:unparam // `fehlend` bleibt hier 2, weil das Deck vier braucht und zwei da sind — die Zahl gehört trotzdem an den Ruf, sonst prüfte er sie nicht sichtbar
func pruefeZuordnung(t *testing.T, roh string, frei, zugeordnet, fehlend int) string {
	t.Helper()

	var stand zuordnungsstand
	if err := json.Unmarshal([]byte(roh), &stand); err != nil {
		t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
	}

	if len(stand.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet eine: %s", len(stand.Zeilen), roh)
	}

	zeile := stand.Zeilen[0]
	if zeile.Frei != frei || zeile.Zugeordnet != zugeordnet || stand.Fehlend != fehlend {
		t.Errorf("frei %d, zugeordnet %d, fehlend %d — erwartet %d, %d, %d",
			zeile.Frei, zeile.Zugeordnet, stand.Fehlend, frei, zugeordnet, fehlend)
	}

	if zugeordnet == 0 {
		return ""
	}

	if len(zeile.Zuordnungen) != 1 {
		t.Fatalf("%d Zuordnungen: %s", len(zeile.Zuordnungen), roh)
	}

	return zeile.Zuordnungen[0].ID
}

// zuordnungsausgabe holt den Stand als JSON — der Rückgabewert 1 heißt „es
// fehlt" und ist kein Fehler.
func zuordnungsausgabe(t *testing.T, umgebung []string, deckID string) string {
	t.Helper()

	aus, klage, code := kommandozeile(t, umgebung, "deck", "zuordnungen", deckID, "--json")
	if code > 1 {
		t.Fatalf("zuordnungen gab %d: %s", code, klage)
	}

	return aus
}

// exemplarAnlegen nimmt ein Exemplar auf und gibt seine Kennung.
func exemplarAnlegen(t *testing.T, umgebung []string, menge int) string {
	t.Helper()

	var angelegt struct {
		ID int64 `json:"id"`
	}

	roh := mussKommandozeile(t, umgebung, "exemplar", druckkennung,
		strconv.Itoa(menge), "nearMint", "en", "nonfoil",
		"", "", "", "", "", "", "", "", "besitz", "--json")

	if err := json.Unmarshal([]byte(roh), &angelegt); err != nil {
		t.Fatalf("die Antwort lesen: %v (%s)", err, roh)
	}

	return strconv.FormatInt(angelegt.ID, 10)
}

func TestDerGanzeZuordnungswegAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	exemplarID := exemplarAnlegen(t, umgebung, 2)

	deckID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "neu", "Aggro", "magic"))
	platzID := strings.TrimSpace(mussKommandozeile(t, umgebung, "deck", "legen", deckID, druckkennung, "4"))

	mussKommandozeile(t, umgebung, "deck", "zuordnen", deckID, platzID, exemplarID, "2")

	// **Der Rückgabewert sagt „es fehlt"** — wie beim Abgleich, und deshalb
	// nicht über `mussKommandozeile`: ein fachliches Nein ist kein Fehler.
	zuordnungID := pruefeZuordnung(t, zuordnungsausgabe(t, umgebung, deckID), 0, 2, 2)

	// **Der Rückweg nennt das Deck.**
	if aus := mussKommandozeile(t, umgebung, "exemplar", "decks", exemplarID); !strings.Contains(aus, "Aggro") {
		t.Errorf("der Rückweg lautet %q", aus)
	}

	// **Zwei fehlen** — die Liste sagt es, und der Rückgabewert auch.
	aus, _, code := kommandozeile(t, umgebung, "deck", "fehlkarten", deckID)
	if code != 1 || !strings.HasPrefix(strings.TrimSpace(aus), "2 ") {
		t.Errorf("fehlkarten gab %d: %q", code, aus)
	}

	// **Mehr als der Besitz geht nicht.**
	if _, _, code := kommandozeile(t, umgebung, "deck", "zuordnen", deckID, platzID, exemplarID, "3"); code == 0 {
		t.Error("die zu große Zuordnung kam durch")
	}

	mussKommandozeile(t, umgebung, "deck", "zuordnung-loesen", deckID, zuordnungID)
	pruefeZuordnung(t, zuordnungsausgabe(t, umgebung, deckID), 2, 0, 2)
}

func TestDerGanzeZuordnungswegUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	exemplarID := zahl(t, exemplarAnlegen(t, umgebung, 2))

	deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
		map[string]any{"name": "Aggro", "spiel": "magic"})

	platz := mussMCP(t, umgebung, "schreiben.deckplatz.legen", map[string]any{
		"deckID": deckID, "printing": druckkennung, "menge": 4, "bereich": "hauptdeck",
	})

	mussMCP(t, umgebung, "schreiben.deck.zuordnen", map[string]any{
		"deckID": deckID, "platzID": platzkennung(t, platz),
		"exemplarID": exemplarID, "menge": 2,
	})

	zuordnungID := pruefeZuordnung(t,
		mussMCP(t, umgebung, "lesen.deck.zuordnungen", map[string]any{"deckID": deckID}), 0, 2, 2)

	mussMCP(t, umgebung, "lesen.exemplar.decks", map[string]any{"exemplarID": exemplarID})
	mussMCP(t, umgebung, "lesen.deck.fehlkarten", map[string]any{"deckID": deckID})

	// **Mehr als der Besitz ist ein Fehler und kein leeres Ergebnis.**
	if roh, istFehler := mcpRuf(t, umgebung, "schreiben.deck.zuordnen", map[string]any{
		"deckID": deckID, "platzID": platzkennung(t, platz),
		"exemplarID": exemplarID, "menge": 3,
	}); !istFehler {
		t.Errorf("die zu große Zuordnung kam durch: %s", roh)
	}

	mussMCP(t, umgebung, "schreiben.deck.zuordnung.loesen", map[string]any{
		"deckID": deckID, "zuordnungID": zahl(t, zuordnungID),
	})
}

func TestDerGanzeZuordnungswegUeberREST(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)
	exemplarID := exemplarAnlegen(t, umgebung, 2)
	an := starteDienst(t, umgebung)

	deckID := kennungAusREST(t, an, "/decks", `{"name":"Aggro","spiel":"magic"}`)

	platz := mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/plaetze",
		`{"printing":"`+druckkennung+`","menge":4,"bereich":"hauptdeck"}`)

	mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/zuordnungen",
		`{"platzID":`+platzkennungText(t, platz)+`,"exemplarID":`+exemplarID+`,"menge":2}`)

	zuordnungID := pruefeZuordnung(t,
		mussREST(t, an, http.MethodGet, "/decks/"+deckID+"/zuordnungen", ""), 0, 2, 2)

	mussREST(t, an, http.MethodGet, "/sammlung/exemplare/"+exemplarID+"/decks", "")
	mussREST(t, an, http.MethodGet, "/decks/"+deckID+"/fehlkarten", "")

	// **Mehr als der Besitz gibt einen Konflikt.**
	if status, antwort := an.ruf(t, http.MethodPut, "/decks/"+deckID+"/zuordnungen",
		`{"platzID":`+platzkennungText(t, platz)+`,"exemplarID":`+exemplarID+`,"menge":3}`); status == http.StatusOK {
		t.Errorf("die zu große Zuordnung kam durch: %s", antwort)
	}

	mussREST(t, an, http.MethodDelete, "/decks/"+deckID+"/zuordnungen/"+zuordnungID, "")
}
