package storage_test

// **Wann der Bestand gesichert wurde** — M01
// (#2345),
// die erste Wanderung nach der Grundlinie (Vermerk 0189, Satz 6).

import (
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// vermerkUm ist ein Vermerk, der jede Prüfung besteht.
func vermerkUm(zeit time.Time, datei string) domain.Sicherungsvermerk {
	return domain.Sicherungsvermerk{
		Gesichert: zeit, Anlass: domain.SicherungVonHand, Flaeche: domain.FlaecheCLI,
		Datei: datei, Groesse: 8192, Pruefsumme: strings.Repeat("0f", 32),
	}
}

func TestDerLetzteVermerkIstDerJuengste(t *testing.T) {
	db := frischeDateidatenbank(t)
	vermerke := storage.NeueSicherungsvermerke(db)
	frueh := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)

	// **Nach der Zeit und nicht nach der Reihenfolge des Eintragens**: zwei
	// Prozesse sichern, und der langsamere vermerkt zuletzt.
	for _, vermerk := range []domain.Sicherungsvermerk{
		vermerkUm(frueh.Add(48*time.Hour), "spaet.db"),
		vermerkUm(frueh, "frueh.db"),
	} {
		if fehler := vermerke.Vermerken(t.Context(), vermerk); fehler != nil {
			t.Fatalf("vermerken: %v", fehler)
		}
	}

	gelesen, _, fehler := vermerke.Letzter(t.Context())
	if fehler != nil || gelesen.Datei != "spaet.db" {
		t.Errorf("der letzte ist %q (%v), erwartet spaet.db", gelesen.Datei, fehler)
	}
}

func TestEinUngueltigerVermerkKommtNichtInDenBestand(t *testing.T) {
	db := frischeDateidatenbank(t)
	vermerke := storage.NeueSicherungsvermerke(db)

	ungueltig := vermerkUm(time.Now(), "/Users/jemand/bestand.db")
	if fehler := vermerke.Vermerken(t.Context(), ungueltig); fehler == nil {
		t.Fatal("ein Pfad statt eines Namens wurde vermerkt")
	}

	if _, bekannt, _ := vermerke.Letzter(t.Context()); bekannt { //nolint:errcheck // nur die Frage, ob einer da ist
		t.Error("der abgewiesene Vermerk steht im Bestand")
	}
}

// **Das Alter fragt nach einer Sicherung, die man wegtragen kann** — OPS01
// (#2347). Die vor einer Wanderung liegt neben dem Bestand, ohne Bilder; sie
// hilft gegen eine missglückte Wanderung und nicht gegen eine kaputte Platte.
func TestDerLetzteVermerkUebergehtDieSicherungVorDerWanderung(t *testing.T) {
	db := frischeDateidatenbank(t)
	vermerke := storage.NeueSicherungsvermerke(db)
	frueh := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)

	vorher := vermerkUm(frueh.Add(time.Hour), "bestand-vor-0.9.0.db")
	vorher.Anlass = domain.SicherungVorWanderung

	for _, vermerk := range []domain.Sicherungsvermerk{vermerkUm(frueh, "hand.msesicherung"), vorher} {
		if fehler := vermerke.Vermerken(t.Context(), vermerk); fehler != nil {
			t.Fatalf("vermerken: %v", fehler)
		}
	}

	gelesen, bekannt, fehler := vermerke.Letzter(t.Context())
	if fehler != nil || !bekannt || gelesen.Datei != "hand.msesicherung" {
		t.Errorf("der letzte ist %q (%v, %v), erwartet hand.msesicherung", gelesen.Datei, bekannt, fehler)
	}
}

// **Die Reihe des Plans** — OPS01 (#2347): nur seine eigenen, die jüngste
// zuerst, mit allem, woran eine Datei wiederzuerkennen ist.
func TestDieVermerkeEinesAnlasses(t *testing.T) {
	db := frischeDateidatenbank(t)
	vermerke := storage.NeueSicherungsvermerke(db)
	frueh := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)

	plan := func(tag int, datei string) domain.Sicherungsvermerk {
		vermerk := vermerkUm(frueh.AddDate(0, 0, tag), datei)
		vermerk.Anlass, vermerk.Flaeche, vermerk.Verschluesselt = domain.SicherungNachPlan, domain.FlaecheFenster, tag == 7

		return vermerk
	}

	for _, vermerk := range []domain.Sicherungsvermerk{
		plan(0, "plan-a.msesicherung"), vermerkUm(frueh.AddDate(0, 0, 3), "hand.msesicherung"),
		plan(7, "plan-b.msesicherung"),
	} {
		if fehler := vermerke.Vermerken(t.Context(), vermerk); fehler != nil {
			t.Fatalf("vermerken: %v", fehler)
		}
	}

	gelesen, fehler := vermerke.Vermerke(t.Context(), domain.SicherungNachPlan)
	if fehler != nil || len(gelesen) != 2 {
		t.Fatalf("%d Vermerke (%v), erwartet die zwei des Plans", len(gelesen), fehler)
	}

	if gelesen[0] != plan(7, "plan-b.msesicherung") || gelesen[1] != plan(0, "plan-a.msesicherung") {
		t.Errorf("gelesen %+v", gelesen)
	}
}
