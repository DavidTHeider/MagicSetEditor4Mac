package render_test

import (
	"image"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// echteSchrift lädt eine Schriftdatei vom System.
//
// # Warum keine im Repository liegt
//
// Schriften sind lizenziert (`DRITTANBIETER.md` §5, AGENTS.md §11). Eine
// mitzuliefern hieße, ihre Lizenz zu prüfen und einzuhalten — für die Prüfung
// eines Rasterers ist das der falsche Preis.
//
// Der Preis dieser Wahl steht dafür hier: **die Prüfung hängt an dieser
// Maschine.** Sie überspringt sich, wo die Datei fehlt, und sie vergleicht
// deshalb keine Goldbilder — was sie prüft, sind Aussagen, die für jede Schrift
// gelten: aus einem Zeichen wird Farbe, aus Leerraum nicht, und ein größerer
// Grad deckt mehr Fläche.
func echteSchrift(t *testing.T) *render.OpenTypeSchrift {
	t.Helper()

	const pfad = "/System/Library/Fonts/Supplemental/Andale Mono.ttf"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Schriftdatei unter %s: %v", pfad, err)
	}

	schrift, err := render.SchriftAusDatei(roh)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return schrift
}

// **Eine Textebene mit gefundener Schrift zeichnet Glyphen** — die Zusage aus
// #105.
func TestEineTextebeneWirdGezeichnet(t *testing.T) {
	schrift := echteSchrift(t)

	plan := &render.Plan{Breite: 120, Hoehe: 40, Ebenen: []render.Ebene{
		{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Links: 4, Oben: 4, Breite: 112, Hoehe: 32},
			Abschnitte: render.Aufrecht("Feuerball"),
			Farbe:      render.Farbe{A: 255},
			Satz:       render.Satz{Schrift: "Andale", Groesse: 12, Ausrichtung: render.LinksOben},
			Deckung:    1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("%v", befunde)
	}

	if gedeckt := gedeckteFlaeche(bild); gedeckt == 0 {
		t.Fatal("kein einziger Bildpunkt trägt Farbe — es wurde nichts gezeichnet")
	}
}

// **Leerraum zeichnet nichts.** Eine Prüfung, die nur „irgendetwas ist da"
// verlangt, bestünde auch ein Rasterer, der den Rahmen ausfüllt.
func TestLeerraumZeichnetNichts(t *testing.T) {
	schrift := echteSchrift(t)

	mitText := gedeckteFlaeche(gezeichnet(t, schrift, "Feuerball", 12))
	nurLeerraum := gedeckteFlaeche(gezeichnet(t, schrift, "     ", 12))

	if nurLeerraum != 0 {
		t.Errorf("%d Bildpunkte aus lauter Leerzeichen", nurLeerraum)
	}

	if mitText == 0 {
		t.Error("aus Text wurde nichts")
	}
}

// **Ein größerer Grad deckt mehr Fläche.** Damit hängt das Ergebnis wirklich an
// der Schriftgröße und nicht an einem festen Kasten.
func TestEinGroessererGradDecktMehr(t *testing.T) {
	schrift := echteSchrift(t)

	klein := gedeckteFlaeche(gezeichnet(t, schrift, "Feuerball", 8))
	gross := gedeckteFlaeche(gezeichnet(t, schrift, "Feuerball", 16))

	if gross <= klein {
		t.Errorf("Grad 16 deckt %d Punkte, Grad 8 deckt %d", gross, klein)
	}
}

// **Ein Zeichen, das die Schrift nicht führt, wird benannt** (§8) — und hält
// den Rest nicht auf.
func TestEinFehlendesZeichenWirdBenannt(t *testing.T) {
	schrift := echteSchrift(t)

	plan := &render.Plan{Breite: 120, Hoehe: 40, Ebenen: []render.Ebene{
		{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Breite: 120, Hoehe: 40},
			Abschnitte: render.Aufrecht("A字B"),
			Farbe:      render.Farbe{A: 255},
			Satz:       render.Satz{Schrift: "Andale", Groesse: 12},
			Deckung:    1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "fehlen Zeichen") {
		t.Fatalf("%v", befunde)
	}

	// Und A und B stehen trotzdem da.
	if gedeckteFlaeche(bild) == 0 {
		t.Error("wegen eines fehlenden Zeichens wurde gar nichts gezeichnet")
	}
}

// gezeichnet rastert einen Text und gibt das Bild.
func gezeichnet(t *testing.T, schrift render.Glyphenschrift, text string, grad float64) *image.NRGBA {
	t.Helper()

	plan := &render.Plan{Breite: 200, Hoehe: 60, Ebenen: []render.Ebene{
		{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Breite: 200, Hoehe: 60},
			Abschnitte: render.Aufrecht(text),
			Farbe:      render.Farbe{A: 255},
			Satz:       render.Satz{Schrift: "Andale", Groesse: grad},
			Deckung:    1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("%v", befunde)
	}

	return bild
}

// gedeckteFlaeche zählt die Bildpunkte, die Farbe tragen.
func gedeckteFlaeche(bild *image.NRGBA) int {
	gezaehlt := 0

	for y := bild.Bounds().Min.Y; y < bild.Bounds().Max.Y; y++ {
		for x := bild.Bounds().Min.X; x < bild.Bounds().Max.X; x++ {
			if bild.NRGBAAt(x, y).A > 0 {
				gezaehlt++
			}
		}
	}

	return gezaehlt
}

// **Eine unbrauchbare Schriftdatei wird abgewiesen**, nicht halb gelesen.
func TestEineUnbrauchbareSchriftdateiWirdAbgewiesen(t *testing.T) {
	if _, err := render.SchriftAusDatei([]byte("das ist keine Schrift")); err == nil {
		t.Error("eine Zeichenkette wurde als Schriftdatei angenommen")
	}
}

// **Text, der über den Rand hinausragt, wird beschnitten** — und stürzt nicht
// ab. Der Zweig steht im Zeichner und wäre sonst eine Behauptung.
func TestTextUeberDemRandWirdBeschnitten(t *testing.T) {
	schrift := echteSchrift(t)

	// Der Rahmen liegt am rechten Rand, der Text passt nicht.
	plan := &render.Plan{Breite: 30, Hoehe: 20, Ebenen: []render.Ebene{
		{
			Art:        render.ArtText,
			Rahmen:     render.Rechteck{Links: 20, Oben: 10, Breite: 100, Hoehe: 40},
			Abschnitte: render.Aufrecht("Feuerball wirken"),
			Farbe:      render.Farbe{A: 255},
			Satz:       render.Satz{Schrift: "Andale", Groesse: 14},
			Deckung:    1,
		},
	}}

	bild, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"Andale": schrift}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 0 {
		t.Fatalf("%v", befunde)
	}

	// Etwas steht da — und das Bild hat seine Größe behalten.
	if bild.Bounds().Dx() != 30 || bild.Bounds().Dy() != 20 {
		t.Errorf("%v", bild.Bounds())
	}
}

// Und eine einzelne Zeile heißt „eine Zeile", nicht „1 Zeilen".
func TestEineEinzelneZeileHeisstEineZeile(t *testing.T) {
	plan := &render.Plan{Breite: 400, Hoehe: 20, Ebenen: []render.Ebene{
		{
			Art: render.ArtText, Rahmen: render.Rechteck{Breite: 400, Hoehe: 20},
			Abschnitte: render.Aufrecht("kurz"), Satz: render.Satz{Schrift: "MPlantin", Groesse: 5},
			Deckung: 1,
		},
	}}

	_, befunde, err := render.NeuerRasterer(bilder{}).
		MitSchriften(schriften{"MPlantin": festeBreite{}}).
		Zeichne(t.Context(), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("abgewiesen: %v", err)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "eine Zeile") {
		t.Fatalf("%v", befunde)
	}
}

// **Eine Sammlung wird auf ihre erste Schrift gelesen.**
//
// Auf macOS liegt der halbe Grundbestand als `.ttc`; ohne diesen Zweig fände die
// Suche dort fast nichts.
func TestEineSammlungWirdAufIhreErsteSchriftGelesen(t *testing.T) {
	const pfad = "/System/Library/Fonts/Helvetica.ttc"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Sammlung unter %s: %v", pfad, err)
	}

	schrift, err := render.SchriftAusDatei(roh)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if schrift.Breite('A') <= 0 || schrift.Zeilenhoehe() <= 0 {
		t.Errorf("Breite %g, Zeilenhöhe %g", schrift.Breite('A'), schrift.Zeilenhoehe())
	}
}

// Eine Sammlung, die keine ist, wird abgewiesen.
func TestEineKaputteSammlungWirdAbgewiesen(t *testing.T) {
	if _, err := render.SchriftAusDatei([]byte("ttcf und dann nichts")); err == nil {
		t.Error("eine erfundene Sammlung wurde angenommen")
	}
}

// **Und eine Sammlung, deren erste Schrift sich nicht lesen lässt, wird
// benannt.**
//
// Das ist kein erfundener Fall: `HelveticaNeue.ttc` liegt auf jedem macOS, und
// `x/image` weist sie mit „unsupported cmap encodings" ab. Der Vermerk
// [0008](../../docs/entscheidungen/0008-schriftrasterung.md) nennt genau das als
// die Grenze der gewählten Bibliothek — hier steht sie als Testfall.
func TestEineSammlungMitUnlesbarerErsterSchriftWirdBenannt(t *testing.T) {
	const pfad = "/System/Library/Fonts/HelveticaNeue.ttc"

	roh, err := os.ReadFile(pfad)
	if err != nil {
		t.Skipf("keine Sammlung unter %s: %v", pfad, err)
	}

	gelesen, err := render.SchriftAusDatei(roh)
	if err == nil {
		t.Skipf("diese Fassung von x/image liest %s — dann ist die Grenze gefallen", pfad)
	}

	if gelesen != nil {
		t.Error("es kam eine Schrift und ein Fehler")
	}

	if !strings.Contains(err.Error(), "Sammlung") {
		t.Errorf("die Meldung nennt die Sammlung nicht: %v", err)
	}
}

// **Dieselbe Schrift, viele Zeichner.** Mit `-race` ist das der Wächter über
// den Puffer von `sfnt`, den eine Schrift allein besitzt.
//
// Der erste Anlauf ließ acht Zeichner laufen und meldete ohne Schloss **keinen**
// Wettlauf — acht Goroutinen kollidieren nicht zuverlässig. Deshalb greifen hier
// zusätzlich Breite und Maske unmittelbar ineinander: das ist der Zugriff, an
// dem die laufende Anwendung abgestürzt ist.
func TestEineSchriftHaeltVielenZeichnernStand(t *testing.T) {
	schrift := echteSchrift(t)

	var warten sync.WaitGroup

	for lauf := range 32 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			for _, zeichen := range "Feuerball" {
				if schrift.Breite(zeichen) < 0 {
					t.Errorf("Breite %g", schrift.Breite(zeichen))
				}

				if _, _, da := schrift.Maske(zeichen, float64(8+lauf%8)); !da {
					t.Errorf("kein Glyph für %q", zeichen)
				}
			}
		}()
	}

	for grad := range 8 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			plan := &render.Plan{Breite: 200, Hoehe: 60, Ebenen: []render.Ebene{
				{
					Art:        render.ArtText,
					Rahmen:     render.Rechteck{Breite: 200, Hoehe: 60},
					Abschnitte: render.Aufrecht("Feuerball"),
					Farbe:      render.Farbe{A: 255},
					Satz:       render.Satz{Schrift: "Andale", Groesse: float64(8 + grad)},
					Deckung:    1,
				},
			}}

			if _, _, err := render.NeuerRasterer(bilder{}).
				MitSchriften(schriften{"Andale": schrift}).
				Zeichne(t.Context(), plan, 1, render.GueteMittel); err != nil {
				t.Errorf("zeichnen: %v", err)
			}
		}()
	}

	warten.Wait()
}
