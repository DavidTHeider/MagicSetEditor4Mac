//go:build fremdpakete

// **Wo die Textzeilen stehen — bei uns und bei MSE2** —
// #1029,
// Werkstück Ω.3.
//
// # Die Frage
//
// Bei 87 von 680 kanonischen Karten ist unser Text gegen MSE2s versetzt, und
// der Versatz hängt an der **Vorlage**. Fünf Ursachen sind gegen den Quelltext
// ausgeschlossen (Ausrichtung, Polster, die drei Zeilenfaktoren, die Dehnung,
// leere Schlusszeilen). Übrig bleibt die **Zeilenhöhe** — und die ist bei uns
// gemessen (MPlantin 14 gibt 16,25 auf Viertelpunkte), bei MSE2 nicht.
//
// **Das Orakel hilft nicht**: MSE-Script hat keine Funktion für Textmaße
// (`src/script/functions/*.cpp`, nachgesehen am 2026-08-25). Was bleibt, ist
// MSEs eigenes Bild.
//
// # Wie hier gemessen wird
//
// Die **Herkunft** sagt, welche Bildpunkte unser Rasterer als Text gemalt hat.
// Ihr umschließendes Rechteck ist der Textkasten. In diesem Rechteck wird für
// **beide** Bilder je Zeile gezählt, wie viel Tinte darin steht; daraus werden
// die **Bänder** — zusammenhängende Zeilen mit Tinte.
//
// Der Abstand zweier Bandanfänge ist die Zeilenhöhe mal Faktor, und der
// Abstand des ersten Bandes zur Kastenoberkante ist die Lage.
//
// # Was diese Messung nicht sagt
//
// **Sie findet Bänder, keine Grundlinien.** Ein Band beginnt bei der obersten
// Tinte — bei einer Zeile mit `T` höher als bei einer aus `a`, `e`, `o`. Für
// den **Abstand zweier Bänder** hebt sich das auf, solange beide Zeilen
// Oberlängen haben; für die Lage des ersten nicht.
//
// **Und sie misst unseren Kasten.** Wo MSEs Text weiter reicht als unserer,
// fällt er aus dem Rechteck — die Zahl ist dann eine über unseren Ausschnitt.
package app

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// tintengrenze ist, ab wann ein Bildpunkt als Tinte zählt.
//
// **Der Abstand zum hellsten Punkt des Kastens**, nicht ein fester Wert: ein
// Textkasten ist bei `m15-altered` fast weiß und bei `archenemy` dunkel, und
// eine feste Schwelle fände dort entweder alles oder nichts.
const tintengrenze = 96

// TestDieZeilenlageGegenMSE2 misst die Bänder einer benannten Karte.
//
// **Ein Messlauf, kein Wächter.** Er sichert nichts zu; er legt zwei Zahlen
// nebeneinander, damit ein Mensch sie liest. Eine Zusicherung käme erst, wenn
// klar ist, welche Zahl die richtige ist — und das ist die offene Frage.
func TestDieZeilenlageGegenMSE2(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	// **Je Vorlage eine Karte aus der Versatztabelle.** Die Namen stehen dort
	// mit ihrer Vorlage; hier sind sie der Einstieg.
	for _, fall := range []struct{ satz, karte string }{
		{"Archenemy", "Dance, Pathetic Marionette"},
		{"Archenemy", "Bow to My Command"},
		{"Vanguard", "Barrin"},
		{"Black", "No Mercy"},
		{"Black", "Western Paladin"},
	} {
		t.Run(fall.satz+"/"+fall.karte, func(t *testing.T) {
			messeZeilenlage(t, anwendung, fall.satz, fall.karte)
		})
	}
}

// messeZeilenlage legt die Bänder beider Bilder nebeneinander.
func messeZeilenlage(t *testing.T, anwendung *App, satz, kartenname string) {
	t.Helper()

	karte, da := karteMitNamen(t, anwendung, satz, kartenname)
	if !da {
		t.Skipf("%s steht nicht im Bestand", kartenname)
	}

	referenz, err := ladeBild(filepath.Join(referenzort(t, satz),
		referenzdatei(kartenname)))
	if err != nil {
		t.Skipf("kein Referenzbild: %v", err)
	}

	vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
	if vorschau.Fehler != nil {
		t.Fatalf("Vorschau: %+v", vorschau.Fehler)
	}

	unsere := bildAus(t, anwendung, vorschau.Vorschau.Bild)

	// **Der Kasten kommt aus dem Plan, nicht aus der Herkunft.**
	//
	// Die Herkunft sagt nur „hier stand Text" — und das gilt für den
	// Kartennamen, die Typzeile, den Illustrator und den Regeltext
	// gleichermaßen. Ein Rechteck über sie alle umfasst fast die ganze Karte,
	// und dann hat jede Zeile irgendwo Tinte: die Bänder liefen zu einem
	// einzigen zusammen.
	//
	// Der Plan weiß, **welches Feld** eine Ebene malt.
	kasten, hat := feldkasten(t, anwendung, karte.ID, "text")
	if !hat {
		t.Skip("die Karte malt kein Feld text")
	}

	t.Logf("Kasten des Feldes `text` aus dem Plan: x %d…%d, y %d…%d",
		kasten.Min.X, kasten.Max.X, kasten.Min.Y, kasten.Max.Y)

	unsereBaender := baender(unsere, kasten)
	mse2Baender := baender(referenz, kasten)

	t.Logf("  unsere Bänder: %s", bandtext(unsereBaender))
	t.Logf("  MSE2s  Bänder: %s", bandtext(mse2Baender))
	t.Logf("  unsere Abstände: %s", abstandtext(unsereBaender))
	t.Logf("  MSE2s  Abstände: %s", abstandtext(mse2Baender))

	meldeUnsereZeilen(t, anwendung, karte.ID)

	if len(unsereBaender) > 0 && len(mse2Baender) > 0 {
		t.Logf("  erstes Band: bei uns %d, bei MSE2 %d — Unterschied %+d",
			unsereBaender[0].Min, mse2Baender[0].Min,
			unsereBaender[0].Min-mse2Baender[0].Min)
	}
}

// meldeUnsereZeilen schreibt unsere **ungerundeten** Zeilenoberkanten.
//
// **Sie sind die Antwort auf die Frage, die die Bänder offen lassen.** Ein
// gezeichneter Abstand von 17 statt 16 kann eine zu große Zeilenhöhe sein oder
// eine Nachkommastelle, die über die Rundung kippt — und das unterscheidet nur
// die ungerundete Zahl.
func meldeUnsereZeilen(t *testing.T, anwendung *App, karteID string) {
	t.Helper()

	plan := planZurKarte(t, anwendung, karteID)

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		return
	}

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText || ebene.Feld != "text" {
			continue
		}

		gesetzt, err := rasterer.GesetzterZug(ebene)
		if err != nil {
			t.Logf("  unser Satz: %v", err)

			continue
		}

		t.Logf("  unser Satz: Kasten oben %.3f hoch %.3f, Zug hoch %.4f, %d Zeilen",
			ebene.Rahmen.Oben, ebene.Rahmen.Hoehe, gesetzt.Hoehe,
			len(gesetzt.Zeilen))

		for i, zeile := range gesetzt.Zeilen {
			abstand := 0.0
			if i > 0 {
				abstand = zeile.Oben - gesetzt.Zeilen[i-1].Oben
			}

			t.Logf("     %2d: oben %8.4f  hoch %7.4f  Abstand %7.4f  %.24q",
				i, zeile.Oben, zeile.Hoehe, abstand, zeile.Text)
		}
	}
}

// feldkasten gibt das Rechteck der Textebene eines benannten Feldes.
//
// **Die Vereinigung aller Ebenen dieses Feldes.** `magic-m15` malt den
// Regeltext in mehreren Ebenen — Regeltext und Flavortext teilen sich den
// Kasten —, und eine davon allein wäre ein Ausschnitt.
func feldkasten(
	t *testing.T, anwendung *App, karteID, feld string,
) (image.Rectangle, bool) {
	t.Helper()

	plan := planZurKarte(t, anwendung, karteID)

	linksOben := image.Pt(1<<30, 1<<30)
	rechtsUnten := image.Pt(0, 0)
	gefunden := false

	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtText || ebene.Feld != feld {
			continue
		}

		// **Eine Ebene ohne Fläche malt nicht** — MSE nimmt Breite oder Höhe
		// null als „wird nicht gezeichnet" (`ValueViewer::isVisible`).
		if ebene.Rahmen.Breite <= 0 || ebene.Rahmen.Hoehe <= 0 {
			continue
		}

		gefunden = true
		linksOben.X = min(linksOben.X, int(ebene.Rahmen.Links))
		linksOben.Y = min(linksOben.Y, int(ebene.Rahmen.Oben))
		rechtsUnten.X = max(rechtsUnten.X,
			int(ebene.Rahmen.Links+ebene.Rahmen.Breite))
		rechtsUnten.Y = max(rechtsUnten.Y,
			int(ebene.Rahmen.Oben+ebene.Rahmen.Hoehe))
	}

	if !gefunden {
		return image.Rectangle{}, false
	}

	return image.Rectangle{Min: linksOben, Max: rechtsUnten}, true
}

// band ist ein zusammenhängender Zeilenbereich mit Tinte.
type band struct{ Min, Max int }

// baender findet die Zeilenbereiche mit Tinte in einem Rechteck.
//
// **Die Schwelle hängt am Kasten**: gemessen wird gegen den hellsten Punkt
// darin, damit ein dunkler Textkasten dieselbe Aussage gibt wie ein heller.
func baender(bild image.Image, kasten image.Rectangle) []band {
	hellster := hellsterPunkt(bild, kasten)

	var (
		gefunden []band
		offen    = -1
	)

	for y := kasten.Min.Y; y < kasten.Max.Y; y++ {
		hatTinte := false

		for x := kasten.Min.X; x < kasten.Max.X; x++ {
			if hellster-helligkeit(bild, x, y) >= tintengrenze {
				hatTinte = true

				break
			}
		}

		switch {
		case hatTinte && offen < 0:
			offen = y

		case !hatTinte && offen >= 0:
			gefunden = append(gefunden, band{offen, y})
			offen = -1
		}
	}

	if offen >= 0 {
		gefunden = append(gefunden, band{offen, kasten.Max.Y})
	}

	return gefunden
}

// hellsterPunkt gibt die größte Helligkeit im Rechteck.
func hellsterPunkt(bild image.Image, kasten image.Rectangle) int {
	hellster := 0

	for y := kasten.Min.Y; y < kasten.Max.Y; y++ {
		for x := kasten.Min.X; x < kasten.Max.X; x++ {
			hellster = max(hellster, helligkeit(bild, x, y))
		}
	}

	return hellster
}

// helligkeit gibt den Mittelwert der drei Kanäle, 0…255.
func helligkeit(bild image.Image, x, y int) int {
	r, g, b, _ := bild.At(x, y).RGBA()

	return (int(r>>8) + int(g>>8) + int(b>>8)) / 3
}

// bandtext schreibt die Bänder als Liste.
func bandtext(baender []band) string {
	if len(baender) == 0 {
		return "keine"
	}

	teile := make([]string, 0, len(baender))
	for _, eines := range baender {
		teile = append(teile, fmt.Sprintf("%d…%d", eines.Min, eines.Max))
	}

	return strings.Join(teile, "  ")
}

// abstandtext schreibt die Abstände zwischen den Bandanfängen.
func abstandtext(baender []band) string {
	if len(baender) < 2 {
		return "—"
	}

	teile := make([]string, 0, len(baender)-1)
	for i := 1; i < len(baender); i++ {
		teile = append(teile, fmt.Sprintf("%d", baender[i].Min-baender[i-1].Min))
	}

	return strings.Join(teile, "  ")
}

// karteMitNamen sucht eine Karte des Satzes beim Namen.
func karteMitNamen(
	t *testing.T, anwendung *App, satz, gesucht string,
) (KarteDTO, bool) {
	t.Helper()

	geoeffnet := anwendung.SetOeffnen(context.Background(),
		filepath.Join(kanonort(), satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("%s einfahren: %+v", satz, geoeffnet.Fehler)
	}

	antwort := anwendung.Karten(context.Background(), geoeffnet.SetID)
	if antwort.Fehler != nil {
		t.Fatalf("Karten: %+v", antwort.Fehler)
	}

	for _, karte := range antwort.Karten {
		if kartenname(karte) == gesucht {
			return karte, true
		}
	}

	return KarteDTO{}, false
}
