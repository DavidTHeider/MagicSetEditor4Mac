package service_test

import (
	"context"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// planDouble gibt einen Plan, der ohne Schriften auskommt.
//
// **Eine Fläche und sonst nichts.** Der Behelfsplan trägt Textebenen, und die
// brauchen eine Schriftquelle; hier geht es um den Weg vom Bild zur Datei, und
// ein Test, der dabei über eine fehlende Schrift stolpert, prüft die falsche
// Sache.
type planDouble struct {
	breite, hoehe float64
	fehler        error
}

func (p planDouble) Kartenplan(context.Context, *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	if p.fehler != nil {
		return nil, nil, p.fehler
	}

	breite, hoehe := p.breite, p.hoehe
	if breite == 0 {
		breite, hoehe = 100, 140
	}

	return &render.Plan{
		Breite: breite, Hoehe: hoehe,
		Ebenen: []render.Ebene{{
			Art:     render.ArtFlaeche,
			Rahmen:  render.Rechteck{Breite: breite, Hoehe: hoehe / 2},
			Farbe:   render.Farbe{R: 200, G: 40, B: 40, A: 255},
			Deckung: 1,
		}},
	}, nil, nil
}

// kartenDouble gibt eine einzelne Karte.
type kartenDouble struct {
	leserDouble

	fehlt bool
}

func (k kartenDouble) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	if k.fehlt {
		return nil, domain.ErrNotFound
	}

	for _, karte := range k.karten {
		if karte.ID == karteID {
			return karte, nil
		}
	}

	return nil, domain.ErrNotFound
}

func bildausfuhr(t *testing.T, plan planDouble) *service.Bildausfuhr {
	t.Helper()

	return service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball")}, plan, render.NeuerRasterer(nil))
}

// **Ein Kartenbild kommt als PNG heraus — und zwar als eines.**
//
// Geprüft wird die Datei, nicht der Aufruf: ein Test, der nur zusieht, dass ein
// Kodierer gerufen wurde, bliebe grün, wenn dabei null Bytes entstehen.
func TestEinKartenbildWirdAlsPNGGeschrieben(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "001-Feuerball.png")

	ergebnis, err := bildausfuhr(t, planDouble{}).
		Schreibe(t.Context(), 1, ort, "001-Feuerball.png", service.FormatPNG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	datei, err := os.Open(ziel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	bild, err := png.Decode(datei)
	if err != nil {
		t.Fatalf("das ist kein PNG: %v", err)
	}

	if bild.Bounds().Dx() != 100 || bild.Bounds().Dy() != 140 {
		t.Errorf("%v, erwartet 100×140", bild.Bounds())
	}

	if ergebnis.Breite != 100 || ergebnis.Hoehe != 140 {
		t.Errorf("gemeldet %d×%d", ergebnis.Breite, ergebnis.Hoehe)
	}
}

// **Die Zielbreite bestimmt die Auflösung, nicht ein Faktor.**
//
// „750 Bildpunkte breit" ist etwas, das jemand nachrechnen kann; „viermal" ist
// es nur, wenn man die Größe des Plans kennt — und die steht in einer Vorlage,
// die niemand aufmacht.
func TestDieZielbreiteBestimmtDieGroesse(t *testing.T) {
	ergebnis, err := bildausfuhr(t, planDouble{}).
		Schreibe(t.Context(), 1, t.TempDir(), "gross.png", service.FormatPNG, 750, domain.BildkonfliktUeberschreiben,
			render.GueteMittel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// 750 / 100 = 7,5 — und die Höhe folgt im selben Verhältnis.
	if ergebnis.Breite != 750 || ergebnis.Hoehe != 1050 {
		t.Errorf("%d×%d, erwartet 750×1050", ergebnis.Breite, ergebnis.Hoehe)
	}
}

// **JPG kennt keine Durchsichtigkeit — also wird sie auf Weiß gelegt.**
//
// Wer sie stehen ließe, bekäme dort, wo die Karte durchsichtig ist, Schwarz:
// `image/jpeg` schreibt die Farbwerte, ohne den Alphakanal zu berücksichtigen,
// und ein unbeschriebenes NRGBA ist durchsichtiges Schwarz.
func TestEinKartenbildWirdAlsJPGGeschrieben(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "001-Feuerball.jpg")

	if _, err := bildausfuhr(t, planDouble{}).
		Schreibe(t.Context(), 1, ort, "001-Feuerball.jpg", service.FormatJPG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	datei, err := os.Open(ziel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	bild, err := jpeg.Decode(datei)
	if err != nil {
		t.Fatalf("das ist kein JPG: %v", err)
	}

	// Die untere Hälfte trägt keine Ebene — dort ist der Plan durchsichtig.
	unten := bild.At(50, 120)
	r, g, b, _ := unten.RGBA()

	if r < 0xf000 || g < 0xf000 || b < 0xf000 {
		t.Errorf("die durchsichtige Hälfte ist %v — erwartet Weiß", unten)
	}
}

// Ein Verzeichnis, das es noch nicht gibt, entsteht.
func TestDasZielverzeichnisEntsteht(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bilder", "tiefer")
	ziel := filepath.Join(ort, "001.png")

	if _, err := bildausfuhr(t, planDouble{}).
		Schreibe(t.Context(), 1, ort, "001.png", service.FormatPNG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}
}

// **Ein Befund des Rasterers hält das Schreiben nicht auf** (§8).
//
// Eine Karte mit elf von zwölf Ebenen und einer Meldung ist brauchbar; eine
// leere Karte ohne Meldung ist es nicht.
func TestEinBefundHaeltDasSchreibenNichtAuf(t *testing.T) {
	// Ohne Schriftquelle meldet der Rasterer je Textebene einen Befund.
	ausfuhr := service.NeueBildausfuhr(
		kartenDouble{leserDouble: karten("Feuerball")},
		service.NeuerPlanbauer(), render.NeuerRasterer(nil))

	ort := t.TempDir()
	ziel := filepath.Join(ort, "mitbefund.png")

	ergebnis, err := ausfuhr.Schreibe(t.Context(), 1, ort, "mitbefund.png", service.FormatPNG, 0, domain.BildkonfliktUeberschreiben, render.GueteMittel)
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if len(ergebnis.Befunde) == 0 {
		t.Error("kein Befund — dass der Text fehlt, bliebe unsichtbar")
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt trotzdem: %v", err)
	}
}
