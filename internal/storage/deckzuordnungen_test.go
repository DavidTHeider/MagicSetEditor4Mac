package storage_test

// **Die Zuordnung im Bestand** — Werkstück U9 von Phase 5.1
// (#1420):
// je Platz und Exemplar eine Zeile, die Menge ersetzbar, und sie verschwindet
// mit dem Deck, dem Platz oder dem Exemplar.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zuordnungsstand baut ein Deck mit einem Platz und ein Exemplar desselben Drucks.
func zuordnungsstand(t *testing.T) (db *sql.DB, deckID, platzID, exemplarID int64) {
	t.Helper()

	db = frischeDatenbank(t)
	decks := storage.NeueDeckablage(db)
	deck := deckAnlegen(t, decks)

	platz, err := decks.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: probedruck, Menge: 4, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("Platz legen: %v", err)
	}

	exemplarID, err = storage.NeuerSammlungsspeicher(db).Anlegen(probekontext(t), probeexemplar())
	if err != nil {
		t.Fatalf("Exemplar anlegen: %v", err)
	}

	return db, deck.ID, platz.ID, exemplarID
}

func TestEineZuordnungWirdAbgelegtErsetztUndGeloest(t *testing.T) {
	db, deckID, platzID, exemplarID := zuordnungsstand(t)
	ablage := storage.NeueDeckzuordnungen(db)

	erste, err := ablage.Setzen(probekontext(t), domain.Deckzuordnung{
		DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: 2,
	})
	if err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	// **Dieselbe Zeile noch einmal ist eine neue Menge, keine zweite Zeile.**
	zweite, err := ablage.Setzen(probekontext(t), domain.Deckzuordnung{
		DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: 3,
	})
	if err != nil {
		t.Fatalf("zweites Setzen: %v", err)
	}

	if zweite.ID != erste.ID || zweite.Menge != 3 {
		t.Errorf("zweite Zuordnung %+v, erste %+v", zweite, erste)
	}

	jeDeck, err := ablage.JeDeck(probekontext(t), deckID)
	if err != nil {
		t.Fatalf("JeDeck: %v", err)
	}

	if len(jeDeck) != 1 || jeDeck[0].Menge != 3 || jeDeck[0].ExemplarID != exemplarID {
		t.Errorf("JeDeck: %+v", jeDeck)
	}

	// **Der Rückweg**: vom Exemplar zu seinen Decks.
	jeExemplar, err := ablage.JeExemplar(probekontext(t), exemplarID)
	if err != nil {
		t.Fatalf("JeExemplar: %v", err)
	}

	if len(jeExemplar) != 1 || jeExemplar[0].DeckID != deckID {
		t.Errorf("JeExemplar: %+v", jeExemplar)
	}

	zugeordnet, err := ablage.ZugeordnetJeExemplar(probekontext(t))
	if err != nil {
		t.Fatalf("ZugeordnetJeExemplar: %v", err)
	}

	if zugeordnet[exemplarID] != 3 {
		t.Errorf("zugeordnet: %v", zugeordnet)
	}

	if fehler := ablage.Loesen(probekontext(t), deckID, erste.ID); fehler != nil {
		t.Fatalf("Loesen: %v", fehler)
	}

	if fehler := ablage.Loesen(probekontext(t), deckID, erste.ID); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("zweites Lösen: %v, erwartet ErrNotFound", fehler)
	}

	danach, err := ablage.JeDeck(probekontext(t), deckID)
	if err != nil || len(danach) != 0 {
		t.Errorf("nach dem Lösen: %+v, %v", danach, err)
	}
}

func TestZweiDecksTeilenSichDieZuordnungenEinesExemplars(t *testing.T) {
	db, deckID, platzID, exemplarID := zuordnungsstand(t)
	decks := storage.NeueDeckablage(db)
	ablage := storage.NeueDeckzuordnungen(db)

	zweites, err := decks.Anlegen(probekontext(t), domain.Deck{Name: "Zweites", Spiel: "magic"})
	if err != nil {
		t.Fatalf("zweites Deck: %v", err)
	}

	zweiterPlatz, err := decks.PlatzSetzen(probekontext(t), zweites.ID, domain.Deckplatz{
		Printing: probedruck, Menge: 1, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("zweiter Platz: %v", err)
	}

	for _, zuordnung := range []domain.Deckzuordnung{
		{DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: 2},
		{DeckID: zweites.ID, PlatzID: zweiterPlatz.ID, ExemplarID: exemplarID, Menge: 1},
	} {
		if _, fehler := ablage.Setzen(probekontext(t), zuordnung); fehler != nil {
			t.Fatalf("Setzen: %v", fehler)
		}
	}

	zugeordnet, err := ablage.ZugeordnetJeExemplar(probekontext(t))
	if err != nil || zugeordnet[exemplarID] != 3 {
		t.Errorf("zugeordnet %v, %v — erwartet 3", zugeordnet, err)
	}

	// **Eine Zuordnung gehört zu ihrem Deck**: über ein fremdes Deck ist sie nicht zu lösen.
	jeDeck, jefehler := ablage.JeDeck(probekontext(t), deckID)
	if jefehler != nil {
		t.Fatalf("JeDeck: %v", jefehler)
	}

	if err := ablage.Loesen(probekontext(t), zweites.ID, jeDeck[0].ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Lösen über das fremde Deck: %v", err)
	}
}

func TestEineZuordnungVerschwindetMitDeckPlatzOderExemplar(t *testing.T) {
	faelle := map[string]func(t *testing.T, db *sql.DB, deckID, platzID, exemplarID int64){
		"das Deck": func(t *testing.T, db *sql.DB, deckID, _, _ int64) {
			if err := storage.NeueDeckablage(db).Loeschen(probekontext(t), deckID); err != nil {
				t.Fatalf("Deck löschen: %v", err)
			}
		},
		"der Platz": func(t *testing.T, db *sql.DB, deckID, platzID, _ int64) {
			if err := storage.NeueDeckablage(db).PlatzLoeschen(probekontext(t), deckID, platzID); err != nil {
				t.Fatalf("Platz löschen: %v", err)
			}
		},
		"das Exemplar": func(t *testing.T, db *sql.DB, _, _, exemplarID int64) {
			if err := storage.NeuerSammlungsspeicher(db).Loeschen(probekontext(t), exemplarID, 1); err != nil {
				t.Fatalf("Exemplar löschen: %v", err)
			}
		},
	}

	for name, weg := range faelle {
		t.Run(name, func(t *testing.T) {
			db, deckID, platzID, exemplarID := zuordnungsstand(t)
			ablage := storage.NeueDeckzuordnungen(db)

			if _, err := ablage.Setzen(probekontext(t), domain.Deckzuordnung{
				DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: 1,
			}); err != nil {
				t.Fatalf("Setzen: %v", err)
			}

			weg(t, db, deckID, platzID, exemplarID)

			zugeordnet, err := ablage.ZugeordnetJeExemplar(probekontext(t))
			if err != nil || len(zugeordnet) != 0 {
				t.Errorf("nach dem Löschen: %v, %v", zugeordnet, err)
			}
		})
	}
}

func TestDieZuordnungenMeldenEinenGeschlossenenBestand(t *testing.T) {
	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	ablage := storage.NeueDeckzuordnungen(db)
	ctx := probekontext(t)

	if _, err := ablage.Setzen(ctx, domain.Deckzuordnung{DeckID: 1, PlatzID: 1, ExemplarID: 1, Menge: 1}); err == nil {
		t.Error("Setzen kam durch")
	}

	if err := ablage.Loesen(ctx, 1, 1); err == nil {
		t.Error("Loesen kam durch")
	}

	if _, err := ablage.JeDeck(ctx, 1); err == nil {
		t.Error("JeDeck kam durch")
	}

	if _, err := ablage.JeExemplar(ctx, 1); err == nil {
		t.Error("JeExemplar kam durch")
	}

	if _, err := ablage.ZugeordnetJeExemplar(ctx); err == nil {
		t.Error("ZugeordnetJeExemplar kam durch")
	}

	// **Und ohne Profil im Kontext gibt es keine Summe** (0066).
	if _, err := storage.NeueDeckzuordnungen(frischeDatenbank(t)).ZugeordnetJeExemplar(context.Background()); err == nil {
		t.Error("die Summe ohne Profil kam durch")
	}
}

// **Eine ungültige Zuordnung kommt nicht in den Bestand.**
func TestEineUngueltigeZuordnungWirdAbgewiesen(t *testing.T) {
	db, deckID, platzID, exemplarID := zuordnungsstand(t)

	_, err := storage.NeueDeckzuordnungen(db).Setzen(probekontext(t), domain.Deckzuordnung{
		DeckID: deckID, PlatzID: platzID, ExemplarID: exemplarID, Menge: 0,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Eine Zeile, die nicht in ihre Felder passt, ist ein Fehler** — für jeden der drei Lesewege.
func TestEineUnlesbareZuordnungIstEinFehler(t *testing.T) {
	db, deckID, _, exemplarID := zuordnungsstand(t)
	ablage := storage.NeueDeckzuordnungen(db)

	for _, satz := range []string{
		`DROP TABLE deckzuordnungen`,
		`CREATE TABLE deckzuordnungen (id INTEGER PRIMARY KEY, deck_id INTEGER, platz_id INTEGER, exemplar_id, menge TEXT)`,
		`INSERT INTO deckzuordnungen (deck_id, platz_id, exemplar_id, menge) VALUES (` +
			fmt.Sprint(deckID) + `, 1, ` + fmt.Sprint(exemplarID) + `, 'keine Zahl')`,
	} {
		mussAusfuehren(t, db, satz)
	}

	if _, err := ablage.JeDeck(probekontext(t), deckID); err == nil {
		t.Error("JeDeck las die kaputte Zeile")
	}

	if _, err := ablage.JeExemplar(probekontext(t), exemplarID); err == nil {
		t.Error("JeExemplar las die kaputte Zeile")
	}

	// **Auch die Summe**: ohne `exemplar_id` als Zahl ist die Gruppe keine.
	mussAusfuehren(t, db, `UPDATE deckzuordnungen SET exemplar_id = 'x'`)

	if _, err := ablage.ZugeordnetJeExemplar(probekontext(t)); err == nil {
		t.Error("ZugeordnetJeExemplar las die kaputte Summe")
	}
}
