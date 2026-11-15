// **Die Marken für die Flächen ohne Fenster** —
// #1208.
package app_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Naht reicht alle vier Wege an dieselbe Markenpflege durch.**
//
// **Über einen echten Bestand**, weil sie ein Composition Root ist: was sie
// zusagt, ist gerade, dass Ablage und Kartenleser aus **dieser** Datenbank
// kommen. Ein Doppel dahinter prüfte die Verdrahtung nicht.
func TestDieMarkennahtReichtDurch(t *testing.T) {
	db := frischerBestand(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Karten lesen: %v", lesefehler)
	}

	karteID := karten[0].ID

	naht := app.Markennaht(db)

	if fehler := naht.MarkeSetzen(probekontext(t), setID, karteID, "Fertig"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	anKarte, lesefehler2 := naht.KartenMarken(probekontext(t), setID, karteID)
	if lesefehler2 != nil || len(anKarte) != 1 || anKarte[0] != "fertig" {
		t.Fatalf("an der Karte: %v (%v)", anKarte, lesefehler2)
	}

	imSet, setfehler := naht.Setmarken(probekontext(t), setID)
	if setfehler != nil || len(imSet) != 1 {
		t.Fatalf("im Set: %v (%v)", imSet, setfehler)
	}

	if fehler := naht.MarkeEntfernen(probekontext(t), setID, karteID, "FERTIG"); fehler != nil {
		t.Fatalf("entfernen: %v", fehler)
	}

	danach, endfehler := naht.KartenMarken(probekontext(t), setID, karteID)
	if endfehler != nil || len(danach) != 0 {
		t.Errorf("nach dem Abnehmen: %v (%v)", danach, endfehler)
	}
}
