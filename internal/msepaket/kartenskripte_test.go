package msepaket_test

// **`add cards script:` — Karten aus einem Skript** — Werkstück B von Phase
// 4.11 (#1337,
// #1349).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `src/data/add_cards_script.hpp`, Zeilen 12–26 (Fassung f9d9356 vom
// 2021-07-06):
//
//	class AddCardsScript : public IntrusivePtrBase<AddCardsScript> {
//	public:
//	  String           name;
//	  String           description;
//	  Scriptable<bool> enabled;
//	  OptionalScript   script;
//
// Vier Angaben, und `enabled` ist **selbst ein Skript** — es entscheidet je
// Set, ob der Eintrag zur Wahl steht.
//
// # Und was im Pack steht
//
// `magic.mse-game/add_cards_scripts` führt **zwölf** Blöcke (gezählt am
// 2026-08-31; ein dreizehnter steht auskommentiert da). Ihre Namen tragen ein
// `&` vor dem Buchstaben, der das Tastenkürzel gibt — `&Basic Lands`. Das ist
// wxWidgets-Schreibweise und gehört **nicht** in den angezeigten Namen.

import (
	"testing"
)

// **Ein `add cards script:` wird mit allen vier Angaben gelesen.**
func TestEinKartenskriptWirdGelesen(t *testing.T) {
	t.Parallel()

	paket := spielMitBlock(t, `add cards script:
	name: &Basic Lands
	description: Adds 5 basic lands to the set
	enabled: { true }
	script:
		[ new_card([name: "Plains"])
		, new_card([name: "Island"])
		]
add cards script:
	name: Leerer Zug
	script:
		[ new_card() ]
`)

	skripte := paket.Kartenskripte()

	if len(skripte) != 2 {
		t.Fatalf("%d Skripte, erwartet waren zwei", len(skripte))
	}

	erste := skripte[0]

	// **Das `&` ist ein Tastenkürzel und kein Name.** Wer es stehen ließe,
	// zeigte „&Basic Lands" in einem Menü, das keine wxWidgets-Kürzel kennt.
	if erste.Name != "Basic Lands" {
		t.Errorf("der Name lautet %q", erste.Name)
	}

	if erste.Kuerzel != "B" {
		t.Errorf("das Kürzel lautet %q, erwartet war B", erste.Kuerzel)
	}

	if erste.Beschreibung != "Adds 5 basic lands to the set" {
		t.Errorf("die Beschreibung lautet %q", erste.Beschreibung)
	}

	// **Mit den geschweiften Klammern.** Sie sind die Form, in der dieses
	// Projekt jedes Skript führt (`stilmasse.go`: „die Klammern sind kein
	// Kunstgriff, sondern die Gleichmachung der beiden Formen"), und der
	// Auswerter erwartet sie. Der erste Anlauf dieses Laufs verlangte `true`
	// und lag falsch — nicht der Leser.
	if erste.Bedingung != "{ true }" {
		t.Errorf("die Bedingung lautet %q", erste.Bedingung)
	}

	if erste.Skript == "" {
		t.Error("das Skript fehlt")
	}

	// **Ohne `enabled:` steht der Eintrag zur Wahl.** So ist es im Altcode:
	// `Scriptable<bool> enabled` ohne Angabe ist wahr, und ein Eintrag, der
	// ohne Grund fehlte, wäre schlimmer als einer zu viel.
	if skripte[1].Bedingung != "" {
		t.Errorf("die zweite Bedingung lautet %q, erwartet war leer",
			skripte[1].Bedingung)
	}

	if skripte[1].Name != "Leerer Zug" || skripte[1].Kuerzel != "" {
		t.Errorf("das zweite Skript lautet %+v", skripte[1])
	}
}

// **Das Kaufmanns-Und in seinen vier Gestalten.**
//
// wxWidgets nimmt `&` als Kürzelmarke; `&&` ist ein echtes Und. Der Leser muss
// beide unterscheiden — und zwei Randfälle dazu, die im Pack nicht vorkommen
// und trotzdem im Code stehen: ein `&` am Ende und ein zweites im Namen.
//
// **Sie stehen hier, weil der Zweig sonst unbetreten bliebe** und §3 keinen
// zulässt, den kein Lauf erreicht.
func TestDasKaufmannsUndInSeinenGestalten(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		roh     string
		name    string
		kuerzel string
	}{
		"das übliche":                {"&Basic Lands", "Basic Lands", "B"},
		"ohne Kürzel":                {"Basic Lands", "Basic Lands", ""},
		"mitten im Wort":             {"Basic &Lands", "Basic Lands", "L"},
		"ein echtes Und":             {"Rock && Roll", "Rock & Roll", ""},
		"beides":                     {"&Rock && Roll", "Rock & Roll", "R"},
		"am Ende markiert es nichts": {"Basic&", "Basic", ""},

		// **Nur das erste markiert** — wie bei wxWidgets. Ein zweites wäre
		// eine Angabe, die niemand auflösen kann.
		"zwei Marken": {"&Basic &Lands", "Basic Lands", "B"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			paket := spielMitBlock(t, "add cards script:\n\tname: "+fall.roh+
				"\n\tscript:\n\t\t[ new_card() ]\n")

			skripte := paket.Kartenskripte()
			if len(skripte) != 1 {
				t.Fatalf("%d Skripte", len(skripte))
			}

			if skripte[0].Name != fall.name || skripte[0].Kuerzel != fall.kuerzel {
				t.Errorf("%q gab Name %q und Kürzel %q — erwartet %q und %q",
					fall.roh, skripte[0].Name, skripte[0].Kuerzel,
					fall.name, fall.kuerzel)
			}
		})
	}
}
