package vorschauweg_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// **Ein Fach hält, was es zählt** — E06
// (#2278).
//
// Beide Fächer bekommen die Bytes eines `bytes.Buffer`, in den eben ein PNG
// gepackt wurde — ein Stück eines Arrays, das beim Wachsen verdoppelt wurde
// und bis zu doppelt so groß ist wie das Bild. Das Texturfach zählte `len`
// gegen sein Budget und hielt `cap`: im Zwei-Stunden-Lauf trug
// `App.Spieltextur` mehr als die Hälfte des Zuwachses im Go-Heap
// (`go tool pprof -base` über zwei Abzüge der Leitung).
func TestEinFachHaeltNurDieBytesDesBilds(t *testing.T) {
	puffer := make([]byte, 100_000, 400_000)

	texturen := vorschauweg.NeuesTexturfach(1 << 20)
	texturen.Hinterlege("karte", puffer, 10, 10)

	gehalten, da := texturen.Hole("karte")
	if !da || cap(gehalten) > len(puffer)+len(puffer)/8 {
		t.Errorf("das Texturfach hält %d Bytes für ein Bild von %d", cap(gehalten), len(puffer))
	}

	var vorschauen vorschauweg.Fach

	adresse := vorschauen.Hinterlege(puffer)

	gezeigt, gefunden := vorschauen.Hole(strings.TrimSuffix(strings.TrimPrefix(adresse, vorschauweg.Pfad), ".png"))
	if !gefunden || cap(gezeigt) > len(puffer)+len(puffer)/8 {
		t.Errorf("das Vorschaufach hält %d Bytes für ein Bild von %d", cap(gezeigt), len(puffer))
	}
}
