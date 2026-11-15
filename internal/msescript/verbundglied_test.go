package msescript_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Ein reflektierter Verbund findet sein Glied über den Präfix**
// (#494).
//
//	template <typename T>
//	void handle(const Char* name, const T& object) {
//	  if (!gdm.result() && canonical_name_compare(target_name, name)) {
//	    gdm.handle(object);
//	  }
//	}
//
// `src/util/io/get_member.hpp`, Zeilen 91–96, und dazu:
//
//	bool canonical_name_compare(StringView as, const Char* b) {
//	  for (String::const_iterator a_it = as.begin(); a_it != as.end(); ++a_it, ++b) {
//	    if (*a_it != *b && !(*a_it == '_' && *b == ' ')) return false;
//	  }
//	  return true;
//	}
//
// `src/util/string.cpp`, Zeilen 339–347, Fassung f9d9356 vom 2021-07-06.
//
// **Die Schleife läuft über den gesuchten Namen und gibt am Ende `true`** —
// ohne zu prüfen, ob der Gliedname zu Ende ist. `separator` findet damit
// `separators`.
//
// # Warum es nicht für jede Abbildung gilt
//
// MSE benutzt diesen Vergleich für **reflektierte C++-Verbunde**, nicht für
// Skriptabbildungen; dort vergleicht es genau. Ein Kartenfeld `text` fände
// sonst `text_box_mana_symbols`.
//
// # Umfang, gemessen
//
// Über das ganze Pack greifen 28 Stellen auf `layout.separator` zu (Einzahl)
// und 35 auf `layout.separators`. **Ein anderer Präfixzugriff auf `layout`
// kommt nicht vor** — die Regel ist trotzdem die Regel.
func TestEinVerbundFindetSeinGliedUeberDenPraefix(t *testing.T) {
	t.Parallel()

	verbund := msescript.Wert{
		Art:     msescript.ArtAbbildung,
		Verbund: true,
		Eintraege: []msescript.Eintrag{
			{Name: "blocks", Wert: msescript.Zahl(1)},
			{Name: "separators", Wert: msescript.Zahl(2)},
			{Name: "mit luecke", Wert: msescript.Zahl(3)},
		},
	}

	faelle := []struct {
		name     string
		gesucht  string
		erwartet float64
		gibt     bool
	}{
		{"der genaue Name", "separators", 2, true},
		{"der Präfix", "separator", 2, true},
		{"ein kürzerer Präfix", "sep", 2, true},
		{"ein anderes Glied", "blocks", 1, true},
		{"länger als das Glied", "separatorsss", 0, false},
		{"gar kein Glied", "hoehe", 0, false},
		// **Ein `_` im gesuchten Namen trifft ein Leerzeichen im Glied** —
		// MSEs Ausnahme in derselben Zeile.
		{"Unterstrich gegen Leerzeichen", "mit_luecke", 3, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			wert, gibt := verbund.Eintrag(fall.gesucht)
			if gibt != fall.gibt {
				t.Fatalf("gefunden %v, erwartet %v", gibt, fall.gibt)
			}

			if gibt && wert.Zahl != fall.erwartet {
				t.Errorf("Wert %g, erwartet %g", wert.Zahl, fall.erwartet)
			}
		})
	}
}

// TestEineGewoehnlicheAbbildungVergleichtGenau ist die Gegenprobe.
//
// **Ohne sie fände ein Kartenfeld `text` das Feld `text_box_mana_symbols`** —
// und niemand sähe, warum plötzlich der Name einer Symbolschrift im Textkasten
// steht.
func TestEineGewoehnlicheAbbildungVergleichtGenau(t *testing.T) {
	t.Parallel()

	abbildung := msescript.Wert{
		Art: msescript.ArtAbbildung,
		Eintraege: []msescript.Eintrag{
			{Name: "text_box_mana_symbols", Wert: msescript.Text("magic-mana-small")},
		},
	}

	if _, gibt := abbildung.Eintrag("text"); gibt {
		t.Error("`text` hat `text_box_mana_symbols` gefunden — eine gewöhnliche " +
			"Abbildung vergleicht genau")
	}
}
