package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Fehlerwege des Kartenkopierens**
// (#625).
//
// Paketintern und über den Stolpertreiber, aus demselben Grund wie überall
// hier: mit einer funktionierenden Datenbank sind sie nicht erreichbar, weil
// sie funktioniert.
//
// **Und sie gehören geprüft, weil ein halber Stand hier eine halbe Karte
// wäre.** Scheitert das Schreiben eines Feldwerts und die Karte bleibt
// trotzdem stehen, hat jemand eine Kopie, die keine ist — und der Verlauf
// nennt sie vollständig.
func TestDasKopierenMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht":  "",
		"das Lesen der Karte scheitert":  "stolper-fragen",
		"das Lesen der Felder scheitert": "stolper-lesen",
		"das Set liest sich nicht":       "stolper-karte-setid",
		"das Einreihen scheitert":        "stolper-karte-einreihen",
		"ein Feldwert scheitert":         "stolper-karte-feldeinsetzen",
		"der Verlauf scheitert":          "stolper-karte-verlauf",
		"das Kürzen scheitert":           "stolper-karte-kuerzen",
		"der Commit scheitert":           "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if _, err := NeuerKartenschreiber(db).KarteKopieren(probekontext(t), 1); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Und die des Anlegens** — es geht seit
// #626
// denselben Weg, nur ohne Werte.
func TestDasAnlegenMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht": "",
		"das Einreihen scheitert":       "stolper-karte-einreihen",
		"der Verlauf scheitert":         "stolper-karte-verlauf",
		"das Kürzen scheitert":          "stolper-karte-kuerzen",
		"der Commit scheitert":          "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if _, err := NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), 1); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}

// **Und die des Reihenanlegens** — [Kartenschreiber.KartenAnlegen] geht denselben Weg,
// nur mehrmals unter einer Gruppenmarke
// (#1337).
//
// **Der Commit ist der Grund, warum dieser Lauf hier steht** und nicht bei den
// übrigen: er lässt sich von außen nicht zum Scheitern bringen. Die anderen
// beiden Wege — geschlossene Datenbank, Set ohne Eintrag — stehen in
// `kartenanlegen_test.go`, wo sie hingehören.
func TestDasReihenanlegenMeldetSeineFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Transaktion beginnt nicht": "",
		"das Einreihen scheitert":       "stolper-karte-einreihen",
		"der Verlauf scheitert":         "stolper-karte-verlauf",
		"das Kürzen scheitert":          "stolper-karte-kuerzen",
		"der Commit scheitert":          "stolper-karte-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db := geschlosseneDatenbank(t)

			if treiber != "" {
				geoeffnet, err := sql.Open(treiber, "")
				if err != nil {
					t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
				}

				t.Cleanup(func() { _ = db.Close() })

				db = geoeffnet
			}

			if _, err := NeuerKartenschreiber(db).KartenAnlegen(probekontext(t), 1,
				[][]domain.Feldstand{{{Feld: "name", Wert: "Plains"}}}); err == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}
