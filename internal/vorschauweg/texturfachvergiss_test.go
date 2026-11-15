package vorschauweg_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Vergessen gibt den Platz frei** — E06 (#2278): ein Bild, auf das keine
// Karte mehr zeigt, wirft der Dienst weg, statt es bis zum Budget liegen zu
// lassen.
func TestVergessenGibtDenPlatzFrei(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1 << 20)
	fach.Hinterlege("a", make([]byte, 100), 1, 1)
	fach.Hinterlege("b", make([]byte, 50), 1, 1)

	fach.Vergiss("a")

	if _, _, _, da := fach.Adresse("a"); da {
		t.Error("a liegt nach dem Vergessen noch da")
	}

	if fach.Belegt() != 50 {
		t.Errorf("belegt %d, erwartet 50", fach.Belegt())
	}

	// **Was es nicht gibt, zu vergessen, ändert nichts.**
	fach.Vergiss("gibt-es-nicht")

	if _, _, _, da := fach.Adresse("b"); !da || fach.Belegt() != 50 {
		t.Errorf("b fehlt oder belegt %d", fach.Belegt())
	}
}
