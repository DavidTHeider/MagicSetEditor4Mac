package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Marke wird über dieselbe Suchfläche gefunden wie ein Feldwert** —
// #1208.
//
// Die Frage ist dieselbe: „wo steht das Wort?" Ob im Regeltext oder an der
// Karte, entscheidet der Suchende — der Treffer sagt es ihm über sein Feld.
func TestEineMarkeWirdMitgefunden(t *testing.T) {
	db, ablage, setID, karteID := markenstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, "prüfen"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	ergebnis, err := storage.NeuerSetLeser(db).
		Suche(probekontext(t), domain.SearchQuery{Text: "Prüfen", SetID: setID})
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if len(ergebnis.Hits) != 1 {
		t.Fatalf("%d Treffer: %+v", len(ergebnis.Hits), ergebnis.Hits)
	}

	treffer := ergebnis.Hits[0]
	if treffer.Field != domain.Markenfeld || treffer.Value != "prüfen" ||
		treffer.CardID != karteID {
		t.Errorf("Treffer: %+v", treffer)
	}
}

// **Ein Feldfilter schließt die Marken aus — außer er meint sie.**
func TestEinFeldfilterEntscheidetUeberDieMarken(t *testing.T) {
	db, ablage, setID, karteID := markenstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, "prüfen"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	leser := storage.NeuerSetLeser(db)

	for _, fall := range []struct {
		name, feld string
		will       int
	}{
		{name: "ohne Filter mitgefunden", feld: "", will: 1},
		{name: "in einem Kartenfeld nicht", feld: "name", will: 0},
		{name: "im Markenfeld nur dort", feld: domain.Markenfeld, will: 1},
	} {
		t.Run(fall.name, func(t *testing.T) {
			ergebnis, err := leser.Suche(probekontext(t),
				domain.SearchQuery{Text: "prüfen", Field: fall.feld, SetID: setID})
			if err != nil {
				t.Fatalf("suchen: %v", err)
			}

			if len(ergebnis.Hits) != fall.will {
				t.Errorf("%d Treffer, erwartet %d: %+v",
					len(ergebnis.Hits), fall.will, ergebnis.Hits)
			}
		})
	}
}

// **Ein Text, der keine Suche hergibt, findet auch keine Marke.**
//
// Ein leeres FTS5-Muster heißt, dass nichts Suchbares dasteht; eine
// Markensuche darüber gäbe **alles** zurück.
func TestEinLeeresMusterFindetAuchKeineMarke(t *testing.T) {
	db, ablage, setID, karteID := markenstand(t)

	if err := ablage.Setzen(probekontext(t), karteID, "prüfen"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	ergebnis, err := storage.NeuerSetLeser(db).
		Suche(probekontext(t), domain.SearchQuery{Text: "   ", SetID: setID})
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	if len(ergebnis.Hits) != 0 {
		t.Errorf("%d Treffer: %+v", len(ergebnis.Hits), ergebnis.Hits)
	}
}
