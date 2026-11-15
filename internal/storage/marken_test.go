package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Marken an Karten** —
// #1208.
func TestEineMarkeWirdGesetztUndGelesen(t *testing.T) {
	_, ablage, setID, karteID := markenstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, "fertig"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// **Zweimal dieselbe Marke ist kein Fehler, sondern nichts.**
	if err := ablage.Setzen(probekontext(t), karteID, "fertig"); err != nil {
		t.Fatalf("zweites Setzen: %v", err)
	}

	marken, err := ablage.AnKarte(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(marken) != 1 || marken[0] != "fertig" {
		t.Fatalf("Marken der Karte: %v", marken)
	}

	imSet, err := ablage.ImSet(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Set lesen: %v", err)
	}

	if len(imSet) != 1 || imSet[0] != "fertig" {
		t.Errorf("Marken des Sets: %v", imSet)
	}
}

// **Und sie lässt sich wieder abnehmen.**
func TestEineMarkeLaesstSichEntfernen(t *testing.T) {
	_, ablage, _, karteID := markenstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, "fertig"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if err := ablage.Entfernen(probekontext(t), karteID, "fertig"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	marken, err := ablage.AnKarte(probekontext(t), karteID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(marken) != 0 {
		t.Errorf("Marken: %v", marken)
	}
}

// **Eine Marke, die gar nicht dasteht, abzunehmen ist ein Befund** — sonst
// meldete eine Oberfläche Erfolg für eine Handlung, die nichts getan hat (§8).
func TestEineFehlendeMarkeAbzunehmenIstEinBefund(t *testing.T) {
	_, ablage, _, karteID := markenstand(t)

	err := ablage.Entfernen(probekontext(t), karteID, "gibtesnicht")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Entfernen = %v, erwartet ErrNotFound", err)
	}
}

// **Eine Marke an einer Karte, die es nicht gibt, wird abgewiesen.**
//
// Der Fremdschlüssel greift — seit `foreign_keys(1)` gesetzt wird.
func TestEineMarkeOhneKarteWirdAbgewiesen(t *testing.T) {
	_, ablage, _, _ := markenstand(t)

	err := ablage.Setzen(probekontext(t), 4711, "fertig")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Setzen = %v, erwartet ErrNotFound", err)
	}
}

// **Je Karte gesammelt** — für die Kartenliste, in **einem** Ruf.
func TestDieMarkenEinesSetsKommenJeKarte(t *testing.T) {
	_, ablage, setID, karteID := markenstand(t)

	for _, marke := range []string{"prüfen", "fertig"} {
		if err := ablage.Setzen(probekontext(t), karteID, marke); err != nil {
			t.Fatalf("setzen: %v", err)
		}
	}

	jeKarte, err := ablage.JeKarte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **In der Reihenfolge der Marke**, damit zwei Läufe dasselbe zeigen.
	if len(jeKarte[karteID]) != 2 ||
		jeKarte[karteID][0] != "fertig" || jeKarte[karteID][1] != "prüfen" {
		t.Errorf("je Karte: %v", jeKarte)
	}
}

// markenstand baut Bestand, Set und eine Karte.
func markenstand(t *testing.T) (db *sql.DB, ablage *storage.Marken, setID, karteID int64) {
	t.Helper()

	db = frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
		},
	}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Karten lesen: %v", err)
	}

	return db, storage.NeueMarken(db), setID, karten[0].ID
}

// **Jeder Weg meldet, wenn der Bestand nicht mehr da ist** (§8).
//
// **Eine geschlossene Datenbank und kein kaputtes SQL**: sie ist der Fall, der
// im Betrieb vorkommt — der Bestand wird zugemacht, während noch jemand liest.
func TestJederMarkenwegMeldetEinenBestandsfehler(t *testing.T) {
	db, ablage, setID, karteID := markenstand(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	for name, ruf := range map[string]func() error{
		"Setzen":    func() error { return ablage.Setzen(probekontext(t), karteID, "a") },
		"Entfernen": func() error { return ablage.Entfernen(probekontext(t), karteID, "a") },
		"AnKarte": func() error {
			_, err := ablage.AnKarte(probekontext(t), karteID)

			return err
		},
		"ImSet": func() error {
			_, err := ablage.ImSet(probekontext(t), setID)

			return err
		},
		"JeKarte": func() error {
			_, err := ablage.JeKarte(probekontext(t), setID)

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := ruf(); err == nil {
				t.Error("der geschlossene Bestand kam ohne Fehler durch")
			}
		})
	}
}
