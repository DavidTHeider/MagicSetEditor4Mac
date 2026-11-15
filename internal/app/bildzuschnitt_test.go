package app_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Der Ausschnitt wird gewählt, nicht zugeteilt** —
// #576.
//
// Bis hierher schnitt das Einsetzen mit MSEs **Vorgabewahl** zu und war damit
// fertig. Wer ein Bild einsetzte, dessen Motiv nicht oben links sitzt, bekam
// den falschen Teil und hatte keinen Weg, das zu ändern.
//
// MSE öffnet dafür `ImageSliceWindow` (`src/gui/image_slice_window.cpp`, Zeilen
// 68–360, Fassung f9d9356 vom 2021-07-06): links das Ursprungsbild mit einem
// verschiebbaren Rahmen, rechts die Vorschau in Feldgröße.
//
// **Damit zerfällt der Weg in zwei Schritte.** `BildSetzen` und `BildAusDatei`
// **bereiten** vor — sie lesen die Datei und geben zurück, was das Fenster
// zeigen muss; `BildEinsetzen` **legt ab**. Das ist eine benannte Änderung an
// zwei bestehenden Bindungen (§15, Satz 4): vorher schrieben beide selbst.
func TestDieVorbereitungGibtQuelleZielUndVorgabewahl(t *testing.T) {
	wurzel := t.TempDir()
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	schreiber := &schreibprobe{}
	anwendung := bildanwendung(t, wurzel, schreiber, quelle)

	antwort := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Der Pfad kommt zurück**, denn der zweite Schritt braucht ihn — und nur
	// der Dialog kennt ihn.
	if antwort.Pfad != quelle {
		t.Errorf("Pfad %q", antwort.Pfad)
	}

	// **Die Quelle als Data-URI**, dieselbe Übereinkunft wie bei der
	// Kartenvorschau: unmittelbar in ein `src` einsetzbar.
	if !strings.HasPrefix(antwort.Quelle, "data:image/png;base64,") {
		t.Errorf("Quelle beginnt mit %.40q", antwort.Quelle)
	}

	// **Die Maße der Quelle sind die der Datei**, nicht die des Vorschaubildes.
	// Die Auswahl wird in ihnen gerechnet — ein Fenster, das in seinen eigenen
	// Punkten rechnete, gäbe bei jedem Maßstab ein anderes Bild.
	if antwort.QuelleBreite != 400 || antwort.QuelleHoehe != 100 {
		t.Errorf("Quelle %d × %d", antwort.QuelleBreite, antwort.QuelleHoehe)
	}

	if antwort.ZielBreite != 100 || antwort.ZielHoehe != 80 {
		t.Errorf("Ziel %d × %d", antwort.ZielBreite, antwort.ZielHoehe)
	}

	// **Und die Vorgabewahl ist MSEs** — 400 × 100 auf ein Feld 100 × 80 gibt
	// 125 × 100, oben links verankert ([render.Zuschnittflaeche]).
	vorgabe := render.Zuschnittflaeche(image.Rect(0, 0, 400, 100), 100, 80)
	if antwort.Auswahl.X != 0 || antwort.Auswahl.Y != 0 ||
		antwort.Auswahl.Breite != vorgabe.Dx() || antwort.Auswahl.Hoehe != vorgabe.Dy() {
		t.Errorf("Auswahl %+v, erwartet %v", antwort.Auswahl, vorgabe)
	}

	// **Vorbereiten schreibt nicht.** Wer abbricht, hat nichts geändert.
	if schreiber.feld != "" {
		t.Errorf("geschrieben wurde %q = %q", schreiber.feld, schreiber.wert)
	}
}

// **Abgebrochen ist kein Fehler und kein leeres Bild.**
func TestEineAbgebrocheneWahlBereitetNichtsVor(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	antwort := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if antwort.Fehler != nil || antwort.Pfad != "" || antwort.Quelle != "" {
		t.Errorf("%+v", antwort)
	}
}

// **Der zweite Schritt legt ab, was gewählt wurde.**
func TestDasEinsetzenNimmtDieGewaehlteFlaeche(t *testing.T) {
	wurzel := t.TempDir()
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	// **Ein Bild, dessen Hälften sich nicht verwechseln lassen.** Der erste
	// Anlauf nahm die Farbrampe aus [schreibeJPEG] — sie färbt nach `x % 256`
	// und **schlägt** bei 256 um: bei Punkt 277 stand wieder ein niedriger
	// Rotwert, und der Test las daraus „die linke Hälfte".
	halbbild := filepath.Join(t.TempDir(), "haelften.png")
	schreibeHaelften(t, halbbild)

	schreiber := &schreibprobe{}
	anwendung := bildanwendung(t, wurzel, schreiber, halbbild)

	// **Die rechte Hälfte**, nicht die Vorgabewahl von links.
	antwort := anwendung.BildEinsetzen(t.Context(), "1", "1", "kunst", halbbild,
		app.AuswahlDTO{X: 275, Y: 0, Breite: 125, Hoehe: 100})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if schreiber.wert != "kunst1" {
		t.Errorf("geschrieben wurde %q", schreiber.wert)
	}

	abgelegt := filepath.Join(wurzel, "7", "kunst1")

	datei, err := os.Open(abgelegt) //nolint:gosec // eben geschriebener Testpfad
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // Testpfad

	bild, _, err := image.Decode(datei)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if bild.Bounds().Dx() != 100 || bild.Bounds().Dy() != 80 {
		t.Fatalf("abgelegt ist %v", bild.Bounds())
	}

	// **Und es ist wirklich die rechte Hälfte** — die weiße.
	if r, _, _, _ := bild.At(2, 40).RGBA(); r>>8 < 200 {
		t.Errorf("die linke, schwarze Hälfte wurde abgelegt: Rot %d", r>>8)
	}
}

// schreibeHaelften legt ein Bild an: links schwarz, rechts weiß.
func schreibeHaelften(t *testing.T, pfad string) {
	const breite, hoehe = 400, 100

	t.Helper()

	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))

	for x := range breite {
		farbe := color.RGBA{A: 255}
		if x >= breite/2 {
			farbe = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		}

		for y := range hoehe {
			bild.Set(x, y, farbe)
		}
	}

	datei, err := os.Create(pfad) //nolint:gosec // Testpfad
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // Testpfad

	if err := png.Encode(datei, bild); err != nil {
		t.Fatalf("kodieren: %v", err)
	}
}

// **Wer nichts ändert, bekommt genau das bisherige Ergebnis.**
//
// Das ist der Abnahmekasten aus #576, der am leichtesten stillschweigend
// bricht: die Vorgabewahl ist portiert, und ein Dialog davor darf sie nicht
// verschieben.
func TestOhneAenderungBleibtEsBeimBisherigenBild(t *testing.T) {
	wurzel := t.TempDir()
	quelle := filepath.Join(t.TempDir(), "urlaub.jpg")
	schreibeJPEG(t, quelle)

	anwendung := bildanwendung(t, wurzel, &schreibprobe{}, quelle)

	vorbereitet := anwendung.BildSetzen(t.Context(), "1", "kunst")
	if vorbereitet.Fehler != nil {
		t.Fatalf("%+v", vorbereitet.Fehler)
	}

	if antwort := anwendung.BildEinsetzen(
		probekontext(t), "1", "1", "kunst", quelle, vorbereitet.Auswahl,
	); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	ueberDenDialog := lieszurueck(t, filepath.Join(wurzel, "7", "kunst1"))

	// Und derselbe Weg ohne Dialog — die Vorgabewahl unmittelbar.
	zweiter := t.TempDir()
	ohneDialog := bildanwendung(t, zweiter, &schreibprobe{}, quelle)

	if antwort := ohneDialog.BildEinsetzen(probekontext(t), "1", "1", "kunst", quelle,
		app.AuswahlDTO{Breite: 125, Hoehe: 100}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if gab := lieszurueck(t, filepath.Join(zweiter, "7", "kunst1")); gab != ueberDenDialog {
		t.Error("der Dialog verschiebt die Vorgabewahl")
	}
}

// lieszurueck gibt den Inhalt einer abgelegten Bilddatei.
func lieszurueck(t *testing.T, pfad string) string {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // eben geschriebener Testpfad
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	return string(roh)
}

// **Die Grenzregel liegt in Go, weil sie portiert ist.**
//
// Eine zweite Fassung in TypeScript wäre eine zweite Antwort auf dieselbe Frage
// und liefe auseinander, sobald jemand eine davon nachzieht (§15). Der Preis
// steht daneben: das Fenster fragt bei jeder Bewegung über die Bindung.
func TestDieZuschnittgrenzeReichtMSEsRegelDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	faelle := map[string]struct {
		auswahl   app.AuswahlDTO
		frei      bool
		bevorzugt string
		erwartet  app.AuswahlDTO
	}{
		// 100 × 100 passt in die Quelle, die Ecke rutscht auf (300, 0) — und
		// **danach** greift das Seitenverhältnis: `diff = 100·80 − 100·100`
		// ist negativ, also gibt die Höhe nach, `100 − 2000/100 = 80`.
		"eine Auswahl über den Rand hinaus": {
			auswahl:  app.AuswahlDTO{X: 380, Y: 90, Breite: 100, Hoehe: 100},
			erwartet: app.AuswahlDTO{X: 300, Y: 0, Breite: 100, Hoehe: 80},
		},
		"die gezogene Breite bleibt": {
			auswahl:   app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 50},
			bevorzugt: "breite",
			erwartet:  app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 160},
		},
		"die gezogene Höhe bleibt": {
			auswahl:   app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 50},
			bevorzugt: "hoehe",
			// `diff = 200·80 − 50·100 = 11000`, `11000/80 = 137` — ganzzahlig
			// **abgeschnitten**, nicht gerundet: 200 − 137 = 63.
			erwartet: app.AuswahlDTO{X: 0, Y: 0, Breite: 63, Hoehe: 50},
		},
		// **Ein unbekanntes Wort heißt „keines".** Die Bindung nimmt eine
		// Zeichenkette entgegen; ein Fehlerweg dafür hätte keinen Nutzen.
		"ein unbekanntes Wort": {
			auswahl:   app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 50},
			bevorzugt: "quatsch",
			erwartet:  app.AuswahlDTO{X: 0, Y: 0, Breite: 63, Hoehe: 50},
		},
		// **Frei ist MSEs „force to fit"** — das Seitenverhältnis fällt weg.
		"ohne Bindung ans Seitenverhältnis": {
			auswahl:  app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 50},
			frei:     true,
			erwartet: app.AuswahlDTO{X: 0, Y: 0, Breite: 200, Hoehe: 50},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gab := anwendung.Zuschnittgrenze(400, 100, 100, 80, fall.auswahl, fall.frei, fall.bevorzugt)
			if gab != fall.erwartet {
				t.Errorf("gab %+v, erwartet %+v", gab, fall.erwartet)
			}
		})
	}
}

// **Eine unlesbare Kennung wird auch beim Ablegen abgewiesen.**
//
// Dieselbe Wache für beide Schritte, an einer Stelle ([App.bildziel]) — sonst
// prüfte der erste, und der zweite ließe durch, was ihn unmittelbar erreicht.
func TestBildEinsetzenWeistEineUnlesbareKennungAb(t *testing.T) {
	anwendung := bildanwendung(t, t.TempDir(), &schreibprobe{}, "")

	if antwort := anwendung.BildEinsetzen(t.Context(), "1", "keine zahl", "kunst", "/tmp/egal.png",
		app.AuswahlDTO{Breite: 10, Hoehe: 10}); antwort.Fehler == nil {
		t.Error("eine unlesbare Kennung kam durch")
	}
}

// **Und das Ziehen ebenso** — dieselbe Regel, dieselbe Naht.
//
// Das Fenster rechnet Bildschirmpunkte in Quellpunkte um und reicht sie
// herein; was MSE daraus macht, steht in Go
// (`ImageSliceSelector::onMotion`, Zeilen 609–646, Fassung f9d9356 vom
// 2021-07-06).
func TestDasZuschnittziehenReichtMSEsRegelDurch(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")
	ganz := app.AuswahlDTO{Breite: 200, Hoehe: 160}

	t.Run("verschieben", func(t *testing.T) {
		t.Parallel()

		gab := anwendung.Zuschnittziehen(400, 300, 100, 80, ganz, false, 50, 20, 0, 0)
		if gab != (app.AuswahlDTO{X: 50, Y: 20, Breite: 200, Hoehe: 160}) {
			t.Errorf("gab %+v", gab)
		}
	})

	t.Run("eine Ecke ziehen", func(t *testing.T) {
		t.Parallel()

		gab := anwendung.Zuschnittziehen(400, 300, 100, 80, ganz, false, 40, 40, 1, 1)
		if gab.Breite <= ganz.Breite || gab.Hoehe <= ganz.Hoehe {
			t.Errorf("die Ecke vergrößerte nicht: %+v", gab)
		}
	})
}
