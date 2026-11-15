package mtgjson_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// **Warum es diese Prüfung überhaupt gibt.**
//
// Der Rundlauf allein sagt nichts über Gültigkeit. Die Gegenprobe zeigt es: nimmt
// man die Rückwandlungen aus der Zuordnung heraus — schreibt also `2R` statt
// `{2}{R}` und `mythic rare` statt `mythic` —, bleibt er **grün**. Er muss es:
// beide Richtungen benutzen dieselbe Tabelle, und was die eine verbiegt, biegt
// die andere zurück. Ein Rundlauf prüft Selbstkonsistenz, nicht Wahrheit.
//
// Deshalb verlangt #142 ausdrücklich eine Prüfung „gegen das Schema, nicht gegen
// unsere Vorstellung davon".
//
// **Ein maschinenlesbares Schema gibt es nicht.** Nachgesehen am 2026-08-15:
// weder mtgjson.com noch `mtgjson/mtgjson` auf GitHub veröffentlichen eines; die
// Datenmodelle stehen als Dokumentation da. Was hier steht, ist der nächste
// Ersatz und er ist prüfbar: eine Tabelle bekannter Schlüssel samt ihrer Art,
// **und derselbe Prüfer läuft hinter dem Build-Tag `fremdpakete` gegen eine
// bezogene Datei**. Wäre die Tabelle erfunden, fiele die echte Datei durch.
func TestDieAusgabeIstGueltigesMTGJSON(t *testing.T) {
	quelle, err := mtgjson.Lies(pfad("probe.json"), mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	inhalt, _ := mtgjson.ZuordnungMagic.Schreibe(alsDomaene(quelle)[0])

	if befunde := mtgjson.Pruefe(inhalt, "ausgabe.json"); len(befunde) > 0 {
		t.Errorf("die eigene Ausgabe ist nicht gültig:\n%s\n%s", befunde, inhalt)
	}
}

// Der Prüfer weist ab, was er abweisen soll.
//
// **Die Gegenprobe zum Prüfer selbst** (§10): ein Prüfer, der nur an gültigen
// Daten läuft, ist von einem grünen `return nil` nicht zu unterscheiden.
func TestDerPrueferWeistUngueltigesAb(t *testing.T) {
	faelle := map[string]struct {
		inhalt   string
		enthaelt string
	}{
		"Manakosten ohne Klammern": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":[{"name":"A","manaCost":"2R"}]}}`,
			enthaelt: "manaCost",
		},
		"eine Seltenheit, die nur MSE kennt": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":[{"name":"A","rarity":"mythic rare"}]}}`,
			enthaelt: "rarity",
		},
		"eine Typliste als Zeichenkette": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":[{"name":"A","types":"Creature"}]}}`,
			enthaelt: "types",
		},
		"eine Karte ohne Namen": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":[{"manaCost":"{R}"}]}}`,
			enthaelt: "name",
		},
		"ein Set ohne Karten": {
			inhalt:   `{"data":{"code":"X","name":"X"}}`,
			enthaelt: "cards",
		},
		// **Ohne Umschlag gilt der Inhalt selbst** — geändert am 2026-08-24
		// (Zweitblick über Phase V).
		//
		// Hier stand `enthaelt: "data"`: der Prüfer meldete „Es fehlt der
		// Umschlag". Das war **falsch** — der Leser deutet eine Datei ohne
		// `data` seit jeher als Set (siehe `setsAus`), und
		// `docs/FORMATE.md` führt drei Gestalten, von denen zwei ohne
		// Umschlag auskommen.
		//
		// **Die Änderung ist eine Semantikänderung** (§15, Regel 4): der
		// Prüfer meldet jetzt, was an dem Inhalt fehlt — hier `cards` —,
		// statt den fehlenden Umschlag zu beklagen. Das ist die Meldung, die
		// weiterhilft.
		"ohne Umschlag wird der Inhalt selbst gedeutet": {
			inhalt:   `{"meta":{"version":"5"}}`,
			enthaelt: "cards",
		},
		"kein JSON": {
			inhalt:   `{ kaputt`,
			enthaelt: "JSON",
		},
		"ein data, das kein Objekt ist": {
			inhalt:   `{"data": [1, 2]}`,
			enthaelt: "weder ein Set noch",
		},
		"ein Eintrag der Abbildung, der kein Set ist": {
			inhalt:   `{"data":{"AAA":"kein Set"}}`,
			enthaelt: "AAA",
		},
		"cards, das keine Liste ist": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":"drei"}}`,
			enthaelt: "keine Liste von Karten",
		},
		"eine Zahl als Zeichenkette": {
			inhalt:   `{"data":{"code":"X","name":"X","cards":[{"name":"A","convertedManaCost":"3"}]}}`,
			enthaelt: "convertedManaCost",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			befunde := mtgjson.Pruefe([]byte(fall.inhalt), "probe.json")
			if len(befunde) == 0 {
				t.Fatalf("kein Befund für %s", fall.inhalt)
			}

			if !enthaelt(befunde, fall.enthaelt) {
				t.Errorf("Befunde: %v — %q kommt darin nicht vor", befunde, fall.enthaelt)
			}
		})
	}
}

// Eine AllSets-Datei wird genauso geprüft wie eine einzelne.
func TestDerPrueferLiestAuchEineAllSetsDatei(t *testing.T) {
	gut := `{"data":{"AAA":{"code":"AAA","name":"A","cards":[{"name":"Eine"}]}}}`
	if befunde := mtgjson.Pruefe([]byte(gut), "probe.json"); len(befunde) > 0 {
		t.Errorf("Befunde: %v, erwartet keine", befunde)
	}

	schlecht := `{"data":{"AAA":{"code":"AAA","name":"A","cards":[{"name":"Eine","types":"Creature"}]}}}`
	if befunde := mtgjson.Pruefe([]byte(schlecht), "probe.json"); len(befunde) == 0 {
		t.Error("kein Befund — eine AllSets-Datei wird nicht geprüft")
	}
}

// Ein unbekannter Schlüssel ist **kein** Befund.
//
// MTGJSON wächst; eine Prüfung, die jeden neuen Schlüssel abweist, wäre nach dem
// nächsten Erscheinen rot, ohne dass sich hier etwas geändert hätte. Geprüft
// wird, was wir kennen — nicht, dass wir alles kennen.
func TestEinUnbekannterSchluesselIstKeinBefund(t *testing.T) {
	inhalt := `{"data":{"code":"X","name":"X","cards":[{"name":"A","ganzNeu":{"was":"auch immer"}}]}}`

	if befunde := mtgjson.Pruefe([]byte(inhalt), "probe.json"); len(befunde) > 0 {
		t.Errorf("Befunde: %v, erwartet keine", befunde)
	}
}

// enthaelt sagt, ob eine Befundliste einen Text führt.
func enthaelt(befunde []domain.Befund, text string) bool {
	for _, befund := range befunde {
		if strings.Contains(befund.String(), text) {
			return true
		}
	}

	return false
}

// **Jeder Befund nennt seine Stelle — auch bei einem Set ohne `code`.**
//
// # Der erste Fund eines Fuzzlaufs in diesem Projekt
//
// Am 2026-08-29 fand `FuzzLiesInhalt` die Eingabe `{"":{}}`: ein Set, dessen
// Schlüssel leer ist und das kein `code` trägt. Der Befund lautete „das Set
// führt kein »cards«" — **ohne Ort**. Bei dreihundert Sets sagt das nicht,
// welches gemeint ist, und §8 verlangt die Stelle: Datei, Zeile, Feld.
//
// **Der Fall bleibt als Testfall**, wie das Issue es verlangt
// (#1289) —
// hier mit Namen, und daneben im Korpus unter
// `testdata/fuzz/FuzzLiesInhalt/`.
func TestEinBefundNenntSeineStelleAuchOhneCode(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		inhalt   string
		erwartet string
	}{
		{"Schlüssel und Code fehlen", `{"":{}}`, "(Set ohne Code und ohne Namen)"},
		{"nur der Schlüssel ist da", `{"PRB":{}}`, "PRB"},
		{"der Code schlägt den Schlüssel", `{"PRB":{"code":"XYZ"}}`, "XYZ"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			befunde := mtgjson.Pruefe([]byte(fall.inhalt), "probe.json")
			if len(befunde) == 0 {
				t.Fatalf("kein Befund für %s", fall.inhalt)
			}

			for _, befund := range befunde {
				if befund.Ort == "" {
					t.Errorf("ein Befund ohne Ort: %q", befund.Meldung)
				}
			}

			if befunde[0].Ort != fall.erwartet {
				t.Errorf("der Ort ist %q — erwartet war %q",
					befunde[0].Ort, fall.erwartet)
			}
		})
	}
}
