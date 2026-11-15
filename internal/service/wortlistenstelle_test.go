package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Wortlistenstellen eines Feldwerts** —
// #1074.
//
// Portiert aus `TextValueEditor::findWordLists`, `src/gui/value/text.cpp`,
// Zeilen 1349–1377, Fassung f9d9356 vom 2021-07-06.
func TestWortlistenstellenWerdenGefunden(t *testing.T) {
	faelle := []struct {
		name   string
		roh    string
		listen []string
		texte  []string
	}{
		{
			name:   "eine Stelle",
			roh:    "<word-list-land>Forest</word-list-land>",
			listen: []string{"land"},
			texte:  []string{"Forest"},
		},
		{
			name: "zwei Stellen mit einem Trenner dazwischen",
			roh: "<word-list-type-en>Creature</word-list-type-en>" +
				"<sep-soft> — </sep-soft>" +
				"<word-list-class-en>Elf</word-list-class-en>",
			listen: []string{"type-en", "class-en"},
			texte:  []string{"Creature", "Elf"},
		},
		{
			name:   "eine leere Stelle",
			roh:    "<word-list-land></word-list-land>",
			listen: []string{"land"},
			texte:  []string{""},
		},
		{
			// **Der Inhalt wird entkleidet.** `current =
			// untag(value.substr(start, end - start))` — `selection()`,
			// Zeile 252. Was zwischen den Marken steht, ist Text.
			name:   "Auszeichnung im Inhalt",
			roh:    "<word-list-land><i>Forest</i></word-list-land>",
			listen: []string{"land"},
			texte:  []string{"Forest"},
		},
		{
			// **Ohne Schluss bricht die Suche ab.** `if (type_end == npos ||
			// end == npos) return;` — Zeile 1359. Nicht „überspringen": MSE
			// hört auf, und was danach käme, ist ebenfalls weg.
			name:   "ohne schließende Marke",
			roh:    "<word-list-land>Forest<word-list-type-en>Creature</word-list-type-en>",
			listen: nil,
			texte:  nil,
		},
		{
			name:   "gar keine Marke",
			roh:    "Forest",
			listen: nil,
			texte:  nil,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			stellen := service.Wortlistenstellen(fall.roh)

			if len(stellen) != len(fall.listen) {
				t.Fatalf("%d Stellen, erwartet %d: %+v", len(stellen), len(fall.listen), stellen)
			}

			for i, stelle := range stellen {
				if stelle.Liste != fall.listen[i] || stelle.Text != fall.texte[i] {
					t.Errorf("Stelle %d ist %+v, erwartet %q/%q",
						i, stelle, fall.listen[i], fall.texte[i])
				}

				if roh := fall.roh[stelle.Anfang:stelle.Ende]; roh == "" {
					t.Errorf("Stelle %d spannt nichts auf: %+v", i, stelle)
				}
			}
		})
	}
}

// **Die Grenzen umfassen die Marken.**
//
// `WordListPos(pos, end, …)` — `pos` steht auf dem `<` der öffnenden Marke,
// `end` hinter dem `>` der schließenden (`match_close_tag_end`,
// `src/util/tagged_string.cpp`, Zeilen 256–258, Fassung f9d9356 vom
// 2021-07-06). Der **Inhalt** liegt dazwischen, und ihn ersetzt eine Wahl.
func TestDieGrenzenEinerStelle(t *testing.T) {
	roh := "vorn<word-list-land>Forest</word-list-land>hinten"

	stellen := service.Wortlistenstellen(roh)
	if len(stellen) != 1 {
		t.Fatalf("%d Stellen", len(stellen))
	}

	if roh[stellen[0].Anfang:stellen[0].Ende] != "<word-list-land>Forest</word-list-land>" {
		t.Errorf("die Marken liegen nicht in den Grenzen: %q",
			roh[stellen[0].Anfang:stellen[0].Ende])
	}

	if roh[stellen[0].Textanfang:stellen[0].Textende] != "Forest" {
		t.Errorf("der Inhalt liegt bei %q",
			roh[stellen[0].Textanfang:stellen[0].Textende])
	}
}

// **Eine gleichnamige Marke darin zählt als Verschachtelung.**
//
// `match_close_tag` zählt Ebenen, und `tag_type_at` schneidet am ersten `>`
// **oder `-`** ab (`src/util/tagged_string.cpp`, Zeilen 224–228 und 235–253,
// Fassung f9d9356 vom 2021-07-06): für `<word-list-land>` ist die Art `word`.
// Ein `<word-list-x>` darin öffnet also eine zweite Ebene, und erst das zweite
// `</word…` schließt die erste.
func TestEineVerschachtelteMarkeZaehlt(t *testing.T) {
	roh := "<word-list-a><word-list-b>X</word-list-b></word-list-a>"

	stellen := service.Wortlistenstellen(roh)

	if len(stellen) != 1 {
		t.Fatalf("%d Stellen: %+v", len(stellen), stellen)
	}

	if stellen[0].Liste != "a" || stellen[0].Text != "X" {
		t.Errorf("die Stelle ist %+v", stellen[0])
	}

	// **Und die innere wird nicht noch einmal gefunden**: die Suche geht bei
	// `end` weiter (Zeile 1376), nicht bei `pos + 1`.
	if stellen[0].Ende != len(roh) {
		t.Errorf("die Stelle endet bei %d statt %d", stellen[0].Ende, len(roh))
	}
}
