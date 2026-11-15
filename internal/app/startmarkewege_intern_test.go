package app

// **Die Fehlerwege der Startmarke** — F04 (#2343). Ein Start, der seine
// Marke nicht legen kann, sagt es; eine Marke, die sich nicht lesen lässt,
// ist ein Abbruch ohne Zeit, denn liegen geblieben ist sie trotzdem.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDerBeginnScheitertLautAnJederStelle(t *testing.T) {
	alsDatei := filepath.Join(t.TempDir(), "keinordner")
	if err := os.WriteFile(alsDatei, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	ohneSchreibrecht := t.TempDir()
	if err := os.Chmod(ohneSchreibrecht, 0o500); err != nil { //nolint:gosec // der Ordner eines Tests
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(ohneSchreibrecht, 0o700) }) //nolint:errcheck,gosec // zum Aufräumen

	markeAlsOrdner := t.TempDir()
	// **Mit Inhalt**: ein leerer Ordner gälte als unlesbare Marke und würde
	// abgeräumt — dann ließe sich die eigene doch legen.
	if err := os.MkdirAll(filepath.Join(markeAlsOrdner, "fenster-815.json", "darin"), 0o700); err != nil {
		t.Fatal(err)
	}

	// **Die Spur gibt es schon**: seit der Sperre über den Ordner scheitert ein
	// Ordner ohne Schreibrecht schon an ihr, und die Spur ist ausschließlich
	// neu (`O_EXCL`).
	spurDa := t.TempDir()
	if err := os.WriteFile(filepath.Join(spurDa, fmt.Sprintf("absturz-%d-815.txt", jetzt.UnixNano())), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	fall := map[string]struct {
		ordner  string
		ausgabe func(*os.File) error
	}{
		"der Ordner ist eine Datei":         {alsDatei, ohneAusgabe},
		"der Ordner nimmt keine Datei an":   {ohneSchreibrecht, ohneAusgabe},
		"die Spur lässt sich nicht anlegen": {spurDa, ohneAusgabe},
		"die Spur lässt sich nicht richten": {t.TempDir(), func(*os.File) error { return errors.New("nein") }},
		"die Marke lässt sich nicht legen":  {markeAlsOrdner, ohneAusgabe},
	}

	for name, f := range fall {
		t.Run(name, func(t *testing.T) {
			if _, err := sitzungBeginnen(f.ordner, 815, jetzt, func(int) bool { return false }, f.ausgabe); err == nil {
				t.Error("der Beginn gelang trotzdem")
			}
		})
	}
}

func TestEinUnlesbareMarkeIstEinAbbruchOhneZeit(t *testing.T) {
	ordner := t.TempDir()
	if err := os.WriteFile(filepath.Join(ordner, "fenster-4711.json"), []byte("{kaputt"), 0o600); err != nil {
		t.Fatal(err)
	}

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return true }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	if len(beginn.abbrueche) != 1 || !beginn.abbrueche[0].Gestartet.IsZero() {
		t.Errorf("Abbrüche %+v, erwartet einer ohne Zeit", beginn.abbrueche)
	}
}

func TestEinAbbruchMitLeererOderFehlenderSpurIstOhneSpur(t *testing.T) {
	ordner := t.TempDir()
	if err := os.WriteFile(filepath.Join(ordner, "absturz-1-4711.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	legeMarke(t, ordner, 4711, "absturz-1-4711.txt")
	legeMarke(t, ordner, 4712, "absturz-1-4712.txt")

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	for _, abbruch := range beginn.abbrueche {
		if abbruch.MitSpur {
			t.Errorf("ein Abbruch ohne Inhalt galt als mit Spur: %+v", abbruch)
		}
	}

	if _, err := os.Stat(filepath.Join(ordner, "absturz-1-4711.txt")); !os.IsNotExist(err) {
		t.Error("die leere Spur des Abbruchs liegt noch")
	}
}

func TestDasEndeMeldetEineMarkeDieNichtWeggeht(t *testing.T) {
	ordner := t.TempDir()

	beginn, err := sitzungBeginnen(ordner, 815, jetzt, func(int) bool { return false }, ohneAusgabe)
	if err != nil {
		t.Fatal(err)
	}

	marke := filepath.Join(ordner, "fenster-815.json")
	if err := os.Remove(marke); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(marke, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(marke, "darin"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := beginn.ende(); err == nil {
		t.Error("eine Marke, die nicht wegging, galt als geräumt")
	}
}

func TestDieEchtenPlattformwege(t *testing.T) {
	if !prozessLebt(os.Getpid()) || !prozessLebt(1) {
		t.Error("ein laufender Prozess galt als tot — der eigene oder launchd (EPERM)")
	}

	if prozessLebt(1 << 30) {
		t.Error("ein Prozess, den es nicht gibt, galt als lebendig")
	}

	datei, err := os.CreateTemp(t.TempDir(), "spur")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // ein Testgriff

	if err := setzeAbsturzausgabe(datei); err != nil {
		t.Fatal(err)
	}

	if err := setzeAbsturzausgabe(nil); err != nil {
		t.Fatal(err)
	}
}
