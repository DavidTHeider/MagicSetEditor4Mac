package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func angelegteKarte(t *testing.T) (*storage.SetLeser, *storage.Kartenschreiber, int64) {
	t.Helper()

	db := frischeDatenbank(t)

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(), nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	karten, err := leser.Karten(probekontext(t), 1)
	if err != nil || len(karten) == 0 {
		t.Fatalf("%v / %v", karten, err)
	}

	return leser, storage.NeuerKartenschreiber(db), karten[0].ID
}

func TestEinNeuerFeldwertKommtHinzu(t *testing.T) {
	leser, schreiber, karteID := angelegteKarte(t)

	if err := schreiber.FeldSetzen(probekontext(t), karteID, "typ", "Zauber"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	karte, err := leser.KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// Der neue Wert kommt hinten an — die Reihenfolge zählt für den Rundlauf.
	if len(karte.Values) != 2 || karte.Values[1].Field != "typ" {
		t.Fatalf("%+v", karte.Values)
	}

	if karte.Values[1].Text != "Zauber" {
		t.Errorf("%q", karte.Values[1].Text)
	}
}

// **Ein zweites Schreiben überschreibt, statt am Schlüssel zu scheitern.**
//
// Zwischen einem Lesen und einem Schreiben liegt Zeit, und in ihr kann ein
// anderer Vorgang dasselbe Feld anlegen. `ON CONFLICT` macht daraus einen
// Schritt.
func TestEinZweitesSchreibenUeberschreibt(t *testing.T) {
	leser, schreiber, karteID := angelegteKarte(t)

	for _, wert := range []string{"erst", "dann"} {
		if err := schreiber.FeldSetzen(probekontext(t), karteID, "name", wert); err != nil {
			t.Fatalf("schreiben: %v", err)
		}
	}

	karte, err := leser.KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karte.Values) != 1 || karte.Values[0].Text != "dann" {
		t.Errorf("%+v", karte.Values)
	}
}

// Eine Karte, die es nicht gibt, ist ein `ErrNotFound` — kein Konflikt.
//
// Der Unterschied entscheidet, ob das Frontend nachlädt oder eine Meldung
// zeigt (§8).
func TestEineUnbekannteKarteIstNichtGefunden(t *testing.T) {
	_, schreiber, _ := angelegteKarte(t)

	err := schreiber.FeldSetzen(probekontext(t), 999, "name", "egal")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
	}
}

func TestKarteLesenMeldetEineUnbekannteKarte(t *testing.T) {
	leser, _, _ := angelegteKarte(t)

	if _, err := leser.KarteLesen(probekontext(t), 999); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
	}
}

// Eine Karte ohne Feldwerte lässt sich lesen — sie ist eine Karte.
func TestEineKarteOhneWerteLaesstSichLesen(t *testing.T) {
	db := frischeDatenbank(t)

	set := probeSet()
	set.Cards = []*domain.Card{{}}

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), 1)
	if err != nil || len(karten) != 1 {
		t.Fatalf("%v / %v", karten, err)
	}

	karte, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), karten[0].ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karte.Values) != 0 {
		t.Errorf("%+v", karte.Values)
	}
}

func TestDerSchreiberMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if err := storage.NeuerKartenschreiber(db).
		FeldSetzen(probekontext(t), 1, "name", "egal"); err == nil {
		t.Error("eine geschlossene Datenbank wurde angenommen")
	}

	if _, err := storage.NeuerSetLeser(db).KarteLesen(probekontext(t), 1); err == nil {
		t.Error("KarteLesen kam ohne Fehler zurück")
	}
}

// **Eine gelesene Karte sagt, zu welchem Set sie gehört.**
//
// # Warum das eine Spalte wert ist
//
// Ohne sie findet niemand den **Stil** der Karte: `Vorschau` und `Bildausfuhr`
// bekommen eine Kartenkennung und sonst nichts, und der Stil steht am Set.
// Genau daran hing #206
// — die Anwendung zeigte den Behelfsplan, weil der Weg zum Stilpaket an dieser
// Stelle abriss.
func TestEineGeleseneKarteNenntIhrSet(t *testing.T) {
	leser, _, karteID := angelegteKarte(t)

	karte, err := leser.KarteLesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if karte.SetID == 0 {
		t.Error("die gelesene Karte nennt kein Set — der Weg zum Stil reißt hier ab")
	}
}
