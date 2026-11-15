package vorschauweg_test

// **Beide Fächer zählen** — OPS-08.AK-01
// (#1725):
// das der Vorschauen, wie oft ein Bild abgeholt wurde und wie oft es schon
// fort war, das der Spielbilder, wie oft eine Adresse schon dastand.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

func TestDasVorschaufachZaehltUndLeert(t *testing.T) {
	t.Parallel()

	var fach vorschauweg.Fach

	adresse := fach.Hinterlege([]byte("ein Bild"))
	marke := strings.TrimSuffix(strings.TrimPrefix(adresse, vorschauweg.Pfad), ".png")

	fach.Hole(marke)
	fach.Hole("gibtsnicht")

	zahlen := fach.Zahlen()
	if zahlen.Treffer != 1 || zahlen.Fehlgriffe != 1 || zahlen.Eintraege != 1 ||
		zahlen.Bytes != int64(len("ein Bild")) || zahlen.GrenzeEintraege != vorschauweg.Faecher {
		t.Errorf("%+v", zahlen)
	}

	fach.Leere()

	if _, da := fach.Hole(marke); da || fach.Zahlen().Eintraege != 0 {
		t.Errorf("nach dem Leeren: %+v", fach.Zahlen())
	}
}

func TestDasTexturfachZaehlt(t *testing.T) {
	t.Parallel()

	fach := vorschauweg.NeuesTexturfach(1 << 20)
	fach.Adresse("a")
	fach.Hinterlege("a", []byte("abcd"), 2, 2)
	fach.Adresse("a")

	zahlen := fach.Zahlen()
	if zahlen.Treffer != 1 || zahlen.Fehlgriffe != 1 || zahlen.Eintraege != 1 || zahlen.Bytes != 4 ||
		zahlen.GrenzeBytes != 1<<20 {
		t.Errorf("%+v", zahlen)
	}
}
