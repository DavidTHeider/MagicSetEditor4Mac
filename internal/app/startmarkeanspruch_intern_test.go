package app

// **Eine Sitzung erkennt man an ihrem gehaltenen Anspruch, nicht an der
// Prozessnummer** — Befund 11 des Zweitblicks über Phase 1 von 0.9.0 (F04,
// #2343). Eine Prozessnummer wird wiederverwendet: eine Marke mit der eigenen
// verschwand ohne Hinweis, und bei einer fremden genügte irgendein lebender
// Prozess. Jetzt hält jede Sitzung eine Sperre (`flock`) neben ihrer Marke,
// und das Betriebssystem gibt sie mit dem Prozess frei — auch nach SIGKILL.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// legeMarkeMitAnspruch legt eine Marke, wie sie eine Sitzung mit Anspruch
// schreibt — ohne den Anspruch zu halten.
func legeMarkeMitAnspruch(t *testing.T, ordner string, pid int) {
	t.Helper()

	roh, err := json.Marshal(sitzungsmarke{PID: pid, Gestartet: jetzt.Add(-time.Hour), Fassung: "0.9.0", Gesperrt: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ordner, "fenster-"+strconv.Itoa(pid)+".json"), roh, 0o600); err != nil {
		t.Fatal(err)
	}
}

func immerLebend(int) bool { return true }

func TestEineMarkeMitDerEigenenPIDIstEinAbbruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarke(t, ordner, 815, "")

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, immerLebend, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 {
		t.Errorf("Abbrüche %+v — die Marke einer früheren Sitzung mit derselben Nummer verschwand", beginn.abbrueche)
	}

	_ = beginn.ende() //nolint:errcheck // Aufräumen
}

func TestEineWiederverwendetePIDVerdecktKeinenAbbruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarkeMitAnspruch(t, ordner, 4711)

	// **Ein lebender Prozess mit derselben Nummer** — aber keiner, der den
	// Anspruch hält.
	beginn, err := sitzungBeginnen(ordner, 815, jetzt, immerLebend, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 {
		t.Errorf("Abbrüche %+v — ein fremder Prozess mit derselben Nummer verdeckte den Abbruch", beginn.abbrueche)
	}

	_ = beginn.ende() //nolint:errcheck // Aufräumen
}

func TestEinGehaltenerAnspruchIstKeinAbbruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarkeMitAnspruch(t, ordner, 4711)

	griff, err := dateisperre.Oeffnen(filepath.Join(ordner, "fenster-4711.sperre"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	if sperrfehler := griff.Allein(); sperrfehler != nil {
		t.Fatal(sperrfehler)
	}

	// **Auch wenn die Nummer tot scheint**: der Anspruch entscheidet.
	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 0 {
		t.Errorf("Abbrüche %+v — eine gehaltene Sitzung galt als abgebrochen", beginn.abbrueche)
	}

	_ = beginn.ende() //nolint:errcheck // Aufräumen
}

func TestZweiSitzungenSehenEinanderAlsLebend(t *testing.T) {
	ordner := t.TempDir()
	tot := func(int) bool { return false }

	erste, err := sitzungBeginnen(ordner, 815, jetzt, tot, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	zweite, err := sitzungBeginnen(ordner, 816, jetzt, tot, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(zweite.abbrueche) != 0 {
		t.Errorf("die zweite sah die erste als abgebrochen: %+v", zweite.abbrueche)
	}

	if err := erste.ende(); err != nil {
		t.Fatal(err)
	}

	if err := zweite.ende(); err != nil {
		t.Fatal(err)
	}

	if rest, _ := filepath.Glob(filepath.Join(ordner, "fenster-*")); len(rest) != 0 { //nolint:errcheck // ein festes Muster
		t.Errorf("nach beiden Schlüssen liegt noch %v", rest)
	}
}

func TestDerBeginnScheitertAmAnspruchLaut(t *testing.T) {
	alsOrdner := t.TempDir()
	if err := os.MkdirAll(filepath.Join(alsOrdner, "fenster-815.sperre", "darin"), 0o700); err != nil {
		t.Fatal(err)
	}

	gehalten := t.TempDir()

	griff, err := dateisperre.Oeffnen(filepath.Join(gehalten, "fenster-815.sperre"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = griff.Schliessen() }) //nolint:errcheck // Aufräumen

	if sperrfehler := griff.Allein(); sperrfehler != nil {
		t.Fatal(sperrfehler)
	}

	for name, ordner := range map[string]string{
		"der Anspruch lässt sich nicht anlegen": alsOrdner,
		"der Anspruch ist schon gehalten":       gehalten,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := sitzungBeginnen(ordner, 815, jetzt, immerLebend, ohneAusgabe); err == nil {
				t.Error("der Beginn gelang trotzdem")
			}

			if marken, _ := filepath.Glob(filepath.Join(ordner, "fenster-*.json")); len(marken) != 0 { //nolint:errcheck // ein festes Muster
				t.Errorf("ohne Anspruch liegt eine Marke: %v", marken)
			}
		})
	}
}

// **Eine Sperrdatei, die sich nicht öffnen lässt, hält keinen Anspruch** —
// die Sitzung dahinter gilt als abgebrochen, statt für immer als lebend.
func TestEineUnoeffnbareSperreHaeltKeinenAnspruch(t *testing.T) {
	ordner := t.TempDir()
	legeMarkeMitAnspruch(t, ordner, 4711)

	if err := os.MkdirAll(filepath.Join(ordner, "fenster-4711.sperre", "darin"), 0o700); err != nil {
		t.Fatal(err)
	}

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, immerLebend, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 {
		t.Errorf("Abbrüche %+v", beginn.abbrueche)
	}

	_ = beginn.ende() //nolint:errcheck // Aufräumen
}
