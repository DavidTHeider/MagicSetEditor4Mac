package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **`<b>` schneidet den Zug in Abschnitte, so wie `<i>` es tut.**
//
// MSE zählt beide gleich — `bold += 1` beim Öffnen, `-= 1` beim Schließen
// (`src/render/text/element.cpp`, Zeilen 78–79, Fassung f9d9356 vom
// 2021-07-06). Verschieden ist erst, **wie** aus dem Schnitt eine Schriftdatei
// wird ([0023](../../docs/entscheidungen/0023-fett-wird-ueber-den-namen-gesucht.md)).
func TestFetteStellenWerdenEigeneAbschnitte(t *testing.T) {
	faelle := []struct {
		name    string
		text    string
		fette   []bool
		texte   []string
		befunde bool
	}{
		{
			name:  "eine fette Stelle mitten im Text",
			text:  "Tappe: <b>Feuer</b>ball",
			fette: []bool{false, true, false},
			texte: []string{"Tappe: ", "Feuer", "ball"},
		},
		{
			// **Gezählt und nicht geschaltet** — dieselbe Regel wie bei `<i>`.
			name:  "verschachtelt",
			text:  "<b>a<b>b</b>c</b>d",
			fette: []bool{true, false},
			texte: []string{"abc", "d"},
		},
		{
			// **`<bullet>` setzt **nicht** fett.** Hier stand das Gegenteil,
			// mit Vermerk 0023 als Deckung — und der Vermerk hat `is_tag` nur
			// zur Hälfte gelesen:
			//
			//	return is_substr(str, pos, tag)
			//	    && pos+tag.size() < str.size()
			//	    && is_tag_end_char(str[pos+tag.size()]);
			//
			// `src/util/tagged_string.cpp`, Zeilen 270–272, Fassung f9d9356
			// vom 2021-07-06. Hinter dem Präfix muss ein `>`, `-`, `:` oder
			// ein Leerzeichen stehen (ebd. Zeilen 28–30); in `<bullet>` steht
			// dort ein `u`.
			//
			// **Die Gegenprobe steht im Altcode selbst**: `element.cpp:171`
			// führt einen eigenen `</bullet`-Zweig, und den erreichte niemand,
			// wenn `<b` ihn vorher schluckte
			// (#429).
			name:  "bullet setzt nicht fett",
			text:  "a<bullet>b</bullet>c",
			fette: []bool{false, false, false},
			texte: []string{"a", "b", "c"},
			// **Und es wird nicht mehr gemeldet**: der Rand des
			// Aufzählungspunktes steht seit
			// #431.
		},
		{
			name:  "fett und kursiv zugleich",
			text:  "<b><i>beides</i></b>",
			fette: []bool{true},
			texte: []string{"beides"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			abschnitte := textabschnitte(t, fall.text, fall.befunde)

			if len(abschnitte) != len(fall.fette) {
				t.Fatalf("%d Abschnitte, erwartet %d: %+v",
					len(abschnitte), len(fall.fette), abschnitte)
			}

			for stelle, abschnitt := range abschnitte {
				if abschnitt.Fett != fall.fette[stelle] ||
					abschnitt.Text != fall.texte[stelle] {
					t.Errorf("Abschnitt %d ist %q (fett: %v), erwartet %q (fett: %v)",
						stelle, abschnitt.Text, abschnitt.Fett,
						fall.texte[stelle], fall.fette[stelle])
				}
			}
		})
	}
}

// **Und `<i>` allein macht nichts fett.** Die Gegenprobe: ohne sie bestünde
// jeder Fall, der alles fett setzte.
func TestKursivAlleinMachtNichtsFett(t *testing.T) {
	abschnitte := textabschnitte(t, "<i>nur kursiv</i>", false)

	if len(abschnitte) != 1 {
		t.Fatalf("%d Abschnitte, erwartet einer: %+v", len(abschnitte), abschnitte)
	}

	if abschnitte[0].Fett || !abschnitte[0].Kursiv {
		t.Errorf("der Abschnitt ist %+v, erwartet kursiv und nicht fett", abschnitte[0])
	}
}

// textabschnitte baut den Plan zu einem Regeltext und gibt seine Abschnitte.
func textabschnitte(t *testing.T, text string, erwarteteBefunde bool) []render.Abschnitt {
	t.Helper()

	feld := msepaket.Feldstil{
		Feld: "text", Links: "0", Oben: "0", Breite: "200", Hoehe: "60",
		Schrift: msepaket.Schrift{
			Name: "Probe", Groesse: "10", KursivName: "Probe Kursiv",
		},
	}

	plan, befunde := gedeutetMitVorlage(t,
		kartemit("text", text), planvorlage(bildstil(feld), nil))
	if (len(befunde) > 0) != erwarteteBefunde {
		t.Fatalf("Befunde %v, erwartet: %t", befunde, erwarteteBefunde)
	}

	for stelle := range plan.Ebenen {
		if plan.Ebenen[stelle].Art == render.ArtText {
			return plan.Ebenen[stelle].Abschnitte
		}
	}

	t.Fatal("der Plan führt keine Textebene")

	return nil
}
