package msepaket_test

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein zweiter Griff nach demselben Bild geht nicht mehr auf die Platte.**
//
// # Was das kostet, wenn es fehlt
//
// Gemessen an einem echten Kartenrahmen aus `M15-Magic-Pack`
// (`bildmessung_test.go`, hinter `fremdpakete`): **13 ms je Lauf**, 1,18 s für
// neunzig Karten — und das für **ein** Bild. Eine Karte trägt ein Dutzend
// Ebenen (#187).
//
// # Wie der Test es feststellt
//
// Er **löscht die Datei** zwischen den beiden Griffen. Ein Zähler an einer
// Vorrichtung prüfte die Vorrichtung; eine gelöschte Datei prüft die Platte.
func TestEinZweiterGriffGehtNichtAufDiePlatte(t *testing.T) {
	wurzel := t.TempDir()
	pfad := filepath.Join(wurzel, "probe-bild.mse-style", "rahmen.png")
	legePNG(t, pfad, color.NRGBA{R: 42, A: 255})

	quelle := msepaket.NeueBildquelle(wurzel)

	if _, err := quelle.Bild("/probe-bild.mse-style/rahmen.png"); err != nil {
		t.Fatalf("erster Griff: %v", err)
	}

	if err := os.Remove(pfad); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	bild, err := quelle.Bild("/probe-bild.mse-style/rahmen.png")
	if err != nil {
		t.Fatalf("zweiter Griff: %v", err)
	}

	if ecke(t, bild).R != 42 {
		t.Errorf("Ecke %+v", ecke(t, bild))
	}
}

// **Ein Fehlschlag wird nicht gemerkt.**
//
// Das ist die Antwort auf die Frage, wann der Speicher ungültig wird — und sie
// ist dieselbe, die `schriften.paketdatei` seit #140 gibt: wer ein Paket
// **während** der Sitzung bezieht, findet es beim nächsten Versuch. Ein
// gemerkter Fehlschlag machte daraus einen, der bis zum Neustart gilt.
func TestEinFehlschlagWirdNichtGemerkt(t *testing.T) {
	wurzel := t.TempDir()
	quelle := msepaket.NeueBildquelle(wurzel)

	if _, err := quelle.Bild("/spaeter.mse-style/rahmen.png"); !errors.Is(err, msepaket.ErrKeinBild) {
		t.Fatalf("Fehler %v, erwartet ErrKeinBild", err)
	}

	// Jetzt kommt das Paket — so, wie ein Bezug es während der Sitzung ablegt.
	legePNG(t, filepath.Join(wurzel, "spaeter.mse-style", "rahmen.png"),
		color.NRGBA{G: 99, A: 255})

	bild, err := quelle.Bild("/spaeter.mse-style/rahmen.png")
	if err != nil {
		t.Fatalf("nach dem Bezug: %v", err)
	}

	if ecke(t, bild).G != 99 {
		t.Errorf("Ecke %+v", ecke(t, bild))
	}
}

// **Zwei Leser gleichzeitig**, und der Lauf hält es aus.
//
// Der Bogenleger rastert Karten nacheinander; das muss nicht so bleiben, und
// eine `map` ohne Sperre wäre dann kein langsamer Lauf, sondern ein Absturz.
// **Mit `-race` gefahren** — ohne ihn prüft dieser Test, dass zwei Läufe
// einander zufällig nicht treffen.
func TestZweiLeserGleichzeitig(t *testing.T) {
	wurzel := t.TempDir()

	for nummer := range 4 {
		legePNG(t, filepath.Join(wurzel,
			fmt.Sprintf("paket%d.mse-style", nummer), "rahmen.png"),
			color.NRGBA{R: uint8(nummer), A: 255})
	}

	quelle := msepaket.NeueBildquelle(wurzel)

	var warten sync.WaitGroup

	for lauf := range 8 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			verweis := fmt.Sprintf("/paket%d.mse-style/rahmen.png", lauf%4)

			for range 20 {
				if _, err := quelle.Bild(verweis); err != nil {
					t.Errorf("%s: %v", verweis, err)

					return
				}
			}
		}()
	}

	warten.Wait()
}

// **Die Obergrenze räumt ganz, nicht klug.**
//
// Eine Verdrängung nach Alter verlangt Buchführung über jeden Zugriff, und der
// Nutzen wäre einer für einen Fall, den es nicht gibt: ein Druckbogen benutzt
// ein gutes Dutzend Bilder immer wieder. Wer die Grenze **erreicht**, hat
// keinen Arbeitssatz mehr, der hineinpasst — und dann verschiebt eine
// Verdrängungsstrategie das Problem, statt es zu lösen.
//
// Der Test füllt über die Grenze und stellt fest, dass ein früher gelesenes
// Bild wieder von der Platte kommt: die gelöschte Datei fehlt jetzt.
func TestDieObergrenzeRaeumtGanz(t *testing.T) {
	wurzel := t.TempDir()

	erstes := filepath.Join(wurzel, "erstes.mse-style", "rahmen.png")
	legePNG(t, erstes, color.NRGBA{B: 7, A: 255})

	quelle := msepaket.NeueBildquelle(wurzel)

	if _, err := quelle.Bild("/erstes.mse-style/rahmen.png"); err != nil {
		t.Fatalf("erster Griff: %v", err)
	}

	if err := os.Remove(erstes); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	for nummer := range msepaket.GrenzeBildspeicher {
		name := fmt.Sprintf("fuellung%d.mse-style", nummer)
		legePNG(t, filepath.Join(wurzel, name, "rahmen.png"), color.NRGBA{A: 255})

		if _, err := quelle.Bild("/" + name + "/rahmen.png"); err != nil {
			t.Fatalf("Füllung %d: %v", nummer, err)
		}
	}

	if _, err := quelle.Bild("/erstes.mse-style/rahmen.png"); !errors.Is(err, msepaket.ErrKeinBild) {
		t.Errorf("Fehler %v, erwartet ErrKeinBild — der Speicher hat nicht geräumt", err)
	}
}

// **Der Arbeitssatz eines ganzen Sets passt hinein** — PERF02
// (#2352).
//
// Ein Bogen braucht ein Dutzend Bilder, ein Set mehr: über `Black` sammelten
// sich 255 verschiedene, und unter einer Grenze von 64 räumte die Quelle so
// oft, dass die Prüfung aller 3.788 Karten 28.837 Mal auf die Platte ging.
// **Hundert Bilder bleiben**, auch wenn ihre Dateien danach fehlen.
func TestEinArbeitssatzVonHundertBildernBleibt(t *testing.T) {
	wurzel := t.TempDir()
	quelle := msepaket.NeueBildquelle(wurzel)

	for nummer := range 100 {
		name := fmt.Sprintf("satz%d.mse-style", nummer)
		legePNG(t, filepath.Join(wurzel, name, "rahmen.png"), color.NRGBA{A: 255})

		if _, err := quelle.Bild("/" + name + "/rahmen.png"); err != nil {
			t.Fatalf("Bild %d: %v", nummer, err)
		}
	}

	if err := os.RemoveAll(wurzel); err != nil {
		t.Fatal(err)
	}

	for nummer := range 100 {
		if _, err := quelle.Bild(fmt.Sprintf("/satz%d.mse-style/rahmen.png", nummer)); err != nil {
			t.Fatalf("Bild %d kam nicht aus dem Speicher: %v", nummer, err)
		}
	}
}
