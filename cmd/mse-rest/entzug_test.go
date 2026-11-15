package main

// **Der Dienst schreibt nicht mit den Rechten vom Start** — ID-03.AK-02
// (#1702).
//
// Dieselbe Lage wie beim MCP-Server (`cmd/mse-mcp/entzug_test.go`): die
// Verbindung steht seit dem Start, jede Anfrage arbeitet im Profil, das der
// Bestand **jetzt** zeigt. Zwischen zwei Anfragen bewegt der Test den Zeiger
// als Leser in ein fremdes Profil, wie das Fenster es täte.

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerDienstFragtJedeSchreibendeAnfrageNeu(t *testing.T) {
	zuhause := t.TempDir()
	pfad := filepath.Join(zuhause, "bestand.db")

	t.Setenv("HOME", zuhause)
	t.Setenv("MSE_BESTAND", pfad)

	var bobsProfil int64

	aufBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		if fehler := storage.Migriere(ctx, db, nil); fehler != nil {
			t.Fatalf("migrieren: %v", fehler)
		}

		bob, fehler := storage.NeueNutzerablage(db).NutzerAnlegen(ctx, domain.Nutzer{Name: "Bob"})
		if fehler != nil {
			t.Fatalf("Bob: %v", fehler)
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

	ctx, beenden := context.WithCancel(probekontext(t))
	angesagt := &sofort{ziel: &bytes.Buffer{}}
	fertig := make(chan error, 1)

	go func() { fertig <- lauf(ctx, []string{"-port", "0"}, angesagt) }()

	adresse, marke := warteAufAnsage(t, angesagt)

	anlegen := func(name string) int {
		t.Helper()

		anfrage, fehler := http.NewRequestWithContext(ctx, http.MethodPost,
			"http://"+adresse+"/decks", strings.NewReader(`{"name":"`+name+`","spiel":"magic"}`))
		if fehler != nil {
			t.Fatalf("Anfrage bauen: %v", fehler)
		}

		anfrage.Header.Set("Authorization", "Bearer "+marke)
		anfrage.Header.Set("Content-Type", "application/json")

		antwort, rufehler := http.DefaultClient.Do(anfrage)
		if rufehler != nil {
			t.Fatalf("fragen: %v", rufehler)
		}

		defer schliesse(t, antwort.Body)

		return antwort.StatusCode
	}

	// **Im eigenen Profil darf er** — sonst bewiese der Lauf nichts.
	if status := anlegen("Eigenes"); status >= http.StatusBadRequest {
		t.Fatalf("im eigenen Profil abgewiesen: %d", status)
	}

	aufBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		if fehler := storage.NeueNutzerablage(db).ZeigerSetzen(ctx, 1, bobsProfil); fehler != nil {
			t.Fatalf("Zeiger: %v", fehler)
		}
	})

	if status := anlegen("Unberechtigt"); status != http.StatusForbidden {
		t.Errorf("als Leser im fremden Profil: %d, erwartet %d", status, http.StatusForbidden)
	}

	beenden()

	if fehler := <-fertig; fehler != nil {
		t.Errorf("der Dienst endete mit einem Fehler: %v", fehler)
	}

	// **An der Wirkung** (AK-04): in Bobs Profil steht kein Deck.
	aufBestand(t, pfad, func(ctx context.Context, db *sql.DB) {
		decks, fehler := storage.NeueDeckablage(db).Decks(storage.MitProfil(ctx, bobsProfil), "")
		if fehler != nil || len(decks) != 0 {
			t.Errorf("in Bobs Profil: %v (%v)", decks, fehler)
		}
	})
}

// aufBestand öffnet den Bestand für eine Arbeit und schließt ihn danach.
func aufBestand(t *testing.T, pfad string, arbeit func(context.Context, *sql.DB)) {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	arbeit(probekontext(t), db)

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatalf("schließen: %v", schliessfehler)
	}
}
