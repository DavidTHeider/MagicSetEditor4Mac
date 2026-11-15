// Der Probelauf an den Nähten ohne Fenster — UXR-F7
// (#2015).
//
// **Gegen den echten Lauf desselben Sets**: der Probelauf sagt, was der Lauf
// danach tut, und die Zahl, die er nennt, ist die, die danach auf der Platte
// liegt.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Probelauf des Bogens nennt, was der Lauf legt.**
func TestProbeBogenNenntDieZahlenDesLegens(t *testing.T) {
	werk, setID := bogenwerk(t)

	probe, fehler := werk.ProbeBogen(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Probelauf: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "bogen.pdf")

	gelegt, fehler := werk.LegeBogen(probekontext(t), setID, ziel)
	if fehler != nil {
		t.Fatalf("legen: %v", fehler)
	}

	if probe.Karten != gelegt.Karten || probe.Seiten != gelegt.Seiten {
		t.Errorf("Probelauf %d Karten, %d Seiten; Lauf %d und %d",
			probe.Karten, probe.Seiten, gelegt.Karten, gelegt.Seiten)
	}

	// **Das Standardmaß der Flächen ohne Fenster**: A4, neun je Seite, 300 dpi.
	if probe.JeSeite != 9 || probe.DPI != 300 || probe.Seitenbreite != 210 ||
		probe.Seitenhoehe != 297 || probe.Kartenbreite != 63 || !probe.Schnittmarken {
		t.Errorf("%+v", probe)
	}
}

// **Dasselbe Nein: ein Set, das es nicht gibt.**
func TestProbeBogenFuerEinSetDasEsNichtGibt(t *testing.T) {
	werk, _ := bogenwerk(t)

	_, probefehler := werk.ProbeBogen(probekontext(t), 4711)
	_, lauffehler := werk.LegeBogen(probekontext(t), 4711, filepath.Join(t.TempDir(), "b.pdf"))

	if !errors.Is(probefehler, domain.ErrNotFound) || lauffehler == nil ||
		probefehler.Error() != lauffehler.Error() {
		t.Errorf("Probelauf %v, Lauf %v", probefehler, lauffehler)
	}
}

// **Der Probelauf der Vorlage nennt genau die Größe, die danach auf der Platte
// liegt — und legt vorher nichts an.**
func TestDerVorlagenwegProbelaufNenntDieGroesseDerDatei(t *testing.T) {
	weg, setID := vorlagenset(t)

	ziel := filepath.Join(t.TempDir(), "ausfuhr.txt")

	probe, fehler := weg.Probelauf(probekontext(t), setID, "probe-liste.mse-export-template", ziel, nil)
	if fehler != nil {
		t.Fatalf("Probelauf: %v (%v)", fehler, probe.Befunde)
	}

	if _, err := os.Stat(ziel); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("der Probelauf hat die Datei angelegt (%v)", err)
	}

	if _, err := weg.Ausfuehren(probekontext(t), setID, "probe-liste.mse-export-template",
		ziel, nil); err != nil {
		t.Fatalf("ausführen: %v", err)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Datei fehlt: %v", err)
	}

	if probe.Bytes != len(roh) || probe.Karten != 2 || probe.Dateiname != "ausfuhr.txt" {
		t.Errorf("Probelauf %+v, Datei mit %d Bytes", probe, len(roh))
	}

	if probe.Befunde == nil {
		t.Error("Befunde ist nil statt einer leeren Liste")
	}
}

// **Dasselbe Nein: eine Vorlage, die es nicht gibt.**
func TestDerVorlagenwegProbelaufWeistEineUnbekannteVorlageAb(t *testing.T) {
	weg, setID := vorlagenset(t)

	_, probefehler := weg.Probelauf(probekontext(t), setID, "gibtsnicht.mse-export-template", "", nil)
	_, lauffehler := weg.Ausfuehren(probekontext(t), setID, "gibtsnicht.mse-export-template",
		filepath.Join(t.TempDir(), "raus.txt"), nil)

	if !errors.Is(probefehler, domain.ErrNotFound) || lauffehler == nil ||
		probefehler.Error() != lauffehler.Error() {
		t.Errorf("Probelauf %v, Lauf %v", probefehler, lauffehler)
	}
}
