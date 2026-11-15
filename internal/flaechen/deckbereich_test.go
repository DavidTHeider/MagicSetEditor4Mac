package flaechen_test

// **Der Bereichswechsel über Kommandozeile, MCP und Dienst** — Befund 7 des
// Zweitblicks über Phase 5.1
// (#1405).
//
// # Warum es ihn gibt
//
// Weil das Fenster den Wechsel bis zum 2026-09-10 aus **zwei** Rufen gebaut
// hat — legen, dann nehmen — und ein gescheiterter zweiter Ruf dieselbe Karte
// in beiden Bereichen liegen ließ. Der Weg ist jetzt einer, und §16 verlangt,
// dass er an jeder Fläche derselbe ist.
//
// # Was er nicht hält
//
// **Das Fenster.** Er fährt es nicht. Hier stand, die Wails-Bindung laufe in
// keinem Go-Prozess — seit C04 ruft `fensterAus` sie über `NeueBindungen`;
// welche Flächen ein Lauf anspricht, misst `FLAECHENPARITAET.md`.
// Der Kopf hieß bis zum 2026-09-10 „über alle vier Flächen" und behauptete
// damit eine, die er nicht fährt — angestrichen vom Nachlauf zum Zweitblick
// über Phase 5.1 (Befund 2), und dieselbe Klarstellung steht seit Phase W in
// `pack_test.go`. Die Bindung prüft `internal/app/deckbindungen_test.go`,
// **samt ihrer Antwort**: sie ließ sich wirkungslos machen, ohne dass ein
// Lauf rot wurde.

import (
	"net/http"
	"strings"
	"testing"
)

// **Derselbe Wechsel an drei Flächen, und derselbe fachliche Fehler.**
//
// **Der Fehler wird benannt und nicht nur gezählt** (Befund 2 des
// Nachlaufs): „irgendein Misserfolg" ließe einen Weg durchgehen, der an
// der Kennung scheitert statt am Bereich.
func TestDerBereichswechselGehtUeberAlleFlaechen(t *testing.T) {
	t.Parallel()

	t.Run("Kommandozeile", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		deckID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "neu", "Aggro", "magic"))
		platzID := strings.TrimSpace(mussKommandozeile(t, umgebung,
			"deck", "legen", deckID, druckkennung, "4", "hauptdeck"))

		if aus := mussKommandozeile(t, umgebung,
			"deck", "verschieben", deckID, platzID, "sideboard"); !strings.Contains(aus, "sideboard") {
			t.Errorf("Ausgabe %q nennt den Bereich nicht", aus)
		}

		// **Und der alte Platz ist weg** — genau ein Platz bleibt.
		if aus := mussKommandozeile(t, umgebung, "deck", deckID); strings.Count(aus, "sideboard") != 1 ||
			strings.Contains(aus, "hauptdeck") {
			t.Errorf("Deck nach dem Wechsel:\n%s", aus)
		}

		// **Ein Bereich, den es nicht gibt, wird abgewiesen** — an jeder
		// Fläche derselbe fachliche Fehler.
		_, klage, code := kommandozeile(t, umgebung,
			"deck", "verschieben", deckID, platzID, "kiste")

		switch {
		case code == 0:
			t.Error("ein unbekannter Bereich wurde angenommen")
		case !strings.Contains(klage, "Hauptdeck oder im Sideboard"):
			t.Errorf("Meldung %q nennt den Grund nicht", klage)
		}
	})

	t.Run("MCP", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)

		deckID := kennungAusMCP(t, umgebung, "schreiben.deck.anlegen",
			map[string]any{"name": "Aggro", "spiel": "magic"})
		platz := mussMCP(t, umgebung, "schreiben.deckplatz.legen", map[string]any{
			"deckID": deckID, "printing": druckkennung,
			"menge": 4, "bereich": "hauptdeck",
		})

		verschoben := mussMCP(t, umgebung, "schreiben.deckplatz.verschieben", map[string]any{
			"deckID": deckID, "platzID": platzkennung(t, platz), "bereich": "sideboard",
		})

		if !strings.Contains(verschoben, "sideboard") {
			t.Errorf("Antwort %q nennt den Bereich nicht", verschoben)
		}

		roh, istFehler := mcpRuf(t, umgebung, "schreiben.deckplatz.verschieben",
			map[string]any{
				"deckID": deckID, "platzID": platzkennung(t, platz), "bereich": "kiste",
			})

		switch {
		case !istFehler:
			t.Error("ein unbekannter Bereich wurde angenommen")
		case !strings.Contains(roh, "Hauptdeck oder im Sideboard"):
			t.Errorf("Meldung %q nennt den Grund nicht", roh)
		}
	})

	t.Run("REST", func(t *testing.T) {
		t.Parallel()

		umgebung := deckumgebung(t)
		an := starteDienst(t, umgebung)

		deckID := kennungAusREST(t, an, "/decks", `{"name":"Aggro","spiel":"magic"}`)
		platz := mussREST(t, an, http.MethodPut, "/decks/"+deckID+"/plaetze",
			`{"printing":"`+druckkennung+`","menge":4,"bereich":"hauptdeck"}`)

		weg := "/decks/" + deckID + "/plaetze/" + platzkennungText(t, platz) + "/bereich"

		if antwort := mussREST(t, an, http.MethodPut, weg,
			`{"bereich":"sideboard"}`); !strings.Contains(antwort, "sideboard") {
			t.Errorf("Antwort %q nennt den Bereich nicht", antwort)
		}

		code, antwort := an.ruf(t, http.MethodPut, weg, `{"bereich":"kiste"}`)

		switch {
		case code == http.StatusOK:
			t.Error("ein unbekannter Bereich wurde angenommen")
		case !strings.Contains(antwort, "Hauptdeck oder im Sideboard"):
			t.Errorf("Antwort %q nennt den Grund nicht", antwort)
		}
	})
}
