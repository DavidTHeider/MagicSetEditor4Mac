package storage

// **Der Haushalt des zweiten Faktors** — SEC-08, SEC-09, SEC-10
// (#2149,
// #2150,
// #2151,
// Vermerk 0178).
//
// Gegen eine echte Datei, und wo es um Gleichzeitigkeit geht, über **zwei**
// Vorräte auf derselben Datei — für SQLite dasselbe wie zwei Prozesse.

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiVorraete öffnet einen frisch migrierten Bestand zweimal.
func zweiVorraete(t *testing.T) (eins, zwei *sql.DB) {
	t.Helper()

	quelle := "file:" + filepath.Join(t.TempDir(), "bestand.db")
	ctx := probekontext(t)

	var err error

	eins, err = Oeffnen(ctx, quelle)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = eins.Close() })

	if fehler := Migriere(ctx, eins, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	zwei, err = Oeffnen(ctx, quelle)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = zwei.Close() })

	return eins, zwei
}

func mitTOTP(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO nutzerfaktoren (nutzer_id, art) VALUES (1, 'totp')`); err != nil {
		t.Fatal(err)
	}
}

func TestEinTOTPSchrittGiltEinmal(t *testing.T) {
	eins, zwei := zweiVorraete(t)
	ctx := probekontext(t)
	mitTOTP(t, eins)

	ablage := NeueNutzerablage(eins)

	for _, fall := range []struct {
		name    string
		schritt int64
		gilt    bool
	}{
		{"der erste", 100, true},
		{"derselbe noch einmal", 100, false},
		{"ein älterer im Fenster", 99, false},
		{"der nächste", 101, true},
	} {
		gilt, err := ablage.SchrittVerbrauchen(ctx, 1, fall.schritt)
		if err != nil || gilt != fall.gilt {
			t.Errorf("%s: gilt %v (%v), erwartet %v", fall.name, gilt, err, fall.gilt)
		}
	}

	if gilt, err := ablage.SchrittVerbrauchen(ctx, 2, 500); err != nil || gilt {
		t.Errorf("ohne Faktor: gilt %v (%v)", gilt, err)
	}

	// **Zwei Vorräte, ein Schritt: höchstens einer gilt.**
	var (
		warte  sync.WaitGroup
		gelten = make(chan bool, 2)
	)

	for _, db := range []*sql.DB{eins, zwei} {
		warte.Go(func() {
			gilt, err := NeueNutzerablage(db).SchrittVerbrauchen(ctx, 1, 200)
			if err != nil {
				t.Error(err)
			}

			gelten <- gilt
		})
	}

	warte.Wait()
	close(gelten)

	if anzahl := zaehleWahre(gelten); anzahl != 1 {
		t.Errorf("derselbe Schritt galt %d-mal", anzahl)
	}
}

func zaehleWahre(werte <-chan bool) (anzahl int) {
	for wert := range werte {
		if wert {
			anzahl++
		}
	}

	return anzahl
}

func TestFehlversucheHabenEinBudget(t *testing.T) {
	eins, _ := zweiVorraete(t)
	ctx := probekontext(t)
	ablage := NeueNutzerablage(eins)
	jetzt := time.Date(2026, 10, 7, 4, 0, 0, 0, time.UTC)

	beginne := func(wann time.Time) time.Time {
		t.Helper()

		bis, err := ablage.FaktorversuchBeginnen(ctx, 1, wann)
		if err != nil {
			t.Fatal(err)
		}

		return bis
	}

	// **Fünf Versuche gehen**, und der fünfte sperrt schon für die übrigen.
	for versuch := range 5 {
		if bis := beginne(jetzt); !bis.IsZero() {
			t.Fatalf("Versuch %d gesperrt bis %v", versuch+1, bis)
		}
	}

	if bis := beginne(jetzt.Add(29 * time.Second)); !bis.Equal(jetzt.Add(30 * time.Second)) {
		t.Errorf("der sechste: gesperrt bis %v", bis)
	}

	// **Nach der Sperre geht einer — und die nächste ist doppelt so lang.**
	if bis := beginne(jetzt.Add(30 * time.Second)); !bis.IsZero() {
		t.Errorf("nach der Sperre: gesperrt bis %v", bis)
	}

	if bis := beginne(jetzt.Add(31 * time.Second)); !bis.Equal(jetzt.Add(90 * time.Second)) {
		t.Errorf("danach: gesperrt bis %v", bis)
	}

	// **Ein gelungener Versuch setzt zurück.**
	if err := ablage.FaktorversuchGelungen(ctx, 1); err != nil {
		t.Fatal(err)
	}

	if bis := beginne(jetzt.Add(31 * time.Second)); !bis.IsZero() {
		t.Errorf("nach dem Erfolg: gesperrt bis %v", bis)
	}
}

func TestGleichzeitigeVersucheTeilenSichDasBudget(t *testing.T) {
	eins, zwei := zweiVorraete(t)
	ctx := probekontext(t)
	jetzt := time.Date(2026, 10, 7, 4, 0, 0, 0, time.UTC)

	for range 4 {
		if _, err := NeueNutzerablage(eins).FaktorversuchBeginnen(ctx, 1, jetzt); err != nil {
			t.Fatal(err)
		}
	}

	var (
		warte   sync.WaitGroup
		erlaubt = make(chan bool, 2)
	)

	for _, db := range []*sql.DB{eins, zwei} {
		warte.Go(func() {
			bis, err := NeueNutzerablage(db).FaktorversuchBeginnen(ctx, 1, jetzt)
			if err != nil {
				t.Error(err)
			}

			erlaubt <- bis.IsZero()
		})
	}

	warte.Wait()
	close(erlaubt)

	if anzahl := zaehleWahre(erlaubt); anzahl != 1 {
		t.Errorf("beim fünften Versuch gingen %d gleichzeitig durch", anzahl)
	}
}

func TestEinVersuchFuerNiemandenIstNichtGefunden(t *testing.T) {
	db := offeneDatenbank(t)

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	if _, err := NeueNutzerablage(db).FaktorversuchBeginnen(context.Background(), 99, time.Now()); err == nil {
		t.Error("ein Versuch für einen Nutzer, den es nicht gibt")
	}
}

func TestDerHaushaltMeldetSeineFehler(t *testing.T) {
	eins, _ := zweiVorraete(t)
	ctx := probekontext(t)
	ablage := NeueNutzerablage(eins)

	// **Eine Tafel, die fehlt, ist kein „noch kein Versuch".**
	if _, err := eins.ExecContext(ctx, `DROP TABLE faktorhaushalt`); err != nil {
		t.Fatal(err)
	}

	if _, err := ablage.FaktorversuchBeginnen(ctx, 1, time.Now()); err == nil {
		t.Error("ohne Tafel angemeldet")
	}

	if err := eins.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := ablage.FaktorversuchBeginnen(ctx, 1, time.Now()); err == nil {
		t.Error("ohne Bestand angemeldet")
	}

	if _, err := ablage.SchrittVerbrauchen(ctx, 1, 1); err == nil {
		t.Error("ohne Bestand verbraucht")
	}
}

// **Ein TOTP-Faktor entsteht mit seinem Schritt** (Zweitblick über F,
// Befund F1): getrennt angelegt und verbraucht, galt der Code der
// Bestätigung zwischen beiden Schritten — und nach einem Fehler dazwischen
// stand der Faktor ohne verbrauchten Schritt da.
func TestEinTOTPFaktorEntstehtMitSeinemSchritt(t *testing.T) {
	eins, _ := zweiVorraete(t)
	ctx := probekontext(t)
	ablage := NeueNutzerablage(eins)

	if _, err := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: 1, Art: domain.FaktorTOTP, LetzterSchritt: 100,
	}); err != nil {
		t.Fatal(err)
	}

	for schritt, gilt := range map[int64]bool{100: false, 101: true} {
		if galt, err := ablage.SchrittVerbrauchen(ctx, 1, schritt); err != nil || galt != gilt {
			t.Errorf("Schritt %d: gilt %v (%v), erwartet %v", schritt, galt, err, gilt)
		}
	}
}
