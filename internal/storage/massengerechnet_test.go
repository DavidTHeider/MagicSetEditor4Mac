package storage_test

// **Eine Massenänderung, die rechnet** — #2516 (Rest aus #2396). Ein
// gerechneter Wert fällt je Karte anders aus; die Ablage schreibt deshalb je
// Karte eigene Werte, und doch in **einem** Schritt im Verlauf, mit der
// Erwartung aus der Vorschau (U6, #1417) und dem Stand, aus dem gerechnet
// wurde (#2533).

import (
	"errors"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// gerechnet gibt je Karte das gesetzte Feld und seinen Leser, wie der
// Feldweg sie nachzieht.
func gerechnet(karten []int64) []domain.Kartenschritt {
	schritte := make([]domain.Kartenschritt, 0, len(karten))
	for i, karte := range karten {
		typ := "[Artefakt " + strconv.Itoa(i) + "]"
		schritte = append(schritte, domain.Kartenschritt{KarteID: karte, Werte: []domain.Value{
			{Field: "typ", Text: typ}, {Field: "echo", Text: typ + "!", Nachgezogen: true},
		}})
	}

	return schritte
}

func TestEineGerechneteMassenaenderungIstEinSchritt(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	gruppe, fehler := storage.NeuerKartenschreiber(db).FelderGerechnetSetzen(probekontext(t), gerechnet(karten), nil, "typ")
	if fehler != nil || gruppe == "" {
		t.Fatalf("setzen: %q, %v", gruppe, fehler)
	}

	if typ, echo := werte(t, db, setID, "typ"), werte(t, db, setID, "echo"); typ[1] != "[Artefakt 1]" || echo[2] != "[Artefakt 2]!" {
		t.Fatalf("nach dem Setzen typ=%v, echo=%v", typ, echo)
	}

	leser := storage.NeuerVerlaufsleser(db)

	verlauf, err := leser.Verlauf(probekontext(t), setID)
	if err != nil || len(verlauf) != 6 {
		t.Fatalf("%d Einträge (%v), erwartet sechs", len(verlauf), err)
	}

	for _, eintrag := range verlauf {
		if eintrag.Gruppe != gruppe {
			t.Errorf("%+v gehört nicht zur Gruppe %q", eintrag, gruppe)
		}
	}

	// **Ein ⌘Z nimmt alles zurück**, das Gesetzte und das Nachgezogene.
	if _, gab, err := leser.Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	for _, feld := range []string{"typ", "echo"} {
		for i, wert := range werte(t, db, setID, feld) {
			if wert != "«fehlt»" {
				t.Errorf("Karte %d: %s steht auf %q", i, feld, wert)
			}
		}
	}
}

// **Die Erwartung aus der Vorschau gilt weiter** — und der Stand der Rechnung
// je Karte; weicht eines ab, steht nichts geschrieben.
func TestEineGerechneteMassenaenderungHaeltErwartungUndStand(t *testing.T) {
	for _, fall := range []struct {
		name     string
		erwartet func(karten []int64) []domain.Feldvorwert
		stand    string
	}{
		{name: "ein anderer Altwert", erwartet: func(karten []int64) []domain.Feldvorwert {
			return []domain.Feldvorwert{{KarteID: karten[0]}, {KarteID: karten[1], Wert: "Zauber", GabEs: true}, {KarteID: karten[2]}}
		}},
		{name: "ein überholter Stand", stand: "überholt"},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, setID, karten := bestandMitDreiKarten(t)
			schritte := gerechnet(karten)
			schritte[2].Erwartet = fall.stand

			var erwartet []domain.Feldvorwert
			if fall.erwartet != nil {
				erwartet = fall.erwartet(karten)
			}

			_, fehler := storage.NeuerKartenschreiber(db).FelderGerechnetSetzen(probekontext(t), schritte, erwartet, "typ")
			if !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("Fehler %v, erwartet ErrConflict", fehler)
			}

			if typ := werte(t, db, setID, "typ"); typ[0] != "«fehlt»" || typ[1] != "«fehlt»" {
				t.Errorf("geschrieben trotz Konflikt: %v", typ)
			}
		})
	}
}

func TestEineGerechneteMassenaenderungOhneKarteIstNichts(t *testing.T) {
	db, _, karten := bestandMitDreiKarten(t)
	schreiber := storage.NeuerKartenschreiber(db)

	if gruppe, fehler := schreiber.FelderGerechnetSetzen(probekontext(t), nil, nil, "typ"); gruppe != "" || fehler != nil {
		t.Errorf("leer: %q, %v", gruppe, fehler)
	}

	schritte := append(gerechnet(karten[:1]), domain.Kartenschritt{KarteID: 9999, Werte: []domain.Value{{Field: "typ", Text: "x"}}})
	if _, fehler := schreiber.FelderGerechnetSetzen(probekontext(t), schritte, nil, "typ"); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("fehlende Karte: %v", fehler)
	}

	if _, fehler := schreiber.FelderGerechnetSetzen(probekontext(t), gerechnet(karten), nil, ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("leerer Feldname: %v", fehler)
	}
}
