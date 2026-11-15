// **Eine Ersetzung über ein Set ist ein Schritt** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # Der Unterschied zur Massenänderung daneben
//
// `FelderSetzen` schreibt **einen** Wert auf viele Karten. Eine Ersetzung
// schreibt auf jede Karte einen **anderen** — den ihren, mit der ersetzten
// Stelle darin. Ein Aufruf je Karte gäbe je Karte einen Schritt, und ⌘Z nähme
// dreihundertmal eine Karte zurück.
package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Verschiedene Werte, eine Gruppe, ein ⌘Z.**
func TestEineErsetzungIstEinSchritt(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	fehler := storage.NeuerKartenschreiber(db).WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "name", Wert: "Feuerkugel", Vorher: "Feuerball"},
		{KarteID: karten[2], Feld: "name", Wert: "Steinmauer", Vorher: "Steinwand"},
	})
	if fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if gelesen := werte(t, db, setID, "name"); len(gelesen) != 3 ||
		gelesen[0] != "Feuerkugel" || gelesen[1] != "Eisball" ||
		gelesen[2] != "Steinmauer" {
		t.Fatalf("nach der Ersetzung %v", gelesen)
	}

	leser := storage.NeuerVerlaufsleser(db)

	verlauf, err := leser.Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 2 {
		t.Fatalf("%d Einträge, erwartet zwei", len(verlauf))
	}

	if verlauf[0].Gruppe == "" || verlauf[0].Gruppe != verlauf[1].Gruppe {
		t.Fatalf("Gruppen: %q / %q", verlauf[0].Gruppe, verlauf[1].Gruppe)
	}

	// **Ein ⌘Z nimmt beide zurück** — und die unberührte Karte bleibt.
	if _, gab, err := leser.Zuruecknehmen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("zurücknehmen: %v / %v", gab, err)
	}

	if gelesen := werte(t, db, setID, "name"); gelesen[0] != "Feuerball" ||
		gelesen[1] != "Eisball" || gelesen[2] != "Steinwand" {
		t.Errorf("nach der Rücknahme %v", gelesen)
	}
}

// **Und wieder vorwärts stellt beide wieder her.**
//
// Der Zweig steht neben dem darüber: eine Rücknahme, die sich nicht
// wiederholen lässt, ist eine halbe.
func TestEineZurueckgenommeneErsetzungLaesstSichWiederherstellen(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	schreiber := storage.NeuerKartenschreiber(db)
	if err := schreiber.WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "name", Wert: "Feuerkugel", Vorher: "Feuerball"},
		{KarteID: karten[2], Feld: "name", Wert: "Steinmauer", Vorher: "Steinwand"},
	}); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	if _, _, err := leser.Zuruecknehmen(probekontext(t), setID); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if _, gab, err := leser.Wiederherstellen(probekontext(t), setID); err != nil || !gab {
		t.Fatalf("wiederherstellen: %v / %v", gab, err)
	}

	if gelesen := werte(t, db, setID, "name"); gelesen[0] != "Feuerkugel" ||
		gelesen[2] != "Steinmauer" {
		t.Errorf("nach dem Wiederherstellen %v", gelesen)
	}
}

// **Eine Karte, die es nicht gibt, bricht den ganzen Zug ab.**
//
// §8 und dieselbe Zusage wie bei `FelderSetzen`: eine Ersetzung, die zur
// Hälfte durchgeht, hinterlässt einen Bestand, von dem niemand sagen kann,
// welche Hälfte.
func TestEineFehlendeKarteBrichtDieErsetzungAb(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	fehler := storage.NeuerKartenschreiber(db).WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "name", Wert: "Feuerkugel", Vorher: "Feuerball"},
		{KarteID: 9999, Feld: "name", Wert: "Nichts"},
	})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("Fehler: %v", fehler)
	}

	if gelesen := werte(t, db, setID, "name"); gelesen[0] != "Feuerball" {
		t.Errorf("die erste Karte steht auf %q", gelesen[0])
	}
}

// **Ein unbrauchbarer Feldname wird abgewiesen** — wie auf jedem anderen Weg.
func TestEinUnbrauchbaresFeldBrichtDieErsetzungAb(t *testing.T) {
	db, _, karten := bestandMitDreiKarten(t)

	fehler := storage.NeuerKartenschreiber(db).WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "", Wert: "X"},
	})
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("Fehler: %v", fehler)
	}
}

// **Zwei Felder derselben Karte nennen sie in der Vorprüfung einmal.**
//
// Ohne diese Zusammenfassung stünde eine Karte, die es nicht gibt, mehrfach in
// der Fehlermeldung — und eine Ersetzung trifft oft `name` **und** `rule text`
// derselben Karte.
func TestZweiFelderDerselbenKarteNennenSieEinmal(t *testing.T) {
	db, setID, karten := bestandMitDreiKarten(t)

	fehler := storage.NeuerKartenschreiber(db).WerteSetzen(probekontext(t), []domain.Feldwechsel{
		{KarteID: karten[0], Feld: "name", Wert: "Feuerkugel", Vorher: "Feuerball"},
		{KarteID: karten[0], Feld: "text", Wert: "Brennt."},
	})
	if fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 2 || verlauf[0].Gruppe != verlauf[1].Gruppe {
		t.Errorf("%d Einträge, Gruppen %q/%q",
			len(verlauf), verlauf[0].Gruppe, verlauf[1].Gruppe)
	}
}

// **Ohne Wechsel geschieht nichts, und das ist kein Fehler.**
func TestOhneWechselSchreibtDieErsetzungNichts(t *testing.T) {
	db, setID, _ := bestandMitDreiKarten(t)

	if err := storage.NeuerKartenschreiber(db).
		WerteSetzen(probekontext(t), nil); err != nil {
		t.Fatalf("Fehler: %v", err)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 0 {
		t.Errorf("%d Einträge, erwartet keine", len(verlauf))
	}
}
