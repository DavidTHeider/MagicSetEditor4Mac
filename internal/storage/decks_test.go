package storage_test

// **Der Deckspeicher** — Werkstück 2 von Phase 4.6
// (#1293),
// nach Vermerk 0061.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// TestEinDeckLaesstSichAnlegenUndWiederfinden ist die Grundzusage.
func TestEinDeckLaesstSichAnlegenUndWiederfinden(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Deck{
		Name: "Mein Deck", Spiel: "magic", Notiz: "eine Notiz",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.ID == 0 {
		t.Error("das angelegte Deck hat keine Kennung")
	}

	if angelegt.Fassung != 1 {
		t.Errorf("Fassung %d, erwartet 1", angelegt.Fassung)
	}

	gefunden, err := ablage.Deck(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gefunden.Name != "Mein Deck" || gefunden.Spiel != "magic" {
		t.Errorf("%+v", gefunden)
	}

	if gefunden.Notiz != "eine Notiz" {
		t.Errorf("Notiz %q", gefunden.Notiz)
	}
}

// **Ein Deck, das es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlendesDeckIstEinNein(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := storage.NeueDeckablage(db).Deck(probekontext(t), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Zwei Decks desselben Spiels mit demselben Namen sind ein Konflikt.**
//
// Sie wären zwei Dinge, die niemand auseinanderhält — dieselbe Überlegung wie
// bei den Sets.
func TestZweiDecksMitDemselbenNamenSindEinKonflikt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	if _, err := ablage.Anlegen(probekontext(t), domain.Deck{
		Name: "Aggro", Spiel: "magic",
	}); err != nil {
		t.Fatalf("das erste: %v", err)
	}

	_, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Aggro", Spiel: "magic"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", err)
	}
}

// **Über Spiele hinweg ist derselbe Name in Ordnung.**
//
// „Aggro" für Magic und „Aggro" für ein anderes Spiel sind zwei Decks.
func TestDerselbeNameInZweiSpielenGehet(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	for _, spiel := range []string{"magic", "space"} {
		if _, err := ablage.Anlegen(probekontext(t), domain.Deck{
			Name: "Aggro", Spiel: spiel,
		}); err != nil {
			t.Fatalf("%s: %v", spiel, err)
		}
	}
}

// **Ein Platz mit einem Druck lässt sich legen und wiederfinden.**
func TestEinDeckplatzLaesstSichLegen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: "a1b2c3d4-0000-0000-0000-000000000000",
		Menge:    4, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	gefunden := deckLesen(t, ablage, deck.ID)

	if len(gefunden.Plaetze) != 1 {
		t.Fatalf("%d Plätze", len(gefunden.Plaetze))
	}

	if gefunden.Plaetze[0].Menge != 4 {
		t.Errorf("Menge %d", gefunden.Plaetze[0].Menge)
	}
}

// **Derselbe Bezug im selben Bereich zweimal ist eine höhere Menge, kein
// zweiter Platz.**
func TestDerselbeBezugZweimalIstEineHoehereMenge(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	platz := domain.Deckplatz{
		Printing: "a1b2c3d4-0000-0000-0000-000000000000",
		Menge:    2, Bereich: domain.BereichHauptdeck,
	}

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, platz); err != nil {
		t.Fatalf("der erste: %v", err)
	}

	platz.Menge = 4

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, platz); err != nil {
		t.Fatalf("der zweite: %v", err)
	}

	gefunden := deckLesen(t, ablage, deck.ID)

	if len(gefunden.Plaetze) != 1 {
		t.Fatalf("%d Plätze, erwartet 1", len(gefunden.Plaetze))
	}

	if gefunden.Plaetze[0].Menge != 4 {
		t.Errorf("Menge %d, erwartet 4", gefunden.Plaetze[0].Menge)
	}
}

// **Und derselbe Bezug in beiden Bereichen sind zwei Plätze** — bei Magic der
// Normalfall.
func TestDerselbeBezugInBeidenBereichenSindZwei(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	for _, bereich := range []domain.Deckbereich{
		domain.BereichHauptdeck, domain.BereichSideboard,
	} {
		if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
			Printing: "a1b2c3d4-0000-0000-0000-000000000000",
			Menge:    2, Bereich: bereich,
		}); err != nil {
			t.Fatalf("%s: %v", bereich, err)
		}
	}

	gefunden := deckLesen(t, ablage, deck.ID)

	if len(gefunden.Plaetze) != 2 {
		t.Errorf("%d Plätze, erwartet 2", len(gefunden.Plaetze))
	}
}

// **Ein Platz lässt sich herausnehmen.**
func TestEinDeckplatzLaesstSichHerausnehmen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	gelegt, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: "a1b2c3d4-0000-0000-0000-000000000000",
		Menge:    1, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if err := ablage.PlatzLoeschen(probekontext(t), deck.ID, gelegt.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	gefunden := deckLesen(t, ablage, deck.ID)

	if len(gefunden.Plaetze) != 0 {
		t.Errorf("%d Plätze", len(gefunden.Plaetze))
	}
}

// **Ein Deck lässt sich löschen — und nimmt seine Plätze mit.**
func TestEinGeloeschtesDeckNimmtSeinePlaetzeMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	deck := deckAnlegen(t, ablage)

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: "a1b2c3d4-0000-0000-0000-000000000000",
		Menge:    1, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	if err := ablage.Loeschen(probekontext(t), deck.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	var gezaehlt int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM deckplaetze`).Scan(&gezaehlt); err != nil {
		t.Fatalf("nachzählen: %v", err)
	}

	if gezaehlt != 0 {
		t.Errorf("%d Plätze übrig", gezaehlt)
	}
}

// **Die Liste gibt die Decks eines Spiels.**
func TestDieListeGibtDieDecksEinesSpiels(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)

	for _, fall := range [][2]string{{"A", "magic"}, {"B", "magic"}, {"C", "space"}} {
		if _, err := ablage.Anlegen(probekontext(t), domain.Deck{
			Name: fall[0], Spiel: fall[1],
		}); err != nil {
			t.Fatalf("%s: %v", fall[0], err)
		}
	}

	decks, err := ablage.Decks(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(decks) != 2 {
		t.Errorf("%d Decks, erwartet 2", len(decks))
	}

	// **Ohne Spiel alle** — die Übersicht über den ganzen Bestand.
	alle, err := ablage.Decks(probekontext(t), "")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(alle) != 3 {
		t.Errorf("%d Decks, erwartet 3", len(alle))
	}
}

// deckAnlegen legt ein Deck an — und bricht ab, wenn das schon scheitert.
//
// **Zwei Helfer statt `_`**: ein verworfener Fehler in einer Vorbereitung
// lässt den eigentlichen Fall an einer Stelle scheitern, an der niemand ihn
// sucht.
func deckAnlegen(t *testing.T, ablage *storage.Deckablage) domain.Deck {
	t.Helper()

	deck, err := ablage.Anlegen(probekontext(t), domain.Deck{Name: "Deck", Spiel: "magic"})
	if err != nil {
		t.Fatalf("das Deck anlegen: %v", err)
	}

	return deck
}

// deckLesen liest ein Deck — und bricht ab, wenn das scheitert.
func deckLesen(t *testing.T, ablage *storage.Deckablage, kennung int64) domain.Deck {
	t.Helper()

	deck, err := ablage.Deck(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("das Deck lesen: %v", err)
	}

	return deck
}
