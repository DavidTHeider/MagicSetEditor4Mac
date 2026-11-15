// **Ein Paket mit einer ersetzten Datei** — Werkstück 3 von Phase 3.3
// (#1164).
//
// # Wofür
//
// Für die Vorschau einer **ungespeicherten** Änderung. §6 verlangt: „Eine
// Änderung an einem Feld ist in der Vorschau zu sehen, ohne dass jemand etwas
// bestätigt" — und Speichern ist eine Bestätigung. Eine Vorlage, die man erst
// schreiben muss, um sie zu sehen, macht aus jedem Versuch eine Änderung an
// allem, was dieses Paket benutzt.
//
// # Warum ein Ersatz und keine Kopie
//
// Eine Kopie des Pakets kostete bei jedem Tastendruck alles, was darin liegt —
// `magic-m15.mse-style` trägt vier Bilder, andere Pakete Hunderte. Der Ersatz
// legt **eine** Datei über den Baum und lässt den Rest, wo er ist.
package msepaket_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Die ersetzte Datei kommt aus dem Speicher.**
func TestEineErsetzteDateiKommtAusDemSpeicher(t *testing.T) {
	t.Parallel()

	for name, anlegen := range map[string]func(*testing.T) string{
		"Verzeichnis": probeverzeichnis,
		"Archiv":      probearchiv,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			paket, err := msepaket.OeffneMitErsatz(anlegen(t), "style",
				[]byte("mse version: 2.0.0\ngame: ersetzt\n"))
			if err != nil {
				t.Fatalf("öffnen: %v", err)
			}

			t.Cleanup(func() {
				if fehler := paket.Schliessen(); fehler != nil {
					t.Errorf("schließen: %v", fehler)
				}
			})

			inhalt, lesefehler := paket.Datei("style")
			if lesefehler != nil {
				t.Fatalf("lesen: %v", lesefehler)
			}

			if string(inhalt) != "mse version: 2.0.0\ngame: ersetzt\n" {
				t.Errorf("gelesen %q", inhalt)
			}
		})
	}
}

// **Auf der Platte steht davon nichts.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn der
// Ersatz die Datei einfach schriebe — und dann wäre jede Vorschau eine
// Änderung an allem, was dieses Paket benutzt.
func TestEinErsatzSchreibtNichts(t *testing.T) {
	t.Parallel()

	ort := probeverzeichnis(t)

	paket, err := msepaket.OeffneMitErsatz(ort, "style", []byte("ersetzt"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := paket.Schliessen(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if aufDerPlatte := dateiAus(t, ort, "style"); aufDerPlatte == "ersetzt" {
		t.Error("der Ersatz landete auf der Platte")
	}
}

// **Die anderen Dateien kommen weiterhin aus dem Paket.**
func TestDieAnderenDateienBleibenDieDesPakets(t *testing.T) {
	t.Parallel()

	paket, err := msepaket.OeffneMitErsatz(probeverzeichnis(t), "style", []byte("x"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	inhalt, lesefehler := paket.Datei("unterordner/skript")
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(inhalt) != "1 + 1\n" {
		t.Errorf("gelesen %q", inhalt)
	}
}

// **Der Kopf wird aus dem Ersatz gelesen, nicht aus der Datei.**
//
// Das ist der eigentliche Zweck: `Oeffne` liest beim Öffnen den Paketkopf, und
// wer eine Vorschau der geänderten Vorlage will, will ihn aus dem geänderten
// Text.
func TestDerKopfKommtAusDemErsatz(t *testing.T) {
	t.Parallel()

	paket, err := msepaket.OeffneMitErsatz(probeverzeichnis(t), "style",
		[]byte("mse version: 2.0.0\ngame: ersetzt\nshort name: Neu\n"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if paket.Spiel != "ersetzt" {
		t.Errorf("das Spiel ist %q", paket.Spiel)
	}
}

// **Ein Ersatz für eine Datei, die es nicht gibt, legt sie an.**
//
// Eine Vorlage kann eine Datei nennen, die noch niemand geschrieben hat; wer
// sie im Fenster anlegt, will sie vor dem Speichern sehen.
func TestEinErsatzFuerEineNeueDateiLegtSieAn(t *testing.T) {
	t.Parallel()

	paket, err := msepaket.OeffneMitErsatz(probeverzeichnis(t), "neu", []byte("frisch"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	inhalt, lesefehler := paket.Datei("neu")
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(inhalt) != "frisch" {
		t.Errorf("gelesen %q", inhalt)
	}
}

// **Ein Ausbruch wird auch hier abgewiesen.**
func TestEinAusbrechenderErsatzWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	if _, err := msepaket.OeffneMitErsatz(
		probeverzeichnis(t), "../draussen", []byte("x")); err == nil {
		t.Error("der Ausbruch kam durch")
	}
}

// **Ein Paket, das es nicht gibt, meldet sich.**
func TestEinFehlendesPaketMeldetSichBeimErsatz(t *testing.T) {
	t.Parallel()

	if _, err := msepaket.OeffneMitErsatz(
		filepath.Join(t.TempDir(), "nichts.mse-style"), "style", []byte("x")); err == nil {
		t.Error("das fehlende Paket kam durch")
	}
}

// **Und ein Ersatz, der den Kopf unlesbar macht, ebenso.**
//
// Der Kopf wird beim Öffnen gelesen — aus dem **Ersatz**, nicht von der
// Platte. Wer beim Tippen eine Zeile zerlegt, bekommt hier den Fehler und
// nicht eine Vorschau nach altem Stand.
func TestEinUnlesbarerErsatzMeldetSich(t *testing.T) {
	t.Parallel()

	// **Ein Ersatz an der Hauptdatei, den der Zerteiler abweist.** Welche
	// Gestalt ihn abweist, sagt `zerteiler.go`; hier zählt, dass der Fehler
	// **beim Öffnen** kommt und nicht später als leere Vorschau.
	_, err := msepaket.OeffneMitErsatz(probeverzeichnis(t), "style",
		[]byte(": ohne Schlüssel\n"))
	if err == nil {
		t.Error("der unlesbare Ersatz kam durch")
	}
}
