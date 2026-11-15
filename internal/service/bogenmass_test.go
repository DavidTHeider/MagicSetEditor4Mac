package service_test

// **Ein Druckbogen hält seine Maße auf ein Zehntel Millimeter** — I04
// (#2262).
//
// Gemessen wird am geschriebenen PDF ([vermiss]) und gegen die **Zusage** des
// Bogens, nicht gegen seine Rechnung: Seite so groß wie das Papier, jede Karte
// samt Anschnitt so groß wie bestellt, Nachbarn ohne Lücke, gleiche Ränder,
// jede Schnittmarke auf einer Schnittkante und so viele Bildpunkte, wie die
// Auflösung auf dieser Breite verlangt. **Ein Zehntel Millimeter** ist die
// Grenze; darüber ist es ein Befund mit seiner Abweichung.
//
// # Was er nicht belegt
//
// **Einen Ausdruck.** Was ein Drucker aus 63,0 mm macht, misst nur ein Lineal
// auf Papier — das bleibt Handprüfung, und hier steht nichts, was sie ersetzt.

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// zehntel ist die Grenze aus I04.
const zehntel = 0.1

// massbefunde prüft vermessene Seiten gegen die Maße eines Bogens.
func massbefunde(seiten []vermesseneSeite, masse service.Bogenmasse, karten int) []string {
	var befunde []string

	daneben := func(was string, ist, soll float64) {
		if math.Abs(ist-soll) > zehntel {
			befunde = append(befunde, fmt.Sprintf("%s: %.3f mm statt %.3f mm (%+.3f)", was, ist, soll, ist-soll))
		}
	}

	zellbreit := masse.Kartenbreite + 2*masse.Anschnitt
	zellhoch := masse.Kartenhoehe + 2*masse.Anschnitt
	jeSeite := int((masse.Seitenbreite-2*masse.Rand)/zellbreit) * int((masse.Seitenhoehe-2*masse.Rand)/zellhoch)

	if soll := (karten + jeSeite - 1) / jeSeite; len(seiten) != soll {
		befunde = append(befunde, fmt.Sprintf("%d Seiten statt %d", len(seiten), soll))
	}

	gezaehlt := 0

	for nummer, seite := range seiten {
		ort := "Seite " + strconv.Itoa(nummer+1)

		daneben(ort+", Breite", seite.breite, masse.Seitenbreite)
		daneben(ort+", Höhe", seite.hoehe, masse.Seitenhoehe)

		for platz, bild := range seite.bilder {
			karte := fmt.Sprintf("%s, Karte %d", ort, platz+1)

			daneben(karte+", Breite samt Anschnitt", bild.breite, zellbreit)
			daneben(karte+", Höhe samt Anschnitt", bild.hoehe, zellhoch)
			daneben(karte+", Breite der Bildpunkte", float64(bild.punkteBreit)*25.4/float64(masse.DPI), zellbreit)
			daneben(karte+", Höhe der Bildpunkte", float64(bild.punkteHoch)*25.4/float64(masse.DPI), zellhoch)

			if bild.links < masse.Rand-zehntel || bild.unten < masse.Rand-zehntel ||
				bild.links+bild.breite > masse.Seitenbreite-masse.Rand+zehntel ||
				bild.unten+bild.hoehe > masse.Seitenhoehe-masse.Rand+zehntel {
				befunde = append(befunde, karte+" liegt im Rand")
			}
		}

		gezaehlt += len(seite.bilder)
		befunde = append(befunde, rasterbefunde(ort, seite, masse, daneben)...)
		befunde = append(befunde, markenbefunde(ort, seite, masse)...)
	}

	if gezaehlt != karten {
		befunde = append(befunde, fmt.Sprintf("%d Karten statt %d", gezaehlt, karten))
	}

	return befunde
}

// rasterbefunde prüft Nachbarn und Ränder: Karten stoßen aneinander, und der
// Rest der Seite liegt zu gleichen Teilen links und rechts, oben und unten.
func rasterbefunde(ort string, seite vermesseneSeite, masse service.Bogenmasse, daneben func(string, float64, float64)) []string {
	if len(seite.bilder) == 0 {
		return []string{ort + " ohne Karte"}
	}

	links, rechts := math.Inf(1), math.Inf(-1)
	unten, oben := math.Inf(1), math.Inf(-1)

	for platz, bild := range seite.bilder {
		links, rechts = min(links, bild.links), max(rechts, bild.links+bild.breite)
		unten, oben = min(unten, bild.unten), max(oben, bild.unten+bild.hoehe)

		if platz > 0 && math.Abs(bild.unten-seite.bilder[platz-1].unten) <= zehntel {
			vorher := seite.bilder[platz-1]
			daneben(fmt.Sprintf("%s, Lücke vor Karte %d", ort, platz+1), bild.links-(vorher.links+vorher.breite), 0)
		}
	}

	// **Die Ränder nur auf einer vollen Seite**: eine halbe ist oben bündig,
	// und das ist gewollt.
	zellbreit := masse.Kartenbreite + 2*masse.Anschnitt
	zellhoch := masse.Kartenhoehe + 2*masse.Anschnitt
	spalten := int((masse.Seitenbreite - 2*masse.Rand) / zellbreit)
	zeilen := int((masse.Seitenhoehe - 2*masse.Rand) / zellhoch)

	if len(seite.bilder) == spalten*zeilen {
		daneben(ort+", linker gegen rechten Rand", links, seite.breite-rechts)
		daneben(ort+", unterer gegen oberen Rand", unten, seite.hoehe-oben)
	}

	return nil
}

// markenbefunde prüft, dass jede Schnittkante ihre Marken hat und jede Marke
// auf einer Schnittkante steht, am Papierrand und vier Millimeter lang.
//
// **Mindestens zwei je Kante und nicht genau zwei**: ohne Anschnitt teilen sich
// zwei Nachbarn ihre Kante, und der Bogen zieht ihre Marken für jede Spalte —
// eine doppelt gezogene Linie liegt auf derselben Stelle und misst dasselbe.
func markenbefunde(ort string, seite vermesseneSeite, masse service.Bogenmasse) []string {
	if !masse.Schnittmarken {
		if len(seite.linien) > 0 {
			return []string{fmt.Sprintf("%s: %d Linien ohne Schnittmarken", ort, len(seite.linien))}
		}

		return nil
	}

	var senkrecht, waagerecht []schnittkante

	for _, bild := range seite.bilder {
		senkrecht = mitKante(senkrecht, bild.links+masse.Anschnitt)
		senkrecht = mitKante(senkrecht, bild.links+bild.breite-masse.Anschnitt)
		waagerecht = mitKante(waagerecht, bild.unten+masse.Anschnitt)
		waagerecht = mitKante(waagerecht, bild.unten+bild.hoehe-masse.Anschnitt)
	}

	var befunde []string

	for _, linie := range seite.linien {
		laenge := math.Hypot(linie.x2-linie.x1, linie.y2-linie.y1)
		amRand := math.Min(math.Min(linie.y1, linie.y2), math.Min(linie.x1, linie.x2)) <= zehntel ||
			math.Max(linie.y1, linie.y2) >= seite.hoehe-zehntel || math.Max(linie.x1, linie.x2) >= seite.breite-zehntel

		switch {
		case math.Abs(laenge-4) > zehntel || !amRand:
			befunde = append(befunde, fmt.Sprintf("%s: eine Marke von %.2f mm, nicht am Rand: %+v", ort, laenge, linie))
		case linie.x1 == linie.x2 && zaehle(senkrecht, linie.x1):
		case linie.y1 == linie.y2 && zaehle(waagerecht, linie.y1):
		default:
			befunde = append(befunde, fmt.Sprintf("%s: eine Marke neben jeder Schnittkante: %+v", ort, linie))
		}
	}

	for _, kante := range append(senkrecht, waagerecht...) {
		if kante.marken < 2 {
			befunde = append(befunde, fmt.Sprintf("%s: die Kante bei %.2f mm hat %d Marken", ort, kante.stelle, kante.marken))
		}
	}

	return befunde
}

// schnittkante ist eine Kante mit den Marken, die auf ihr stehen.
type schnittkante struct {
	stelle float64
	marken int
}

// mitKante nimmt eine Kante auf — eine, die schon dasteht, nicht noch einmal.
func mitKante(kanten []schnittkante, stelle float64) []schnittkante {
	for _, kante := range kanten {
		if math.Abs(kante.stelle-stelle) <= zehntel {
			return kanten
		}
	}

	return append(kanten, schnittkante{stelle: stelle})
}

// zaehle schreibt eine Marke ihrer Kante gut und sagt, ob es eine gab.
func zaehle(kanten []schnittkante, stelle float64) bool {
	for nummer := range kanten {
		if math.Abs(kanten[nummer].stelle-stelle) <= zehntel {
			kanten[nummer].marken++

			return true
		}
	}

	return false
}

// bogenVermessen legt einen Bogen mit einer Vorlage im Verhältnis der Karte.
func bogenVermessen(t *testing.T, masse service.Bogenmasse, anzahl int) []vermesseneSeite {
	t.Helper()

	namen := make([]string, 0, anzahl)
	for nummer := range anzahl {
		namen = append(namen, "Karte "+strconv.Itoa(nummer+1))
	}

	leger := service.NeuerBogenleger(kartenDouble{leserDouble: karten(namen...)},
		planDouble{breite: masse.Kartenbreite * 10, hoehe: masse.Kartenhoehe * 10},
		render.NeuerRasterer(nil))
	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	if _, fehler := leger.Lege(t.Context(), 7, ziel, masse, domain.Kartenauswahl{}); fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	roh, fehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	seiten, fehler := vermiss(roh)
	if fehler != nil {
		t.Fatalf("vermessen: %v", fehler)
	}

	return seiten
}

// **Die Gegenproben stehen in `bogenmassgegenprobe_test.go`.** Hier stand eine
// einzige, die zugleich eine zu schmale Karte, eine verrutschte Marke und
// fehlende Marken prüfte und mindestens drei Befunde verlangte; die fehlenden
// Marken allein erfüllten das, auch mit abgeschalteter Breitenprüfung (J05,
// #2269).

// **Der Bogen hält seine Maße** — vier Papierlagen, drei Kartenkohorten, mit
// und ohne Anschnitt, je eine volle Seite und eine angefangene.
func TestDerBogenHaeltSeineMasseAufEinZehntel(t *testing.T) {
	t.Parallel()

	papiere := []struct {
		name          string
		breite, hoehe float64
	}{
		{"A4 hoch", 210, 297},
		{"A4 quer", 297, 210},
		{"Letter hoch", 215.9, 279.4},
		{"Letter quer", 279.4, 215.9},
	}

	kohorten := []struct {
		name                     string
		breite, hoehe, anschnitt float64
	}{
		{"Standard 63 × 88", 63, 88, 0},
		{"Standard mit 3 mm Anschnitt", 63, 88, 3},
		{"Mini 41 × 63", 41, 63, 0},
		{"Tarot 70 × 120", 70, 120, 0},
	}

	for _, papier := range papiere {
		for _, kohorte := range kohorten {
			t.Run(papier.name+", "+kohorte.name, func(t *testing.T) {
				t.Parallel()

				masse := service.StandardBogen()
				masse.Seitenbreite, masse.Seitenhoehe = papier.breite, papier.hoehe
				masse.Kartenbreite, masse.Kartenhoehe = kohorte.breite, kohorte.hoehe
				masse.Anschnitt = kohorte.anschnitt

				// **300 dpi, die Vorgabe des Bogens**: bei 150 ist schon ein
				// halber Bildpunkt 0,085 mm, und ein Zehntel ließe sich nicht
				// mehr von der Rundung trennen.

				zellbreit := masse.Kartenbreite + 2*masse.Anschnitt
				zellhoch := masse.Kartenhoehe + 2*masse.Anschnitt
				jeSeite := int((masse.Seitenbreite-2*masse.Rand)/zellbreit) *
					int((masse.Seitenhoehe-2*masse.Rand)/zellhoch)

				for _, befund := range massbefunde(bogenVermessen(t, masse, jeSeite+1), masse, jeSeite+1) {
					t.Error(befund)
				}
			})
		}
	}
}
