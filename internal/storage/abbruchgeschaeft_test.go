package storage_test

import (
	"database/sql"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Prozess, der mitten im Geschäft stirbt, hinterlässt den alten Stand**
// — ENG-05.AK-04 (#1730).
//
// # Was die übrigen Läufe nicht zeigen
//
// `absturz_test.go` tötet einen Dienst **nach** einer quittierten Antwort
// und sieht, dass sie bleibt. `vollerbestand_test.go` lässt ein Schreiben
// **scheitern**. Hier stirbt ein Prozess, während sein Geschäft **offen**
// ist: tausend Decks sind geschrieben, ein vorhandenes umbenannt, nichts ist
// festgeschrieben — und kein Aufräumen läuft mehr.
//
// # Was er zeigt
//
// Danach ist die Datei heil (`integrity_check`), kein Fremdschlüssel zeigt
// ins Leere, **nichts** vom offenen Geschäft steht da, und der nächste
// schreibt sofort — keine Sperre hat den Prozess überlebt.
//
// # Wie das Kind sein Ziel erfährt
//
// **Über seine Argumente nach `--`, nicht über die Umgebung**: eine neue
// Umgebungsvariable gehörte in den Plan der Messläufe
// (`docs/belege/fremdpakete.plan.json`), und ein geänderter Plan machte den
// Beleg ungültig.
func TestEinAbbruchMittenImGeschaeftLaesstDenAltenStand(t *testing.T) {
	t.Parallel()

	ctx := probekontext(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")
	bereit := ort + ".bereit"

	vorher := zaehleDecks(t, ort, func(db *sql.DB) {
		if _, fehler := storage.NeueDeckablage(db).Anlegen(ctx, domain.Deck{Name: "Alt", Spiel: "magic"}); fehler != nil {
			t.Fatalf("das alte Deck: %v", fehler)
		}
	})

	// #nosec G204 G702 -- das eigene Testprogramm, mit festem Lauf
	kind := exec.CommandContext(t.Context(), os.Args[0],
		"-test.run=^TestAbbruchKindHaeltEinGeschaeftOffen$", "--", ort, bereit)

	if fehler := kind.Start(); fehler != nil {
		t.Fatalf("das Kind starten: %v", fehler)
	}

	warteAuf(t, bereit)

	if fehler := kind.Process.Kill(); fehler != nil {
		t.Fatalf("das Kind töten: %v", fehler)
	}

	_ = kind.Wait() //nolint:errcheck // getötet — sein Ausgang ist der Grund

	nachher := zaehleDecks(t, ort, func(db *sql.DB) {
		var integritaet string
		if fehler := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integritaet); fehler != nil ||
			integritaet != "ok" {
			t.Errorf("die Integrität nach dem Abbruch: %q (%v)", integritaet, fehler)
		}

		verwaist, fehler := db.QueryContext(ctx, "PRAGMA foreign_key_check")
		if fehler != nil {
			t.Fatalf("die Fremdschlüssel prüfen: %v", fehler)
		}

		defer func() { _ = verwaist.Close() }() //nolint:errcheck // nur gelesen

		if verwaist.Next() {
			t.Error("nach dem Abbruch zeigt ein Fremdschlüssel ins Leere")
		}

		if lesefehler := verwaist.Err(); lesefehler != nil {
			t.Errorf("die Fremdschlüssel prüfen: %v", lesefehler)
		}

		decks, fehler := storage.NeueDeckablage(db).Decks(ctx, "")
		if fehler != nil || len(decks) != 1 || decks[0].Name != "Alt" {
			t.Errorf("nach dem Abbruch stehen %d Decks da (%v), erwartet nur das alte", len(decks), fehler)
		}
	})

	if nachher != vorher {
		t.Errorf("%d Decks nach dem Abbruch, %d vorher", nachher, vorher)
	}
}

// TestAbbruchKindHaeltEinGeschaeftOffen ist das Kind — es schreibt in einem
// offenen Geschäft, meldet sich und wartet, bis es getötet wird.
//
// **Übersprungen, wenn es kein Kind ist**, und ausdrücklich: ein Lauf, der
// hier grün meldete, ohne zu schreiben, stünde als bestanden da.
func TestAbbruchKindHaeltEinGeschaeftOffen(t *testing.T) {
	if len(flag.Args()) != 2 {
		t.Skip("läuft nur als Kindprozess von TestEinAbbruchMittenImGeschaeftLaesstDenAltenStand")
	}

	ort, bereit := flag.Args()[0], flag.Args()[1]
	ctx := probekontext(t)

	db, fehler := storage.Oeffnen(ctx, ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	geschaeft, fehler := db.BeginTx(ctx, nil)
	if fehler != nil {
		t.Fatalf("das Geschäft beginnen: %v", fehler)
	}

	if _, fehler := geschaeft.ExecContext(ctx, `UPDATE decks SET name = 'Neu' WHERE name = 'Alt'`); fehler != nil {
		t.Fatalf("umbenennen: %v", fehler)
	}

	for nummer := range 1000 {
		if _, fehler := geschaeft.ExecContext(ctx,
			`INSERT INTO decks (name, spiel, profil_id) VALUES (?, 'magic', 1)`,
			"Offen "+string(rune('A'+nummer%26))+string(rune('A'+nummer/26))); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	if fehler := os.WriteFile(bereit, nil, 0o600); fehler != nil {
		t.Fatalf("sich melden: %v", fehler)
	}

	time.Sleep(time.Hour)
}

// zaehleDecks öffnet den Bestand, lässt eine Arbeit daran tun und zählt dann
// die Decks — und schreibt eines, zum Beleg, dass keine Sperre übrig ist.
func zaehleDecks(t *testing.T, ort string, arbeit func(db *sql.DB)) int {
	t.Helper()

	ctx := probekontext(t)

	db, fehler := storage.Oeffnen(ctx, ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // am Ende eines Laufs

	if migrationsfehler := storage.Migriere(ctx, db, nil); migrationsfehler != nil {
		t.Fatalf("migrieren: %v", migrationsfehler)
	}

	arbeit(db)

	var zahl int
	if zaehlfehler := db.QueryRowContext(ctx, `SELECT count(*) FROM decks`).Scan(&zahl); zaehlfehler != nil {
		t.Fatalf("zählen: %v", zaehlfehler)
	}

	// **Und der Nächste schreibt sofort** — in einer eigenen Transaktion, die
	// nach dem Zählen zurückgenommen wird, damit die Zahl stimmt.
	geschaeft, fehler := db.BeginTx(ctx, nil)
	if fehler != nil {
		t.Fatalf("nach dem Abbruch ließ sich nicht schreiben: %v", fehler)
	}

	if _, schreibfehler := geschaeft.ExecContext(ctx,
		`INSERT INTO decks (name, spiel, profil_id) VALUES ('Probe', 'magic', 1)`); schreibfehler != nil &&
		!errors.Is(schreibfehler, domain.ErrDoppelt) {
		t.Errorf("nach dem Abbruch ließ sich nicht schreiben: %v", schreibfehler)
	}

	_ = geschaeft.Rollback() //nolint:errcheck // nur geprobt

	return zahl
}

// warteAuf wartet, bis eine Datei erscheint.
func warteAuf(t *testing.T, pfad string) {
	t.Helper()

	frist := time.Now().Add(30 * time.Second)

	for time.Now().Before(frist) {
		if _, fehler := os.Stat(pfad); fehler == nil {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("das Kind hat sich in 30 Sekunden nicht gemeldet")
}
