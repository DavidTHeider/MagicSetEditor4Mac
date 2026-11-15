package service_test

// **Das Profilbild lesen und in die Schranken bringen** — Werkstück 8 von
// Phase 4.12
// (#1351).
//
// **Eine eigene Datei neben `nutzerbild_test.go`**, damit beide unter der
// Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort steht, wie
// ein Bild **hinkommt**, hier, wie es wieder **herauskommt**.

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// bildflaeche baut ein Bild der gewünschten Maße.
func bildflaeche(breite, hoehe int) image.Image {
	bild := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))
	bild.SetNRGBA(0, 0, color.NRGBA{R: 200, A: 255})

	return bild
}

// **Ein Bild kommt so zurück, wie es hineinging.**
func TestEinProfilbildLaesstSichLesen(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	name, fehler := ablage.NutzerbildAblegen(7, bildflaeche(20, 10))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.NutzerbildLesen(name)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen == nil {
		t.Fatal("es kam kein Bild zurück")
	}

	if masse := gelesen.Bounds(); masse.Dx() != 20 || masse.Dy() != 10 {
		t.Errorf("die Maße sind %dx%d", masse.Dx(), masse.Dy())
	}
}

// **Ein großes Bild wird auf die Kante gebracht — mit seinem
// Seitenverhältnis.**
//
// **Beim Ablegen und nicht beim Lesen**: ein Profilbild ist ein Abzeichen
// neben einem Namen, und wer die Originalgröße behielte, entschlüsselte für
// jede Anzeige ein Megabyte.
func TestEinGrossesProfilbildWirdVerkleinert(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	faelle := []struct {
		name                  string
		breite, hoehe         int
		zielbreite, zielhoehe int
	}{
		{"breiter als hoch", 1000, 500, service.Nutzerbildkante, service.Nutzerbildkante / 2},
		{"höher als breit", 500, 1000, service.Nutzerbildkante / 2, service.Nutzerbildkante},
		{"quadratisch", 800, 800, service.Nutzerbildkante, service.Nutzerbildkante},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			name, fehler := ablage.NutzerbildAblegen(7, bildflaeche(fall.breite, fall.hoehe))
			if fehler != nil {
				t.Fatalf("ablegen: %v", fehler)
			}

			gelesen, lesefehler := ablage.NutzerbildLesen(name)
			if lesefehler != nil {
				t.Fatalf("lesen: %v", lesefehler)
			}

			masse := gelesen.Bounds()
			if masse.Dx() != fall.zielbreite || masse.Dy() != fall.zielhoehe {
				t.Errorf("die Maße sind %dx%d statt %dx%d",
					masse.Dx(), masse.Dy(), fall.zielbreite, fall.zielhoehe)
			}
		})
	}
}

// **Ein sehr flaches Bild behält mindestens eine Zeile.**
//
// **Ganzzahlig gerechnet ergäbe ein 1000×1 sonst die Höhe null** — und ein
// Bild ohne Fläche ist keines.
func TestEinSehrFlachesProfilbildBehaeltEineZeile(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	name, fehler := ablage.NutzerbildAblegen(7, bildflaeche(1000, 1))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	gelesen, lesefehler := ablage.NutzerbildLesen(name)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if masse := gelesen.Bounds(); masse.Dy() < 1 {
		t.Errorf("die Höhe ist %d", masse.Dy())
	}
}

// **Ein fehlendes Bild ist kein Fehler, sondern keines.**
//
// **Dieselbe Zusage wie bei [Vorschaubildwerk.Kartenbild]**: `nil` ohne
// Fehler heißt „es gibt keines", und das ist der Normalfall — die meisten
// Menschen haben kein Profilbild. Ein Fehler wäre eine Meldung, die niemand
// braucht.
func TestEinFehlendesProfilbildIstKeinFehler(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		wurzel, name string
	}{
		"die Datei liegt nicht da": {wurzel: t.TempDir(), name: "nutzer-7.png"},
		"kein Name":                {wurzel: t.TempDir(), name: ""},
		"nur Leerzeichen":          {wurzel: t.TempDir(), name: "   "},
		"keine Wurzel":             {wurzel: "", name: "nutzer-7.png"},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			bild, fehler := service.NeueBildablage(fall.wurzel).NutzerbildLesen(fall.name)
			if fehler != nil {
				t.Errorf("gemeldet wurde: %v", fehler)
			}

			if bild != nil {
				t.Error("es kam ein Bild zurück")
			}
		})
	}
}

// **Eine Datei, die kein Bild ist, meldet sich — und schweigt nicht.**
//
// **Der Unterschied zum fehlenden Bild ist die ganze Zusage**: „es gibt
// keines" ist eine Auskunft, „da liegt etwas Kaputtes" ein Befund. Wer beides
// gleich behandelte, ließe eine beschädigte Ablage unbemerkt.
func TestEineUnlesbareProfilbilddateiMeldetSich(t *testing.T) {
	t.Parallel()

	wurzel := t.TempDir()
	ablage := service.NeueBildablage(wurzel)

	if err := os.MkdirAll(ablage.Nutzerverzeichnis(), 0o750); err != nil {
		t.Fatalf("Verzeichnis anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ablage.Nutzerverzeichnis(), "nutzer-7.png"),
		[]byte("kein Bild"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, fehler := ablage.NutzerbildLesen("nutzer-7.png"); fehler == nil {
		t.Error("die kaputte Datei kam durch")
	}
}

// **Ein Abzeichen liegt neben dem Profilbild und trägt seine Profilkennung.**
//
// **Der Name ist die ganze Zusage**: `profil-7-….png` neben `nutzer-3-….png`. Ein
// Abzeichen gehört einem **Profil**, ein Bild einem **Menschen**, und wer
// beide gleich benennte, überschriebe das eine mit dem anderen.
func TestEinAbzeichenLiegtNebenDemProfilbild(t *testing.T) {
	t.Parallel()

	ablage := service.NeueBildablage(t.TempDir())

	bild, fehler := ablage.NutzerbildAblegen(3, bildflaeche(20, 20))
	if fehler != nil {
		t.Fatalf("ablegen: %v", fehler)
	}

	badge, badgefehler := ablage.ProfilbadgeAblegen(7, bildflaeche(20, 20))
	if badgefehler != nil {
		t.Fatalf("ablegen: %v", badgefehler)
	}

	if !strings.HasPrefix(bild, "nutzer-3-") || !strings.HasPrefix(badge, "profil-7-") {
		t.Errorf("die Namen sind %q und %q", bild, badge)
	}

	// **Beide liegen wirklich da** — und beide lassen sich lesen.
	for _, name := range []string{bild, badge} {
		gelesen, lesefehler := ablage.NutzerbildLesen(name)
		if lesefehler != nil || gelesen == nil {
			t.Errorf("%s: %v", name, lesefehler)
		}
	}

	// **Und das Wegnehmen des einen lässt das andere liegen.**
	if err := ablage.ProfilbadgeWegnehmen(7); err != nil {
		t.Fatalf("wegnehmen: %v", err)
	}

	if gelesen, _ := ablage.NutzerbildLesen(badge); gelesen != nil { //nolint:errcheck // hier zählt nur, ob es weg ist
		t.Error("das Abzeichen liegt noch")
	}

	if gelesen, _ := ablage.NutzerbildLesen(bild); gelesen == nil { //nolint:errcheck // dito
		t.Error("das Profilbild ging mit")
	}
}
