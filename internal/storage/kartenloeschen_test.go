// **Mehrere Karten in einem Zug löschen** — Befund 4 des Zweitblicks über
// Phase 3.7
// (#1168).
//
// # Warum es diesen Weg gibt
//
// `Ausschneiden` löschte Karte für Karte, jede in ihrer eigenen Transaktion.
// Scheiterte die dritte, waren die ersten beiden fort — und die Oberfläche
// bekam einen Fehler, den sie als „nichts geschrieben" behandelte. **MSE2
// bündelt** dagegen: `AddCardAction` nimmt beim Einfügen alle Karten zugleich
// (`src/gui/control/card_list.cpp`, Zeile 170, Fassung `f9d9356` vom
// 2021-07-06), und dasselbe gilt für das Löschen einer Auswahl.
package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dreiKarten legt ein Set mit drei Karten an und gibt ihre Kennungen.
func dreiKarten(t *testing.T) (db *sql.DB, setID int64, kennungen []int64) {
	t.Helper()

	db = frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
			{Values: []domain.Value{{Field: "name", Text: "Zweite"}}},
			{Values: []domain.Value{{Field: "name", Text: "Dritte"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	kennungen = make([]int64, 0, len(karten))
	for _, karte := range karten {
		kennungen = append(kennungen, karte.ID)
	}

	return db, setID, kennungen
}

// **Alle genannten Karten sind danach fort.**
func TestMehrereKartenGehenInEinemZug(t *testing.T) {
	db, setID, kennungen := dreiKarten(t)
	erste, zweite, dritte := kennungen[0], kennungen[1], kennungen[2]

	schreiber := storage.NeuerKartenschreiber(db)

	if err := schreiber.KartenLoeschen(probekontext(t), []int64{erste, dritte}); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 1 || karten[0].ID != zweite {
		t.Fatalf("übrig sind %d Karten, erwartet nur die zweite", len(karten))
	}

	_ = setID
}

// **Scheitert eine, bleibt keine gelöscht.**
//
// Das ist die Zusage, für die es diesen Weg gibt: ein halber Schnitt ist der
// Zustand, den niemand mehr auflösen kann — die Ablage trägt alle Karten, das
// Set nur noch einen Teil.
func TestEinGescheiterterZugLoeschtGarNichts(t *testing.T) {
	db, setID, kennungen := dreiKarten(t)
	erste, zweite := kennungen[0], kennungen[1]

	schreiber := storage.NeuerKartenschreiber(db)

	// **Die dritte Kennung gibt es nicht** — der Zug muss abbrechen und die
	// erste stehen lassen.
	err := schreiber.KartenLoeschen(probekontext(t), []int64{erste, zweite, 99999})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Fehler = %v, erwartet ErrNotFound", err)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(karten) != 3 {
		t.Fatalf("übrig sind %d Karten, erwartet alle drei", len(karten))
	}
}

// **Der ganze Zug ist ein Schritt im Verlauf.**
//
// Ein ⌘Z holt alle Karten zurück, nicht die letzte. Dieselbe Gestalt wie beim
// Ersetzen über ein Set (Werkstück 2).
func TestEinZugIstEinSchrittImVerlauf(t *testing.T) {
	db, setID, kennungen := dreiKarten(t)
	erste, zweite := kennungen[0], kennungen[1]

	schreiber := storage.NeuerKartenschreiber(db)

	if err := schreiber.KartenLoeschen(probekontext(t), []int64{erste, zweite}); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	_, bewegt, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if !bewegt {
		t.Fatal("nichts bewegt")
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(karten) != 3 {
		t.Fatalf("nach einem ⌘Z sind %d Karten da, erwartet alle drei", len(karten))
	}
}
