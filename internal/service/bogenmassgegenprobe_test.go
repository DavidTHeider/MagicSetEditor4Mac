package service_test

// **Die Gegenproben der Maßprüfung** — I04
// (#2262),
// Befund des Zweitblicks J05 (#2269).
//
// `bogenmass_test.go` belegt, dass ein richtiger Bogen keinen Befund bringt.
// Hier steht die andere Hälfte: **jede Prüfung greift für sich**. Die erste
// Fassung dieser Probe baute *einen* Bogen mit einer zu schmalen Karte, einer
// verrutschten Marke und fehlenden Marken und verlangte mindestens drei
// Befunde — die fehlenden Marken allein erfüllten das, auch wenn die
// Breitenprüfung gar nichts tat. Gegenprobe der Gegenprobe: mit abgeschalteter
// Breitenprüfung blieb sie grün.
//
// Jetzt geht jeder Fall von **einem gültigen Bogen** aus ([einKartenbogen])
// und ändert genau eine Sache; verlangt wird der eine Befund, der dazugehört —
// und von der Grundlage, dass der unveränderte Bogen **keinen** bringt.
//
// # Was er nicht belegt
//
// **Dass `massbefunde` alles findet, was an einem Bogen falsch sein kann.**
// Er belegt die hier genannten Abweichungen, und jede, die dazukommt, braucht
// ihren eigenen Fall.

import (
	"image"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bogenfall ist, was an einem gültigen Bogen anders ist. Der Nullwert ändert
// nichts.
type bogenfall struct {
	// breite und hoehe sind das Maß der Karte in Millimetern; null heißt das
	// Sollmaß.
	breite, hoehe float64

	// punkteBreit sind die Bildpunkte in der Breite; null heißt die Zahl, die
	// 300 dpi auf 63 mm verlangen.
	punkteBreit int

	// versatz schiebt die untere Marke der linken Schnittkante nach rechts, in
	// Millimetern.
	versatz float64

	// marke ist die Länge einer Marke in Millimetern; null heißt vier.
	marke float64

	// markenAbstand zieht die Marken vom Papierrand weg nach innen.
	markenAbstand float64

	// ohneMarken lässt alle Marken weg.
	ohneMarken bool
}

// einKartenbogen schreibt eine Karte auf eine A4-Seite, so wie der Bogenleger
// den ersten Platz eines 3 × 3-Rasters setzt, und liest das Dokument zurück.
//
// **Die Marken folgen dem Bild, nicht dem Sollmaß**: wer nur die Skalierung
// ändert, soll genau einen Befund über die Skalierung bekommen. Ständen die
// Marken weiter am Soll, käme für jede abweichende Karte ein zweiter Befund
// dazu — und der erste ließe sich nicht mehr von seinem Folgeschaden trennen.
func einKartenbogen(t *testing.T, fall bogenfall) []vermesseneSeite {
	t.Helper()

	masse := service.StandardBogen()
	breite, hoehe := masse.Kartenbreite, masse.Kartenhoehe

	if fall.breite != 0 {
		breite = fall.breite
	}

	if fall.hoehe != 0 {
		hoehe = fall.hoehe
	}

	punkte := fall.punkteBreit
	if punkte == 0 {
		punkte = 744
	}

	marke := fall.marke
	if marke == 0 {
		marke = 4
	}

	// **Platz eins von drei mal drei**: (210 − 3·63)/2 von links, (297 − 3·88)/2
	// von oben. Der Bogenleger rechnet es genauso; weicht diese Rechnung ab,
	// meldet schon die Grundprobe einen Befund.
	links := (masse.Seitenbreite - 3*masse.Kartenbreite) / 2
	oben := (masse.Seitenhoehe - 3*masse.Kartenhoehe) / 2
	unten := masse.Seitenhoehe - oben - masse.Kartenhoehe

	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(pdf.MM(masse.Seitenbreite), pdf.MM(masse.Seitenhoehe))
	seite.Bild(image.NewRGBA(image.Rect(0, 0, punkte, 1039)),
		pdf.MM(links), pdf.MM(unten), pdf.MM(breite), pdf.MM(hoehe))

	if !fall.ohneMarken {
		abstand := fall.markenAbstand

		for nummer, x := range []float64{links, links + breite} {
			versatz := 0.0
			if nummer == 0 {
				versatz = fall.versatz
			}

			seite.Linie(pdf.MM(x+versatz), pdf.MM(abstand), pdf.MM(x+versatz), pdf.MM(abstand+marke), 0.25, 0.4)
			seite.Linie(pdf.MM(x), pdf.MM(masse.Seitenhoehe-marke-abstand), pdf.MM(x),
				pdf.MM(masse.Seitenhoehe-abstand), 0.25, 0.4)
		}

		for _, y := range []float64{unten, unten + hoehe} {
			seite.Linie(pdf.MM(abstand), pdf.MM(y), pdf.MM(abstand+marke), pdf.MM(y), 0.25, 0.4)
			seite.Linie(pdf.MM(masse.Seitenbreite-marke-abstand), pdf.MM(y),
				pdf.MM(masse.Seitenbreite-abstand), pdf.MM(y), 0.25, 0.4)
		}
	}

	seiten, fehler := vermiss(dokument.Bytes())
	if fehler != nil {
		t.Fatalf("vermessen: %v", fehler)
	}

	return seiten
}

// **Die Grundlage**: der gültige Bogen bringt keinen Befund. Ohne sie
// bewiese keiner der Fälle darunter etwas — ein Bogen, der von Haus aus
// Befunde hat, erfüllte jede Forderung nach mindestens einem.
func TestEinGueltigerKartenbogenBringtKeinenBefund(t *testing.T) {
	t.Parallel()

	seiten := einKartenbogen(t, bogenfall{})

	for _, befund := range massbefunde(seiten, service.StandardBogen(), 1) {
		t.Error(befund)
	}
}

// **Jede Abweichung bringt genau ihren Befund** — und keinen anderen.
//
// `erwartet` ist ein Stück des Befundtexts; `anzahl` die Zahl der Befunde
// insgesamt. Zwei Befunde, wo einer verlangt ist, sind ein Folgeschaden, den
// der Fall nicht trennt — und ein Bogen mit drei Fehlern erfüllte jede Schwelle.
func TestJedeAbweichungBringtGenauIhrenBefund(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		fall     bogenfall
		anzahl   int
		erwartet []string
	}{
		// Die Skalierung der Karte, je Achse und Richtung.
		{"Karte um drei Zehntel zu schmal", bogenfall{breite: 62.7}, 1, []string{"Breite samt Anschnitt"}},
		{"Karte um drei Zehntel zu breit", bogenfall{breite: 63.3}, 1, []string{"Breite samt Anschnitt"}},
		{"Karte um drei Zehntel zu niedrig", bogenfall{hoehe: 87.7}, 1, []string{"Höhe samt Anschnitt"}},
		{"Karte um drei Zehntel zu hoch", bogenfall{hoehe: 88.3}, 1, []string{"Höhe samt Anschnitt"}},

		// Die Grenze liegt bei einem Zehntel, nicht weiter und nicht enger.
		{"ein halbes Zehntel Abweichung gilt noch", bogenfall{breite: 63.05}, 0, nil},
		{"anderthalb Zehntel fallen auf", bogenfall{breite: 63.15}, 1, []string{"Breite samt Anschnitt"}},

		// Die Bildpunkte gegen die Auflösung, ohne dass die Karte falsch steht.
		{"zu wenige Bildpunkte für 300 dpi", bogenfall{punkteBreit: 700}, 1, []string{"Breite der Bildpunkte"}},

		// Die Marken, **eigens** und nicht im Verbund mit der Skalierung.
		{
			"eine Marke um drei Zehntel verrutscht",
			bogenfall{versatz: 0.3},
			2,
			[]string{"eine Marke neben jeder Schnittkante", "hat 1 Marken"},
		},
		{"Marken fehlen ganz", bogenfall{ohneMarken: true}, 4, []string{"hat 0 Marken"}},
		// Acht Marken, die nicht gelten, und vier Kanten, die dadurch keine haben.
		{"Marken zu kurz", bogenfall{marke: 3}, 12, []string{"eine Marke von 3.00 mm"}},
		{"Marken nicht am Papierrand", bogenfall{markenAbstand: 5}, 12, []string{"nicht am Rand"}},
	}

	for _, faellchen := range faelle {
		t.Run(faellchen.name, func(t *testing.T) {
			t.Parallel()

			befunde := massbefunde(einKartenbogen(t, faellchen.fall), service.StandardBogen(), 1)

			if len(befunde) != faellchen.anzahl {
				t.Errorf("%d Befunde, erwartet %d: %v", len(befunde), faellchen.anzahl, befunde)
			}

			for _, stueck := range faellchen.erwartet {
				if !strings.Contains(strings.Join(befunde, "\n"), stueck) {
					t.Errorf("kein Befund mit %q: %v", stueck, befunde)
				}
			}
		})
	}
}
