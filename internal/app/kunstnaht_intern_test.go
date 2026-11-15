// Der Wächter über die Naht, die #380
// gefehlt hat: **das Bild der Karte, durch die laufende Anwendung.**
//
// Fünf Teile liegen dazwischen, und jeder war für sich geprüft: der Setkopf
// kennt seine Herkunft, die Vorlage trägt sie, der Feldstil baut den Verweis,
// die Bildquelle löst ihn auf, der Rasterer zeichnet ihn. **Dazwischen nichts**
// — genau die Lage, in der in Phase 6 dreimal ein fertiges Werkstück unsichtbar
// blieb.
//
// Er läuft **ohne** fremdes Paket: Spiel und Stil kommen aus `saat/`, das Set
// und seine Illustration schreibt der Test selbst (§11).
package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// kunstfarbe ist die Farbe der Illustration, die dieser Test malt.
//
// **Reines Magenta**, weil es auf einer Karte aus der Saat sonst nirgends
// vorkommt: der Grund ist weiß, die Schrift schwarz. Wer diese Farbe im Bild
// findet, hat das Bild der Karte gefunden und nicht irgendetwas.
var kunstfarbe = color.NRGBA{R: 255, G: 0, B: 255, A: 255}

// **Die Illustration einer Karte steht in der Vorschau.**
//
// # Die Gegenprobe, die dahintersteht
//
// [TestOhneEigenesBildBleibtDasKunstfensterLeer] ist sie: dieselbe Karte ohne
// Dateinamen, und dann ist kein einziger Bildpunkt magenta. Ohne sie wäre
// dieser Wächter auch dann grün, wenn er die falsche Fläche zählte.
func TestDasBildDerKarteStehtInDerVorschau(t *testing.T) {
	anwendung := mitSaat(t)

	vorschau := ersteVorschau(t, anwendung, kunstset(t, "bild.png"))

	bild := bildAus(t, anwendung, vorschau.Vorschau.Bild)

	gezaehlt := zaehleFarbe(bild, kunstfarbe)

	t.Logf("%d magentafarbene Bildpunkte, %d Befunde", gezaehlt, len(vorschau.Vorschau.Befunde))

	for _, befund := range vorschau.Vorschau.Befunde {
		t.Logf("  %s", befund)
	}

	// **Das Kunstfenster misst 315 × 230 = 72.450 Bildpunkte**, und die
	// Illustration wird darauf **gestreckt** — sie füllt es also ganz. Die
	// Schwelle liegt bei neun Zehnteln davon und lässt Raum für die Ränder, die
	// eine Streckung an den Kanten weichzeichnen könnte.
	if erwartet := 315 * 230 * 9 / 10; gezaehlt < erwartet {
		t.Errorf("%d magentafarbene Bildpunkte, erwartet mindestens %d — "+
			"das Bild der Karte ist nicht angekommen; Befunde: %v",
			gezaehlt, erwartet, vorschau.Vorschau.Befunde)
	}
}

// **Ohne Dateinamen bleibt das Kunstfenster leer** — die Gegenprobe.
func TestOhneEigenesBildBleibtDasKunstfensterLeer(t *testing.T) {
	anwendung := mitSaat(t)

	vorschau := ersteVorschau(t, anwendung, kunstset(t, ""))

	if gezaehlt := zaehleFarbe(bildAus(t, anwendung, vorschau.Vorschau.Bild), kunstfarbe); gezaehlt != 0 {
		t.Errorf("%d magentafarbene Bildpunkte, obwohl die Karte kein Bild führt", gezaehlt)
	}
}

// ersteVorschau öffnet ein Set und gibt die Vorschau seiner ersten Karte.
func ersteVorschau(t *testing.T, anwendung *App, ort string) VorschauAntwort {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(t.Context(), ort)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if len(karten.Karten) == 0 {
		t.Fatal("keine Karte im geöffneten Set")
	}

	vorschau := anwendung.Vorschau(t.Context(), karten.Karten[0].ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	return vorschau
}

// kunstset schreibt ein Set auf `probe-kunst` — samt seiner Illustration.
//
// **Die Datei liegt im Set-Paket**, so wie MSE es hält: `getLocalPackage()`.
// Ein leerer Name heißt „die Karte führt kein Bild" und ist die Gegenprobe.
func kunstset(t *testing.T, datei string) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "kunstprobe.mse-set")
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\ngame: probe\nstylesheet: probe-kunst\n" +
		"set info:\n\ttitle: Mit Bild\ncard:\n\tname: Feuerball\n\tfarbe: rot\n"
	if datei != "" {
		inhalt += "\tkunst: " + datei + "\n"

		legeKunstbild(t, filepath.Join(ort, datei))
	}

	if err := os.WriteFile(filepath.Join(ort, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// legeKunstbild malt eine einfarbige Illustration.
//
// **Klein und einfarbig**: gemessen wird, ob sie ankommt und das Fenster füllt,
// nicht wie sie skaliert. Vier Bildpunkte reichen dafür — und sie zeigen
// zugleich, dass gestreckt wird: eingepasst bliebe das Fenster zu drei Vierteln
// frei, weil es breiter als hoch ist.
func legeKunstbild(t *testing.T, pfad string) {
	t.Helper()

	gemalt := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := range 2 {
		for x := range 2 {
			gemalt.SetNRGBA(x, y, kunstfarbe)
		}
	}

	var puffer bytes.Buffer
	if err := png.Encode(&puffer, gemalt); err != nil {
		t.Fatalf("kodieren: %v", err)
	}

	if err := os.WriteFile(pfad, puffer.Bytes(), 0o600); err != nil {
		t.Fatalf("ablegen: %v", err)
	}
}

// bildAus holt das PNG über die Adresse aus dem DTO.
//
// **Über den echten Weg und nicht am Rasterer vorbei**: genau diese Adresse
// bekommt das Fenster, und ein Fehler auf dem Weg dorthin — eine falsche Marke,
// ein vergessener Medientyp, ein zu kleines Fach — wäre sonst unsichtbar.
//
// **Seit #1270
// braucht er die Anwendung dazu.** Vorher stand das Bild als Data-URI in der
// Antwort und der Test kam mit der Zeichenkette aus; jetzt liegt es im Fach und
// wird über den Assethandler abgeholt — dieselbe Bewegung wie im Fenster.
//
// **Hier und nicht in `rahmennaht_intern_test.go`**, wo er herkommt: der dortige
// Lauf hängt am Build-Tag `fremdpakete` und liefe damit in der Pipeline nie.
func bildAus(t *testing.T, anwendung *App, adresse string) image.Image {
	t.Helper()

	if !strings.HasPrefix(adresse, Bildwegpfad) {
		t.Fatalf("die Vorschau trägt keine Bildadresse: %.40s…", adresse)
	}

	bild, err := png.Decode(bytes.NewReader(bildausDemWeg(t, anwendung, adresse)))
	if err != nil {
		t.Fatalf("Vorschaubild lesen: %v", err)
	}

	return bild
}

// zaehleFarbe zählt die Bildpunkte einer Farbe.
func zaehleFarbe(bild image.Image, gesucht color.NRGBA) int {
	var gezaehlt int

	rahmen := bild.Bounds()

	for y := rahmen.Min.Y; y < rahmen.Max.Y; y++ {
		for x := rahmen.Min.X; x < rahmen.Max.X; x++ {
			rot, gruen, blau, alpha := bild.At(x, y).RGBA()

			gelesen := color.NRGBA{
				R: uint8(rot >> 8), G: uint8(gruen >> 8),
				B: uint8(blau >> 8), A: uint8(alpha >> 8),
			}
			if gelesen == gesucht {
				gezaehlt++
			}
		}
	}

	return gezaehlt
}
