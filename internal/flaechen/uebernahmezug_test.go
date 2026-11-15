// **Eine gescheiterte Übernahme hinterlässt kein Deck — an jeder Fläche** —
// #1820,
// Befund 6 des Zweitblicks über Phase 6.1.
//
// # Die Abnahme
//
// **„Ein Lauf, der beim dritten Platz scheitern lässt: danach gibt es kein
// neues Deck — an allen vier Flächen."** Das Fenster prüft `internal/app`;
// hier stehen Kommandozeile, MCP und Dienst, über **einen** Bestand.
//
// # Wie der dritte Platz scheitert
//
// **Ein Auslöser, den nur dieser Lauf setzt**: er bricht das Einfügen eines
// Platzes ab, sobald das Deck schon zwei trägt. Das fremde Deck bekommt seine
// drei vorher; getroffen wird nur das neue.
//
// # Die Gegenprobe
//
// **Ohne Auslöser glückt die Übernahme** — das halten die Läufe in
// `herkunft_test.go` an denselben drei Flächen.
package flaechen_test

import (
	"net/http"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// fremdesDeckMitDreiPlaetzen legt im Profil eines zweiten Nutzers ein Deck
// mit drei Plätzen an, gibt mir die Freigabe und stellt mich zurück.
//
// **Über die Kommandozeile, für jede Fläche**: wer das Deck drüben anlegt,
// ist nicht die Frage — die Frage ist, wer es holt.
func fremdesDeckMitDreiPlaetzen(t *testing.T, umgebung []string) (fremdes, deck string) {
	t.Helper()

	eigenes := mussKommandozeile(t, umgebung, "profil", "--json")
	ich := eignerkennung(t, eigenes)
	meines := profilkennung(t, eigenes)
	fremder, fremdes := fremdesProfilAnDerZeile(t, umgebung)

	mussKommandozeile(t, umgebung, "nutzer-wechseln", fremder)

	deck = deckkennung(t,
		mussKommandozeile(t, umgebung, "deck", "neu", "Nachtlauf", "probe", "--json"))

	for _, druck := range []string{
		"a1b2c3d4-0000-4000-8000-00000000000a",
		"a1b2c3d4-0000-4000-8000-00000000000b",
		"a1b2c3d4-0000-4000-8000-00000000000c",
	} {
		mussKommandozeile(t, umgebung, "deck", "legen", deck, druck, "1")
	}

	mussKommandozeile(t, umgebung, "freigabe-setzen", fremdes, ich, "mitarbeiter")
	mussKommandozeile(t, umgebung, "nutzer-wechseln", ich)
	mussKommandozeile(t, umgebung, "profil", meines)

	return fremdes, deck
}

// derDritteScheitert setzt den Auslöser in den Bestand dieses Laufs.
func derDritteScheitert(t *testing.T, umgebung []string) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	if _, fehler := db.ExecContext(t.Context(),
		`CREATE TRIGGER dritter BEFORE INSERT ON deckplaetze
		 WHEN (SELECT COUNT(*) FROM deckplaetze WHERE deck_id = NEW.deck_id) >= 2
		 BEGIN SELECT RAISE(ABORT, 'der dritte Platz scheitert'); END`); fehler != nil {
		t.Fatalf("den Auslöser setzen: %v", fehler)
	}
}

func TestEineGescheiterteUebernahmeHinterlaesstKeinDeck(t *testing.T) {
	t.Parallel()

	for _, flaeche := range []struct {
		name string

		// hole übernimmt das fremde Deck und sagt, ob es **scheiterte**.
		hole func(t *testing.T, umgebung []string, fremdes, deck string) bool
	}{
		{"an der Kommandozeile", func(t *testing.T, umgebung []string, fremdes, deck string) bool {
			_, _, code := kommandozeile(t, umgebung, "deck-uebernehmen", fremdes, deck, "--json")

			return code != 0
		}},
		{"über MCP", func(t *testing.T, umgebung []string, fremdes, deck string) bool {
			_, istFehler := mcpRuf(t, umgebung, "schreiben.deck.uebernehmen",
				map[string]any{"quellProfil": fremdes, "quellDeck": deck})

			return istFehler
		}},
		{"über den Dienst", func(t *testing.T, umgebung []string, fremdes, deck string) bool {
			status, _ := starteDienst(t, umgebung).ruf(t, http.MethodPost, "/decks/uebernehmen",
				`{"quellProfil":"`+fremdes+`","quellDeck":"`+deck+`"}`)

			return status >= http.StatusBadRequest
		}},
	} {
		t.Run(flaeche.name, func(t *testing.T) {
			t.Parallel()

			umgebung := eigenerBestand(t)
			fremdes, deck := fremdesDeckMitDreiPlaetzen(t, umgebung)
			vorher := deckeDerLage(t, umgebung)

			derDritteScheitert(t, umgebung)

			if !flaeche.hole(t, umgebung, fremdes, deck) {
				t.Fatal("die Übernahme scheiterte nicht — der Auslöser griff nicht")
			}

			if nachher := deckeDerLage(t, umgebung); nachher != vorher {
				t.Errorf("%d Decks im eigenen Profil, vorher %d — ein halbes Deck blieb stehen",
					nachher, vorher)
			}
		})
	}
}
