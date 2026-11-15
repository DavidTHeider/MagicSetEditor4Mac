package service_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Marke ohne Ende gibt keine Stelle.**
//
// `tag_type_at` gibt die leere Zeichenkette, wenn weder `>` noch `-` folgt
// (`src/util/tagged_string.cpp`, Zeilen 224–228, Fassung f9d9356 vom
// 2021-07-06) — und ohne Art ist keine schließende Marke zu suchen.
func TestEineMarkeOhneEnde(t *testing.T) {
	faelle := map[string]string{
		"nichts folgt":         "<word-list-land",
		"kein Schluss im Text": "<word-list-land>Forest",
		"ein fremder Schluss":  "<word-list-land>Forest</sep>",
	}

	for name, roh := range faelle {
		t.Run(name, func(t *testing.T) {
			if stellen := service.Wortlistenstellen(roh); len(stellen) != 0 {
				t.Errorf("gefunden wurde %+v", stellen)
			}
		})
	}
}

// **Die geschützten Zeichen kommen im Vergleichstext zurück.**
//
// `untag` bildet `ESCAPED_LANGLE` auf `<` ab und `CONNECTION_SPACE` auf ein
// Leerzeichen (`src/util/tagged_string.cpp`, Zeilen 15–19, Fassung f9d9356 vom
// 2021-07-06). Ohne das verglichen wir einen Steuercode mit einem Wort.
func TestDerVergleichstextIstEntschuetzt(t *testing.T) {
	stellen := service.Wortlistenstellen("<word-list-a>3 \u0001 4\ueb00x</word-list-a>")

	if len(stellen) != 1 {
		t.Fatalf("%d Stellen", len(stellen))
	}

	if stellen[0].Text != "3 < 4 x" {
		t.Errorf("der Text ist %q", stellen[0].Text)
	}
}

// **Zwei Ebenen tiefer schließt erst die dritte Marke.**
//
// Der Zähler in `match_close_tag` (Zeilen 235–253) geht auf und wieder zu;
// eine Suche, die bei der ersten schließenden aufhörte, endete zu früh.
func TestZweiEbenenVerschachtelt(t *testing.T) {
	roh := "<word-list-a><word-list-b><word-list-c>X</word-list-c></word-list-b></word-list-a>"

	stellen := service.Wortlistenstellen(roh)

	if len(stellen) != 1 || stellen[0].Ende != len(roh) || stellen[0].Text != "X" {
		t.Errorf("gefunden wurde %+v", stellen)
	}
}
