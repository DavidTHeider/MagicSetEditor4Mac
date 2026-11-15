// **Was 0.7.0 geschrieben hat, öffnet diese Fassung ohne einen Schritt** —
// J02 (#2266,
// Vermerk 0189).
//
// Seit es nur die Grundlinie gibt, ist 0.7.0 die Brücke: dort steigt jeder
// ältere Bestand auf, und was dort steht, führt genau die Schritte, die die
// Grundlinie enthält. **Die Sicherung hat 0.7.0 selbst geschrieben**, aus dem
// Tag `v0.7.0` — wie, steht in `testdaten/fassung-0.7.0/LIESMICH.md`.
package app_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// alleWanderungen zählt, was diese Fassung kennt: die Schritte der
// Grundlinie und jede Wanderung danach.
func alleWanderungen(t *testing.T) []string {
	t.Helper()

	umfang, fehler := os.ReadFile(filepath.Join("..", "storage", "grundlinie", "umfang.txt"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	var namen []string

	for _, zeile := range strings.Split(string(umfang), "\n") {
		if name, _, ganz := strings.Cut(zeile, " "); ganz && !strings.HasPrefix(zeile, "#") {
			namen = append(namen, name)
		}
	}

	danach, fehler := filepath.Glob(filepath.Join("..", "storage", "migrationen", "*.sql"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	for _, pfad := range danach {
		namen = append(namen, filepath.Base(pfad))
	}

	slices.Sort(namen)

	return namen
}

func TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet(t *testing.T) {
	bestand := filepath.Join(t.TempDir(), "bestand.db")
	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.SicherungZurueckspielen(t.Context(),
		ausFassung(t, "0.7.0", "sicherung.zip"), ""); antwort.Fehler != nil {
		t.Fatalf("die Sicherung aus 0.7.0 ließ sich nicht zurückspielen: %+v", antwort.Fehler)
	}

	sets := anwendung.Sets(t.Context())
	if sets.Fehler != nil {
		t.Fatalf("Sets lesen: %+v", sets.Fehler)
	}

	namen, karten := make([]string, 0, len(sets.Sets)), 0
	for _, set := range sets.Sets {
		namen = append(namen, set.Name)
		karten += set.Karten
	}

	slices.Sort(namen)

	if strings.Join(namen, "|") != "Altes Probeset|Mein Probeset" || karten != 4 {
		t.Errorf("nach dem Zurückspielen: Sets %q mit %d Karten", namen, karten)
	}

	if decks := anwendung.Decks(t.Context(), "probe"); decks.Fehler != nil ||
		len(decks.Decks) != 1 || decks.Decks[0].Name != "Altes Deck" {
		t.Errorf("das Deck aus 0.7.0: %+v", decks)
	}

	if got := merkmal(t, anwendung); got != "en" {
		t.Errorf("die Sprache aus 0.7.0 ist %q, erwartet war %q", got, "en")
	}

	// **Auf der Grundlinie und heil** — die Buchführung nennt genau, was
	// diese Fassung kennt, und Seiten wie Fremdschlüssel sind in Ordnung.
	db, fehler := storage.OeffnenNurLesend(t.Context(), bestand)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur gelesen

	verbucht, fehler := storage.AngewendeteMigrationen(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if erwartet := alleWanderungen(t); !slices.Equal(verbucht, erwartet) {
		t.Errorf("verbucht sind %d Wanderungen, diese Fassung kennt %d", len(verbucht), len(erwartet))
	}

	var integritaet string
	if pruefung := db.QueryRowContext(t.Context(), "PRAGMA integrity_check").Scan(&integritaet); pruefung != nil ||
		integritaet != "ok" {
		t.Errorf("die Integrität: %q (%v)", integritaet, pruefung)
	}

	verwaist, fehler := db.QueryContext(t.Context(), "PRAGMA foreign_key_check")
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = verwaist.Close() }() //nolint:errcheck // nur gelesen

	if verwaist.Next() {
		t.Error("ein Fremdschlüssel zeigt ins Leere")
	}

	if lesefehler := verwaist.Err(); lesefehler != nil {
		t.Errorf("die Fremdschlüssel prüfen: %v", lesefehler)
	}
}
