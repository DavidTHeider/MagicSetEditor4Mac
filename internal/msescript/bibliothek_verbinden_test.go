package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`combined_editor` setzt mehrere Felder zu einem zusammen.**
//
// # Wozu
//
// Das Feld `text` von `magic.mse-game` trägt `save value: false` und entsteht
// erst daraus:
//
//	combined_editor(field1: card.rule_text, separator: "<line>\n</line>",
//	                field2: card.flavor_text)
//
// Regeltext, Blockgrenze, Flavortext — der ganze Textkasten einer Karte.
//
// # Die Regel
//
//	new_value = value_parts.front();
//	for (i = 1 …) new_value += "<sep>" + separators[i-1] + "</sep>" + value_parts[i];
//
// `src/script/functions/editor.cpp`, Zeilen 96–116, Fassung f9d9356 vom
// 2021-07-06. Der Trenner steht **in** `<sep>`-Klammern; MSE schneidet daran
// beim Bearbeiten, gezeichnet wird sein Inhalt
// (#345).
func TestCombinedEditorSetztFelderZusammen(t *testing.T) {
	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{
			name:     "zwei Felder mit Trenner",
			quelle:   `combined_editor(field1: "a", separator: "<line>\n</line>", field2: "b")`,
			erwartet: "a<sep><line>\n</line></sep>b",
		},
		{
			name: "drei Felder",
			// **`separator` und `separator1`**, nicht `separator2`: MSE hängt
			// die Nummer ab dem **zweiten** Trenner an und beginnt bei null
			// (`String name = _("separator"); if (i > 0) name = name << i;`).
			quelle:   `combined_editor(field1: "a", separator: "-", field2: "b", separator1: "+", field3: "c")`,
			erwartet: "a<sep>-</sep>b<sep>+</sep>c",
		},
		{
			// **Ein leeres Feld bekommt trotzdem seinen Trenner** — solange
			// niemand `hide_when_empty` sagt. Das ist die Vorgabe.
			name:     "ein leeres zweites Feld",
			quelle:   `combined_editor(field1: "a", separator: "-", field2: "")`,
			erwartet: "a<sep>-</sep>",
		},
		{
			// **`hide_when_empty` greift nur, solange **alles** leer ist.** MSE
			// prüft `value_parts[i].second && new_value_empty && hide_when_empty`
			// — steht vorne schon Text, kommt der Trenner trotzdem.
			name:     "hide_when_empty bei lauter leeren Feldern",
			quelle:   `combined_editor(field1: "", separator: "-", field2: "", hide_when_empty: true)`,
			erwartet: "",
		},
		{
			name:     "hide_when_empty greift nicht hinter einem gefüllten Feld",
			quelle:   `combined_editor(field1: "a", separator: "-", field2: "", hide_when_empty: true)`,
			erwartet: "a<sep>-</sep>",
		},
		{
			name:     "ein einziges Feld bleibt es selbst",
			quelle:   `combined_editor(field1: "a")`,
			erwartet: "a",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Ohne ein einziges Feld ist es ein Fehler.**
//
//	if (values.empty()) throw ScriptError("No fields specified for combined_editor");
//
// `src/script/functions/editor.cpp`, Zeilen 38–40. Ein leeres Ergebnis wäre
// eine erfundene Antwort auf eine Vorlage, die falsch geschrieben ist.
//
// Fassung f9d9356 vom 2021-07-06.
func TestCombinedEditorOhneFeldIstEinFehler(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `combined_editor(separator: "-")`, msescript.Bibliothek(),
	); err == nil {
		t.Error("kein Fehler ohne Feld")
	}
}

// **Zu wenige Trenner sind ein Fehler.**
//
//	if (separators.size() < values.size() - 1) throw ScriptError(…)
//
// Zeilen 50–52. Ohne die Prüfung stünde ein Feld ohne Trenner am nächsten —
// zwei Texte, die aussehen wie einer.
func TestCombinedEditorMitZuWenigTrennernIstEinFehler(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `combined_editor(field1: "a", field2: "b")`, msescript.Bibliothek(),
	); err == nil {
		t.Error("kein Fehler ohne Trenner")
	}
}

// **`soft_before_empty` macht den Trenner vor einem leeren Feld weich.**
//
//	if (value_parts[i].second && new_value_empty && hide_when_empty) {
//	  // no separator
//	} else if (value_parts[i].second && soft_before_empty) {
//	  new_value += _("<sep-soft>") + separators[i - 1] + _("</sep-soft>");
//	  new_value_empty = false;
//	} else {
//	  new_value += _("<sep>") + separators[i - 1] + _("</sep>");
//	  new_value_empty = false;
//	}
//
// `src/script/functions/editor.cpp`, Zeilen 96–116, Fassung f9d9356 vom
// 2021-07-06.
//
// **Drei Fälle, und die Reihenfolge ist die Aussage**: ganz weg nur, wenn auch
// alles davor leer ist; sonst weich; sonst gewöhnlich.
//
// `magic.mse-game` baut die Typzeile mit beiden Angaben. Bei einer Zauberei
// ohne Untertyp steht der Gedankenstrich damit **weich** da — und weicher Text
// zählt bei der Ausrichtung nicht mit
// (#345).
func TestSoftBeforeEmptyMachtDenTrennerWeich(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name: "vor einem leeren Feld wird der Trenner weich",
			quelle: `combined_editor(field1: "Sorcery", separator: " — ", field2: "",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "Sorcery<sep-soft> — </sep-soft>",
		},
		{
			name: "mit Inhalt bleibt er gewöhnlich",
			quelle: `combined_editor(field1: "Creature", separator: " — ", field2: "Goblin",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "Creature<sep> — </sep>Goblin",
		},
		{
			// **Ganz weg nur, wenn auch alles davor leer ist.**
			name: "alles leer heißt kein Trenner",
			quelle: `combined_editor(field1: "", separator: " — ", field2: "",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "",
		},
		{
			// **Ohne `soft_before_empty` bleibt es beim gewöhnlichen Trenner.**
			name: "ohne die Angabe",
			quelle: `combined_editor(field1: "Sorcery", separator: " — ", field2: "",
				hide_when_empty: true)`,
			erwartet: "Sorcery<sep> — </sep>",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}

// **Leer heißt „ohne Zeichen zwischen den Marken", nicht „die leere Kette".**
//
//	nv.second = index_to_untagged(nv.first, nv.first.size()) == 0;
//
// `src/script/functions/editor.cpp`, Zeile 93, Fassung f9d9356 vom
// 2021-07-06. MSE zählt die Zeichen **außerhalb** der Marken; diese Umsetzung
// verglich mit `""`.
//
// # Was das gekostet hat
//
// `magic.mse-game` gibt einen leeren Untertyp nicht als leere Kette zurück:
//
//	if sub_types == "" then "<word-list-{list_type_first}></word-list-{list_type_first}>"
//
// `magic.mse-game/script`, Zeile 2363 — ein leeres Wortlistenfeld für die
// Eingabe. Der Vergleich mit `""` sah darin einen Inhalt, nahm den **harten**
// Trenner, und der Gedankenstrich stand auf jeder Karte ohne Untertyp.
//
// **Gemessen über das kanonische Set `Black`**, alle 3.788 Karten, Stil
// `magic-m15`: 1.308 Typzeilen endeten sichtbar auf `—` — `"Instant — "`,
// `"Enchantment — "`, `"Sorcery — "`. Gemeldet aus der laufenden Anwendung am
// 2026-08-21 — #725.
func TestNurMarkenGiltAlsLeer(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name: "ein leeres Wortlistenfeld ist leer",
			quelle: `combined_editor(field1: "Instant", separator: " — ",
				field2: "<word-list-spell></word-list-spell>",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "Instant<sep-soft> — </sep-soft><word-list-spell></word-list-spell>",
		},
		{
			name: "mit einem Wort darin ist es nicht leer",
			quelle: `combined_editor(field1: "Creature", separator: " — ",
				field2: "<word-list-race>Goblin</word-list-race>",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "Creature<sep> — </sep><word-list-race>Goblin</word-list-race>",
		},
		{
			// **Auch das erste Feld zählt so.** Sind beide nur Marken, fällt
			// der Trenner ganz weg.
			name: "beide nur Marken heißt kein Trenner",
			quelle: `combined_editor(field1: "<i></i>", separator: " — ",
				field2: "<word-list-spell></word-list-spell>",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "<i></i><word-list-spell></word-list-spell>",
		},
		{
			// **Ein Kleinerzeichen ohne Klammer verschluckt den Rest** —
			// `skip_tag` gibt `npos`, und die Zählschleife endet daran
			// (`index_to_untagged`, `tagged_string.cpp`, Zeilen 543–556).
			name: "ein offenes Kleinerzeichen zählt nicht mit",
			quelle: `combined_editor(field1: "Instant", separator: " — ",
				field2: "<offen",
				soft_before_empty: true, hide_when_empty: true)`,
			erwartet: "Instant<sep-soft> — </sep-soft><offen",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, msescript.Bibliothek())
			if err != nil {
				t.Fatalf("ausrechnen: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
