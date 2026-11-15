package storage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Übernahme eines Decks ist ein Zug** —
// #1820,
// Befund 6 des Zweitblicks über Phase 6.1.
//
// # Der Befund
//
// `DeckUebernehmen` legte das Deck an und schrieb danach **einzeln** Plätze
// und Herkunft. Scheiterte ein späterer Schritt, blieb das angelegte Deck
// stehen — halb, und niemand erfuhr es.

func uebernahmeplaetze() []domain.Deckplatz {
	return []domain.Deckplatz{
		{Printing: "a1b2c3d4-0000-4000-8000-00000000000a", Menge: 4, Bereich: domain.BereichHauptdeck, Bezeichnung: "Blitz"},
		{Printing: "a1b2c3d4-0000-4000-8000-00000000000b", Menge: 2, Bereich: domain.BereichHauptdeck, Bezeichnung: "Glut"},
		{Printing: "a1b2c3d4-0000-4000-8000-00000000000c", Menge: 1, Bereich: domain.BereichSideboard, Bezeichnung: "Asche"},
	}
}

var uebernahmeherkunft = domain.Herkunft{
	Art: domain.HerkunftProfil, Quelle: "Zweiter · Mein Profil", QuellProfil: 2,
}

func TestEineUebernahmeLegtDeckPlaetzeUndHerkunftAn(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	angelegt, fehler := ablage.Uebernehmen(probekontext(t),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe", Notiz: "vom Zweiten"},
		uebernahmeplaetze(), uebernahmeherkunft, 41)
	if fehler != nil {
		t.Fatalf("übernehmen: %v", fehler)
	}

	// **Die Antwort zählt wie das Lesen** — bis FMT-04 kam sie mit null
	// Karten im Hauptdeck zurück, und wer sie anzeigte, zeigte eine Null.
	if angelegt.Hauptdeck+angelegt.Sideboard == 0 {
		t.Errorf("die Antwort zählt nichts: %d im Hauptdeck, %d im Sideboard", angelegt.Hauptdeck, angelegt.Sideboard)
	}

	gelesen, lesefehler := ablage.Deck(probekontext(t), angelegt.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen.Name != "Nachtlauf" || gelesen.Notiz != "vom Zweiten" || len(gelesen.Plaetze) != 3 {
		t.Errorf("zurück kam %+v", gelesen)
	}

	herkunft, herkunftfehler := storage.NeueHerkunftsablage(db).Deckherkunft(
		probekontext(t), angelegt.ID)
	if herkunftfehler != nil || herkunft.Quelle != uebernahmeherkunft.Quelle ||
		herkunft.QuellProfil != 2 {
		t.Errorf("Herkunft = %+v (%v)", herkunft, herkunftfehler)
	}
}

// **Scheitert ein Schritt, bleibt nichts** — weder das Deck noch ein Platz
// noch eine Herkunft. Die Auslöser stehen nur in diesem Lauf: sie lassen
// den dritten Platz scheitern, oder die Herkunft.
func TestEineGescheiterteUebernahmeHinterlaesstNichts(t *testing.T) {
	for _, fall := range []struct {
		name     string
		ausloser string
	}{
		{
			"beim dritten Platz",
			`CREATE TRIGGER dritter BEFORE INSERT ON deckplaetze
			 WHEN (SELECT COUNT(*) FROM deckplaetze WHERE deck_id = NEW.deck_id) >= 2
			 BEGIN SELECT RAISE(ABORT, 'der dritte Platz scheitert'); END`,
		},
		{
			"beim Deck selbst",
			`CREATE TRIGGER deck BEFORE INSERT ON decks
			 BEGIN SELECT RAISE(ABORT, 'das Deck scheitert'); END`,
		},
		{
			"bei der Herkunft",
			`CREATE TRIGGER herkunft BEFORE INSERT ON deck_herkunft
			 BEGIN SELECT RAISE(ABORT, 'die Herkunft scheitert'); END`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db := frischeDatenbank(t)
			if _, fehler := db.ExecContext(probekontext(t), fall.ausloser); fehler != nil {
				t.Fatalf("Auslöser: %v", fehler)
			}

			_, fehler := storage.NeueDeckablage(db).Uebernehmen(probekontext(t),
				domain.Deck{Name: "Nachtlauf", Spiel: "probe"},
				uebernahmeplaetze(), uebernahmeherkunft, 41)
			if fehler == nil {
				t.Fatal("die Übernahme scheiterte nicht")
			}

			for tabelle, erwartet := range map[string]int{
				"decks": 0, "deckplaetze": 0, "deck_herkunft": 0,
			} {
				if gezaehlt := zeilen(t, db, tabelle); gezaehlt != erwartet {
					t.Errorf("%s: %d Zeilen, erwartet %d — ein halbes Deck", tabelle, gezaehlt, erwartet)
				}
			}
		})
	}
}

// **Eine ungültige Herkunft wird vor dem Zug abgewiesen** — sie ginge sonst
// als Datenbankfehler hinaus, und ein Deck wäre schon angelegt.
func TestEineUebernahmeMitUngueltigerHerkunftLegtNichtsAn(t *testing.T) {
	db := frischeDatenbank(t)

	_, fehler := storage.NeueDeckablage(db).Uebernehmen(probekontext(t),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"},
		uebernahmeplaetze(), domain.Herkunft{Art: "erfunden"}, 41)
	if fehler == nil {
		t.Fatal("eine ungültige Herkunft ging durch")
	}

	if gezaehlt := zeilen(t, db, "decks"); gezaehlt != 0 {
		t.Errorf("%d Decks, erwartet keines", gezaehlt)
	}
}

func zeilen(t *testing.T, db *sql.DB, tabelle string) int {
	t.Helper()

	var anzahl int
	if fehler := db.QueryRowContext(probekontext(t),
		"SELECT COUNT(*) FROM "+tabelle).Scan(&anzahl); fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	return anzahl
}

// **Ohne Profil im Kontext gibt es keinen Ort** — Vermerk 0066.
func TestEineUebernahmeOhneProfilLegtNichtsAn(t *testing.T) {
	db := frischeDatenbank(t)

	_, fehler := storage.NeueDeckablage(db).Uebernehmen(context.Background(),
		domain.Deck{Name: "Nachtlauf", Spiel: "probe"},
		uebernahmeplaetze(), uebernahmeherkunft, 41)
	if fehler == nil {
		t.Fatal("ohne Profil ging die Übernahme durch")
	}

	if gezaehlt := zeilen(t, db, "decks"); gezaehlt != 0 {
		t.Errorf("%d Decks, erwartet keines", gezaehlt)
	}
}
