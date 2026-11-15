package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`remove_choice` und `require_exclusive_choice` sieben eine Wahlliste.**
//
// # Was eine Wahlliste ist
//
// Ein `choice`-Feld mit Mehrfachauswahl trägt seine Werte als Text:
// `"a, b, c"`. `filter_choices` geht die Liste durch, merkt sich, welche der
// genannten Wahlen vorkommen, und setzt sie am Ende neu zusammen
// (`src/script/functions/editor.cpp`, Zeilen 248–308, Fassung f9d9356 vom
// 2021-07-06).
//
//   - `remove_choice` ruft es mit `min = max = 0` — die genannten Wahlen
//     fallen weg.
//   - `require_exclusive_choice` mit `min = max = 1` — genau eine bleibt, und
//     `last_change:` sagt, welche bevorzugt wird.
//
// **Zusammen 4.750 Befunde** über die 190 Stile von `M15-Magic-Pack`, sobald
// die einzeiligen Wertskripte laufen
// (#345).
func TestWahlenWerdenGesiebt(t *testing.T) {
	faelle := []struct {
		name, quelle, erwartet string
	}{
		{
			name:     "eine Wahl fällt weg",
			quelle:   `remove_choice(input: "a, b, c", choice: "b")`,
			erwartet: "a, c",
		},
		{
			name:     "mehrere Wahlen über `choices:`",
			quelle:   `remove_choice(input: "a, b, c", choices: "a, c")`,
			erwartet: "b",
		},
		{
			// **Die unbeteiligten stehen vorn, die behaltenen hinten** — MSE
			// setzt erst die nicht genannten zusammen und hängt die gemerkten
			// an (Zeilen 264–268 und 299–305).
			name:     "genau eine bleibt",
			quelle:   `require_exclusive_choice(input: "x, a, b", choices: "a, b")`,
			erwartet: "x, a",
		},
		{
			// **`last_change:` wird bevorzugt** — beim Wegnehmen bleibt sie
			// stehen (Zeile 294).
			name:     "die zuletzt geänderte bleibt",
			quelle:   `require_exclusive_choice(input: "a, b", choices: "a, b", last_change: "b")`,
			erwartet: "b",
		},
		{
			// **Ist keine da, kommt die bevorzugte dazu** (Zeilen 271–288).
			name:     "eine fehlende wird ergänzt",
			quelle:   `require_exclusive_choice(input: "x", choices: "a, b", last_change: "b")`,
			erwartet: "x, b",
		},
		{
			// **Ohne Wahlen bleibt die Eingabe** (Zeile 249).
			name:     "ohne Wahlen bleibt alles",
			quelle:   `remove_choice(input: "a, b")`,
			erwartet: "a, b",
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

// **Die alte Schreibweise mit `choice1:`, `choice2:` gilt weiter.**
//
// `read_choices_param` liest erst `choices:`, dann `choice:`, dann die
// nummerierten (Zeilen 310–328).
func TestDieAlteSchreibweiseDerWahlenGiltWeiter(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `remove_choice(input: "a, b, c", choice1: "a", choice2: "c")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "b" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "b")
	}
}

// **Beide brauchen einen Text.**
//
// MSE liest ihn als `SCRIPT_PARAM_C(String, input)` ohne Vorgabe; ohne ihn ist
// die Vorlage falsch geschrieben.
func TestDieWahlsiebeVerlangenIhrenText(t *testing.T) {
	for _, quelle := range []string{
		`remove_choice(choice: "a")`,
		`require_exclusive_choice(choices: "a, b")`,
	} {
		t.Run(quelle, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek(),
			); err == nil {
				t.Error("kein Fehler ohne Text")
			}
		})
	}
}

// **Ohne `last_change:` wird die erste Wahl ergänzt.**
//
// `filter_choices` sucht erst die bevorzugte und geht danach der Reihe nach
// (Zeilen 271–288). Ohne bevorzugte trifft der zweite Durchgang.
func TestOhneBevorzugteWahlKommtDieErste(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(), `require_exclusive_choice(input: "x", choices: "a, b")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "x, a" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "x, a")
	}
}

// **Beim Kürzen bleibt die bevorzugte stehen, und die nicht gemerkten werden
// übersprungen.**
//
//	if (!seen[i]) continue;
//	if (max > 0 && choices[i] == prefered) continue;
//
// Zeilen 292–296: drei Wahlen sind da, eine soll bleiben, und die bevorzugte
// steht **vorn** — gekürzt wird von hinten, also fällt sie zuletzt.
func TestBeimKuerzenBleibtDieBevorzugte(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(),
		`require_exclusive_choice(input: "a, b, c", choices: "a, b, c", last_change: "a")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "a" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "a")
	}
}

// **Eine Wahl, die gar nicht dasteht, wird beim Kürzen übersprungen.**
//
//	if (!seen[i]) continue;
//
// Zeile 293. Gekürzt wird von hinten; steht die letzte genannte Wahl nicht im
// Text, muss die Schleife über sie hinweggehen, statt sie zu zählen.
func TestBeimKuerzenWerdenFehlendeWahlenUebersprungen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(
		t.Context(),
		`require_exclusive_choice(input: "a, b", choices: "a, b, c")`,
		msescript.Bibliothek())
	if err != nil {
		t.Fatalf("ausrechnen: %v", err)
	}

	if wert.AlsText() != "a" {
		t.Errorf("ergibt %q, erwartet %q", wert.AlsText(), "a")
	}
}

// **`exclusive_choice` lässt höchstens eine Wahl stehen, `forward_editor` ist
// `combined_editor`.**
//
//	SCRIPT_FUNCTION(exclusive_choice) {
//	  …
//	  SCRIPT_RETURN(filter_choices(input, choices, 0, 1, last_change));
//	}
//
//	ctx.setVariable(_("forward_editor"), script_combined_editor); // compatability
//
// `src/script/functions/editor.cpp`, Zeilen 340–346 und 390, Fassung f9d9356
// vom 2021-07-06.
//
// **`exclusive_choice` gegen `require_exclusive_choice`**: die Grenzen sind
// 0…1 statt 1…1 — es darf **keine** Wahl übrig bleiben, aber nicht zwei.
//
// Zusammen 56 Befunde über die 190 Stile von `M15-Magic-Pack`
// (#270).
func TestExclusiveChoiceUndForwardEditor(t *testing.T) {
	faelle := []struct{ name, quelle, erwartet string }{
		{
			name:     "zwei Wahlen werden auf eine gekürzt",
			quelle:   `exclusive_choice("a, b", choices: "a, b")`,
			erwartet: "a",
		},
		{
			// **Keine Wahl bleibt keine** — anders als bei
			// `require_exclusive_choice`, das eine erzwingt.
			name:     "keine Wahl bleibt keine",
			quelle:   `exclusive_choice("", choices: "a, b")`,
			erwartet: "",
		},
		{
			name:     "und die erzwungene Fassung füllt auf",
			quelle:   `require_exclusive_choice("", choices: "a, b")`,
			erwartet: "a",
		},
		{
			// **`last_change:` entscheidet, welche bleibt** — dieselbe Regel
			// wie bei `require_exclusive_choice`.
			name:     "die zuletzt geänderte Wahl bleibt",
			quelle:   `exclusive_choice("a, b", choices: "a, b", last_change: "b")`,
			erwartet: "b",
		},
		{
			name:     "forward_editor ist combined_editor",
			quelle:   `forward_editor(field1: "Creature", separator: " — ", field2: "Goblin")`,
			erwartet: "Creature<sep> — </sep>Goblin",
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

// **`exclusive_choice` zählt seine Stellung.**
//
// Ohne Eingabe ist es ein Fehler und keine leere Antwort — dieselbe Regel wie
// bei jeder anderen eingebauten Funktion (§8).
func TestExclusiveChoiceZaehltSeineStellung(t *testing.T) {
	if _, err := msescript.AuswertenQuelle(
		t.Context(), `exclusive_choice()`, msescript.Bibliothek()); err == nil {
		t.Error("erwartet einen Fehler")
	}
}
