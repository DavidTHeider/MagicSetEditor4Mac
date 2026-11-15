// **Was in keiner Protokollzeile stehen darf** —
// #857.
package app_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// protokollLesen legt ein Protokoll an, schreibt hinein und gibt den Inhalt.
func protokollLesen(t *testing.T, schreiben func(*slog.Logger)) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "mse.log")

	schreiber, schliessen, fehler := app.Protokoll(pfad)
	if fehler != nil {
		t.Fatalf("Protokoll: %v", fehler)
	}

	schreiben(schreiber)

	if schliessfehler := schliessen(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}

	roh, fehler := os.ReadFile(pfad) //#nosec G304 -- ein Pfad aus t.TempDir
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	return string(roh)
}

// **Ein Merkmal, das niemand angemeldet hat, verliert seinen Wert.**
//
// Das ist die Zusage, um derentwillen es die Liste gibt: eine vergessene
// Anmeldung kostet eine unlesbare Zeile und nicht einen ausgeplauderten
// Kartentext. Der umgekehrte Nullwert — „unbekannt heißt erlaubt" — bräche
// beim ersten neuen Aufruf still.
func TestEinUnbekanntesMerkmalVerliertSeinenWert(t *testing.T) {
	inhalt := protokollLesen(t, func(schreiber *slog.Logger) {
		schreiber.Info("eine Karte", "kartenname", "Feuerball")
	})

	if strings.Contains(inhalt, "Feuerball") {
		t.Errorf("der Kartenname steht im Protokoll:\n%s", inhalt)
	}

	// **Der Name bleibt stehen.** Ohne ihn sähe niemand, dass etwas fehlt —
	// und eine stille Lücke ist schlimmer als eine sichtbare.
	if !strings.Contains(inhalt, "kartenname=") {
		t.Errorf("der Merkmalsname fehlt ebenfalls:\n%s", inhalt)
	}

	if !strings.Contains(inhalt, app.Nichterklaert) {
		t.Errorf("der Ersatz fehlt:\n%s", inhalt)
	}
}

// **Und auch dann, wenn es über `With` gebunden wurde.**
//
// Derselbe Nebenkanal, an dem die Schwärzung des Benutzerpfads beim ersten
// Anlauf vorbeiging (#848). Bei einer Regel mit zwei Wegen bleibt der zweite
// gern im alten Zustand zurück.
func TestEinGebundenesUnbekanntesMerkmalVerliertSeinenWertAuch(t *testing.T) {
	inhalt := protokollLesen(t, func(schreiber *slog.Logger) {
		schreiber.With("regeltext", "Fliegend, Erstschlag").Info("eine Karte")
	})

	if strings.Contains(inhalt, "Fliegend") {
		t.Errorf("der Regeltext steht im Protokoll:\n%s", inhalt)
	}
}

// **Ein angemeldetes Merkmal behält seinen Wert.**
func TestEinAngemeldetesMerkmalBehaeltSeinenWert(t *testing.T) {
	inhalt := protokollLesen(t, func(schreiber *slog.Logger) {
		schreiber.Info("gescheitert", "grund", "kein Speicher", "anzahl", 3)
	})

	for _, teil := range []string{"grund=", "kein Speicher", "anzahl=3"} {
		if !strings.Contains(inhalt, teil) {
			t.Errorf("%q fehlt im Protokoll:\n%s", teil, inhalt)
		}
	}
}

// **Das Benutzerverzeichnis bleibt geschwärzt** — beide Regeln zugleich.
func TestDerBenutzerpfadBleibtGeschwaerzt(t *testing.T) {
	zuhause, fehler := os.UserHomeDir()
	if fehler != nil || zuhause == "" {
		t.Skip("kein Benutzerverzeichnis")
	}

	inhalt := protokollLesen(t, func(schreiber *slog.Logger) {
		schreiber.Info("offen", "ort", filepath.Join(zuhause, "Library", "x.db"))
	})

	if strings.Contains(inhalt, zuhause) {
		t.Errorf("das Benutzerverzeichnis steht im Protokoll:\n%s", inhalt)
	}

	if !strings.Contains(inhalt, "~/Library/x.db") {
		t.Errorf("der geschwärzte Pfad fehlt:\n%s", inhalt)
	}
}
