package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Karte lässt sich anlegen** —
// #562.
//
// # Die Lücke
//
// Der Bestand konnte Sets **einlesen** und Feldwerte ändern. Eine neue Karte
// gab es nirgends: nicht im Speicher, nicht im Dienst, nicht in der
// Oberfläche. Wer ein Set anlegen wollte, musste es woanders schreiben und hier
// öffnen.
//
// # Die Stelle ist die letzte
//
// MSE hängt eine neue Karte hinten an (`Set::addCard`, `src/data/set.cpp`,
// Fassung f9d9356 vom 2021-07-06). Die Reihenfolge ist eine Spalte mit
// `UNIQUE (set_id, reihenfolge)`; die nächste Zahl kommt aus der Datenbank,
// nicht aus einer Zählung im Programm — zwei Fenster auf demselben Bestand
// bekämen sonst dieselbe.
func TestEineKarteLaesstSichAnlegen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	schreiber := storage.NeuerKartenschreiber(db)
	leser := storage.NeuerSetLeser(db)

	setID, err := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Erste"},
		}}},
	}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	karteID, err := schreiber.KarteAnlegen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karte anlegen: %v", err)
	}

	if karteID == 0 {
		t.Fatal("keine Kennung")
	}

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet zwei", len(karten))
	}

	// **Hinten angehängt**, mit der nächsten Nummer.
	neue := karten[1]
	if neue.ID != karteID {
		t.Errorf("die neue Karte steht nicht am Ende: %+v", karten)
	}

	if neue.Number != 1 {
		t.Errorf("Reihenfolge %d, erwartet 1", neue.Number)
	}

	// **Ohne Feldwerte.** Was eine Karte trägt, entscheidet das Spielpaket
	// über seine Vorgabewerte — der Speicher erfindet keine.
	if len(neue.Values) != 0 {
		t.Errorf("%d Feldwerte, erwartet keine", len(neue.Values))
	}
}

// **Und in einem leeren Set beginnt sie bei null.**
func TestDieErsteKarteEinesSetsBeginntBeiNull(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	schreiber := storage.NeuerKartenschreiber(db)
	leser := storage.NeuerSetLeser(db)

	setID, err := speicher.SetAnlegen(probekontext(t),
		&domain.Set{Name: "Leer", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	if _, fehler := schreiber.KarteAnlegen(probekontext(t), setID); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	karten, err := leser.Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 || karten[0].Number != 0 {
		t.Errorf("%+v", karten)
	}
}

// **Ein Set, das es nicht gibt, ist ein Befund und kein stiller Fehlschlag.**
func TestEineKarteInEinemFremdenSetWirdAbgewiesen(t *testing.T) {
	schreiber := storage.NeuerKartenschreiber(frischeDatenbank(t))

	if _, err := schreiber.KarteAnlegen(probekontext(t), 4711); err == nil {
		t.Error("kein Fehler")
	}
}
