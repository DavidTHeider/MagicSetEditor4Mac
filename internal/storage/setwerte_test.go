package storage_test

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Setangaben lassen sich lesen.**
//
// # Woher diese Prüfung kommt
//
// Aus dem Zweitblick über Phase 9. `set_values` wurde seit Phase 2
// **geschrieben und nie gelesen** — eine Tabelle, die nur hineingeht. Aufgefallen
// ist es an der Ausfuhr: sie baute ein Set ohne seine Angaben, und der Setcode
// ging dabei verloren.
//
// Eine Spalte, die niemand liest, ist eine Zusage ohne Einlösung — dieselbe
// Klasse wie `sets.herkunft` in Phase 2.
func TestSetwerteLassenSichLesen(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "magic",
		Values: []domain.Value{
			{Field: "code", Text: "PRB"},
			{Field: "copyright", Text: "frei"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werte, err := storage.NeuerSetLeser(db).Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **In ihrer Reihenfolge**, wie die Karten: sie ist die der Datei, und ein
	// Rundlauf, der sie umstellt, ändert die Datei ungefragt.
	erwartet := []domain.Value{
		{Field: "code", Text: "PRB"},
		{Field: "copyright", Text: "frei"},
	}

	if len(werte) != len(erwartet) {
		t.Fatalf("Werte: %v, erwartet %v", werte, erwartet)
	}

	for i, wert := range erwartet {
		if werte[i] != wert {
			t.Errorf("Wert %d: %+v, erwartet %+v", i, werte[i], wert)
		}
	}
}

// Ein Set ohne Angaben gibt eine leere Liste, keinen Fehler.
func TestEinSetOhneAngabenGibtEineLeereListe(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Ohne", Game: "magic"}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werte, err := storage.NeuerSetLeser(db).Setwerte(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(werte) != 0 {
		t.Errorf("Werte: %v, erwartet keine", werte)
	}
}
