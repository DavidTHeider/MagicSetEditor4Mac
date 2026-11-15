package storage_test

// **Die Kennungssysteme stehen in Go und in SQLite — und müssen gleich sein**
// (A5, #1444).
//
// # Warum es diesen Wächter gibt
//
// **Weil die Aufzählung zweimal dasteht und einmal auseinandergelaufen ist.**
// `domain.Kennungssysteme()` führt sie in Go, `CHECK (system IN …)` in der
// Wanderung. A5 trug `scrydex` in die eine ein; jeder Lauf in
// `internal/domain` war grün, und beim Schreiben kam „CHECK constraint
// failed" — aus einem Weg, der den Fehler als **Befund** führt und nicht als
// Abbruch. Es wäre also still geblieben.
//
// **Das ist die Falle aus §10**, und sie ist bei den Einstellungen schon
// zweimal zugeschnappt (siehe den Kopf von `einstellung.go`).
//
// # Was er hält
//
// Jedes System aus [domain.Kennungssysteme] lässt sich wirklich schreiben.
//
// # Was er nicht hält
//
// **Die Gegenrichtung.** Ein Wort, das die Wanderung erlaubt und Go nicht
// kennt, fällt hier nicht auf — es käme nie in eine Kennung hinein.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestJedesKennungssystemLaesstSichSchreiben(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)

	systeme := domain.Kennungssysteme()

	if len(systeme) == 0 {
		t.Fatal("es gibt keine Kennungssysteme — der Wächter prüft nichts")
	}

	druck := &domain.Druck{
		Name: "Probe", Setcode: "m11", Nummer: "1", Sprache: "en",
	}

	// **Zwei Systeme verlangen eine UUID** — Scryfall und MTGJSON prüfen die
	// Gestalt ([domain.Druckkennung.Validate]). Der Wert richtet sich danach,
	// damit der Lauf die **Wanderung** prüft und nicht die Validierung.
	for i, system := range systeme {
		wert := string(rune('a'+i)) + "-probe"

		if system == domain.SystemScryfall || system == domain.SystemMTGJSON {
			wert = "1111111" + string(rune('1'+i)) +
				"-1111-4111-8111-111111111111"
		}

		druck.Kennungen = append(druck.Kennungen, domain.Druckkennung{
			System: system, Semantik: domain.SemantikProdukt, Wert: wert,
		})
	}

	if fehler := storage.NeueDrucke(db).Sichern(probekontext(t), druck); fehler != nil {
		t.Fatalf("den Druck mit allen Systemen sichern: %v", fehler)
	}

	gelesen, lesefehler := storage.NeueDrucke(db).Fuer(probekontext(t), druck.ID)
	if lesefehler != nil {
		t.Fatalf("den Druck lesen: %v", lesefehler)
	}

	if len(gelesen.Kennungen) != len(systeme) {
		t.Errorf("%d Kennungen kamen zurück, erwartet %d",
			len(gelesen.Kennungen), len(systeme))
	}
}
