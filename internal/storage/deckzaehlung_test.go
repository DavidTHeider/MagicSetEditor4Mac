package storage_test

// **Was der Deckspeicher zählt** — Werkstück 2 von Phase 4.6
// (#1293).
//
// **Eine eigene Datei**, weil `decks_test.go` über die Zeilengrenze aus §2
// ging. Der Schnitt folgt der Frage: dort steht, **was der Speicher kann**,
// hier, **was er zählt**.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Übersicht zählt die Karten mit — ohne die Plätze zu laden.**
//
// # Der Befund, für den es diesen Lauf gibt
//
// [Deckablage.Decks] gibt die Decks **ohne** ihre Plätze, und das ist richtig:
// für zwanzig Decks wären das zwanzig Abfragen für etwas, das niemand liest.
// Die **Zahlen** liest aber jeder — und `domain.Deck.Zaehle` rechnete über
// eine leere Liste, also immer null.
//
// **Gesehen bei der Sichtprüfung am 2026-09-01** in der laufenden Anwendung:
// ein Deck mit vier Karten stand als „0 im Hauptdeck, 0 im Sideboard" da. Kein
// Test hat es gesehen, weil keiner die Übersicht **nach dem Legen** gelesen
// hat (§3: „was die Tests hier nicht sehen werden").
func TestDieDeckuebersichtZaehltIhreKarten(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	for _, gelegt := range []struct {
		printing string
		menge    int
		bereich  domain.Deckbereich
	}{
		{"a", 4, domain.BereichHauptdeck},
		{"b", 2, domain.BereichHauptdeck},
		{"c", 3, domain.BereichSideboard},
	} {
		if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
			Printing: gelegt.printing, Menge: gelegt.menge, Bereich: gelegt.bereich,
		}); err != nil {
			t.Fatalf("legen: %v", err)
		}
	}

	decks, err := ablage.Decks(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("die Liste: %v", err)
	}

	if len(decks) != 1 {
		t.Fatalf("%d Decks, erwartet eines", len(decks))
	}

	// **Karten, nicht Plätze**: sechs im Hauptdeck aus zwei Plätzen.
	if decks[0].Hauptdeck != 6 || decks[0].Sideboard != 3 {
		t.Errorf("Hauptdeck %d, Sideboard %d — erwartet 6 und 3",
			decks[0].Hauptdeck, decks[0].Sideboard)
	}

	// **Und die Plätze bleiben leer** — das ist die andere Hälfte der Zusage.
	if len(decks[0].Plaetze) != 0 {
		t.Errorf("die Übersicht lädt die Plätze: %+v", decks[0].Plaetze)
	}
}

// **Zweimal derselbe Bezug ist eine höhere Menge, kein Konflikt** — Befund 4
// des Zweitblicks über Phase 4.6.
//
// # Was der Befund sagte, und was davon zutrifft
//
// Er sah, dass Suchen und Schreiben **ohne Transaktion** nebeneinanderstanden,
// und schloss auf ein Zeitfenster: zwei gleichzeitige Rufe lesen beide „nicht
// vorhanden", und der zweite läuft gegen den Eindeutigkeitsindex.
//
// **Der Wortlaut stimmt, die Wirkung nicht.** Gemessen am 2026-09-01:
// `storage.Oeffnen` setzt `db.SetMaxOpenConns(1)` (`oeffnenMit` in
// `internal/storage/datenbank.go`, seit #1485 für jede Datenbank). Mit **einer** Verbindung
// laufen alle Rufe serialisiert, und das Fenster entsteht nicht. Ein Lauf mit
// acht Goroutinen bleibt deshalb grün, auch wenn man die Transaktion
// wieder herausnimmt — er kann seinen Fall nicht enthalten (§10).
//
// # Warum die Transaktion trotzdem dasteht
//
// **Weil sie die Zusage von der Verbindungszahl löst.** Heute hält sie
// `SetMaxOpenConns(1)`; wer diese Zahl je erhöht — für einen Lesepfad, für
// eine zweite Prozessgrenze —, hätte den Befund sonst geerbt, ohne ihn zu
// sehen.
//
// # Was dieser Lauf hält
//
// **Die fachliche Zusage**, die man messen kann: derselbe Bezug im selben
// Bereich ist **ein** Platz mit der neuen Menge. Er hält **nicht**, dass es
// bei mehreren Verbindungen keinen Konflikt gäbe — dafür fehlt die zweite
// Verbindung.
func TestZweimalDerselbeBezugIstEineHoehereMenge(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	for _, menge := range []int{1, 4, 2} {
		if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
			Printing: "a1b2c3d4-0000-4000-8000-00000000000a",
			Menge:    menge, Bereich: domain.BereichHauptdeck,
		}); err != nil {
			t.Fatalf("legen mit Menge %d: %v", menge, err)
		}
	}

	gelesen := deckLesen(t, ablage, deck.ID)

	if len(gelesen.Plaetze) != 1 {
		t.Fatalf("%d Plätze, erwartet einer", len(gelesen.Plaetze))
	}

	if gelesen.Plaetze[0].Menge != 2 {
		t.Errorf("Menge %d, erwartet 2 — die letzte gewinnt",
			gelesen.Plaetze[0].Menge)
	}
}
