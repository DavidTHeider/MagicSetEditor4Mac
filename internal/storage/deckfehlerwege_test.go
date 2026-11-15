package storage_test

// **Die Fehlerwege des Deckspeichers** — Werkstück 2 von Phase 4.6
// (#1293).
//
// # Warum eine eigene Datei
//
// `decks_test.go` stand an der Zeilengrenze aus §2. Der Schnitt folgt der
// Frage: *was der Speicher kann*, steht dort; *was er meldet, wenn es
// schiefgeht*, hier.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Jeder Fehlerweg wird betreten** — ein Bestand, der auf jeden Ruf mit einem
// Fehler antwortet.
//
// **Ohne diese Fälle blieben sie unbetreten**, und ein Fehlerweg, den niemand
// geht, ist keiner (§3). Die Vorrichtung steht in `autoersetzungen_test.go`.
func TestJederDeckwegMeldetEinenKaputtenBestand(t *testing.T) {
	ablage := storage.NeueDeckablage(kaputterBestand(t))

	if _, err := ablage.Anlegen(probekontext(t), domain.Deck{
		Name: "Deck", Spiel: "magic",
	}); err == nil {
		t.Error("anlegen schwieg")
	}

	if _, err := ablage.Deck(probekontext(t), 1); err == nil {
		t.Error("lesen schwieg")
	}

	if _, err := ablage.Decks(probekontext(t), "magic"); err == nil {
		t.Error("die Liste schwieg")
	}

	if _, err := ablage.PlatzSetzen(probekontext(t), 1, domain.Deckplatz{
		Printing: "a1b2c3d4-0000-0000-0000-000000000000",
		Menge:    1, Bereich: domain.BereichHauptdeck,
	}); err == nil {
		t.Error("legen schwieg")
	}

	if err := ablage.PlatzLoeschen(probekontext(t), 1, 1); err == nil {
		t.Error("den Platz löschen schwieg")
	}

	if err := ablage.Loeschen(probekontext(t), 1); err == nil {
		t.Error("löschen schwieg")
	}
}

// **Ein Platz, den es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlenderDeckplatzIstEinNein(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	if err := ablage.PlatzLoeschen(probekontext(t), deck.ID, 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Und ein Deck, das es nicht gibt, ebenso.**
func TestEinFehlendesDeckLaesstSichNichtLoeschen(t *testing.T) {
	db := frischeDatenbank(t)

	if err := storage.NeueDeckablage(db).Loeschen(probekontext(t), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Platz, den der Bestand abweist, kommt als Fehler zurück.**
//
// Die `CHECK`-Bedingungen der Wanderung sind die letzte Aufsicht: ein Platz
// mit beiden Bezügen oder mit keinem kommt hier nicht durch, auch wenn ihn
// jemand am Dienst vorbei legt.
func TestEinPlatzOhneBezugWirdVomBestandAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Menge: 1, Bereich: domain.BereichHauptdeck,
	}); err == nil {
		t.Error("ein Platz ohne Bezug kam durch")
	}
}

// **Eine eigene Karte lässt sich legen** — der andere Zweig von `PlatzSetzen`.
func TestEineEigeneKarteLaesstSichLegen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	karte := kartenkennungFuerDeck(t, db)

	gelegt, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		KarteID: karte, Menge: 3, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if gelegt.ID == 0 {
		t.Error("der Platz hat keine Kennung")
	}

	// **Und ein zweites Mal ist eine höhere Menge** — derselbe Weg wie beim
	// Druck, andere Spalte.
	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		KarteID: karte, Menge: 4, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("das zweite Mal: %v", err)
	}

	gefunden := deckLesen(t, ablage, deck.ID)

	if len(gefunden.Plaetze) != 1 || gefunden.Plaetze[0].Menge != 4 {
		t.Errorf("%+v", gefunden.Plaetze)
	}
}

// kartenkennungFuerDeck legt ein Set mit einer Karte an und gibt die Karte.
func kartenkennungFuerDeck(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	ergebnis, err := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, stil, geoeffnet, profil_id)
		 VALUES ('Probe', 'magic', 'm15', datetime('now'), 1)`)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	setID, err := ergebnis.LastInsertId()
	if err != nil {
		t.Fatalf("die Setkennung: %v", err)
	}

	ergebnis, err = db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (?, 0)`, setID)
	if err != nil {
		t.Fatalf("die Karte anlegen: %v", err)
	}

	karte, err := ergebnis.LastInsertId()
	if err != nil {
		t.Fatalf("die Kartenkennung: %v", err)
	}

	return karte
}

// **Eine Zeile, die sich nicht lesen lässt, ist ein Fehler** — und keine halbe
// Liste.
//
// Der Weg über `zeilen.Scan`: `menge` ist als `INTEGER` erklärt, und ein Text
// darin lässt sich nicht lesen. **Die Tabelle wird neu gelegt, ohne die
// `CHECK`-Bedingungen** — sonst wiese SQLite den Text schon beim Schreiben ab.
// Ein Bestand, den ein anderes Programm angefasst hat, sieht so aus; §8
// verlangt, dass er sich meldet.
func TestEineUnlesbareDeckzeileIstEinFehler(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	for _, satz := range []string{
		`DROP TABLE deckplaetze`,
		`CREATE TABLE deckplaetze (
		   id INTEGER PRIMARY KEY AUTOINCREMENT, deck_id INTEGER,
		   karte_id INTEGER, printing TEXT, druck_id INTEGER, menge TEXT, bereich TEXT, stelle INTEGER)`,
		`INSERT INTO deckplaetze (deck_id, printing, menge, bereich, stelle)
		 VALUES (1, 'a1b2c3d4-0000-0000-0000-000000000000', 'keine Zahl', 'hauptdeck', 0)`,
	} {
		if _, err := db.ExecContext(probekontext(t), satz); err != nil {
			t.Fatalf("%s: %v", satz, err)
		}
	}

	if _, err := ablage.Deck(probekontext(t), deck.ID); err == nil {
		t.Error("die unlesbare Zeile kam durch")
	}
}

// **Und dasselbe für die Deckliste.**
func TestEineUnlesbareDeckuebersichtIstEinFehler(t *testing.T) {
	db := frischeDatenbank(t)

	for _, satz := range []string{
		`DROP TABLE decks`,
		`CREATE TABLE decks (id INTEGER PRIMARY KEY, fassung TEXT,
		   name TEXT, spiel TEXT, notiz TEXT, profil_id INTEGER)`,
		`INSERT INTO decks (id, fassung, name, spiel, notiz, profil_id)
		 VALUES (1, 'keine Zahl', 'Deck', 'magic', '', 1)`,
	} {
		if _, err := db.ExecContext(probekontext(t), satz); err != nil {
			t.Fatalf("%s: %v", satz, err)
		}
	}

	ablage := storage.NeueDeckablage(db)

	if _, err := ablage.Decks(probekontext(t), "magic"); err == nil {
		t.Error("die Liste kam durch")
	}

	if _, err := ablage.Deck(probekontext(t), 1); err == nil {
		t.Error("das einzelne Deck kam durch")
	}
}

// **Ein Deck ohne seine Tabelle meldet sich.**
//
// Der Weg über die zweite Abfrage in `Deck`: die Kopfzeile steht, die Plätze
// lassen sich nicht lesen. Ohne diesen Fall bliebe der Zweig unbetreten — und
// er ist der, an dem ein Deck **ohne** seine Plätze zurückkäme, also eine
// Überschrift ohne Inhalt.
func TestEinDeckOhnePlatztabelleMeldetSich(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Deck", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE deckplaetze`); err != nil {
		t.Fatalf("die Tabelle entfernen: %v", err)
	}

	if _, err := ablage.Deck(probekontext(t), deck.ID); err == nil {
		t.Error("das Deck kam ohne seine Plätze durch")
	}
}

// **Die Wege aus Werkstück 3 melden einen kaputten Bestand ebenso.**
func TestDieNeuenDeckwegeMeldenEinenKaputtenBestand(t *testing.T) {
	db := kaputterBestand(t)

	if _, err := storage.NeueDeckablage(db).Aendern(probekontext(t), domain.Deck{
		ID: 1, Fassung: 1, Name: "Deck", Spiel: "magic",
	}); err == nil {
		t.Error("ändern schwieg")
	}

	if _, err := storage.NeuerSetLeser(db).SpielZurKarte(probekontext(t), 1); err == nil {
		t.Error("das Spiel zur Karte schwieg")
	}
}
