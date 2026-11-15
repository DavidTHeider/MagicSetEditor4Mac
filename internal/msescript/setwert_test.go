package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **`set` ist die Kartenliste und die Setangaben zugleich**
// (#653).
//
// Zwei Gesichter an einem Wert:
//
//	ScriptValueP make_iterator(const Set& set) {
//	  return make_intrusive<ScriptCollectionIterator<vector<CardP>>>(&set.cards);
//	}
//
//	// in scripts, set.something is read from the set_info
//	void reflect_set_info_get_member(GetMember& handler, const IndexMap<FieldP, ValueP>& data) {
//	  REFLECT_NAMELESS(data);
//	}
//
// `src/data/set.cpp`, Zeilen 251–253 und 271–275, Fassung f9d9356 vom
// 2021-07-06. **Durchlaufen** gibt die Karten, **`set.<name>`** die Angabe.
func TestSetTraegtBeideGesichter(t *testing.T) {
	t.Parallel()

	// Wie der Dienst ihn baut: die Karten als Einträge, die Angaben als
	// Glieder.
	set := msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Name: "0", Wert: msescript.Wert{
				Art:       msescript.ArtAbbildung,
				Eintraege: []msescript.Eintrag{{Name: "name", Wert: msescript.Text("Erste")}},
			}},
			{Name: "1", Wert: msescript.Wert{
				Art:       msescript.ArtAbbildung,
				Eintraege: []msescript.Eintrag{{Name: "name", Wert: msescript.Text("Zweite")}},
			}},
		},
		Glieder: []msescript.Eintrag{
			{Name: "title", Wert: msescript.Text("Mein Set")},
			{Name: "set code", Wert: msescript.Text("MS1")},
		},
	}

	faelle := []struct {
		name     string
		quelle   string
		erwartet string
	}{
		{
			name:     "durchlaufen gibt die Karten",
			quelle:   `for each card in set do card.name + ";"`,
			erwartet: "Erste;Zweite;",
		},
		{
			name:     "ein Glied gibt die Setangabe",
			quelle:   `set.title`,
			erwartet: "Mein Set",
		},
		{
			name:     "gezählt werden die Karten",
			quelle:   `length(set)`,
			erwartet: "2",
		},
		{
			name:     "geordnet werden die Karten",
			quelle:   `for each card in sort_list(set, order_by: {card.name}) do card.name + ";"`,
			erwartet: "Erste;Zweite;",
		},
		// **`set.0` gibt keine Karte**, und das ist MSEs Verhalten:
		// `getMember` geht allein ins `set_info`.
		{
			name:     "eine Stelle gibt keine Karte",
			quelle:   `set.0 or else "nichts"`,
			erwartet: "nichts",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			umgebung := msescript.Bibliothek().Darueber()
			umgebung.Setze("set", set)

			wert, err := msescript.AuswertenQuelle(t.Context(), fall.quelle, umgebung)
			if err != nil {
				t.Fatalf("unerwarteter Fehler: %v", err)
			}

			if wert.AlsText() != fall.erwartet {
				t.Errorf("Ergebnis = %q, erwartet %q", wert.AlsText(), fall.erwartet)
			}
		})
	}
}
