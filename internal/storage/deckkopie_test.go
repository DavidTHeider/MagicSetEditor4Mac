package storage_test

// **Ein Deck lässt sich kopieren** — DOM-03.AK-02
// (#1698).
//
// Die Kopie ist ein **eigenes** Deck: Spiel, Notiz, Format und jeder Platz
// in seiner Reihenfolge kommen mit, der Name ist neu. Was an Papier hängt —
// Zuordnungen, Leihen, Marken — bleibt beim Original.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func kopiervorlage(t *testing.T, ablage *storage.Deckablage) domain.Deck {
	t.Helper()

	vorlage, fehler := ablage.Uebernehmen(probekontext(t),
		domain.Deck{Name: "Aschenrat", Spiel: "probe", Notiz: "rot"},
		uebernahmeplaetze(), uebernahmeherkunft, 41)
	if fehler != nil {
		t.Fatalf("die Vorlage: %v", fehler)
	}

	return vorlage
}

func TestEineKopieBringtJedenPlatzInSeinerReihenfolgeMit(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	vorlage := kopiervorlage(t, ablage)

	kopie, fehler := ablage.Kopieren(probekontext(t), vorlage.ID, "Aschenrat II")
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if kopie.ID == vorlage.ID || kopie.Name != "Aschenrat II" ||
		kopie.Spiel != "probe" || kopie.Notiz != "rot" || kopie.Fassung != 1 {
		t.Errorf("die Kopie ist %+v", kopie)
	}

	if kopie.Hauptdeck != 6 || kopie.Sideboard != 1 {
		t.Errorf("gezählt %d/%d, erwartet 6/1", kopie.Hauptdeck, kopie.Sideboard)
	}

	gelesen, lesefehler := ablage.Deck(probekontext(t), kopie.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	erwartet := uebernahmeplaetze()
	if len(gelesen.Plaetze) != len(erwartet) {
		t.Fatalf("%d Plätze, erwartet %d", len(gelesen.Plaetze), len(erwartet))
	}

	for stelle, platz := range gelesen.Plaetze {
		if platz.Printing != erwartet[stelle].Printing || platz.Menge != erwartet[stelle].Menge ||
			platz.Bereich != erwartet[stelle].Bereich {
			t.Errorf("Platz %d ist %+v, erwartet %+v", stelle, platz, erwartet[stelle])
		}

		if platz.ID == vorlage.Plaetze[stelle].ID {
			t.Errorf("Platz %d teilt seine Kennung mit dem Original", stelle)
		}
	}

	// **Das Original bleibt, wie es war** — und die Herkunft gehört ihm.
	original, originalfehler := ablage.Deck(probekontext(t), vorlage.ID)
	if originalfehler != nil || len(original.Plaetze) != 3 || original.Name != "Aschenrat" {
		t.Errorf("das Original ist jetzt %+v", original)
	}

	if gelesen.Herkunft.Art != "" {
		t.Errorf("die Kopie trägt die Herkunft %+v — sie ist hier entstanden", gelesen.Herkunft)
	}
}

// **Das Format reist mit** — wer ein Standard-Deck kopiert, will ein
// Standard-Deck.
func TestEineKopieBehaeltDasFormat(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	vorlage := kopiervorlage(t, ablage)

	format, fehler := storage.NeueFormatablage(db).Anlegen(probekontext(t), domain.Format{
		Name: "Probeformat", Spiel: "probe", Mindestgroesse: 1,
	})
	if fehler != nil {
		t.Fatalf("das Format: %v", fehler)
	}

	vorlage.FormatID = format.ID
	if _, aenderfehler := ablage.Aendern(probekontext(t), vorlage); aenderfehler != nil {
		t.Fatalf("das Format wählen: %v", aenderfehler)
	}

	kopie, kopierfehler := ablage.Kopieren(probekontext(t), vorlage.ID, "Mit Format")
	if kopierfehler != nil || kopie.FormatID != format.ID {
		t.Errorf("Format der Kopie %d, erwartet %d (%v)", kopie.FormatID, format.ID, kopierfehler)
	}
}

// **Ein Name, den es im Spiel schon gibt, ist ein Feldfehler** — dieselbe
// Antwort wie beim Anlegen, und es entsteht nichts.
func TestEineKopieMitVergebenemNamenLegtNichtsAn(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	vorlage := kopiervorlage(t, ablage)

	_, fehler := ablage.Kopieren(probekontext(t), vorlage.ID, "Aschenrat")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "name" {
		t.Fatalf("Fehler = %v, erwartet einen Feldfehler am Namen", fehler)
	}

	if gezaehlt := zeilen(t, db, "decks"); gezaehlt != 1 {
		t.Errorf("%d Decks, erwartet nur das Original", gezaehlt)
	}

	if gezaehlt := zeilen(t, db, "deckplaetze"); gezaehlt != 3 {
		t.Errorf("%d Plätze, erwartet nur die des Originals", gezaehlt)
	}
}

// **Ein fremdes oder fehlendes Deck gibt es nicht** — auch nicht zum
// Kopieren. Sonst holte sich ein Profil die Liste eines anderen.
func TestEinFremdesDeckLaesstSichNichtKopieren(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	vorlage := kopiervorlage(t, ablage)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO profile (id, nutzer_id, name) VALUES (2, 1, 'Zweites')`); fehler != nil {
		t.Fatalf("das zweite Profil: %v", fehler)
	}

	fremd := storage.MitProfil(context.WithoutCancel(probekontext(t)), 2)

	for name, kennung := range map[string]int64{"fremd": vorlage.ID, "fehlend": 999} {
		_, fehler := ablage.Kopieren(fremd, kennung, "Beute")
		if !errors.Is(fehler, domain.ErrNotFound) {
			t.Errorf("%s: Fehler = %v, erwartet ErrNotFound", name, fehler)
		}
	}

	if gezaehlt := zeilen(t, db, "decks"); gezaehlt != 1 {
		t.Errorf("%d Decks, erwartet nur das Original", gezaehlt)
	}
}

// **Scheitert ein Platz, bleibt keine halbe Kopie** — ein Anwendungsfall,
// ein Zug (#1820).
func TestEineGescheiterteKopieHinterlaesstNichts(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	vorlage := kopiervorlage(t, ablage)

	// **Nur die Kopie stolpert** — das Original ist das erste Deck.
	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE TRIGGER zweiter BEFORE INSERT ON deckplaetze
		 WHEN NEW.deck_id <> (SELECT min(id) FROM decks)
		  AND (SELECT COUNT(*) FROM deckplaetze WHERE deck_id = NEW.deck_id) >= 1
		 BEGIN SELECT RAISE(ABORT, 'der zweite Platz scheitert'); END`); fehler != nil {
		t.Fatalf("Auslöser: %v", fehler)
	}

	if _, fehler := ablage.Kopieren(probekontext(t), vorlage.ID, "Halb"); fehler == nil {
		t.Fatal("die Kopie scheiterte nicht")
	}

	if gezaehlt := zeilen(t, db, "decks"); gezaehlt != 1 {
		t.Errorf("%d Decks — eine halbe Kopie blieb stehen", gezaehlt)
	}
}

// **Ohne Profil gibt es keinen Ort** — Vermerk 0066.
func TestEineKopieOhneProfilLegtNichtsAn(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	vorlage := kopiervorlage(t, storage.NeueDeckablage(db))

	if _, fehler := storage.NeueDeckablage(db).Kopieren(
		context.Background(), vorlage.ID, "Ohne"); fehler == nil {
		t.Fatal("ohne Profil ging die Kopie durch")
	}
}
