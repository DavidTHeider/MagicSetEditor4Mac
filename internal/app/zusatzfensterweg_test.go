package app_test

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was das Fenster als Zusatzwert setzt, steht im Bestand** — und die
// Kommandozeile liest es (#2298). Am nativen Fenster (`stilfelder`) sah es
// zuerst anders aus: der Ruf kam gelungen zurück, die Vorschau zeigte den
// Wert, und `mse zusatzfeld` fand ihn nicht. Die Ursache lag im Drehbuch —
// es las über `.go/bin/mse`, das die Saat des Bündels nicht kennt —; dieser
// Lauf hält den Weg selbst fest, über einen echten Bestand.
func TestDasFensterSchreibtEinenZusatzwertInDenBestand(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	bestand, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Datenbankpfad: %v", err)
	}

	db, err := storage.Oeffnen(probekontext(t), bestand)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if migrationsfehler := storage.Migriere(probekontext(t), db, nil); migrationsfehler != nil {
		t.Fatalf("migrieren: %v", migrationsfehler)
	}

	setID, karteID := saatsetImBestand(t, db)
	set, karte := strconv.FormatInt(setID, 10), strconv.FormatInt(karteID, 10)

	anwendung := app.NeueApp(bestand)
	if fehler := anwendung.Starten(t.Context()); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	if antwort := anwendung.ZusatzwertSetzen(t.Context(), set, karte, "vermerk", "Vorabdruck"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	gelesen := anwendung.Kartenzusatzfelder(t.Context(), set, karte)
	if gelesen.Fehler != nil || len(gelesen.Felder) == 0 || gelesen.Felder[0].Wert != "Vorabdruck" {
		t.Errorf("das Fenster liest %+v", gelesen)
	}

	weg, err := app.ZusatznahtUeber(db, domain.UrheberKommandozeile)
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	felder, err := weg.Felder(probekontext(t), setID, karteID)
	if err != nil || len(felder) == 0 || felder[0].Wert != "Vorabdruck" {
		t.Errorf("die Kommandozeile liest %+v, Fehler %v", felder, err)
	}
}
