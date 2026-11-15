package pdf_test

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
)

// **Der äußere Maßstab: Quartz liest, was wir geschrieben haben.**
//
// # Warum das hier steht und nicht nur ein Strukturtest
//
// Der Zweitblick über Phase 9 hat den Satz geprägt, der hier trägt: *ein
// Rundlauf prüft Selbstkonsistenz, nicht Wahrheit*. Ein PDF, das nur unser
// eigener Leser liest, ist kein PDF — es ist eine Datei, über die wir uns mit
// uns selbst einig sind.
//
// `sips` ist der Weg zu einem fremden Leser, den jede Zielmaschine hat: es
// gehört zu macOS, ruft Quartz und gibt die Seitengröße in Punkten zurück. Ein
// kaputtes PDF beantwortet die Frage gar nicht.
//
// **Auf einer anderen Plattform steht der Test still**, und das ist eine Lücke,
// keine Auslassung: gebaut wird für darwin (§12), geprüft wird auf genau einer
// Maschine ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
func quartzMisst(t *testing.T, inhalt []byte) (breite, hoehe float64) {
	t.Helper()

	if runtime.GOOS != "darwin" {
		t.Skip("sips gibt es nur auf macOS — auf dieser Plattform bleibt das PDF ungeprüft")
	}

	ort := filepath.Join(t.TempDir(), "probe.pdf")
	if err := os.WriteFile(ort, inhalt, 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	// #nosec G204 -- der Pfad kommt aus `t.TempDir()`, nicht aus einer Eingabe
	ausgabe, err := exec.CommandContext(t.Context(),
		"sips", "-g", "pixelWidth", "-g", "pixelHeight", ort).CombinedOutput()
	if err != nil {
		t.Fatalf("Quartz liest das PDF nicht: %v\n%s", err, ausgabe)
	}

	return zahlHinter(t, string(ausgabe), "pixelWidth:"), zahlHinter(t, string(ausgabe), "pixelHeight:")
}

// zahlHinter zieht den Wert hinter einer Beschriftung aus der Ausgabe.
func zahlHinter(t *testing.T, ausgabe, beschriftung string) float64 {
	t.Helper()

	for _, zeile := range strings.Split(ausgabe, "\n") {
		gefunden := strings.TrimSpace(zeile)
		if !strings.HasPrefix(gefunden, beschriftung) {
			continue
		}

		wert, err := strconv.ParseFloat(
			strings.TrimSpace(strings.TrimPrefix(gefunden, beschriftung)), 64)
		if err != nil {
			t.Fatalf("%q ist keine Zahl: %v", gefunden, err)
		}

		return wert
	}

	t.Fatalf("%q steht nicht in der Ausgabe:\n%s", beschriftung, ausgabe)

	return 0
}

// **Ein leeres Dokument ist schon ein PDF** — und zwar eines mit der Seitengröße,
// die wir angegeben haben.
//
// A4 sind 210 × 297 mm; in Punkten zu 1/72 Zoll macht das 595,28 × 841,89.
func TestEineLeereSeiteHatDieAngegebeneGroesse(t *testing.T) {
	dokument := pdf.NeuesDokument()
	dokument.NeueSeite(pdf.MM(210), pdf.MM(297))

	var gebaut bytes.Buffer
	if err := dokument.Schreibe(&gebaut); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	breite, hoehe := quartzMisst(t, gebaut.Bytes())

	// Auf ein hundertstel Punkt — mehr sagt `sips` nicht, und mehr braucht
	// niemand: ein hundertstel Punkt ist ein dreitausendstel Millimeter.
	if abstand(breite, 595.28) > 0.01 || abstand(hoehe, 841.89) > 0.01 {
		t.Errorf("%.2f × %.2f Punkte, erwartet 595,28 × 841,89", breite, hoehe)
	}
}

// **Zwei Seiten sind zwei Seiten.** Ein Bogen mit dreihundert Karten hat mehr
// als eine, und ein Seitenbaum, der nur die letzte führt, fällt beim Zählen auf.
func TestMehrereSeitenStehenAlleImDokument(t *testing.T) {
	dokument := pdf.NeuesDokument()
	for range 3 {
		dokument.NeueSeite(pdf.MM(210), pdf.MM(297))
	}

	var gebaut bytes.Buffer
	if err := dokument.Schreibe(&gebaut); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quartzMisst(t, gebaut.Bytes())

	if anzahl := strings.Count(gebaut.String(), "/Type /Page\n"); anzahl != 3 {
		t.Errorf("%d Seiten im Dokument, erwartet 3", anzahl)
	}
}

// **Ein Bild kommt mit — und zwar Bildpunkt für Bildpunkt.**
//
// Ein PDF, das ein Bild „ungefähr" trägt, ist für einen Druckbogen wertlos.
// Geprüft wird deshalb der ausgepackte Datenstrom gegen das, was hineinging.
func TestEinBildKommtUnveraendertInsPDF(t *testing.T) {
	bild := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	bild.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	bild.SetNRGBA(0, 1, color.NRGBA{B: 255, A: 255})
	bild.SetNRGBA(1, 1, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(pdf.MM(210), pdf.MM(297))
	seite.Bild(bild, pdf.MM(10), pdf.MM(10), pdf.MM(63), pdf.MM(88))

	var gebaut bytes.Buffer
	if err := dokument.Schreibe(&gebaut); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quartzMisst(t, gebaut.Bytes())

	roh := ausgepackterStrom(t, gebaut.Bytes())

	// **Zeilenweise von oben nach unten**, wie PDF es erwartet: die erste Zeile
	// des Bilds steht vorn, nicht die unterste.
	erwartet := []byte{
		255, 0, 0, 0, 255, 0,
		0, 0, 255, 255, 255, 255,
	}

	if !bytes.Equal(roh, erwartet) {
		t.Errorf("Bilddaten %v, erwartet %v", roh, erwartet)
	}
}

// abstand ist der Betrag der Differenz.
func abstand(a, b float64) float64 {
	if a > b {
		return a - b
	}

	return b - a
}

// **Eine Schnittmarke ist eine Linie, und sie steht im Inhaltsstrom.**
//
// Geprüft wird der Strom und nicht das gerenderte Bild: ob Quartz sie zeichnet,
// sagt die Sichtprüfung — dass sie mit den richtigen Zahlen dasteht, sagt nur
// ein Blick hinein.
func TestEineLinieStehtImInhalt(t *testing.T) {
	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(pdf.MM(210), pdf.MM(297))
	seite.Linie(pdf.MM(10), pdf.MM(20), pdf.MM(10), pdf.MM(30), 0.5, 0.5)

	var gebaut bytes.Buffer
	if err := dokument.Schreibe(&gebaut); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quartzMisst(t, gebaut.Bytes())

	// 10 mm sind 28,3465 Punkte, 20 mm sind 56,6929.
	if !strings.Contains(gebaut.String(), "28.3465 56.6929 m 28.3465 85.0394 l") {
		t.Errorf("die Linie steht nicht im Inhalt:\n%s", gebaut.String()[:600])
	}
}

// **Ein Ziel, das nicht schreibbar ist, wird gemeldet.**
//
// Eine Platte kann voll sein, und ein PDF, das zur Hälfte dasteht, sieht im
// Dateibrowser aus wie ein ganzes (§8).
func TestEinNichtSchreibbaresZielWirdGemeldet(t *testing.T) {
	dokument := pdf.NeuesDokument()
	dokument.NeueSeite(pdf.MM(210), pdf.MM(297))

	if err := dokument.Schreibe(sperrziel{}); err == nil {
		t.Error("kein Fehler")
	}
}

// sperrziel nimmt nichts an.
type sperrziel struct{}

func (sperrziel) Write([]byte) (int, error) { return 0, errVoll }

var errVoll = errors.New("kein Platz mehr")

// **Ein Maß, das keines ist, wird zur Null — und nicht zu `NaN`.**
//
// # Der zweite Fund eines Fuzzlaufs in diesem Projekt
//
// `strconv.FormatFloat` gibt für `math.Inf(1)` die Zeichenkette `+Inf` und für
// `math.NaN()` das Wort `NaN`. Beides ist in einem PDF-Inhaltsstrom ein
// Syntaxfehler, und zwar einer, der **die ganze Seite** unlesbar macht.
//
// **Woher eine Unzahl kommt**: `card width:` steht in einem fremden
// Stilpaket, und der Bogenleger rechnet daraus Punkte. Gefunden am 2026-08-29
// von `FuzzDokument`, schon aus der Saat
// (#1289).
func TestEineUnzahlWirdZurNull(t *testing.T) {
	t.Parallel()

	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(math.Inf(1), math.NaN())
	seite.Linie(math.NaN(), math.Inf(-1), 1, 1, math.Inf(1), 0)

	roh := dokument.Bytes()

	for _, unzahl := range []string{"NaN", "Inf"} {
		if bytes.Contains(roh, []byte(unzahl)) {
			t.Errorf("im Dokument steht %q — das liest kein PDF-Leser", unzahl)
		}
	}

	if !bytes.HasPrefix(roh, []byte("%PDF-")) {
		t.Error("das Dokument beginnt nicht mit %PDF-")
	}
}
