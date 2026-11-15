package main

// **Die Ansagedatei unter launchd** — SEC02
// (#2355).
//
// Was `app.AnsageInDatei` und `app.FehlerausgabeBegrenzen` selbst tun, prüft
// `internal/app`; zehn echte Starts fährt `internal/flaechen`. Hier steht die
// Verdrahtung: dass der Dienst ohne geschriebene Ansage nicht anläuft, und
// dass die Grenze vor allem anderen greift.
//
// **Nicht parallel**: beide Läufe setzen `os.Stderr`.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Lässt sich die Ansage nicht schreiben, läuft der Dienst nicht an** — und
// die zu große Fehlerausgabe ist bis dahin schon gekürzt.
func TestOhneAnsagedateiKeinDienst(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	fehlerausgabe := grosseFehlerausgabe(t, os.O_RDWR)

	// **Ein Ordner, der eine Datei ist**: dort entsteht keine Ansage.
	blockiert := filepath.Join(zuhause, "blockiert")
	if fehler := os.WriteFile(blockiert, nil, 0o600); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	angesagt := &bytes.Buffer{}

	fehler := lauf(probekontext(t), []string{
		"-port", "0", "-ansage",
		filepath.Join(blockiert, "dienst.txt"),
	}, angesagt)
	if fehler == nil || !strings.Contains(fehler.Error(), "Ansage") {
		t.Fatalf("der Dienst lief ohne Ansagedatei an: %v", fehler)
	}

	// **Und die Marke ging nirgendwohin** — wer sie nicht in der Datei
	// findet, soll sie auch auf der Ausgabe nicht bekommen.
	if strings.Contains(angesagt.String(), "Marke ") {
		t.Errorf("die Marke stand trotzdem auf der Ausgabe:\n%s", angesagt)
	}

	if stand, statfehler := fehlerausgabe.Stat(); statfehler != nil || stand.Size() > app.Fehlerausgabegrenze {
		t.Errorf("die Fehlerausgabe wurde nicht gekürzt: %v, %v", stand, statfehler)
	}
}

// **Eine Fehlerausgabe, die sich nicht kürzen lässt, hält den Dienst nicht
// auf** — sie wird genannt, und der Start geht weiter bis zur Ansage.
func TestEineUnkuerzbareFehlerausgabeHaeltDenDienstNichtAuf(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", filepath.Join(zuhause, "bestand.db"))

	grosseFehlerausgabe(t, os.O_RDONLY)

	// **Bis zur Ansage und nicht weiter** — derselbe blockierte Ordner wie
	// oben: ein Ort, den es geben kann, ließe den Dienst laufen, bis die
	// Frist des Tests endet.
	blockiert := filepath.Join(zuhause, "blockiert")
	if fehler := os.WriteFile(blockiert, nil, 0o600); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	fehler := lauf(probekontext(t), []string{
		"-port", "0", "-ansage",
		filepath.Join(blockiert, "dienst.txt"),
	}, &bytes.Buffer{})
	if fehler == nil || !strings.Contains(fehler.Error(), "Ansage") {
		t.Fatalf("der Start kam nicht bis zur Ansage: %v", fehler)
	}
}

// grosseFehlerausgabe setzt `os.Stderr` auf eine Datei über der Grenze.
func grosseFehlerausgabe(t *testing.T, art int) *os.File {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "dienst.log")
	if fehler := os.WriteFile(pfad, bytes.Repeat([]byte("x"), app.Fehlerausgabegrenze+1), 0o600); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	datei, fehler := os.OpenFile(pfad, art, 0) // #nosec G304 -- Pfad unter t.TempDir
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	vorher := os.Stderr
	os.Stderr = datei

	t.Cleanup(func() {
		os.Stderr = vorher
		_ = datei.Close() //nolint:errcheck,gosec // eine Probedatei
	})

	return datei
}
