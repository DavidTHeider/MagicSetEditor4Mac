package main

// **Der Server schreibt nicht mit den Rechten vom Start** — ID-03.AK-02
// (#1702).
//
// # Der Befund
//
// **Der Server hält eine Verbindung über seine ganze Laufzeit**, aber jeder
// Ruf arbeitet in dem Profil, das der Bestand **gerade** zeigt. Wechselte das
// Fenster nach dem Start als Leser in ein fremdes Profil, schrieb ein
// MCP-Werkzeug dort über die schreibende Verbindung vom Start.
//
// # Wie der Lauf es zeigt
//
// **Mit einer Leitung statt einer fertigen Sitzung**: zwischen zwei Rufen
// bewegt der Test den Zeiger im Bestand, wie das Fenster es täte — über eine
// zweite Verbindung, während der Server läuft.

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerServerFragtJedenSchreibendenRufNeu(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, bobsProfil := bestandMitFremdemProfil(t)

	ein, zuleitung := io.Pipe()
	aus := &mitgeschrieben{}
	fertig := make(chan error, 1)

	go func() { fertig <- lauf(probekontext(t), nil, ein, aus) }()

	sende := func(zeile string, antworten int) string {
		t.Helper()

		if _, fehler := io.WriteString(zuleitung, zeile+"\n"); fehler != nil {
			t.Fatalf("senden: %v", fehler)
		}

		return aus.warteAuf(t, antworten)
	}

	anlegen := func(id int, name string) string {
		return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"tools/call","params":`+
			`{"name":"schreiben.deck.anlegen","arguments":{"name":%q,"spiel":"magic"}}}`, id, name)
	}

	sende(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`, 1)

	// **Im eigenen Profil darf er** — sonst bewiese der Lauf nichts.
	if eigenes := sende(anlegen(2, "Eigenes"), 2); strings.Contains(eigenes, `"isError":true`) {
		t.Fatalf("im eigenen Profil abgewiesen: %s", eigenes)
	}

	zeigerSetzen(t, pfad, 1, bobsProfil)

	fremd := sende(anlegen(3, "Unberechtigt"), 3)
	if !strings.Contains(fremd, `"isError":true`) || !strings.Contains(fremd, "nur gelesen") {
		t.Errorf("als Leser im fremden Profil nicht abgewiesen: %s", fremd)
	}

	if fehler := zuleitung.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := <-fertig; fehler != nil {
		t.Fatalf("lauf: %v", fehler)
	}

	// **An der Wirkung** (AK-04): in Bobs Profil steht kein Deck.
	if decks := deckeIn(t, pfad, bobsProfil); len(decks) != 0 {
		t.Errorf("in Bobs Profil entstanden %v", decks)
	}
}

// bestandMitFremdemProfil legt Bob an und gibt dem ersten Nutzer in dessen
// Profil die Rolle Leser; der Zeiger bleibt im eigenen Profil.
func bestandMitFremdemProfil(t *testing.T) (pfad string, bobsProfil int64) {
	t.Helper()

	pfad, fehler := app.Datenbankpfad()
	if fehler != nil {
		t.Fatalf("Pfad: %v", fehler)
	}

	if verzeichnisfehler := os.MkdirAll(filepath.Dir(pfad), 0o700); verzeichnisfehler != nil {
		t.Fatalf("Verzeichnis: %v", verzeichnisfehler)
	}

	mitBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		if migrationsfehler := storage.Migriere(ctx, db, nil); migrationsfehler != nil {
			t.Fatalf("migrieren: %v", migrationsfehler)
		}

		bob, anlegefehler := storage.NeueNutzerablage(db).NutzerAnlegen(ctx, domain.Nutzer{Name: "Bob"})
		if anlegefehler != nil {
			t.Fatalf("Bob: %v", anlegefehler)
		}

		if lesefehler := db.QueryRowContext(ctx,
			`SELECT id FROM profile WHERE nutzer_id = ?`, bob.ID).Scan(&bobsProfil); lesefehler != nil {
			t.Fatalf("Bobs Profil: %v", lesefehler)
		}

		if setzfehler := storage.NeueFreigabenablage(db).FreigabeSetzen(ctx, domain.Profilfreigabe{
			ProfilID: bobsProfil, NutzerID: 1, Rolle: domain.RolleLeser,
		}); setzfehler != nil {
			t.Fatalf("Freigabe: %v", setzfehler)
		}
	})

	return pfad, bobsProfil
}

// zeigerSetzen bewegt den Zeiger über eine zweite Verbindung — wie das
// Fenster, während der Server läuft.
func zeigerSetzen(t *testing.T, pfad string, nutzerID, profilID int64) {
	t.Helper()

	mitBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		if fehler := storage.NeueNutzerablage(db).ZeigerSetzen(ctx, nutzerID, profilID); fehler != nil {
			t.Fatalf("Zeiger: %v", fehler)
		}
	})
}

// deckeIn gibt die Decks eines Profils.
func deckeIn(t *testing.T, pfad string, profilID int64) (decks []domain.Deck) {
	t.Helper()

	mitBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		gelesen, fehler := storage.NeueDeckablage(db).Decks(storage.MitProfil(ctx, profilID), "")
		if fehler != nil {
			t.Fatalf("Decks: %v", fehler)
		}

		decks = gelesen
	})

	return decks
}

// mitBestand öffnet den Bestand für eine Arbeit und schließt ihn danach.
func mitBestand(t *testing.T, pfad string, arbeit func(context.Context, *sql.DB)) {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	arbeit(probekontext(t), db)

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}
}

// mitgeschrieben sammelt die Antworten des Servers, während er läuft.
type mitgeschrieben struct {
	sperre sync.Mutex
	text   strings.Builder
}

func (m *mitgeschrieben) Write(teil []byte) (int, error) {
	m.sperre.Lock()
	defer m.sperre.Unlock()

	return m.text.Write(teil)
}

// warteAuf wartet, bis so viele Antwortzeilen dastehen, und gibt die letzte.
func (m *mitgeschrieben) warteAuf(t *testing.T, anzahl int) string {
	t.Helper()

	frist := time.Now().Add(10 * time.Second)

	for time.Now().Before(frist) {
		m.sperre.Lock()
		zeilen := strings.Split(strings.TrimSpace(m.text.String()), "\n")
		m.sperre.Unlock()

		if len(zeilen) >= anzahl && zeilen[0] != "" {
			return zeilen[anzahl-1]
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("keine %d. Antwort in zehn Sekunden", anzahl)

	return ""
}
