// **Was beim Start geschah, lässt sich erfragen** —
// #962.
package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// ablegen legt eine Datei an — als Hindernis für `MkdirAll`.
func ablegen(pfad string) error {
	return os.WriteFile(pfad, []byte("im Weg"), 0o600)
}

// **Nach einem geglückten Start sagt der Befund: offen.**
func TestDerStartbefundSagtOffen(t *testing.T) {
	t.Parallel()

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	befund := anwendung.Startbefund()
	if !befund.Offen {
		t.Errorf("der Befund sagt %+v", befund)
	}

	if befund.Grund != "" {
		t.Errorf("ein geglückter Start nennt einen Grund: %q", befund.Grund)
	}
}

// **Und nach einem gescheiterten sagt er, was schiefging.**
//
// # Warum das eine eigene Bindung ist
//
// `main.lauf` kehrte bei einem Fehler aus `Starten` **vor** `wails.Run`
// zurück. Es gab dann kein Fenster, keine Bindung und keinen Knopf — der
// Rückweg, den #940
// gebaut hat, war im gescheiterten Fall **nicht erreichbar**.
//
// Eine Anwendung, die sich bei einem Datenbankfehler wortlos beendet, lässt
// genau die Menschen allein, für die es die Sicherung gibt. Jetzt öffnet das
// Fenster trotzdem, und diese Bindung sagt ihm, was es zeigen soll.
func TestDerStartbefundNenntDenGrund(t *testing.T) {
	t.Parallel()

	// **Eine Datei am Platz des Datenverzeichnisses.** `MkdirAll` scheitert
	// daran, und damit scheitert der Start — ohne dass eine Datenbank im
	// Spiel wäre.
	ort := t.TempDir()

	imWeg := filepath.Join(ort, "imweg")
	if err := ablegen(imWeg); err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	anwendung := app.NeueApp(filepath.Join(imWeg, "unten", "bestand.db"))

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der Start kam durch")
	}

	befund := anwendung.Startbefund()
	if befund.Offen {
		t.Fatalf("der Befund sagt offen: %+v", befund)
	}

	if !strings.Contains(befund.Grund, "Datenverzeichnis") {
		t.Errorf("der Grund nennt nicht, was scheiterte: %q", befund.Grund)
	}
}

// **Ein geglückter zweiter Versuch räumt den Grund weg.**
//
// Sonst zeigte das Fenster nach einem erfolgreichen Zurückspielen weiter die
// Wiederherstellung — und der Weg endete in sich selbst.
func TestEinGeglueckterStartRaeumtDenGrund(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	imWeg := filepath.Join(ort, "imweg")
	if err := ablegen(imWeg); err != nil {
		t.Fatalf("die Datei anlegen: %v", err)
	}

	anwendung := app.NeueApp(filepath.Join(imWeg, "unten", "bestand.db"))

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der Start kam durch")
	}

	zweite := app.NeueApp(filepath.Join(ort, "bestand.db"))
	if err := zweite.Starten(t.Context()); err != nil {
		t.Fatalf("der zweite Start: %v", err)
	}

	t.Cleanup(func() { _ = zweite.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	// **Derselbe Verbund**, nachdem er einmal gescheitert ist.
	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der zweite Anlauf am selben Ort kam durch")
	}

	if anwendung.Startbefund().Offen {
		t.Error("der Befund sagt offen, obwohl der Start wieder scheiterte")
	}
}

// **Eine liegengebliebene Rückkehr steht im Startbefund.**
//
// Befund 10 des Zweitblicks über Phase 5.2: die Epoche entsteht **nach** dem
// Tausch, und wenn sie liegen blieb, trägt eine Marke neben dem Bestand sie
// nach. Was dabei nicht gelingt, gehört ins Fenster — im Protokoll allein
// sähe es niemand.
func TestEineLiegengebliebeneRueckkehrStehtImStartbefund(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	pfad := filepath.Join(ort, "bestand.db")

	// **Eine unlesbare Marke** — sie lässt sich nicht nachtragen, und genau
	// das soll das Fenster erfahren.
	if fehler := os.WriteFile(filepath.Join(ort, "rueckkehr-offen.json"),
		[]byte("{"), 0o600); fehler != nil {
		t.Fatalf("die Marke schreiben: %v", fehler)
	}

	befund := laufendeAnwendung(t, pfad).Startbefund()

	if !befund.Offen {
		t.Fatalf("der Start scheiterte: %+v", befund)
	}

	if !strings.Contains(befund.Rueckspielrest, "Rückkehrmarke") {
		t.Errorf("der Befund sagt nichts über die Marke: %+v", befund)
	}
}
