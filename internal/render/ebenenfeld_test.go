package render_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Eine Ebene sagt, aus welchem Feld sie kommt**
// (#246).
//
// # Wofür
//
// Zwei Dinge hängen daran, und beide waren bis hierher nicht möglich:
//
//  1. **Ein Klick in die Vorschau soll das Feld darunter wählen.** Der Plan
//     kennt die Rechtecke; wem sie gehören, wusste er nicht.
//  2. **Ein Befund soll sein Feld nennen.** „Ebene 11: das Bild … ist nicht
//     lesbar" schickt jemanden auf die Suche — bei 40 Ebenen je Karte.
//
// **Leer ist ein Fall, kein Mangel.** Der weiße Grund, den der Planbauer
// selbst hinzufügt, gehört zu keinem Feld; dann steht die Nummer da wie
// bisher.
func TestEinBefundNenntSeinFeld(t *testing.T) {
	t.Parallel()

	mitFeld := render.Befund{Ebene: 11, Feld: "card image", Meldung: "das Bild fehlt"}
	if !strings.Contains(mitFeld.String(), "card image") {
		t.Errorf("%q nennt das Feld nicht", mitFeld.String())
	}

	// **Und die Nummer bleibt dabei.** Sie ist die einzige Angabe, die eine
	// Ebene eindeutig macht: ein Feld kann mehrere Ebenen ergeben.
	if !strings.Contains(mitFeld.String(), "11") {
		t.Errorf("%q nennt die Ebene nicht", mitFeld.String())
	}

	ohneFeld := render.Befund{Ebene: 0, Meldung: "der Grund ist weiß"}
	if ohneFeld.String() != "Ebene 0: der Grund ist weiß" {
		t.Errorf("%q", ohneFeld.String())
	}
}
