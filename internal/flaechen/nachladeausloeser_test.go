package flaechen_test

// **Ein Nachladefehler an einem echten Bestand** — die Vorrichtung der
// Läufe zu #1970.
//
// # Warum ein Auslöser
//
// **Weil das Nachlesen nach dem Schreiben scheitern soll und das Schreiben
// nicht.** Eine krumme Zeile, die schon dasteht, läse oft auch das Schreiben
// — es prüft vorher, was da ist. Ein Auslöser legt sie **im selben Zug**
// hinter die neue Zeile: geschrieben ist dann, und nur das Nachlesen trifft
// sie. Er läuft gegen SQLite selbst und in jedem Prozess, der den Bestand
// öffnet — auch im Dienst, der schon läuft.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// imBestand führt eine Anweisung am Bestand der Umgebung aus.
func imBestand(t *testing.T, umgebung []string, anweisung string) {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	if _, fehler := db.ExecContext(t.Context(), anweisung); fehler != nil {
		t.Fatalf("%s: %v", anweisung, fehler)
	}
}

// ausDemBestand liest einen Wert am Bestand der Umgebung — dort, wo die
// Fläche ihn nach einem Nachladefehler nicht mehr hergibt.
func ausDemBestand(t *testing.T, umgebung []string, abfrage string, werte ...any) string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), bestandsortAus(t, umgebung))
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	var wert string
	if fehler := db.QueryRowContext(t.Context(), abfrage, werte...).Scan(&wert); fehler != nil {
		t.Fatalf("%s: %v", abfrage, fehler)
	}

	return wert
}
