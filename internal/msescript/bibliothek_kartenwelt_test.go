package msescript_test

// **Die drei Namen, die Set und Editor kennen** — Werkstück A von Phase 4.9
// (#1324).
//
// Was hier geprüft wird, ist der **Stummel**: was die Sprache ohne Set und
// Editor tut. Dass `internal/service` mehr weiß und den Namen überschreibt,
// ist derselbe Weg, den `expand_keywords` seit Phase 4.11 geht.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

func TestDieDreiAusSetUndEditor(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, erwartet string }{
		// **Ohne Statistik ist nichts gebraucht worden.** Eine Vorlage, die
		// daraus eine Liste baut, bekommt eine leere — und keinen Fehler über
		// einen Namen, den sie zu Recht ruft.
		{"keyword_usage ohne Statistik", `keyword_usage("egal")`, ""},
		{"keyword_usage mit unique", `keyword_usage("egal", unique: true)`, ""},

		// **Ohne Gruppen ist eine Wahl ihre eigene Gruppe.**
		{"primary_choice gibt die Wahl", `primary_choice("rare")`, "rare"},
		{"primary_choice einer Zahl", `primary_choice(3)`, "3"},

		// **`new_card` beschreibt eine Karte, es legt keine an.**
		{
			"new_card gibt die Felder zurück",
			`new_card(input: ["name": "Blitz"])["name"]`, "Blitz",
		},
		{
			"new_card stellungsweise",
			`new_card(["name": "Blitz"])["name"]`, "Blitz",
		},
		{
			"new_card mit game: daneben",
			`new_card(game: "magic", input: ["name": "Blitz"])["name"]`, "Blitz",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err != nil {
				t.Fatalf("auswerten: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("%q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Und die Fehlerwege** (§3).
func TestDieDreiMeldenWasFehlt(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct{ name, quelle, nennt string }{
		{"keyword_usage ohne Karte", `keyword_usage()`, "keyword_usage"},
		{"primary_choice ohne Eingabe", `primary_choice()`, "primary_choice"},
		{"new_card ohne Felder", `new_card(game: "magic")`, "new_card"},
		{
			"new_card mit einer Zahl statt Feldern",
			`new_card(input: 3)`, "Abbildung",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			_, err := msescript.AuswertenQuelle(t.Context(), fall.quelle,
				msescript.Bibliothek())
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			if !strings.Contains(err.Error(), fall.nennt) {
				t.Errorf("die Meldung %q nennt %q nicht", err, fall.nennt)
			}
		})
	}
}
