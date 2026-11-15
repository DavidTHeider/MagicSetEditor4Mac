package render_test

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// saatplan ist eine Karte aus **eigener** Saat.
//
// **Nichts Fremdes** (`docs/entscheidungen/0005`): die Farben, Maße und Namen
// hier sind erfunden. Ein Goldbild aus `magic-m15` wäre ein Bild einer fremden
// Vorlage im Repository, und das ist genau das, was 0005 ausschließt.
func saatplan() *render.Plan {
	return &render.Plan{
		Breite: 40, Hoehe: 56,
		Ebenen: []render.Ebene{
			{
				Art: render.ArtFlaeche, Rahmen: render.Rechteck{Breite: 40, Hoehe: 56},
				Farbe: render.Farbe{R: 32, G: 40, B: 48, A: 255}, Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: 3, Oben: 3, Breite: 34, Hoehe: 20},
				Verlauf: &render.Verlauf{Winkel: 90, Stufen: []render.Verlaufsstufe{
					{Stelle: 0, Farbe: render.Farbe{R: 200, G: 170, B: 90, A: 255}},
					{Stelle: 1, Farbe: render.Farbe{R: 90, G: 70, B: 30, A: 255}},
				}},
				Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: 3, Oben: 26, Breite: 34, Hoehe: 27},
				Farbe:  render.Farbe{R: 240, G: 238, B: 230, A: 255}, Deckung: 1,
			},
			{
				Art:    render.ArtFlaeche,
				Rahmen: render.Rechteck{Links: 6, Oben: 30, Breite: 28, Hoehe: 10},
				Farbe:  render.Farbe{R: 60, G: 120, B: 200, A: 128}, Deckung: 1,
			},
		},
	}
}

func gerastert(t *testing.T) image.Image {
	t.Helper()

	bild, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), saatplan(), 2, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	return bild
}

func goldbildpfad(t *testing.T) string {
	t.Helper()

	wurzel, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	return filepath.Join(wurzel, "testdaten", "goldbilder", "saatkarte.png")
}

// Das Goldbild hält den Rasterlauf fest.
//
// Fehlt es, wird es angelegt — beim ersten Lauf und nach einer bewussten
// Änderung, die jemand mit `rm` freigibt. **Das Anlegen macht den Lauf nicht
// grün**: der Test meldet es, damit niemand ein Goldbild aus einem kaputten
// Lauf erbt, ohne es gesehen zu haben.
func TestDerRasterlaufBleibtWieErIst(t *testing.T) {
	pfad := goldbildpfad(t)
	neu := gerastert(t)

	datei, err := os.Open(pfad) //nolint:gosec // fester Pfad im Repository
	if errors.Is(err, os.ErrNotExist) {
		schreibe(t, pfad, neu)
		t.Fatalf("das Goldbild fehlte und wurde angelegt: %s — ansehen und den Lauf wiederholen", pfad)
	}

	if err != nil {
		t.Fatalf("Goldbild nicht lesbar: %v", err)
	}

	defer func() {
		if fehler := datei.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	gold, err := png.Decode(datei)
	if err != nil {
		t.Fatalf("Goldbild nicht lesbar: %v", err)
	}

	vergleich, err := render.Vergleiche(gold, neu)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !vergleich.Innerhalb() {
		ablegen(t, neu, gold)
		t.Errorf("%s — über der Toleranz von %g. Differenzbild in testdaten/abweichungen/",
			vergleich, render.Toleranz)
	}
}

// **Die Gegenprobe**, und sie ist die Anforderung, die #35 zuletzt nennt.
//
// Eine Verschiebung um zwei Bildpunkte macht den Lauf rot. Ohne diesen Fall
// wäre die Toleranz eine Zahl ohne Bedeutung: eine, die groß genug ist, damit
// nie etwas rot wird, sähe genauso aus wie eine, die greift.
func TestEineVerschiebungUmZweiBildpunkteMachtDenLaufRot(t *testing.T) {
	gerade := gerastert(t)

	verschoben := saatplan()
	for i := range verschoben.Ebenen {
		verschoben.Ebenen[i].Rahmen.Links += 1
		verschoben.Ebenen[i].Rahmen.Oben += 1
	}

	// Bei Vergrößerung 2 sind das zwei Bildpunkte.
	anders, befunde, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), verschoben, 2, render.GueteMittel)
	if err != nil || len(befunde) != 0 {
		t.Fatalf("%v / %v", befunde, err)
	}

	vergleich, err := render.Vergleiche(gerade, anders)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if vergleich.Innerhalb() {
		t.Errorf("%s — eine Verschiebung um zwei Bildpunkte blieb innerhalb der Toleranz. "+
			"Dann prüft der Goldbildvergleich nichts", vergleich)
	}
}

// Eine Änderung, die niemand sehen kann, bleibt grün.
//
// Die Gegenprobe nach unten: eine Toleranz, die auf jede Rundung anspringt,
// wird nach dem dritten Fehlalarm abgeschaltet.
func TestEineAenderungUmEinenSchrittBleibtInnerhalb(t *testing.T) {
	gerade := gerastert(t)

	rahmen := gerade.Bounds()
	fast := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			rot, gruen, blau, alpha := gerade.At(x, y).RGBA()
			fast.SetNRGBA(x, y, color.NRGBA{
				R: uint8(rot>>8) + 1, G: uint8(gruen >> 8), //nolint:gosec // 0…255
				B: uint8(blau >> 8), A: uint8(alpha >> 8), //nolint:gosec // 0…255
			})
		}
	}

	vergleich, err := render.Vergleiche(gerade, fast)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !vergleich.Innerhalb() {
		t.Errorf("%s — ein Schritt je Bildpunkt schlug schon an", vergleich)
	}
}

func TestVergleicheWeistUnvergleichbaresAb(t *testing.T) {
	klein := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	gross := image.NewNRGBA(image.Rect(0, 0, 3, 3))

	if _, err := render.Vergleiche(klein, gross); !errors.Is(err, render.ErrPlan) {
		t.Errorf("Fehler %v, erwartet war ErrPlan", err)
	}

	if _, err := render.Vergleiche(nil, gross); !errors.Is(err, render.ErrPlan) {
		t.Errorf("Fehler %v, erwartet war ErrPlan", err)
	}
}

// Das Differenzbild zeigt, **wo** es auseinandergeht.
func TestDasDifferenzbildZeigtDieStelle(t *testing.T) {
	links := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	links.SetNRGBA(0, 0, color.NRGBA{R: 100, A: 255})
	links.SetNRGBA(1, 0, color.NRGBA{R: 100, A: 255})

	rechts := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	rechts.SetNRGBA(0, 0, color.NRGBA{R: 100, A: 255})
	rechts.SetNRGBA(1, 0, color.NRGBA{R: 160, A: 255})

	differenz := render.Differenzbild(links, rechts)

	if got := farbeBei(t, differenz, 0, 0); got.R != 0 {
		t.Errorf("gleiche Stelle %s — sie müsste dunkel sein", got)
	}

	// 60 Schritte Unterschied, achtfach verstärkt und bei 255 begrenzt.
	if got := farbeBei(t, differenz, 1, 0); got.R != 255 {
		t.Errorf("verschiedene Stelle %s — sie müsste hell sein", got)
	}
}

func TestDieDarstellungEinesVergleichs(t *testing.T) {
	mitPunkten := render.Vergleich{Abstand: 1.5, Groesster: 60, Abweichende: 12, Punkte: 100}
	if !strings.Contains(mitPunkten.String(), "0.1200") || mitPunkten.Innerhalb() {
		t.Errorf("%s / innerhalb: %v", mitPunkten, mitPunkten.Innerhalb())
	}

	// Ein Vergleich ohne Punkte kommt aus keiner Messung — die Darstellung darf
	// daran trotzdem nicht durch null teilen.
	ohnePunkte := render.Vergleich{Abstand: 0.5, Groesster: 1}
	if !strings.Contains(ohnePunkte.String(), "0.0000") {
		t.Errorf("%s", ohnePunkte)
	}
}

// ---------------------------------------------------------------- Vorrichtung

func schreibe(t *testing.T, pfad string, bild image.Image) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
		t.Fatalf("Verzeichnis: %v", err)
	}

	datei, err := os.Create(pfad) //nolint:gosec // fester Pfad im Repository
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := png.Encode(datei, bild); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}
}

// ablegen legt Ist, Soll und Differenz nebeneinander.
//
// **Alle drei**, nicht nur die Differenz: wer sie ansieht, will wissen, was
// herauskommen sollte und was herausgekommen ist — die Differenz allein sagt
// nur, dass sie sich unterscheiden.
func ablegen(t *testing.T, neu, gold image.Image) {
	t.Helper()

	wurzel, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	ordner := filepath.Join(wurzel, "testdaten", "abweichungen")

	schreibe(t, filepath.Join(ordner, "saatkarte-ist.png"), neu)
	schreibe(t, filepath.Join(ordner, "saatkarte-soll.png"), gold)
	schreibe(t, filepath.Join(ordner, "saatkarte-differenz.png"), render.Differenzbild(gold, neu))
}

// **Die Schwelle greift auch bei Druckauflösung.**
//
// Der Mittelwert allein verdünnt sich mit der Bildfläche: eine um zwei
// Bildpunkte verschobene Kante reißt ihn in einer kleinen Saat über 1 und liegt
// in einem großen Bild rechnerisch weit darunter. Die erste Gegenprobe belegte
// damit nur die Größe **ihres** Testbilds — angestrichen vom Zweitblick über
// Phase 4.
//
// Deshalb entscheidet zusätzlich die **Stärke** der größten Abweichung, und die
// hat mit der Bildgröße nichts zu tun.
func TestDieSchwelleGreiftAuchBeiDruckaufloesung(t *testing.T) {
	for _, vergroesserung := range []float64{2, 8, 20} {
		gerade, _, err := render.NeuerRasterer(bilder{}).
			Zeichne(t.Context(), saatplan(), vergroesserung, render.GueteMittel)
		if err != nil {
			t.Fatalf("Vergrößerung %g: %v", vergroesserung, err)
		}

		verschoben := saatplan()
		for i := range verschoben.Ebenen {
			verschoben.Ebenen[i].Rahmen.Links += 1
		}

		anders, _, err := render.NeuerRasterer(bilder{}).
			Zeichne(t.Context(), verschoben, vergroesserung, render.GueteMittel)
		if err != nil {
			t.Fatalf("Vergrößerung %g: %v", vergroesserung, err)
		}

		vergleich, err := render.Vergleiche(gerade, anders)
		if err != nil {
			t.Fatalf("vergleichen: %v", err)
		}

		if vergleich.Innerhalb() {
			t.Errorf("bei Vergrößerung %g blieb eine Verschiebung innerhalb: %s",
				vergroesserung, vergleich)
		}
	}
}

// Und eine Rundung um einen Schritt bleibt auch bei Druckauflösung grün.
func TestEineRundungBleibtAuchBeiDruckaufloesungInnerhalb(t *testing.T) {
	gerade, _, err := render.NeuerRasterer(bilder{}).Zeichne(t.Context(), saatplan(), 8, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	rahmen := gerade.Bounds()
	fast := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			rot, gruen, blau, alpha := gerade.At(x, y).RGBA()
			fast.SetNRGBA(x, y, color.NRGBA{
				R: uint8(rot>>8) + 1, G: uint8(gruen >> 8), //nolint:gosec // 0…255
				B: uint8(blau >> 8), A: uint8(alpha >> 8), //nolint:gosec // 0…255
			})
		}
	}

	vergleich, err := render.Vergleiche(gerade, fast)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	if !vergleich.Innerhalb() {
		t.Errorf("%s — eine Rundung schlug bei Druckauflösung an", vergleich)
	}
}
