package storage_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Massenänderung mit Erwartung schreibt nur, was die Vorschau gezeigt
// hat** — Werkstück U6 von Phase 5.1
// (#1417).
//
// Zwischen Vorschau und Anwenden kann jemand anderes schreiben — der
// MCP-Server, ein zweites Fenster. Wer dann den gezeigten Altwert
// überschriebe, überschriebe still etwas, das niemand gesehen hat. Deshalb
// trägt das Anwenden je Karte den Altwert mit, und die Transaktion vergleicht
// **vor** dem ersten Schreiben: weicht eine ab, wird **nichts** geschrieben,
// und der Fehler nennt die abweichenden Karten.
func TestFelderSetzenMitErwartungWeistAbweichendeKartenAb(t *testing.T) {
	leser, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	// Karte 2 hat „Zauber" — die Erwartung sagt etwas anderes.
	_, err := schreiber.FelderSetzenMitErwartung(probekontext(t), []domain.Feldvorwert{
		{KarteID: 1, Wert: "Zauber", GabEs: true},
		{KarteID: 2, Wert: "Kreatur", GabEs: true},
	}, "typ", "Land")
	if err == nil {
		t.Fatal("die Änderung ging trotz abweichender Karte durch")
	}

	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("kein Konflikt: %v", err)
	}

	var massenfehler *domain.BulkError
	if !errors.As(err, &massenfehler) {
		t.Fatalf("kein BulkError: %v", err)
	}

	if len(massenfehler.CardIDs) != 1 || massenfehler.CardIDs[0] != 2 {
		t.Errorf("genannt: %v, erwartet [2]", massenfehler.CardIDs)
	}

	// **Nichts geschrieben** — auch Karte 1 nicht, deren Erwartung stimmte.
	gelesen, lesefehler := leser.KarteLesen(probekontext(t), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if wert, _ := gelesen.Value("typ"); wert != "Zauber" {
		t.Errorf("Karte 1 wurde geschrieben: %q", wert)
	}
}

// **Ein Feld, das es nicht gab, wird als „gab es nicht" erwartet** — und ein
// Wert, der inzwischen dasteht, ist ein Konflikt.
func TestFelderSetzenMitErwartungUnterscheidetLeerVonNichtDa(t *testing.T) {
	_, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)

	// Karte 2 hat kein „text" — das passt.
	vorgang, err := schreiber.FelderSetzenMitErwartung(probekontext(t), []domain.Feldvorwert{
		{KarteID: 2, GabEs: false},
	}, "text", "Neu")
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if vorgang == "" {
		t.Error("kein Vorgang")
	}

	// Jetzt steht „Neu" da — dieselbe Erwartung ist nun falsch.
	if _, fehler := schreiber.FelderSetzenMitErwartung(probekontext(t), []domain.Feldvorwert{
		{KarteID: 2, GabEs: false},
	}, "text", "Noch neuer"); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("kein Konflikt, sondern: %v", fehler)
	}
}

// **Stimmt alles, schreibt sie wie `FelderSetzen`** — in einem Schritt des
// Verlaufs, mit Vorgang.
func TestFelderSetzenMitErwartungSchreibtBeiPassendemStand(t *testing.T) {
	leser, db := suchbestand(t)

	vorgang, err := storage.NeuerKartenschreiber(db).FelderSetzenMitErwartung(probekontext(t),
		[]domain.Feldvorwert{
			{KarteID: 1, Wert: "Zauber", GabEs: true},
			{KarteID: 2, Wert: "Zauber", GabEs: true},
		}, "typ", "Land")
	if err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if vorgang == "" {
		t.Error("kein Vorgang")
	}

	for _, karte := range []int64{1, 2} {
		gelesen, lesefehler := leser.KarteLesen(probekontext(t), karte)
		if lesefehler != nil {
			t.Fatalf("lesen: %v", lesefehler)
		}

		if wert, _ := gelesen.Value("typ"); wert != "Land" {
			t.Errorf("Karte %d: %q", karte, wert)
		}
	}
}

// **Ohne Karten geschieht nichts** — und es gibt keinen Vorgang.
func TestFelderSetzenMitErwartungOhneKartenTutNichts(t *testing.T) {
	_, db := suchbestand(t)

	vorgang, err := storage.NeuerKartenschreiber(db).FelderSetzenMitErwartung(
		probekontext(t), nil, "typ", "Land")
	if err != nil {
		t.Fatalf("Fehler: %v", err)
	}

	if vorgang != "" {
		t.Errorf("Vorgang %q ohne Karten", vorgang)
	}
}

// **Die Fehlerwege der Erwartung**: ein leerer Feldname, eine Karte, die es
// nicht gibt, ein Bestand, der den Stand nicht mehr lesen kann, und einer,
// der gar nicht mehr offen ist.
func TestFelderSetzenMitErwartungMeldetSeineFehlerwege(t *testing.T) {
	_, db := suchbestand(t)
	schreiber := storage.NeuerKartenschreiber(db)
	erwartet := []domain.Feldvorwert{{KarteID: 1, Wert: "Zauber", GabEs: true}}

	if _, err := schreiber.FelderSetzenMitErwartung(probekontext(t), erwartet, "", "Land"); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ein leerer Feldname ging durch: %v", err)
	}

	_, fehlend := schreiber.FelderSetzenMitErwartung(probekontext(t),
		[]domain.Feldvorwert{{KarteID: 999, Wert: "x", GabEs: true}}, "typ", "Land")

	var massenfehler *domain.BulkError
	if !errors.As(fehlend, &massenfehler) || !errors.Is(fehlend, domain.ErrNotFound) {
		t.Errorf("eine fehlende Karte gab %v", fehlend)
	}

	// **Der Stand lässt sich nicht lesen**: ohne die Wertetabelle ist die
	// Erwartung nicht zu prüfen — und das ist ein Fehler an der Karte, kein
	// Konflikt.
	if _, err := db.ExecContext(probekontext(t), `DROP TABLE card_values`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	_, unlesbar := schreiber.FelderSetzenMitErwartung(probekontext(t), erwartet, "typ", "Land")
	if !errors.As(unlesbar, &massenfehler) || errors.Is(unlesbar, domain.ErrConflict) {
		t.Errorf("ohne Wertetabelle kam %v", unlesbar)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if _, zu := schreiber.FelderSetzenMitErwartung(probekontext(t), erwartet, "typ", "Land"); zu == nil {
		t.Error("ein geschlossener Bestand nahm die Änderung an")
	}
}

// **Das Schreiben selbst kann scheitern** — nach geprüfter Erwartung: ein
// Trigger sperrt die Wertetabelle, und die Änderung meldet es, statt halb
// geschrieben dazustehen.
func TestFelderSetzenMitErwartungMeldetEinGesperrtesSchreiben(t *testing.T) {
	leser, db := suchbestand(t)

	if _, err := db.ExecContext(probekontext(t),
		`CREATE TRIGGER sperre BEFORE UPDATE ON card_values
		 BEGIN SELECT RAISE(ABORT, 'gesperrt'); END`); err != nil {
		t.Fatalf("Trigger: %v", err)
	}

	_, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitErwartung(probekontext(t),
		[]domain.Feldvorwert{{KarteID: 1, Wert: "Zauber", GabEs: true}}, "typ", "Land")
	if fehler == nil {
		t.Fatal("das gesperrte Schreiben ging durch")
	}

	gelesen, lesefehler := leser.KarteLesen(probekontext(t), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if wert, _ := gelesen.Value("typ"); wert != "Zauber" {
		t.Errorf("Karte 1 wurde trotzdem geschrieben: %q", wert)
	}
}

// **Auch der Abschluss kann scheitern** — der Stolpertreiber lässt genau den
// Commit fallen, nachdem Kennungen, Stand und Schreiben durch waren.
func TestFelderSetzenMitErwartungMeldetEinenGescheitertenAbschluss(t *testing.T) {
	db, err := sql.Open("stolper-erwartung-commit", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := storage.NeuerKartenschreiber(db).FelderSetzenMitErwartung(probekontext(t),
		[]domain.Feldvorwert{{KarteID: 1}}, "typ", "Land"); fehler == nil {
		t.Error("der gescheiterte Abschluss kam nicht an")
	}
}
