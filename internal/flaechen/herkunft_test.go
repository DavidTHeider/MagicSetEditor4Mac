// **Woher ein Deck kommt — an jeder Fläche** — Werkstück 12 von Phase 4.12
// (#1351).
//
// # Die Abnahme
//
// **Ein Deck, das woanders entstanden ist, gehört danach mir — und sagt, woher
// es kommt.** So steht es in der Kommentierung. Der Ablauf ist an jeder Fläche
// derselbe (§16, letzter Absatz):
//
//  1. ein zweiter Nutzer, in dessen Profil ein Deck liegt,
//  2. eine Freigabe als `mitarbeiter` — ohne sie wird abgewiesen,
//  3. das Deck holen,
//  4. es liegt im **eigenen** Profil und nennt seine Herkunft.
//
// # Warum Schritt 2 nicht weggelassen wird
//
// **Weil ein Lauf, der nur den Erfolgsweg fährt, auch dann grün wäre, wenn
// die Erlaubnis gar nicht geprüft würde** (§10). Die Ablehnung steht vor dem
// Holen, und sie ist die Hälfte, die man auslässt.
package flaechen_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// **Der ganze Weg über die Kommandozeile.**
func TestEineDeckherkunftAnDerKommandozeile(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilAnDerZeile(t, umgebung)

	// **Das Deck entsteht im fremden Profil.**
	// **Als der fremde Mensch, nicht nur in seinem Profil** — Werkstück 17
	// (#1400):
	// eine Freigabe **auf** ein Profil setzt nur, wer es verwaltet.
	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	deck := deckkennung(t,
		mussKommandozeile(t, umgebung, "deck", "neu", "Nachtlauf", "probe", "--json"))

	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "mitarbeiter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)

	// **Und jetzt hole ich es.**
	geholt := mussKommandozeile(t, umgebung,
		"deck-uebernehmen", fremdes, deck, "--json")

	meins := deckkennung(t, geholt)
	if meins == deck {
		t.Errorf("das geholte Deck hat dieselbe Kennung: %s", meins)
	}

	// **Es liegt im eigenen Profil und nennt, woher es kam.**
	aus := mussKommandozeile(t, umgebung, "deck", meins, "--json")

	// **Der Nutzer steht vor dem Profil** — „Mein Profil" allein sagt einem
	// Fremden nichts.
	if !strings.Contains(aus, `"art": "profil"`) ||
		!strings.Contains(aus, `"quelle": "Zweiter · Mein Profil"`) {
		t.Errorf("die Herkunft fehlt: %s", aus)
	}
}

// **Und ohne Freigabe wird abgewiesen.**
//
// **Der Lauf, ohne den der obige nichts belegt** (§10): er zeigt, dass die
// Erlaubnis wirklich geprüft wird und nicht bloß danebensteht.
func TestOhneFreigabeWirdKeinDeckGeholt(t *testing.T) {
	t.Parallel()

	// **Mit eigenem Schlüsselbund**, weil hier ein zweiter Faktor eingerichtet
	// wird — siehe `zweiterfaktor_test.go`.
	umgebung := bestandMitBund(t)

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	meines := profilkennung(t, eigenes)

	angelegt := mussKommandozeile(t, umgebung, "nutzer-neu", "Zweiter", "--json")
	fremder, fremdes := profilEinesFremden(t, angelegt)

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	deck := deckkennung(t,
		mussKommandozeile(t, umgebung, "deck", "neu", "Nachtlauf", "probe", "--json"))

	mussKommandozeile(t, umgebung, "profil", meines)

	// **Der Eigner des fremden Profils bekommt einen zweiten Faktor**, sonst
	// sperrt keine Freigabe: ein Profil ohne Schloss wird nicht bewacht
	// ([0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md)).
	// Genau dieselbe Zusage wie beim Wechsel — und der Grund, aus dem die
	// Übernahme sie über `Zutritt` fragt statt eigens.
	eingerichtet := einrichtungAus(t,
		mussKommandozeile(t, umgebung, "faktor-neu", fremder, "totp", "--json"))
	mussKommandozeile(t, umgebung, "faktor-bestaetigen", fremder,
		bestaetigungskennwort(t, eingerichtet.Geheimnis))

	_, fehler, code := kommandozeile(t, umgebung, "deck-uebernehmen", fremdes, deck)
	if code == 0 {
		t.Fatal("das Deck kam ohne Freigabe durch")
	}

	if !strings.Contains(fehler, "zweiten Faktor") {
		t.Errorf("die Ablehnung lautet %q", fehler)
	}
}

// **Derselbe Weg über MCP.**
func TestEineDeckherkunftUeberMCP(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)

	eigenes := mussMCP(t, umgebung, "lesen.profil", map[string]any{})
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilUeberMCP(t, umgebung)

	// **Als der fremde Mensch, nicht nur in seinem Profil** — Werkstück 17
	// (#1400):
	// eine Freigabe **auf** ein Profil setzt nur, wer es verwaltet.
	mussMCP(t, umgebung, "schreiben.nutzer.wechseln", map[string]any{"nutzerID": fremder})

	deck := deckkennung(t, mussMCP(t, umgebung, "schreiben.deck.anlegen",
		map[string]any{"name": "Nachtlauf", "spiel": "probe"}))

	mussMCP(t, umgebung, "schreiben.freigabe", map[string]any{
		"profilID": fremdes, "nutzerID": ich, "rolle": "mitarbeiter",
	})
	mussMCP(t, umgebung, "schreiben.nutzer.wechseln", map[string]any{"nutzerID": ich})
	mussMCP(t, umgebung, "schreiben.profil.wechseln", map[string]any{"profilID": meines})

	geholt := mussMCP(t, umgebung, "schreiben.deck.uebernehmen",
		map[string]any{"quellProfil": fremdes, "quellDeck": deck})

	if !strings.Contains(geholt, "Nachtlauf") {
		t.Fatalf("das Deck kam nicht an: %s", geholt)
	}

	// **Und `lesen.decks` im eigenen Profil führt es.**
	if aus := mussMCP(t, umgebung, "lesen.decks",
		map[string]any{}); !strings.Contains(aus, "Nachtlauf") {
		t.Errorf("das geholte Deck fehlt in der Liste: %s", aus)
	}
}

// **Und derselbe Weg über den Dienst.**
func TestEineDeckherkunftUeberDenDienst(t *testing.T) {
	t.Parallel()

	umgebung := eigenerBestand(t)
	an := starteDienst(t, umgebung)

	eigenes := mussREST(t, an, http.MethodGet, "/profil", "")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilUeberDenDienst(t, an)

	// **Als der fremde Mensch, nicht nur in seinem Profil** — Werkstück 17
	// (#1400):
	// eine Freigabe **auf** ein Profil setzt nur, wer es verwaltet.
	mussREST(t, an, http.MethodPut, "/nutzer", `{"nutzerID":"`+fremder+`"}`)

	deck := deckkennung(t, mussREST(t, an, http.MethodPost, "/decks",
		`{"name":"Nachtlauf","spiel":"probe"}`))

	mussREST(t, an, http.MethodPut,
		"/profile/"+fremdes+"/freigaben/"+ich, `{"rolle":"mitarbeiter"}`)
	mussREST(t, an, http.MethodPut, "/nutzer", `{"nutzerID":"`+ich+`"}`)
	mussREST(t, an, http.MethodPut, "/profil", `{"profilID":"`+meines+`"}`)

	geholt := mussREST(t, an, http.MethodPost, "/decks/uebernehmen",
		`{"quellProfil":"`+fremdes+`","quellDeck":"`+deck+`"}`)

	if !strings.Contains(geholt, "Nachtlauf") {
		t.Fatalf("das Deck kam nicht an: %s", geholt)
	}

	// **Und es nennt seine Herkunft, wenn man es einzeln liest.**
	meins := uebernommeneKennung(t, geholt)

	if aus := mussREST(t, an, http.MethodGet, "/decks/"+meins,
		""); !strings.Contains(aus, `"quelle":"Zweiter · Mein Profil"`) {
		t.Errorf("die Herkunft fehlt: %s", aus)
	}
}

// deckkennung liest die Kennung aus einer Deckantwort.
//
// **Zwei Gestalten, und beide sind richtig.** Die Kommandozeile umschließt das
// Deck (`{"deck": …}`), weil ihre Antwort auch einen Fehler tragen kann; MCP
// und der Dienst geben es blank, weil dort der Fehler ein anderer Kanal ist —
// eine JSON-RPC-Antwort und ein HTTP-Status. Das ist keine Abweichung, die
// dieser Lauf anstreichen soll: er prüft, ob dieselbe **Fähigkeit** überall
// dasselbe **tut** (§16), nicht ob die Hüllen gleich aussehen.
func deckkennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		Deck struct {
			ID string `json:"id"`
		} `json:"deck"`

		ID string `json:"id"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Deckantwort lesen: %v — %s", err, roh)
	}

	if gelesen.Deck.ID != "" {
		return gelesen.Deck.ID
	}

	if gelesen.ID == "" {
		t.Fatalf("die Antwort nennt kein Deck: %s", roh)
	}

	return gelesen.ID
}

// uebernommeneKennung liest sie aus der Antwort des Dienstes auf eine
// Übernahme — sie trägt `deckID` und kein ganzes Deck.
func uebernommeneKennung(t *testing.T, roh string) string {
	t.Helper()

	var gelesen struct {
		DeckID string `json:"deckID"`
	}

	if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
		t.Fatalf("die Übernahmeantwort lesen: %v — %s", err, roh)
	}

	if gelesen.DeckID == "" {
		t.Fatalf("die Antwort nennt kein Deck: %s", roh)
	}

	return gelesen.DeckID
}
