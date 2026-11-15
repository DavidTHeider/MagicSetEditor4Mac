package storage_test

// **Einen Platz in den anderen Bereich verschieben — in einem Zug** — Befund 7
// des Zweitblicks über Phase 5.1
// (#1405).
//
// # Warum es das gibt
//
// Das Fenster hat den Wechsel bis zum 2026-09-10 in **zwei** Rufen gefahren:
// erst den neuen Platz legen, dann den alten nehmen. Scheitert der zweite,
// liegt dieselbe Karte in **beiden** Bereichen — ein halber Stand, den
// niemand zurücknimmt (§7, Regel 3). Über die Wails-Naht lässt sich das nicht
// heilen; ein Vorgang muss einer sein.

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// deckMitPlatz legt ein Deck mit einem Platz an und gibt beide zurück.
func deckMitPlatz(t *testing.T, ablage *storage.Deckablage, menge int) (
	domain.Deck, domain.Deckplatz,
) {
	t.Helper()

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Probe", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	platz, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: "aaaaaaaa-0000-4000-8000-00000000000a",
		Menge:    menge,
		Bereich:  domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	return deck, platz
}

// **Der Wechsel ist ein Vorgang, und danach gibt es den alten Platz nicht mehr.**
func TestEinPlatzWechseltDenBereichInEinemZug(t *testing.T) {
	ablage := storage.NeueDeckablage(frischeDatenbank(t))
	deck, platz := deckMitPlatz(t, ablage, 4)

	verschoben, err := ablage.PlatzVerschieben(probekontext(t), deck.ID, platz.ID, "sideboard")
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.Bereich != "sideboard" || verschoben.Menge != 4 {
		t.Errorf("%+v", verschoben)
	}

	gelesen, err := ablage.Deck(probekontext(t), deck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Plaetze) != 1 {
		t.Fatalf("%d Plätze, erwartet 1 — der alte muss weg sein", len(gelesen.Plaetze))
	}

	if gelesen.Plaetze[0].Bereich != "sideboard" || gelesen.Plaetze[0].Menge != 4 {
		t.Errorf("%+v", gelesen.Plaetze[0])
	}
}

// **Liegt dort schon derselbe Bezug, werden die Mengen zusammengelegt** — der
// Bestand hält je Bereich **einen** Platz je Bezug, und zwei Zeilen desselben
// Drucks im Sideboard gibt es nicht.
func TestEinWechselAufDenselbenBezugLegtDieMengenZusammen(t *testing.T) {
	ablage := storage.NeueDeckablage(frischeDatenbank(t))
	deck, platz := deckMitPlatz(t, ablage, 4)

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: "aaaaaaaa-0000-4000-8000-00000000000a",
		Menge:    2,
		Bereich:  "sideboard",
	}); err != nil {
		t.Fatalf("zweiten setzen: %v", err)
	}

	verschoben, err := ablage.PlatzVerschieben(probekontext(t), deck.ID, platz.ID, "sideboard")
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.Menge != 6 {
		t.Errorf("Menge %d, erwartet 6", verschoben.Menge)
	}

	gelesen, err := ablage.Deck(probekontext(t), deck.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Plaetze) != 1 || gelesen.Plaetze[0].Menge != 6 {
		t.Errorf("%d Plätze: %+v", len(gelesen.Plaetze), gelesen.Plaetze)
	}
}

// **Ein Wechsel in denselben Bereich ändert nichts** — und er ist kein Fehler.
func TestEinWechselInDenselbenBereichLaesstAllesStehen(t *testing.T) {
	ablage := storage.NeueDeckablage(frischeDatenbank(t))
	deck, platz := deckMitPlatz(t, ablage, 3)

	verschoben, err := ablage.PlatzVerschieben(probekontext(t), deck.ID, platz.ID, "hauptdeck")
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.ID != platz.ID || verschoben.Menge != 3 {
		t.Errorf("%+v", verschoben)
	}
}

// **Einen Platz, den es nicht gibt, verschiebt niemand.**
func TestEinUnbekannterPlatzLaesstSichNichtVerschieben(t *testing.T) {
	ablage := storage.NeueDeckablage(frischeDatenbank(t))
	deck, _ := deckMitPlatz(t, ablage, 1)

	_, err := ablage.PlatzVerschieben(probekontext(t), deck.ID, 9999, "sideboard")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein Bereich, den es nicht gibt, wird abgewiesen** — die Domäne kennt zwei.
func TestEinUnbekannterBereichWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueDeckablage(frischeDatenbank(t))
	deck, platz := deckMitPlatz(t, ablage, 1)

	_, err := ablage.PlatzVerschieben(probekontext(t), deck.ID, platz.ID, "kiste")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", err)
	}
}

// **Ein kaputter Bestand meldet sich auch beim Verschieben** — §3: ein
// Fehlerweg, den niemand geht, ist keiner.
func TestDerBereichswechselMeldetEinenKaputtenBestand(t *testing.T) {
	ablage := storage.NeueDeckablage(kaputterBestand(t))

	if _, err := ablage.PlatzVerschieben(
		probekontext(t), 1, 1, domain.BereichSideboard); err == nil {
		t.Error("verschieben schwieg")
	}
}

// **Eine eigene Karte wechselt den Bereich genauso** — und die Menge geht
// dabei mit dem zusammen, was drüben liegt.
//
// **Der Fall gehört hierher und nicht zum Druck**: der Bestand sucht den Bezug
// über **drei** Spalten (`karte_id`, `printing`, `druck_id`), und wer nur den
// Druck prüft, lässt zwei davon unbetreten.
func TestEineEigeneKarteWechseltDenBereich(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Eigen", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karteID := kartenkennungFuerDeck(t, db)

	platz, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		KarteID: karteID, Menge: 2, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if _, setzfehler := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		KarteID: karteID, Menge: 1, Bereich: domain.BereichSideboard,
	}); setzfehler != nil {
		t.Fatalf("drüben setzen: %v", setzfehler)
	}

	verschoben, err := ablage.PlatzVerschieben(
		probekontext(t), deck.ID, platz.ID, domain.BereichSideboard)
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.Menge != 3 || verschoben.KarteID != karteID {
		t.Errorf("%+v", verschoben)
	}
}

// **Ein Platz auf einen Druck ohne Alias wechselt ebenso** — die dritte
// Bezugsspalte (`druck_id`, seit A2). Ohne diesen Fall bliebe sie unbetreten.
func TestEinVorlaeufigerDruckWechseltDenBereich(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Vorläufig", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	druckID := druckOhneAlias(t, db)

	platz, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		DruckID: druckID, Menge: 2, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("setzen: %v", err)
	}

	verschoben, err := ablage.PlatzVerschieben(
		probekontext(t), deck.ID, platz.ID, domain.BereichSideboard)
	if err != nil {
		t.Fatalf("verschieben: %v", err)
	}

	if verschoben.Bereich != domain.BereichSideboard || verschoben.Menge != 2 {
		t.Errorf("%+v", verschoben)
	}
}

// **Und wenn die Tabelle mitten im Zug verschwindet, meldet jeder Schritt.**
//
// **Vier Wege in einem Lauf**: die Menge drüben lesen, den alten löschen, den
// neuen setzen und den Platz überhaupt lesen. Ein `DROP TABLE` **nach** dem
// Öffnen erreicht sie alle — die Transaktion beginnt, und erst die Anweisung
// darin scheitert.
func TestEinBereichswechselOhneTabelleMeldetJedenSchritt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck, _ := deckMitPlatz(t, ablage, 4)

	if _, err := db.ExecContext(t.Context(), `DROP TABLE deckplaetze`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := ablage.PlatzVerschieben(
		probekontext(t), deck.ID, 1, domain.BereichSideboard); err == nil {
		t.Error("verschieben schwieg")
	}
}

// druckOhneAlias legt eine Druckzeile an, die keine Scryfall-Kennung trägt —
// den Fall »vorläufig« aus A2 (#1441).
func druckOhneAlias(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	ergebnis, err := db.ExecContext(t.Context(),
		`INSERT INTO drucke (name, setcode, nummer, sprache) VALUES (?, ?, ?, ?)`,
		"Eigenbau", "xxx", "1", "de")
	if err != nil {
		t.Fatalf("Druck anlegen: %v", err)
	}

	kennung, idfehler := ergebnis.LastInsertId()
	if idfehler != nil {
		t.Fatalf("Kennung lesen: %v", idfehler)
	}

	return kennung
}
