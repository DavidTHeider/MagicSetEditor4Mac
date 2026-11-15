package storage

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

// **Die Herkunft auf dem Rücknahmeweg meldet jeden Fehler** —
// #1036.
//
// Ginge einer unter, hätte man genau den Zustand, den Migration 0014
// ausschließt: ein zurückgenommener Wert ohne seine Herkunft — oder, schlimmer,
// ein eigener mit einer fremden.
func TestDieHerkunftAufDemRuecknahmewegMeldetJedenWeg(t *testing.T) {
	for _, fall := range []struct {
		name    string
		treiber string
		weg     feldweg
		meldung string
	}{
		{
			"beim Festhalten", "stolper-herkunft-schreiben",
			feldweg{Feld: "kunst", Wert: "x", GabEs: true, Herkunft: "https://scryfall.com/card/lea/232"},
			`die Herkunft von "kunst" festhalten`,
		},
		{
			"beim Löschen", "stolper-herkunft-loeschen",
			feldweg{Feld: "kunst", Wert: "x", GabEs: true},
			`die Herkunft von "kunst" löschen`,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			tx, txfehler := db.BeginTx(probekontext(t), nil)
			if txfehler != nil {
				t.Fatalf("Transaktion: %v", txfehler)
			}

			fehler := kartenwegGehen(probekontext(t), tx, 1, fall.weg)
			if fehler == nil {
				t.Fatal("kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.meldung) {
				t.Errorf("Fehler %v — erwartet war %q", fehler, fall.meldung)
			}

			if !errors.Is(fehler, errGestolpert) {
				t.Errorf("Fehler %v trägt den Treiberfehler nicht", fehler)
			}
		})
	}
}

// **Und das Lesen der vorigen Herkunft ebenso.**
//
// Sie wird **vor** der Änderung gelesen; scheitert das, ist die Herkunft des
// vorigen Werts danach nicht mehr zu rekonstruieren — und die Rücknahme
// stellte den Wert wieder her und seine Nennung nicht.
func TestDasLesenDerVorigenHerkunftMeldetSeinenWeg(t *testing.T) {
	db, err := sql.Open("stolper-herkunft-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	_, _, fehler := NeuerKartenschreiber(db).FeldSetzenMitVorwert(
		probekontext(t), 1, "name", "x")
	if fehler == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(fehler.Error(), `die Herkunft von "name" lesen`) {
		t.Errorf("Fehler %v", fehler)
	}
}
