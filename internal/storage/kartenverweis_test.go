package storage_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der Verweis auf einen echten Druck** — Werkstück 2 von Phase 3.6
// (#1167).

// verweisstand baut Bestand, Set und zwei Karten.
func verweisstand(t *testing.T) (
	db *sql.DB, ablage *storage.Kartenverweise, setID, ersteID, zweiteID int64,
) {
	t.Helper()

	db = frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
			{Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten lesen: %v", err)
	}

	return db, storage.NeueKartenverweise(db), setID, karten[0].ID, karten[1].ID
}

func probeverweis() domain.Kartenverweis {
	return domain.Kartenverweis{
		Printing: "516e137a-d9d9-4ba7-89fb-e680ddb2c2d1",
		Name:     "Lightning Bolt", Setcode: "lea", Nummer: "161",
		Adresse: "https://scryfall.com/card/lea/161",
		Gesetzt: time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC),
	}
}

// **Ein Verweis überlebt den Neustart, und der Zeitpunkt mit ihm.**
//
// Der Zeitpunkt ist nicht Schmuck: er sagt, wie alt der angezeigte Name ist
// (§16, Regel 5).
func TestEinKartenverweisUeberlebtDenNeustart(t *testing.T) {
	_, ablage, _, karteID, _ := verweisstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gelesen, err := ablage.Fuer(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen == nil {
		t.Fatal("es kam kein Verweis zurück")
	}

	if gelesen.Printing != probeverweis().Printing || gelesen.Name != "Lightning Bolt" ||
		gelesen.Setcode != "lea" || gelesen.Nummer != "161" ||
		gelesen.Adresse != probeverweis().Adresse {
		t.Errorf("gelesen wurde %+v", *gelesen)
	}

	if !gelesen.Gesetzt.Equal(probeverweis().Gesetzt) {
		t.Errorf("gesetzt am %v, erwartet %v", gelesen.Gesetzt, probeverweis().Gesetzt)
	}
}

// **Keiner ist kein Fehler.** Die meisten Karten haben keinen Verweis; wer
// danach fragt, bekommt „keinen" und keine Ausnahme.
func TestOhneVerweisKommtKeinerZurueck(t *testing.T) {
	_, ablage, _, karteID, _ := verweisstand(t)

	gelesen, err := ablage.Fuer(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen != nil {
		t.Errorf("es kam %+v zurück", *gelesen)
	}
}

// **Ein zweiter Verweis ersetzt den ersten.** Eine Karte zeigt auf **einen**
// Druck; zwei nebeneinander wären zwei Antworten auf dieselbe Frage.
func TestEinZweiterVerweisErsetztDenErsten(t *testing.T) {
	_, ablage, _, karteID, _ := verweisstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	zweiter := probeverweis()
	zweiter.Printing = "e3285e6b-3e79-4d7c-bf96-d920f973b0d5"
	zweiter.Setcode = "2ed"

	if err := ablage.Setzen(probekontext(t), karteID, zweiter); err != nil {
		t.Fatalf("zweitens setzen: %v", err)
	}

	gelesen, err := ablage.Fuer(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Printing != zweiter.Printing || gelesen.Setcode != "2ed" {
		t.Errorf("gelesen wurde %+v", *gelesen)
	}
}

// **Lösen sagt, ob es einen gab.** Ohne diese Auskunft könnte eine Fläche
// nicht zwischen „gelöst" und „war nie da" unterscheiden — und meldete Erfolg,
// wo nichts geschah.
func TestVerweisLoesenSagtObEinerDaWar(t *testing.T) {
	_, ablage, _, karteID, _ := verweisstand(t)

	if geloest, err := ablage.Loesen(probekontext(t), karteID); err != nil || geloest {
		t.Fatalf("lösen ohne Verweis: %v, %v", geloest, err)
	}

	if err := ablage.Setzen(probekontext(t), karteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	geloest, err := ablage.Loesen(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lösen: %v", err)
	}

	if !geloest {
		t.Error("der vorhandene Verweis wurde nicht als gelöst gemeldet")
	}
}

// **Eine Karte, die es nicht gibt, bekommt keinen Verweis** — und der Fehler
// heißt „nicht gefunden" und nicht „SQL logic error" (§8).
func TestEinVerweisAufEineFehlendeKarteWirdAbgewiesen(t *testing.T) {
	_, ablage, _, _, _ := verweisstand(t)

	err := ablage.Setzen(probekontext(t), 987654, probeverweis())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler war %v — will ErrNotFound", err)
	}
}

// **Die Verweise eines Sets kommen in einem Zug.**
//
// Die Kartenliste zeigt sie an jeder Zeile; eine Abfrage je Karte wäre bei
// einem Set mit 300 Karten 300 Abfragen für eine Ansicht.
func TestDieVerweiseEinesSetsKommenZusammen(t *testing.T) {
	_, ablage, setID, ersteID, zweiteID := verweisstand(t)

	if err := ablage.Setzen(probekontext(t), zweiteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	verweise, err := ablage.JeSet(probekontext(t), setID)
	if err != nil {
		t.Fatalf("je Set: %v", err)
	}

	if len(verweise) != 1 {
		t.Fatalf("es kamen %d Verweise", len(verweise))
	}

	if _, da := verweise[ersteID]; da {
		t.Error("die Karte ohne Verweis steht in der Antwort")
	}

	if verweise[zweiteID].Name != "Lightning Bolt" {
		t.Errorf("die zweite Karte trägt %+v", verweise[zweiteID])
	}
}

// **Eine gelöschte Karte nimmt ihren Verweis mit.** Ohne das bliebe eine Zeile
// zurück, die auf nichts zeigt — und die nächste Karte mit derselben Kennung
// erbte sie.
func TestEineGeloeschteKarteNimmtIhrenVerweisMit(t *testing.T) {
	db, ablage, setID, karteID, _ := verweisstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if err := storage.NeuerKartenschreiber(db).
		KarteLoeschen(probekontext(t), karteID); err != nil {
		t.Fatalf("Karte löschen: %v", err)
	}

	verweise, err := ablage.JeSet(probekontext(t), setID)
	if err != nil {
		t.Fatalf("je Set: %v", err)
	}

	if len(verweise) != 0 {
		t.Errorf("es blieben %d Verweise zurück", len(verweise))
	}
}

// **Der Zeitpunkt kommt in UTC zurück und nicht in der Ortszeit.**
//
// `gesetzt` steht als Unix-Sekunden in der Spalte; wer sie ohne Zone
// zurücklesen ließe, bekäme die des Rechners — und dasselbe Datum sähe in
// Berlin und in Los Angeles verschieden aus (§3: die rohe Form gehört in den
// Test).
func TestDerZeitpunktEinesVerweisesKommtInUTCZurueck(t *testing.T) {
	_, ablage, _, karteID, _ := verweisstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, probeverweis()); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gelesen, err := ablage.Fuer(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if zone, _ := gelesen.Gesetzt.Zone(); zone != "UTC" {
		t.Errorf("der Zeitpunkt kam in %q zurück", zone)
	}
}
