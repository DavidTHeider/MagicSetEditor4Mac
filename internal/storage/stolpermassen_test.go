// **Die Fehlerpfade der Massenänderung** — was `FelderSetzen` und
// `WerteSetzen` melden, wenn der Treiber unterwegs abbricht.
//
// **Eine dritte Datei neben `stolpertreiber_test.go` und
// `stolperabfragen_test.go`**, weil jene mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **wo** gestolpert wird und **was**
// dabei zurückkommt; hier stehen die beiden Wege, die diese Antworten
// **benutzen** — und sie sind mit einer echten SQLite-Datenbank nicht
// herzustellen.
package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Fehlerpfade der Massenänderung.
//
// Sie sind mit einer echten SQLite-Datenbank nicht herzustellen: die Vorprüfung
// steht in derselben Transaktion wie das Schreiben, und was sie durchlässt,
// schreibt sich auch. Im Betrieb entstehen sie an einer vollen Platte oder einer
// abgebrochenen Verbindung.
func TestFelderSetzenMeldetDieFehlerpfade(t *testing.T) {
	faelle := []struct {
		name    string
		treiber string
	}{
		{"die Vorprüfung scheitert", "stolper-fragen"},
		{"eine Kennung ist unlesbar", "stolper-umwandeln"},
		{"ein Lesefehler nach der Schleife", "stolper-lesen"},
		{"das Schreiben scheitert", "stolper-schreiben"},
		{"das Commit scheitert", "stolper-commit"},
		{"der Vorwert ist nicht lesbar", "stolper-massen-vorwert"},
		{"der Verlaufseintrag scheitert", "stolper-massen-verlauf"},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db, err := sql.Open(fall.treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			_, fehler := NeuerKartenschreiber(db).
				FelderSetzen(probekontext(t), []int64{1}, "typ", "Artefakt")
			if fehler == nil {
				t.Error("die Massenänderung kam ohne Fehler zurück")
			}
		})
	}
}

// **Dieselben Wege, dieselbe Zusage — für die Ersetzung** (#1168).
//
// `WerteSetzen` geht durch dieselbe Vorprüfung und dieselbe `massenzeile`;
// eine Ersetzung, die einen dieser Fehler still schluckte, hinterließe einen
// halben Bestand.
func TestWerteSetzenMeldetDieFehlerpfade(t *testing.T) {
	for _, treiber := range []string{
		"stolper-fragen", "stolper-umwandeln", "stolper-lesen",
		"stolper-schreiben", "stolper-commit",
		"stolper-massen-vorwert", "stolper-massen-verlauf",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			// **Der Vorwert ist, was der Treiber liest** (#2534) — sonst hielte
			// schon die Bedingung an, und keine Stolperstelle dahinter würde
			// erreicht.
			if fehler := NeuerKartenschreiber(db).WerteSetzen(probekontext(t),
				[]domain.Feldwechsel{{KarteID: 1, Feld: "typ", Wert: "Artefakt", Vorher: "neu"}},
			); fehler == nil {
				t.Error("die Ersetzung kam ohne Fehler zurück")
			}
		})
	}
}
