package storage_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein voller Datenträger lässt den alten Stand stehen** — FMT-06.AK-03
// (#1695).
//
// # Wie er entsteht, ohne einen Datenträger zu füllen
//
// **SQLite gibt dieselbe Antwort**, wenn die Datei die Seitenzahl erreicht,
// die ihr erlaubt ist: `PRAGMA max_page_count`, Fehlercode `SQLITE_FULL` —
// „database or disk is full". Der Bestand hat **eine** Verbindung
// (`datenbank.go`, `SetMaxOpenConns(1)`), die Grenze gilt also für jede
// Transaktion, die danach kommt.
//
// # Was er zeigt
//
// Ein Set, das nicht mehr hineinpasst, kommt **gar nicht** hinein — kein Set
// mit halben Karten —, und was vorher dastand, steht unverändert da. Die Datei
// ist danach heil (`integrity_check`).
//
// # Und wie er heißt
//
// **`SQLITE_FULL` ist [domain.ErrVoll]** — seit
// #1812.
// Welcher Satz daraus im Fenster wird, prüft `internal/app`.
func TestEinVollerDatentraegerLaesstDenAltenStandStehen(t *testing.T) {
	t.Parallel()

	ctx := probekontext(t)

	db, oeffnungsfehler := storage.Oeffnen(ctx, filepath.Join(t.TempDir(), "bestand.db"))
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	t.Cleanup(func() {
		if schliessfehler := db.Close(); schliessfehler != nil {
			t.Errorf("schließen: %v", schliessfehler)
		}
	})

	if fehler := storage.Migriere(ctx, db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	speicher := storage.NeuerSetSpeicher(db)

	if _, fehler := speicher.SetAnlegen(ctx, setMitKarten("Erstes", 3), nil); fehler != nil {
		t.Fatalf("das erste Set: %v", fehler)
	}

	var seiten int
	if fehler := db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&seiten); fehler != nil {
		t.Fatalf("Seitenzahl: %v", fehler)
	}

	if _, fehler := db.ExecContext(ctx, fmt.Sprintf("PRAGMA max_page_count = %d", seiten)); fehler != nil {
		t.Fatalf("Grenze setzen: %v", fehler)
	}

	_, voll := speicher.SetAnlegen(ctx, setMitKarten("Zweites", 400), nil)
	if voll == nil {
		t.Fatal("das zweite Set passt nicht hinein und kam trotzdem an")
	}

	// **Und er heißt so** (#1812): bis zum 2026-09-23 kam er unverändert
	// durch und im Fenster als „unerwarteter Fehler" an.
	if !errors.Is(voll, domain.ErrVoll) {
		t.Fatalf("erwartet war der volle Datenträger, gekommen ist %v", voll)
	}

	if grund := domain.GrundVon(voll); grund == nil || !strings.Contains(grund.Error(), "full") {
		t.Errorf("der Treibertext fehlt im Grund: %v", grund)
	}

	var pruefung string
	if fehler := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&pruefung); fehler != nil {
		t.Fatalf("Prüfung: %v", fehler)
	}

	if pruefung != "ok" {
		t.Errorf("die Datei ist nicht heil: %s", pruefung)
	}

	var sets, karten int
	if fehler := db.QueryRowContext(ctx, "SELECT count(*) FROM sets").Scan(&sets); fehler != nil {
		t.Fatalf("Sets zählen: %v", fehler)
	}

	if fehler := db.QueryRowContext(ctx, "SELECT count(*) FROM cards").Scan(&karten); fehler != nil {
		t.Fatalf("Karten zählen: %v", fehler)
	}

	if sets != 1 || karten != 3 {
		t.Errorf("nach dem gescheiterten Anlegen %d Sets und %d Karten, erwartet 1 und 3", sets, karten)
	}
}

// setMitKarten baut ein Set mit so vielen Karten, jede mit einem langen Text.
func setMitKarten(name string, anzahl int) *domain.Set {
	set := &domain.Set{Name: name, Game: "magic"}

	for stelle := range anzahl {
		set.Cards = append(set.Cards, &domain.Card{Values: []domain.Value{
			{Field: "name", Text: fmt.Sprintf("%s %d", name, stelle)},
			{Field: "rule text", Text: strings.Repeat("Ein langer Regeltext. ", 40)},
		}})
	}

	return set
}
