package service_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bogenleger baut einen Leger über einem Bestand mit `anzahl` Karten.
func bogenleger(t *testing.T, anzahl int) *service.Bogenleger {
	t.Helper()

	namen := make([]string, 0, anzahl)
	for nummer := range anzahl {
		namen = append(namen, "Karte "+strconv.Itoa(nummer+1))
	}

	return service.NeuerBogenleger(
		kartenDouble{leserDouble: karten(namen...)},
		planDouble{breite: 375, hoehe: 523},
		render.NeuerRasterer(nil))
}

// legeBogen schreibt einen Bogen und gibt Pfad und Inhalt.
func legeBogen(
	t *testing.T, anzahl int, masse service.Bogenmasse,
) (lauf *service.Bogenlauf, inhalt string) {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	lauf, err := bogenleger(t, anzahl).Lege(t.Context(), 7, ziel, masse, domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return lauf, string(roh)
}

// kartenmasse liest die Größen der gesetzten Bilder aus dem Inhaltsstrom.
//
// **Aus dem PDF und nicht aus dem Rückgabewert**: dass der Leger dieselbe Zahl
// zurückgibt, die er gerechnet hat, ist keine Prüfung. Was zählt, steht in der
// Datei.
func kartenmasse(t *testing.T, inhalt string) [][2]float64 {
	t.Helper()

	muster := regexp.MustCompile(`([0-9.]+) 0 0 ([0-9.]+) ([0-9.]+) ([0-9.]+) cm`)

	treffer := muster.FindAllStringSubmatch(inhalt, -1)
	masse := make([][2]float64, 0, len(treffer))

	for _, treffer := range treffer {
		breite, _ := strconv.ParseFloat(treffer[1], 64) //nolint:errcheck // das Muster lässt nur Zahlen zu
		hoehe, _ := strconv.ParseFloat(treffer[2], 64)  //nolint:errcheck // dito

		masse = append(masse, [2]float64{breite, hoehe})
	}

	return masse
}

// **Eine Karte auf dem Bogen ist 63 × 88 mm.**
//
// # Warum das die eine Zusage ist
//
// #50 sagt
// „maßhaltig", und das ist keine Redensart: eine Karte, die 61 mm breit
// herauskommt, passt in keine Hülle und in kein Deck. Gemessen wird in
// **Punkten** — 63 mm sind 178,58, 88 mm sind 249,45.
func TestEineKarteAufDemBogenIst63Mal88Millimeter(t *testing.T) {
	_, inhalt := legeBogen(t, 1, service.StandardBogen())

	masse := kartenmasse(t, inhalt)
	if len(masse) != 1 {
		t.Fatalf("%d Karten im Bogen, erwartet 1", len(masse))
	}

	if abstand(masse[0][0], 178.5827) > 0.001 || abstand(masse[0][1], 249.4488) > 0.001 {
		t.Errorf("%.4f × %.4f Punkte, erwartet 178,5827 × 249,4488", masse[0][0], masse[0][1])
	}
}

// **Der Anschnitt vergrößert das Bild über die Schnittkante hinaus.**
//
// Er ist der Grund, warum eine geschnittene Karte keinen weißen Rand hat: der
// Schnitt liegt bei 63 mm, das Bild reicht weiter. Bei 3 mm Anschnitt wird aus
// 63 × 88 also 69 × 94 mm — 195,59 × 266,46 Punkte.
func TestDerAnschnittVergroessertDasBild(t *testing.T) {
	masse := service.StandardBogen()
	masse.Anschnitt = 3

	_, inhalt := legeBogen(t, 1, masse)

	gesetzt := kartenmasse(t, inhalt)
	if abstand(gesetzt[0][0], 195.5906) > 0.001 || abstand(gesetzt[0][1], 266.4567) > 0.001 {
		t.Errorf("%.4f × %.4f Punkte, erwartet 195,5906 × 266,4567", gesetzt[0][0], gesetzt[0][1])
	}
}

// **Was nicht mehr auf die Seite passt, kommt auf die nächste.**
//
// Auf A4 stehen bei 10 mm Rand drei Spalten und drei Zeilen — neun Karten. Die
// zehnte beginnt eine zweite Seite.
func TestDieZehnteKarteBeginntEineZweiteSeite(t *testing.T) {
	lauf, inhalt := legeBogen(t, 10, service.StandardBogen())

	if lauf.Seiten != 2 {
		t.Errorf("%d Seiten, erwartet 2", lauf.Seiten)
	}

	if lauf.Karten != 10 {
		t.Errorf("%d Karten, erwartet 10", lauf.Karten)
	}

	if anzahl := strings.Count(inhalt, "/Type /Page\n"); anzahl != 2 {
		t.Errorf("%d Seitenobjekte, erwartet 2", anzahl)
	}
}

// **Schnittmarken stehen am Papierrand, nicht auf der Karte.**
//
// Eine Marke über dem Bild wäre eine, die man mitdruckt und danach ansieht.
func TestSchnittmarkenStehenAmRand(t *testing.T) {
	_, mitMarken := legeBogen(t, 1, service.StandardBogen())

	ohne := service.StandardBogen()
	ohne.Schnittmarken = false

	_, ohneMarken := legeBogen(t, 1, ohne)

	mit := strings.Count(mitMarken, " l\nS")
	keine := strings.Count(ohneMarken, " l\nS")

	// Vier Kanten, je zwei Marken: acht Linien für eine einzelne Karte.
	if mit != 8 {
		t.Errorf("%d Schnittmarken, erwartet 8", mit)
	}

	if keine != 0 {
		t.Errorf("%d Schnittmarken ohne Marken", keine)
	}
}

// **Eine Vorlage, deren Verhältnis nicht zur Karte passt, wird benannt.**
//
// Der Bogen zieht sie trotzdem auf 63 × 88 — „maßhaltig" ist die Zusage, und
// eine Karte mit weißem Rand wäre die schlechtere Antwort. Aber verzerrt ist
// verzerrt, und das gehört gesagt (§8).
func TestEineUnpassendeVorlageWirdBenannt(t *testing.T) {
	leger := service.NeuerBogenleger(
		kartenDouble{leserDouble: karten("Breit")},
		planDouble{breite: 400, hoehe: 400},
		render.NeuerRasterer(nil))

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	lauf, err := leger.Lege(t.Context(), 7, ziel, service.StandardBogen(), domain.Kartenauswahl{})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if len(lauf.Befunde) == 0 {
		t.Fatal("kein Befund — die Verzerrung bliebe unsichtbar")
	}
}

// Ein Verhältnis, das um weniger als ein Hundertstel abweicht, ist keiner.
//
// **375 × 523 ist nicht 63 × 88**, sondern 0,15 % daneben. Wer das meldete,
// meldete es bei jeder echten Vorlage — und ein Befund, der immer kommt, wird
// nicht gelesen.
func TestEinWinzigerUnterschiedIstKeinBefund(t *testing.T) {
	lauf, _ := legeBogen(t, 1, service.StandardBogen())

	if len(lauf.Befunde) != 0 {
		t.Errorf("Befunde: %v", lauf.Befunde)
	}
}

// abstand ist der Betrag der Differenz.
func abstand(a, b float64) float64 {
	if a > b {
		return a - b
	}

	return b - a
}
