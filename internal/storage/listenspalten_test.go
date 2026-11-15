// **Was die Kartenliste zeigt und wonach sie ordnet** — Werkstück 3 von
// Phase 3.7
// (#1168).
package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Wahl steht nichts da — und das ist keine leere Wahl.**
//
// Der Unterschied trägt: „niemand hat gewählt" heißt, dass der Vorschlag des
// Spielpakets gilt; „alles abgewählt" hieße eine Liste ohne Spalten.
func TestOhneWahlStehtKeineSpalteImBestand(t *testing.T) {
	db := frischeDatenbank(t)

	spalten, err := storage.NeueListenablage(db).Spalten(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(spalten) != 0 {
		t.Errorf("%+v", spalten)
	}
}

// **Eine Wahl wird geschrieben und kommt zurück — je Spiel.**
func TestEineSpaltenwahlUeberlebtDasSet(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueListenablage(db)

	wahl := []domain.Spaltenwahl{
		{Feld: "name", Sichtbar: true, Stelle: 100},
		{Feld: "casting cost", Sichtbar: true, Stelle: 200},
		{Feld: "rule text", Sichtbar: false, Stelle: 300},
	}

	if err := ablage.SpaltenSetzen(probekontext(t), "magic", wahl); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	gelesen, err := ablage.Spalten(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 3 || gelesen[0].Feld != "name" ||
		gelesen[1].Feld != "casting cost" || !gelesen[1].Sichtbar ||
		gelesen[2].Sichtbar {
		t.Fatalf("%+v", gelesen)
	}

	// **Nach der Stelle geordnet** — der Bestand gibt sie in der Reihenfolge
	// heraus, in der sie stehen sollen; ein Aufrufer, der selbst sortierte,
	// wäre eine zweite Antwort auf dieselbe Frage.
	if gelesen[0].Stelle != 100 || gelesen[2].Stelle != 300 {
		t.Errorf("Stellen %d/%d/%d",
			gelesen[0].Stelle, gelesen[1].Stelle, gelesen[2].Stelle)
	}

	// **Ein anderes Spiel hat seine eigene Wahl.**
	fremd, err := ablage.Spalten(probekontext(t), "vs")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(fremd) != 0 {
		t.Errorf("das fremde Spiel trägt %+v", fremd)
	}
}

// **Eine zweite Wahl ersetzt die erste vollständig.**
//
// **Vollständig und nicht zusammengeführt**: wer eine Spalte abwählt, nimmt
// sie weg; ein `INSERT … ON CONFLICT` allein ließe sie stehen, und niemand
// könnte je eine Spalte loswerden.
func TestEineZweiteSpaltenwahlErsetztDieErste(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueListenablage(db)

	if err := ablage.SpaltenSetzen(probekontext(t), "magic", []domain.Spaltenwahl{
		{Feld: "name", Sichtbar: true, Stelle: 100},
		{Feld: "rule text", Sichtbar: true, Stelle: 200},
	}); err != nil {
		t.Fatalf("erste: %v", err)
	}

	if err := ablage.SpaltenSetzen(probekontext(t), "magic", []domain.Spaltenwahl{
		{Feld: "name", Sichtbar: true, Stelle: 100},
	}); err != nil {
		t.Fatalf("zweite: %v", err)
	}

	gelesen, err := ablage.Spalten(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 1 || gelesen[0].Feld != "name" {
		t.Errorf("%+v", gelesen)
	}
}

// **Eine leere Wahl räumt die alte weg.**
func TestEineLeereSpaltenwahlRaeumtAuf(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueListenablage(db)

	if err := ablage.SpaltenSetzen(probekontext(t), "magic", []domain.Spaltenwahl{
		{Feld: "name", Sichtbar: true, Stelle: 100},
	}); err != nil {
		t.Fatalf("erste: %v", err)
	}

	if err := ablage.SpaltenSetzen(probekontext(t), "magic", nil); err != nil {
		t.Fatalf("räumen: %v", err)
	}

	gelesen, err := ablage.Spalten(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen) != 0 {
		t.Errorf("%+v", gelesen)
	}
}

// **Die Sortierung liegt daneben und wird ebenso je Spiel gemerkt.**
func TestDieSortierungUeberlebtDasSet(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueListenablage(db)

	ordnung, err := ablage.Sortierung(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Ohne Wahl: kein Feld, aufsteigend.**
	//
	//	: sort_cards_ascending(true)
	//
	// `src/data/settings.cpp`, Zeile 72, Fassung f9d9356 vom 2021-07-06.
	if ordnung.Feld != "" || !ordnung.Aufsteigend {
		t.Fatalf("%+v", ordnung)
	}

	if fehler := ablage.SortierungSetzen(probekontext(t), "magic",
		domain.Kartenordnung{Feld: "name", Aufsteigend: false}); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	ordnung, err = ablage.Sortierung(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("nachlesen: %v", err)
	}

	if ordnung.Feld != "name" || ordnung.Aufsteigend {
		t.Errorf("%+v", ordnung)
	}

	// **Und eine zweite Wahl ersetzt sie.**
	if fehler := ablage.SortierungSetzen(probekontext(t), "magic",
		domain.Kartenordnung{Feld: "", Aufsteigend: true}); fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	ordnung, err = ablage.Sortierung(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("nachlesen: %v", err)
	}

	if ordnung.Feld != "" || !ordnung.Aufsteigend {
		t.Errorf("%+v", ordnung)
	}
}

// **Ohne Spielnamen wird abgewiesen** — eine Wahl ohne Spiel gehört niemandem.
func TestOhneSpielnamenWirdDieWahlAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueListenablage(db)

	if err := ablage.SpaltenSetzen(probekontext(t), "  ", nil); err == nil {
		t.Error("die Spaltenwahl kam ohne Spiel durch")
	}

	if err := ablage.SortierungSetzen(probekontext(t), "",
		domain.Kartenordnung{}); err == nil {
		t.Error("die Sortierung kam ohne Spiel durch")
	}

	if _, err := ablage.Spalten(probekontext(t), ""); err == nil {
		t.Error("das Lesen kam ohne Spiel durch")
	}

	if _, err := ablage.Sortierung(probekontext(t), ""); err == nil {
		t.Error("das Lesen der Sortierung kam ohne Spiel durch")
	}
}
