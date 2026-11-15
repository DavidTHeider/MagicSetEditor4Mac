package service

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Schleife endet, auch wenn jeder Name vergeben ist.**
//
// MSE hat hier keine Schranke: dort ist die Dateiliste im Speicher, und `find`
// scheitert nicht (`Package::newFileName`, `src/util/io/package.cpp`, Zeilen
// 248–267, Fassung f9d9356 vom 2021-07-06). Hier fragt jeder Durchlauf die
// Platte — und ein Verzeichnis, das jeden Namen belegt meldet, drehte ohne
// Schranke für immer.
//
// **Ein interner Test, weil er `freierName` unmittelbar ruft.** Hunderttausend
// Dateien anzulegen wäre kein Test, sondern ein Belastungslauf; die Zahl selbst
// ist so hoch, dass sie im Betrieb niemanden trifft — das größte kanonische Set
// trägt 3.837 Bilder.
//
// **Die Schranke wird nicht mehr am Paket gesenkt** (#1163): sie war ein
// Zustand, den alle Läufe teilen, und die Läufe dieses Pakets sind
// nebenläufig. Ein Lauf, der sie herunterdrehte, ließ jeden anderen, der
// gerade ein Bild ablegte, an ihr scheitern — zehn rote Läufe an einem Tag,
// keiner davon mit Schranken befasst. `freierName` bekommt sie jetzt als Wert.
func TestDieSucheNachEinemNamenGibtIrgendwannAuf(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	for _, name := range []string{"image1", "image2", "image3"} {
		if err := os.WriteFile(filepath.Join(ort, name), []byte("belegt"), 0o600); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	_, err := freierName(ort, "image", "", 3)
	if err == nil {
		t.Fatal("die Suche gab nicht auf")
	}

	// **Ein Konflikt und kein Validierungsfehler**: die Eingabe ist in Ordnung,
	// der Ort ist voll.
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("%v — erwartet war ErrConflict", err)
	}
}

// **Und die Schranke gilt auch für das Ablegen selbst** — der Fehler kommt
// durch und wird nicht zu einem leeren Namen.
func TestDasAblegenReichtDieAufgabeDurch(t *testing.T) {
	t.Parallel()

	ablage := NeueBildablage(t.TempDir()).MitHoechsterBildzahl(1)

	bild := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	if _, err := ablage.Ablegen(7, "image", bild); err != nil {
		t.Fatalf("das erste Bild: %v", err)
	}

	if name, err := ablage.Ablegen(7, "image", bild); err == nil {
		t.Errorf("das zweite Bild kam als %q durch", name)
	}
}
