//go:build fremdpakete

// Der letzte offene Kasten von
// #532: **das
// Kunstfenster einer benannten Karte, nicht ein Mittel über 573.**
//
// # Warum das ein eigener Wächter ist
//
// Die Wirkung des Filterwechsels ist über den Kanon gemessen worden — 23,043
// auf 22,646. Das belegt die Richtung und **nicht**, dass genau diese Karte
// besser geworden ist: eine Mittelung versteckt eine einzelne
// Verschlechterung. §10 sagt es als Regel — die Reichweite einer Messung ist
// die Reichweite ihrer Aussage.
//
// # Was er misst
//
// Das Feld `image` von `A-Cloister Gargoyle` aus `White`, 316 × 231
// Bildpunkte bei 29,59 — dieselbe Fläche, an der die Filtermessung vom
// 2026-08-19 erhoben wurde, und **nur** sie. Der Rest der Karte steht nicht
// zur Frage: dort sitzen Rahmen, Satz und Symbole.
//
// # Die Gegenprobe
//
// **Am 2026-08-20 gefahren**: die Weiche in `internal/render/skalierung.go`
// umgedreht — also wieder Flächenmittel beim Verkleinern —, und dieser Lauf
// wird rot mit **3,057**.
//
// Das ist auf drei Stellen genau die Zahl, die die Abnahme von #532 als Marke
// nannte. Der Wächter misst damit nachweislich die Fläche, um die es dort
// ging, und nicht irgendeine.
//
// # Was er nicht prüft
//
//   - **Nicht, dass der Filter richtig gewählt ist.** Das hält
//     `render.TestMSEVerkleinertBilinearUndVergroessertGemittelt` gegen wx.
//   - **Nicht, warum der Rest bleibt.** Er ist benannt: wx' bilineare
//     Arithmetik, gegen die unsere Abtastung 2,305 danebenliegt.
package app

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// kunstfensterkarte ist die Karte, an der #532 gemessen wurde.
const (
	kunstfenstersatz  = "White"
	kunstfensterkarte = "A-Cloister Gargoyle"
	kunstfensterfeld  = "image"
)

// kunstfensterdeckel ist der Abstand, unter dem das Kunstfenster bleibt.
//
// **Die Abnahme aus #532 nannte 3,057 als Marke** und null als Ziel. Gemessen
// sind am 2026-08-20 **2,406** über 72.996 Bildpunkte — unter der Marke, über
// dem Ziel.
//
// **Das Viertelraster aus #546 rührt sie nicht an**: vor und nach der Änderung
// dieselben 2,406, eigens gegengeprüft. Dort steht ein Bild und kein Satz.
//
// **Am 2026-08-19 sind 2,362 notiert worden.** Der Unterschied von 0,044 ist
// **nicht zugeordnet**; dazwischen liegen die Werkstücke #547, #555, #556,
// #557, #565, #576, #584, #587 und #588, und welches davon ihn trägt, ist
// nicht gemessen. Er wird hier genannt und nicht weggerundet (§10).
//
// **Null wird nicht erreicht, und der Grund ist benannt**: es ist wx' bilineare
// Arithmetik, nicht die Filterwahl. Gegen wx' eigenes Ergebnis liegt unsere
// Abtastung bei 2,305 — fast der ganze Rest.
const kunstfensterdeckel = 2.500

// **Das Kunstfenster von `A-Cloister Gargoyle` bleibt unter seinem Deckel.**
func TestDasKunstfensterEinerEchtenKarte(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	if _, err := os.Stat(mse2ausgabe(kunstfenstersatz)); err != nil {
		t.Skip("keine MSE2-Ausgabe — scripts/mse2-referenz.sh White")
	}

	anwendung := gestartetMit(t, func() string { return wurzel })

	geoeffnet := anwendung.SetOeffnen(t.Context(),
		filepath.Join(kanonverzeichnis(), kunstfenstersatz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("%s öffnen: %+v", kunstfenstersatz, geoeffnet.Fehler)
	}

	unser, fenster := kunstfensterUnserer(t, anwendung, geoeffnet.SetID)

	referenz, err := ladeBild(filepath.Join(referenzort(t, kunstfenstersatz),
		referenzdatei(kunstfensterkarte)))
	if err != nil {
		t.Fatalf("Referenz: %v", err)
	}

	ganz, err := render.Ganzabstand(
		ausschnitt(t, unser, fenster), ausschnitt(t, referenz, fenster))
	if err != nil {
		t.Fatalf("Abstand: %v", err)
	}

	abstand := ganz.Abstand()

	t.Logf("%s, Feld %q bei %v: Abstand %.3f über %d Bildpunkte",
		kunstfensterkarte, kunstfensterfeld, fenster, abstand, ganz.Punkte)

	if abstand > kunstfensterdeckel {
		t.Errorf("das Kunstfenster liegt bei %.3f, über dem Deckel %.3f — "+
			"die Filterwahl aus #532 ist nicht mehr eingelöst",
			abstand, kunstfensterdeckel)
	}
}

// kunstfensterUnserer rastert die Karte und gibt Bild und Kunstfenster.
func kunstfensterUnserer(
	t *testing.T, anwendung *App, setID string,
) (image.Image, image.Rectangle) {
	t.Helper()

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	for _, karte := range karten.Karten {
		if kartenname(karte) != kunstfensterkarte {
			continue
		}

		vorschau := anwendung.Vorschau(t.Context(), karte.ID, 1)
		if vorschau.Fehler != nil {
			t.Fatalf("Vorschau: %+v", vorschau.Fehler)
		}

		for _, flaeche := range vorschau.Vorschau.Flaechen {
			if flaeche.Feld != kunstfensterfeld {
				continue
			}

			return bildAus(t, anwendung, vorschau.Vorschau.Bild), image.Rect(
				int(flaeche.Links), int(flaeche.Oben),
				int(flaeche.Links+flaeche.Breite), int(flaeche.Oben+flaeche.Hoehe))
		}

		t.Fatalf("%s hat kein Feld %q", kunstfensterkarte, kunstfensterfeld)
	}

	t.Fatalf("%s führt keine Karte %q", kunstfenstersatz, kunstfensterkarte)

	return nil, image.Rectangle{}
}

// ausschnitt schneidet ein Rechteck aus einem Bild.
//
// **Über `SubImage` und nicht über eine Kopie** — [render.Ganzabstand] rechnet
// mit `Bounds().Min` und kommt mit einem versetzten Ursprung zurecht.
func ausschnitt(t *testing.T, bild image.Image, rahmen image.Rectangle) image.Image {
	t.Helper()

	schneider, kann := bild.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !kann {
		t.Fatalf("%T kann nicht schneiden", bild)
	}

	return schneider.SubImage(rahmen)
}
