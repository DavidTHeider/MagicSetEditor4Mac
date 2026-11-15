package storage

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

// **Die Herkunftszeile scheitert laut, nicht still** —
// #1025.
//
// Sie wird in derselben Transaktion geführt wie der Wert. Ginge ihr Fehler
// unter, hätte man genau den Zustand, den die Tabelle ausschließen soll: ein
// Wert von außen ohne seine Herkunft — oder, schlimmer, ein eigener Wert mit
// einer fremden.
func TestDieHerkunftszeileMeldetJedenWeg(t *testing.T) {
	for _, fall := range []struct {
		name     string
		treiber  string
		herkunft string
		meldung  string
	}{
		{
			"beim Festhalten", "stolper-herkunft-schreiben",
			"https://scryfall.com/card/lea/232", `die Herkunft von "name" festhalten`,
		},
		{"beim Löschen", "stolper-herkunft-loeschen", "", `die Herkunft von "name" löschen`},
	} {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			_, _, fehler := NeuerKartenschreiber(db).Von(fall.herkunft).
				FeldSetzenMitVorwert(probekontext(t), 1, "name", "x")
			if fehler == nil {
				t.Fatalf("kein Fehler")
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

// **Und das Lesen ebenso.**
//
// Drei Wege, die eine funktionierende Datenbank nicht hergibt: die Abfrage
// selbst, das Umwandeln einer Zeile, das Ende der Zeilenfolge. Der letzte ist
// der, den man vergisst — `Next` gibt bei einem Fehler `false` zurück, und wer
// danach nicht `Err` fragt, hält einen Abbruch für ein Ende und gibt eine zu
// kurze Liste als vollständig aus.
func TestDasLesenDerHerkuenfteMeldetJedenWeg(t *testing.T) {
	for _, treiber := range []string{"stolper-fragen", "stolper-lesen", "stolper-umwandeln"} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeuerSetLeser(db).Herkuenfte(probekontext(t), 1); fehler == nil {
				t.Error("kein Fehler")
			}
		})
	}
}
