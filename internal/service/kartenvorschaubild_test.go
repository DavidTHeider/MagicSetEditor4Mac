// **Kleine Bilder für die Kartenliste** — Werkstück 2 von Phase 3.5
// (#1166).
package service_test

import (
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bunteQuelle gibt zu jedem Verweis dasselbe Bild — und merkt sich, wonach
// gefragt wurde.
type bunteQuelle struct {
	gefragt []string
	breite  int
	hoehe   int
	fehler  error
}

func (b *bunteQuelle) Bild(pfad string) (image.Image, error) {
	b.gefragt = append(b.gefragt, pfad)

	if b.fehler != nil {
		return nil, b.fehler
	}

	bild := image.NewNRGBA(image.Rect(0, 0, b.breite, b.hoehe))
	bild.Set(0, 0, color.NRGBA{R: 255, A: 255})

	return bild, nil
}

// **Ein Kartenbild kommt verkleinert zurück — im Seitenverhältnis der Quelle.**
//
// # Warum verkleinert und nicht in Originalgröße
//
// **MSE2 macht es genauso, und die Zahlen stehen dort**: `image.Rescale(36,
// 28)` ohne Kantenglättung, dann `resample(image, 18, 14)` mit
// (`CardThumbnailRequest::generate`, `src/gui/control/image_card_list.cpp`,
// Zeilen 61–70, Fassung `5deaeee` vom 2020-05-16). Eine Kartenliste, die
// dreitausend Illustrationen in voller Auflösung durch die Naht schöbe, wäre
// keine Liste mehr.
//
// **Die Maße hier sind größer** — 18×14 war ein Symbol in einer
// `wxListCtrl` von 2007. Das ist `adapted` und nicht `ported` (§15): die Regel
// „ein verkleinertes Abbild der Illustration" ist übernommen, die Zahl nicht.
func TestEinKartenbildKommtVerkleinert(t *testing.T) {
	t.Parallel()

	quelle := &bunteQuelle{breite: 640, hoehe: 480}

	bild, err := service.NeuesVorschaubildwerk(quelle).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	masse := bild.Bounds().Size()

	if masse.X > service.Vorschaubreite || masse.Y > service.Vorschauhoehe {
		t.Errorf("das Bild ist %v — größer als %d×%d",
			masse, service.Vorschaubreite, service.Vorschauhoehe)
	}

	// **Das Seitenverhältnis bleibt — bis auf einen Bildpunkt.**
	//
	// Es genau zu verlangen ginge nicht: 640:480 auf eine Höhe von 40 gerechnet
	// ergibt 53,33 Punkte Breite, und einen Drittelpunkt gibt es nicht. Ein
	// Test, der Gleichheit fordert, verlangte etwas Unmögliches und würde beim
	// ersten Anlauf »korrigiert«, bis er nichts mehr prüft.
	//
	// **Ein Punkt Abweichung ist die Zusage**, und sie hält, was zählt: ein
	// gestauchtes Bild sagt etwas Falsches über die Illustration.
	if abstand := masse.X*480 - masse.Y*640; abstand > 480 || abstand < -480 {
		t.Errorf("das Verhältnis ist %d:%d — zu weit von 640:480",
			masse.X, masse.Y)
	}

	// **Und die Höhe ist ausgeschöpft.** Ohne diese Zeile wäre auch ein
	// 1×1-Bild im Verhältnis und in den Schranken.
	if masse.Y != service.Vorschauhoehe {
		t.Errorf("die Höhe ist %d statt %d", masse.Y, service.Vorschauhoehe)
	}
}

// **Gefragt wird nach dem Setpaket, denn dort liegt das Bild.**
//
//	auto image_file = getLocalPackage().openIn(value().filename);
//
// `ImageValueViewer::draw`, `src/render/value/image.cpp`, Zeilen 30–33,
// Fassung `f9d9356` vom 2021-07-06.
func TestEinKartenbildWirdImSetpaketGesucht(t *testing.T) {
	t.Parallel()

	quelle := &bunteQuelle{breite: 40, hoehe: 40}

	if _, err := service.NeuesVorschaubildwerk(quelle).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", ""); err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if len(quelle.gefragt) != 1 {
		t.Fatalf("gefragt wurde %d-mal: %v", len(quelle.gefragt), quelle.gefragt)
	}

	if gefragt := quelle.gefragt[0]; gefragt != "setpaket:/wo/das/set.mse-set!kunst.png" {
		t.Errorf("gefragt wurde %q", gefragt)
	}
}

// **Ohne Dateinamen gibt es kein Bild — und keinen Fehler.**
//
//	if (val.filename.empty()) return -1; // no image
//
// `ImageCardList::OnGetItemImage`, `src/gui/control/image_card_list.cpp`,
// Zeile 96, Fassung `5deaeee` vom 2020-05-16. Eine Karte ohne Illustration ist
// der Normalfall und keine Störung.
func TestEineKarteOhneBildGibtNichts(t *testing.T) {
	t.Parallel()

	quelle := &bunteQuelle{breite: 40, hoehe: 40}

	bild, err := service.NeuesVorschaubildwerk(quelle).
		Kartenbild("  ", "/wo/das/set.mse-set", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if bild != nil {
		t.Errorf("es kam ein Bild: %v", bild.Bounds())
	}

	if len(quelle.gefragt) != 0 {
		t.Errorf("die Quelle wurde gefragt: %v", quelle.gefragt)
	}
}

// **Ein Bild aus der Ablage geht dem Herkunftspaket vor.**
//
// Ein Bild, das jemand **in dieser Anwendung** eingesetzt hat, liegt neben dem
// Set ([0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md)).
// Läge das Paket vorn, zeigte die Liste nach dem Einsetzen weiter das alte
// Bild — derselbe Vorrang wie beim Zeichnen, und aus demselben Grund.
func TestEinAbgelegtesBildGehtVor(t *testing.T) {
	t.Parallel()

	ablage := t.TempDir()

	if err := os.WriteFile(filepath.Join(ablage, "kunst.png"), []byte("x"), 0o600); err != nil {
		t.Fatalf("hinlegen: %v", err)
	}

	quelle := &bunteQuelle{breite: 40, hoehe: 40}

	if _, err := service.NeuesVorschaubildwerk(quelle).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", ablage); err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if len(quelle.gefragt) != 1 || quelle.gefragt[0] != "setpaket:"+ablage+"!kunst.png" {
		t.Errorf("gefragt wurde %v — will die Ablage", quelle.gefragt)
	}
}

// **Ohne Setpaket und ohne Ablage gibt es kein Bild.**
//
// Ein Set, das aus keiner Datei stammt, hat keinen Ort, an dem seine Bilder
// lägen. Das ist kein Fehler — die Liste bleibt Text.
func TestOhneOrtGibtEsKeinBild(t *testing.T) {
	t.Parallel()

	quelle := &bunteQuelle{breite: 40, hoehe: 40}

	bild, err := service.NeuesVorschaubildwerk(quelle).Kartenbild("kunst.png", "  ", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if bild != nil || len(quelle.gefragt) != 0 {
		t.Errorf("es kam %v, gefragt wurde %v", bild, quelle.gefragt)
	}
}

// **Ein Lesefehler kommt durch.**
//
// **Und wird nicht verschluckt**: der Aufrufer entscheidet, was er damit tut —
// die Fassade lässt die Zeile ohne Bild stehen, ein Test will ihn sehen.
func TestEinLesefehlerKommtDurch(t *testing.T) {
	t.Parallel()

	gestolpert := errors.New("die Datei ist kaputt")

	_, err := service.NeuesVorschaubildwerk(&bunteQuelle{fehler: gestolpert}).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", "")

	if !errors.Is(err, gestolpert) {
		t.Errorf("der Fehler ist %v", err)
	}
}

// **Ein hochkant liegendes Bild wird an der Höhe begrenzt.**
//
// Die Gegenseite zu [TestEinKartenbildKommtVerkleinert], wo die Breite
// entscheidet. Ohne beide Fälle prüfte die Rechnung nur eine ihrer Richtungen.
func TestEinHochkantbildWirdAnDerHoeheBegrenzt(t *testing.T) {
	t.Parallel()

	bild, err := service.NeuesVorschaubildwerk(&bunteQuelle{breite: 100, hoehe: 400}).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	masse := bild.Bounds().Size()

	if masse.Y != service.Vorschauhoehe || masse.X > service.Vorschaubreite {
		t.Errorf("das Bild ist %v", masse)
	}
}

// **Und ein Bild ohne Fläche gibt keines.**
//
// Eine Datei, die als leeres Bild dekodiert, ist ein fremder Datenfall (§8) —
// er darf nicht in eine Division durch null laufen.
func TestEinBildOhneFlaecheGibtKeines(t *testing.T) {
	t.Parallel()

	bild, err := service.NeuesVorschaubildwerk(&bunteQuelle{}).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if bild != nil {
		t.Errorf("es kam ein Bild: %v", bild.Bounds())
	}
}

// **Ein sehr breites Bild behält mindestens einen Punkt Höhe.**
//
// 4.000 × 1 auf 56 Punkte Breite gerechnet ergäbe null — und ein Bild ohne
// Fläche ist keines. Der Fall ist selten und die Zeile billig; ohne ihn stünde
// ein `max(1, …)` da, das nie etwas tut, und niemand wüsste, ob es stimmt.
func TestEinSehrBreitesBildBehaeltEinenPunkt(t *testing.T) {
	t.Parallel()

	bild, err := service.NeuesVorschaubildwerk(&bunteQuelle{breite: 4000, hoehe: 1}).
		Kartenbild("kunst.png", "/wo/das/set.mse-set", "")
	if err != nil {
		t.Fatalf("Kartenbild: %v", err)
	}

	if masse := bild.Bounds().Size(); masse.Y != 1 || masse.X != service.Vorschaubreite {
		t.Errorf("das Bild ist %v", masse)
	}
}
