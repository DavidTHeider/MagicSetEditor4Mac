// **Die Diagnose** — Werkstück 4 von Phase 3.5
// (#1166).
//
// # Was sie ist und was nicht
//
// Ein Text, den jemand einem Fehlerbericht beilegen kann — **ohne Terminal**.
// Bis hierher stand in der Administration `tail -n 200 …/mse.log`, und die
// Zielgruppe dieser Anwendung kann das nicht (Z9-01 im Audit).
//
// **Sie wird gezeigt, bevor sie irgendwohin geht.** Was hier entsteht, ist
// eine Zeichenkette; wer sie weitergibt, hat sie vorher gelesen.
package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die Diagnose nennt Fassung, Architektur und den Stand des Bestands.**
func TestDieDiagnoseNenntDasNoetige(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	text := diagnosetext(t, app.NeueApp(""))

	for _, erwartet := range []string{app.Fassung, "darwin", "arm64"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Diagnose nennt %q nicht:\n%s", erwartet, text)
		}
	}
}

// **Und sie trägt das Ende des Protokolls.**
func TestDieDiagnoseTraegtDasProtokollende(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	pfad, err := app.Protokollpfad()
	if err != nil {
		t.Fatalf("Protokollpfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	// **Mehr Zeilen als die Grenze**, damit die Begrenzung wirklich greift —
	// und **jede eindeutig**. Der erste Anlauf setzte die Namen aus zwei
	// Restklassen zusammen; sie wiederholten sich alle 130 Zeilen, und die
	// „erste" Zeile stand deshalb auch am Ende. Der Lauf war rot und der Code
	// in Ordnung — eine Vorrichtung, die etwas anderes prüft, als sie sagt.
	zeilen := make([]string, 0, app.Protokollzeilen+5)
	for i := range app.Protokollzeilen + 5 {
		zeilen = append(zeilen, fmt.Sprintf("zeile-%d", i))
	}

	if fehler := os.WriteFile(pfad,
		[]byte(strings.Join(zeilen, "\n")), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	text := diagnosetext(t, app.NeueApp(""))

	// **Die letzte Zeile ist da** — sie ist die interessante.
	if !strings.Contains(text, zeilen[len(zeilen)-1]) {
		t.Errorf("die letzte Protokollzeile fehlt:\n%s", text)
	}

	// **Und die erste nicht.** Ohne diese Hälfte wäre auch ein Weg grün, der
	// das ganze Protokoll anhängt — bei einem Monat Betrieb sind das Megabyte,
	// und niemand liest sie.
	if strings.Contains(text, zeilen[0]) {
		t.Errorf("die Diagnose trägt das ganze Protokoll:\n%s", text)
	}
}

// **Der Pfad des Benutzerverzeichnisses wird zur Tilde.**
//
// Dieselbe Regel wie im Protokoll selbst (`schwaerze`): eine Diagnose wird
// weitergegeben, und `/Users/vorname` sagt jedem, der sie liest, wie jemand
// heißt. **Hier ist sie nötig, obwohl das Protokoll schon geschwärzt ist** —
// die Kopfzeilen entstehen erst hier und gehen durch keinen Protokollschreiber.
func TestDieDiagnoseSchwaerztDasZuhause(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	text := diagnosetext(t, app.NeueApp(""))

	if strings.Contains(text, zuhause) {
		t.Errorf("das Zuhause steht in der Diagnose:\n%s", text)
	}

	if !strings.Contains(text, "~/") {
		t.Errorf("die Tilde fehlt — wurde überhaupt ein Pfad genannt?\n%s", text)
	}
}

// **Ein fehlendes Protokoll ist kein Fehler.**
//
// Eine frisch eingerichtete Anwendung hat keines, und wer dann eine Diagnose
// braucht, braucht sie **besonders**. Sie sagt, dass es keines gibt.
func TestEinFehlendesProtokollBrichtDieDiagnoseNicht(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	antwort := app.NeueApp("").Diagnose(probekontext(t))

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.Contains(antwort.Text, app.Fassung) {
		t.Errorf("die Diagnose ist leer:\n%s", antwort.Text)
	}
}

// diagnosetext holt die Diagnose und scheitert bei einem Fehler.
func diagnosetext(t *testing.T, anwendung *app.App) string {
	t.Helper()

	antwort := anwendung.Diagnose(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Diagnose: %+v", antwort.Fehler)
	}

	return antwort.Text
}

// **Die Diagnose lässt sich sichern — an einen Ort, den jemand wählt.**
func TestDieDiagnoseLaesstSichSichern(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	ziel := filepath.Join(t.TempDir(), "diagnose.txt")

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	// **Erst zeigen, dann sichern.** Die Zusage lautet „was du siehst, gibst
	// du weiter"; ohne Vorschau gibt es nichts zu sichern (Befund 5).
	diagnosetext(t, anwendung)

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, err := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(roh), app.Fassung) {
		t.Errorf("die Datei trägt keine Diagnose:\n%s", roh)
	}
}

// **Ein Abbruch schreibt nichts und ist kein Fehler.**
//
// Wer den Dialog schließt, hat sich entschieden. Eine Meldung darüber wäre
// eine Antwort auf eine Frage, die niemand gestellt hat.
func TestEinAbgebrochenerDialogSchreibtKeineDiagnose(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	verzeichnis := t.TempDir()

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	diagnosetext(t, anwendung)

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil || len(eintraege) != 0 {
		t.Errorf("es wurde geschrieben: %v, %v", eintraege, err)
	}
}

// **Ein unschreibbares Ziel wird gemeldet.**
func TestEinUnschreibbaresZielMeldetSich(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return filepath.Join(t.TempDir(), "gibtesnicht", "diagnose.txt"), nil
	})

	diagnosetext(t, anwendung)

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler == nil {
		t.Error("das unschreibbare Ziel kam durch")
	}
}

// **Und ein Fehler des Dialogs kommt an.**
func TestEinFehlerDerDateiwahlErreichtDieDiagnose(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errDialog
	})

	diagnosetext(t, anwendung)

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler == nil {
		t.Error("der Dialogfehler kam nicht an")
	}
}

// errDialog ist der Fehler, den die Dateiwahl in diesen Läufen wirft.
var errDialog = errors.New("der Dialog ist nicht zu öffnen")

// **Ohne Zuhause steht in der Diagnose, dass es keines gibt.**
//
// # Warum das der wichtigste Fall ist
//
// Eine Anwendung, die ihren Bestand nicht findet, tut gar nichts — und genau
// dann braucht jemand eine Diagnose. Ein Weg, der hier abbräche, ließe den
// Fall unberichtet, für den es ihn gibt (§8).
func TestOhneZuhauseSagtDieDiagnoseEs(t *testing.T) {
	t.Setenv("HOME", "")

	text := diagnosetext(t, app.NeueApp(""))

	// **Die Fassung steht trotzdem da.** Sie ist die erste Frage bei jedem
	// Bericht und hängt an keinem Verzeichnis.
	if !strings.Contains(text, app.Fassung) {
		t.Errorf("die Fassung fehlt:\n%s", text)
	}

	if !strings.Contains(text, "nicht zu lesen") {
		t.Errorf("der Grund fehlt:\n%s", text)
	}

	if !strings.Contains(text, "Kein Protokoll") {
		t.Errorf("das fehlende Protokoll steht nicht da:\n%s", text)
	}
}

// **Ohne Vorschau gibt es nichts zu sichern.**
//
// Die Zusage lautet „was du siehst, gibst du weiter" — sie hat ohne ein
// Gesehenes keinen Inhalt. Angestrichen vom Zweitblick über Phase 3.5 als
// Befund 5: der erste Anlauf erzeugte beim Sichern einen **zweiten** Text.
func TestOhneVorschauGibtEsNichtsZuSichern(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		t.Error("der Dialog wurde geöffnet, obwohl nichts gezeigt worden war")

		return "", nil
	})

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler == nil {
		t.Error("das Sichern ohne Vorschau kam durch")
	}
}

// **Und gesichert wird genau der gezeigte Text.**
//
// # Wie das geprüft wird, ohne die Uhr zu bemühen
//
// Zwischen Anzeigen und Sichern wächst im Betrieb das Protokoll. Hier wächst
// es **absichtlich**: nach der Vorschau kommt eine Zeile dazu. Steht sie in
// der Datei, hat jemand etwas weitergegeben, das er nie gesehen hat.
func TestGesichertWirdDerGezeigteText(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	pfad, err := app.Protokollpfad()
	if err != nil {
		t.Fatalf("Protokollpfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(pfad, []byte("vorher\n"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "diagnose.txt")

	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	gezeigt := diagnosetext(t, anwendung)

	// **Das Protokoll wächst zwischen Anzeigen und Sichern.**
	if fehler := os.WriteFile(pfad,
		[]byte("vorher\nDANACH\n"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if antwort := anwendung.DiagnoseSichern(t.Context()); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	roh, lesefehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if string(roh) != gezeigt {
		t.Errorf("gesichert wurde etwas anderes als gezeigt:\n%s", roh)
	}

	if strings.Contains(string(roh), "DANACH") {
		t.Error("die Datei trägt eine Zeile, die niemand gesehen hat")
	}
}

// **Der Schemastand und die Paketstände stehen darin.**
//
// #1166 verlangt beide ausdrücklich; der erste Anlauf hatte sie beim Zuschnitt
// verloren (Befund 2 des Zweitblicks). **Ohne Bestand und ohne Paketverwaltung
// sagt die Diagnose es** — das ist die Lage, in der sie am nötigsten ist.
func TestDieDiagnoseNenntSchemaUndPakete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	text := diagnosetext(t, app.NeueApp(""))

	for _, erwartet := range []string{"Schema:", "Pakete:"} {
		if !strings.Contains(text, erwartet) {
			t.Errorf("die Diagnose nennt %q nicht:\n%s", erwartet, text)
		}
	}
}
