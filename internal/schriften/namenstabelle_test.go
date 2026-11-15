package schriften_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
)

// **Eine Schrift wird über ihren *eigenen* Namen gefunden, nicht über ihren
// Dateinamen** (#240).
//
// # Der Befund
//
// `magic-m15.mse-style` verlangt `Beleren Bold` und `MPlantin`. Im bezogenen
// `M15-Magic-Pack` liegen sie als `beleren-bold_P1.01.ttf` und `mplantin.ttf` —
// die zweite fände der Dateinamensvergleich, die erste nie.
//
// Am laufenden Fenster war das zu sehen:
//
//	Ebene 6: die Schrift Beleren Bold ist nicht zu finden
//
// Das war der letzte Grund, warum eine gezeichnete Karte keinen Text trug.
//
// # Warum die Namenstabelle und nicht eine Namensliste
//
// Der Dateiname ist eine Angabe des Menschen, der die Datei abgelegt hat; der
// Name **in** der Datei ist die Angabe der Schrift über sich selbst. Genau den
// meint eine Vorlage, denn MSE reicht ihn an das Betriebssystem weiter, und das
// sieht ebenfalls in die Datei.
//
// Eine Tabelle „Dateiname → Schriftname" wäre eine Liste, die mit jeder neuen
// Sammlung veraltet.
func TestEineSchriftWirdUeberIhrenEigenenNamenGefunden(t *testing.T) {
	verzeichnis := t.TempDir()

	// **Eine echte Schriftdatei unter einem irreführenden Namen.** Der
	// Grundbestand von macOS liegt in `/System/Library/Fonts`; welche davon
	// vorhanden ist, entscheidet der Rechner — deshalb wird gesucht statt
	// angenommen.
	quelle, name := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if err := schreibe(t, verzeichnis, "voellig-anders_9.9.ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	suche := schriften.Neue([]string{verzeichnis})

	if _, err := suche.Suche(name, false); err != nil {
		t.Errorf("die Schrift %q wurde unter ihrem eigenen Namen nicht gefunden: %v",
			name, err)
	}
}

// **Der Dateiname gilt weiterhin — und zuerst.**
//
// Die Gegenprobe: wer die Namenstabelle vorzöge, müsste jede Datei jedes Ortes
// öffnen, bevor er die erste Antwort geben kann. Der Dateiname ist der billige
// Weg und trifft in den meisten Fällen zu.
func TestDerDateinameGiltWeiterhin(t *testing.T) {
	verzeichnis := t.TempDir()

	quelle, _ := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if err := schreibe(t, verzeichnis, "Wunschname.ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := schriften.Neue([]string{verzeichnis}).Suche("Wunschname", false); err != nil {
		t.Errorf("%v", err)
	}
}

// **Was es nirgends gibt, bleibt ein benannter Fehler.**
//
// Die zweite Gegenprobe: die Namenstabelle darf nicht dazu führen, dass
// irgendeine Schrift zurückkommt. Eine Karte mit der falschen Schrift sieht
// richtig aus und ist es nicht.
func TestEinUnbekannterNameBleibtEinFehler(t *testing.T) {
	verzeichnis := t.TempDir()

	quelle, _ := ersteSystemschrift(t)

	roh, err := os.ReadFile(quelle) //nolint:gosec // Pfad aus der Verzeichnisliste
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if err := schreibe(t, verzeichnis, "irgendwas.ttf", roh); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if _, err := schriften.Neue([]string{verzeichnis}).Suche("Gibt Es Nicht", false); err == nil {
		t.Error("kein Fehler — irgendeine Schrift kam zurück")
	}
}

// schreibe legt eine Schriftdatei unter einem gewählten Namen ab.
//
// **Über einen Helfer und nicht dreimal direkt**, weil `gosec` sonst dreimal
// eine Wegüberschreitung meldet: es sieht, dass der Inhalt aus einer gelesenen
// Datei stammt, und nicht, dass das Ziel ein `t.TempDir()` ist.
func schreibe(t *testing.T, verzeichnis, name string, inhalt []byte) error {
	t.Helper()

	ziel, err := os.OpenRoot(verzeichnis)
	if err != nil {
		return err
	}

	defer ziel.Close() //nolint:errcheck // ein Testverzeichnis

	datei, err := ziel.Create(name)
	if err != nil {
		return err
	}

	defer datei.Close() //nolint:errcheck // ein Testverzeichnis

	_, err = datei.Write(inhalt)

	return err
}

// ersteSystemschrift gibt eine vorhandene `.ttf` oder `.otf` samt ihrem
// **eigenen** Namen.
func ersteSystemschrift(t *testing.T) (pfad, name string) {
	t.Helper()

	for _, ort := range schriften.Orte() {
		eintraege, err := os.ReadDir(ort)
		if err != nil {
			continue
		}

		for _, eintrag := range eintraege {
			endung := filepath.Ext(eintrag.Name())
			if endung != ".ttf" && endung != ".otf" {
				continue
			}

			voll := filepath.Join(ort, eintrag.Name())

			roh, err := os.ReadFile(voll) //nolint:gosec // Pfad aus der Verzeichnisliste
			if err != nil {
				continue
			}

			if eigener, ok := schriften.NameAus(roh); ok {
				return voll, eigener
			}
		}
	}

	t.Skip("auf diesem Rechner ist keine lesbare .ttf oder .otf zu finden")

	return "", ""
}
