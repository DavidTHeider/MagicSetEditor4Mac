package app

// **Erneut versuchen öffnet denselben Bestand** — F05
// (#2344).
//
// Ein belegter Bestand vergeht von allein; der Rat dazu ist „gleich noch
// einmal". Bis hierher gab es dafür keinen Weg: das Fenster zeigte die
// Wiederherstellung, und wer nicht zurückspielen wollte, musste die
// Anwendung beenden. Jetzt wiederholt `StartWiederholen` den Start — und der
// gescheiterte Versuch davor hat die Datei nicht angefasst.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandMitEinemSet legt einen Bestand mit einem Set an und gibt seinen
// Pfad zurück — geschlossen.
func bestandMitEinemSet(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	erste := NeueApp(pfad)

	if err := erste.Starten(t.Context()); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := erste.db.ExecContext(t.Context(),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES ('Probe', 'probe', '2026-10-14T00:00:00.000Z', 1)`); err != nil {
		t.Fatalf("ein Set anlegen: %v", err)
	}

	if err := erste.Beenden(); err != nil {
		t.Fatalf("beenden: %v", err)
	}

	return pfad
}

func pruefsumme(t *testing.T, pfad string) [32]byte {
	t.Helper()

	roh, err := os.ReadFile(pfad) //nolint:gosec // ein Pfad aus t.TempDir
	if err != nil {
		t.Fatal(err)
	}

	return sha256.Sum256(roh)
}

func TestErneutVersuchenOeffnetDenselbenBestand(t *testing.T) {
	pfad := bestandMitEinemSet(t)
	vorher := pruefsumme(t, pfad)

	anwendung := NeueApp(pfad)
	belegt := true
	anwendung.oeffnen = func(ctx context.Context, quelle string) (*sql.DB, error) {
		if belegt {
			return nil, fmt.Errorf("Datenbank öffnen: %w", domain.MitGrund(domain.ErrBelegt,
				errors.New("database is locked (5) (SQLITE_BUSY)"), "gleich noch einmal"))
		}

		return storage.Oeffnen(ctx, quelle)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen

	if err := anwendung.Starten(t.Context()); err == nil {
		t.Fatal("der belegte Start kam durch")
	}

	if nachher := pruefsumme(t, pfad); nachher != vorher {
		t.Error("der belegte Start hat die Datei verändert")
	}

	belegt = false

	befund := anwendung.StartWiederholen(t.Context())
	if !befund.Offen || befund.Schluessel != "" {
		t.Fatalf("der zweite Versuch öffnete nicht: %+v", befund)
	}

	sets := anwendung.Sets(t.Context())
	if sets.Fehler != nil || len(sets.Sets) != 1 || sets.Sets[0].Name != "Probe" {
		t.Errorf("nicht derselbe Bestand: %+v", sets)
	}
}

func TestErneutVersuchenTutBeiOffenemBestandNichts(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	geoeffnet := 0
	anwendung.oeffnen = func(ctx context.Context, quelle string) (*sql.DB, error) {
		geoeffnet++

		return storage.Oeffnen(ctx, quelle)
	}

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = anwendung.Beenden() }) //nolint:errcheck // Aufräumen

	if befund := anwendung.StartWiederholen(t.Context()); !befund.Offen || geoeffnet != 1 {
		t.Errorf("Befund %+v, %d-mal geöffnet — erwartet offen und einmal", befund, geoeffnet)
	}
}

func TestErneutVersuchenNachDemSchlussOeffnetNicht(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.oeffnen = func(context.Context, string) (*sql.DB, error) {
		return nil, domain.ErrBelegt
	}

	_ = anwendung.Starten(t.Context()) //nolint:errcheck // er scheitert, darum geht es

	if err := anwendung.Beenden(); err != nil {
		t.Fatal(err)
	}

	geoeffnet := false
	anwendung.oeffnen = func(ctx context.Context, quelle string) (*sql.DB, error) {
		geoeffnet = true

		return storage.Oeffnen(ctx, quelle)
	}

	if befund := anwendung.StartWiederholen(t.Context()); befund.Offen || geoeffnet {
		t.Errorf("nach dem Schluss wurde geöffnet: %+v", befund)
	}
}

// **Ein Schluss, der während des Öffnens beginnt**, und dessen Frist abläuft,
// bevor er die Sperre bekommt — Befund 2 des Zweitblicks über Phase 1. Er
// kehrt ohne Abräumen zurück; den eben geöffneten Bestand muss die
// Wiederholung selbst schließen, wie `wiederOeffnen` es tut.
func TestErneutVersuchenSchliesstNachEinemSchlussWaehrendDesOeffnens(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.oeffnen = func(context.Context, string) (*sql.DB, error) {
		return nil, domain.ErrBelegt
	}

	_ = anwendung.Starten(t.Context()) //nolint:errcheck // er scheitert, darum geht es

	anwendung.oeffnen = func(ctx context.Context, quelle string) (*sql.DB, error) {
		// **Hier beginnt der Schluss** — er stellt das Tor und wartet auf
		// eine Sperre, die die Wiederholung hält.
		anwendung.lauf.tor.Store(torEndet)

		return storage.Oeffnen(ctx, quelle)
	}

	befund := anwendung.StartWiederholen(t.Context())

	if anwendung.db != nil {
		t.Errorf("nach dem Schluss blieb der Bestand offen (Befund %+v)", befund)
		_ = anwendung.bestandSchliessen() //nolint:errcheck // Aufräumen
	}
}
