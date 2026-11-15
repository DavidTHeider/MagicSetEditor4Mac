package schriften_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// paketmitSchrift legt ein Paket an, das seine Schrift **mitbringt**.
//
// Nachgebaut nach `M15-Magic-Pack`: dort liegen die Schriften unter
// `Magic - Fonts/`, also **neben** den Paketen und nicht in ihnen.
func paketMitSchrift(t *testing.T, dateiname string) string {
	t.Helper()

	wurzel := t.TempDir()

	unterordner := filepath.Join(wurzel, "M15-Magic-Pack", "Magic - Fonts")
	if err := os.MkdirAll(unterordner, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(unterordner, dateiname),
		systemschrift(t), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return wurzel
}

// **Eine Schrift aus einem bezogenen Paket wird gefunden**, ohne dass sie im
// System installiert ist.
//
// Ohne diesen Weg gilt: wer das Paket bezieht, bekommt Vorlagen, die eine
// Schrift nennen, die niemand installiert hat — und je Textebene einen Befund.
// Die Karte sieht dann leer aus, wo sie beschriftet sein sollte.
func TestEineSchriftAusEinemPaketWirdGefunden(t *testing.T) {
	suche := schriften.Neue(nil).MitPaketen(paketMitSchrift(t, "mplantin.ttf"))

	schrift, err := suche.Suche("MPlantin", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if schrift.Zeilenhoehe() <= 0 {
		t.Errorf("Zeilenhöhe %g", schrift.Zeilenhoehe())
	}
}

// **Die Reihenfolge ist System vor Paket**, und sie ist geprüft.
//
// Was jemand selbst installiert hat, ist die Schrift, die er meint: eine
// Paketschrift gleichen Namens ersetzte sie sonst, ohne dass jemand es
// angeordnet hätte.
func TestDasSystemGehtVorDemPaket(t *testing.T) {
	system := verzeichnisMitSchrift(t, "MPlantin.ttf")
	paket := paketMitSchrift(t, "MPlantin.ttf")

	suche := schriften.Neue([]string{system}).MitPaketen(paket)

	ort, err := suche.Fundort("MPlantin", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if filepath.Dir(ort) != system {
		t.Errorf("gefunden unter %q, erwartet im Systemort %q", ort, system)
	}
}

// **Die Schriften bleiben, wo sie liegen.**
//
// Sie ins System zu kopieren wäre eine Installation, die niemand angeordnet
// hat, und sie überlebte die Deinstallation der Anwendung.
func TestEinePaketschriftWirdNichtInsSystemKopiert(t *testing.T) {
	system := t.TempDir()
	paket := paketMitSchrift(t, "mplantin.ttf")

	if _, err := schriften.Neue([]string{system}).MitPaketen(paket).Suche("MPlantin", false); err != nil {
		t.Fatalf("suchen: %v", err)
	}

	eintraege, err := os.ReadDir(system)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(eintraege) != 0 {
		t.Errorf("im Systemort liegt jetzt etwas: %v", eintraege)
	}
}

// Ein Paketverzeichnis, das es nicht gibt, ist kein Fehler.
//
// Wer nie ein Paket bezogen hat, hat keines — und eine Schriftsuche, die daran
// scheiterte, ließe die Anwendung ohne Schriften dastehen.
func TestEinFehlendesPaketverzeichnisIstKeinFehler(t *testing.T) {
	system := verzeichnisMitSchrift(t, "MPlantin.ttf")

	suche := schriften.Neue([]string{system}).
		MitPaketen(filepath.Join(t.TempDir(), "gibtsnicht"))

	if _, err := suche.Suche("MPlantin", false); err != nil {
		t.Errorf("suchen: %v", err)
	}
}

// Ohne Paketwurzel bleibt der Suchweg der alte.
func TestOhnePaketwurzelBleibtEsBeimSystem(t *testing.T) {
	suche := schriften.Neue([]string{verzeichnisMitSchrift(t, "MPlantin.ttf")})

	if _, err := suche.Suche("Gibtsnicht", false); err == nil {
		t.Error("eine Schrift, die es nicht gibt, wurde gefunden")
	}
}

// Eine Schrift, die es weder im System noch im Paket gibt, wird benannt.
func TestEineImPaketFehlendeSchriftWirdBenannt(t *testing.T) {
	suche := schriften.Neue([]string{t.TempDir()}).MitPaketen(paketMitSchrift(t, "mplantin.ttf"))

	if _, err := suche.Suche("Beleren", false); err == nil {
		t.Error("kein Fehler")
	}
}

// Ein Verzeichnis der Paketwurzel, das sich nicht lesen lässt, hält nicht an.
//
// **Die Wurzel gehört den Nutzenden**, und ein Bezug kann halb daliegen. Wer
// daran scheiterte, fände auch die Schriften nicht, die lesbar sind.
func TestEinUnlesbaresPaketverzeichnisHaeltNichtAn(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root ist jedes Verzeichnis lesbar")
	}

	wurzel := paketMitSchrift(t, "mplantin.ttf")

	zu := filepath.Join(wurzel, "zu")
	if err := os.Mkdir(zu, 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chmod(zu, 0o700); err != nil { // #nosec G302 -- Verzeichnis
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	if _, err := schriften.Neue(nil).MitPaketen(wurzel).Suche("MPlantin", false); err != nil {
		t.Errorf("suchen: %v", err)
	}
}

// Eine Datei, die keine Schrift ist, wird übergangen.
func TestEineFremdeDateiImPaketWirdUebergangen(t *testing.T) {
	wurzel := paketMitSchrift(t, "mplantin.ttf")

	if err := os.WriteFile(filepath.Join(wurzel, "LIESMICH.txt"), []byte("nichts"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := schriften.Neue(nil).MitPaketen(wurzel).Suche("LIESMICH", false); err == nil {
		t.Error("eine Textdatei ist als Schrift durchgegangen")
	}
}

// Der Fundort einer Schrift, die es nicht gibt, ist ein Fehler.
func TestEinFundortOhneSchriftIstEinFehler(t *testing.T) {
	if _, err := schriften.Neue(nil).Fundort("Gibtsnicht", false); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Paketschrift zeichnet auch.**
//
// Gefunden zu werden ist die halbe Zusage. Wer nur `Zeilenhoehe() > 0` prüft,
// weiß, dass eine Datei gelesen wurde — nicht, dass daraus Glyphen entstehen.
// Genau das ist der Punkt von #140: ohne diesen Weg sieht die Karte leer aus,
// wo sie beschriftet sein sollte.
func TestEinePaketschriftZeichnetAuch(t *testing.T) {
	schrift, err := schriften.Neue(nil).
		MitPaketen(paketMitSchrift(t, "mplantin.ttf")).Suche("MPlantin", false)
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	glyphen, ist := schrift.(render.Glyphenschrift)
	if !ist {
		t.Fatal("die gefundene Schrift kann keine Glyphen")
	}

	maske, _, da := glyphen.Maske('A', 48)
	if !da {
		t.Fatal("»A« ist nicht in der Schrift")
	}

	flaeche := maske.Bounds()
	if flaeche.Dx() <= 0 || flaeche.Dy() <= 0 {
		t.Errorf("die Maske ist leer: %v", flaeche)
	}

	// Und sie ist nicht durchweg durchsichtig: ein „A" hat Deckung.
	gedeckt := false

	for y := flaeche.Min.Y; y < flaeche.Max.Y && !gedeckt; y++ {
		for x := flaeche.Min.X; x < flaeche.Max.X; x++ {
			if _, _, _, deckung := maske.At(x, y).RGBA(); deckung > 0 {
				gedeckt = true

				break
			}
		}
	}

	if !gedeckt {
		t.Error("die Maske ist überall durchsichtig — es wird nichts gezeichnet")
	}
}
