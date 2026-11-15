package storage_test

// **Eine Ersetzung schreibt nur über dem Text, aus dem sie rechnete** —
// Nachlauf über das Gesamtreview über 0.9.0, Befund 2
// (#2534).
//
// `sammle` rechnete die neuen Texte aus gelesenen Karten, und `WerteSetzen`
// schrieb sie ohne Vergleich: ergänzte ein anderer Klient nach dem Lesen
// etwas am Feld, schrieb die Ersetzung ihren vorberechneten Text darüber,
// und beide meldeten Erfolg.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineErsetzungSchreibtNurUeberIhremVorwert(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	// **Der zweite Wechsel rechnete aus einem Text, der nicht mehr dasteht** —
	// und dann schreibt auch der erste nicht.
	fehler := storage.NeuerKartenschreiber(db).WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "name", Wert: "Feuerkugel", Vorher: "Feuerball"},
		{KarteID: karten[2], Feld: "name", Wert: "Steinmauer", Vorher: "Steinwand, dazwischen ergänzt"},
	})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("setzen: %v, erwartet ErrConflict", fehler)
	}

	if gelesen := werte(t, db, setID, "name"); gelesen[0] != "Feuerball" || gelesen[2] != "Steinwand" {
		t.Errorf("nach dem Konflikt %v — erwartet unverändert", gelesen)
	}
}
