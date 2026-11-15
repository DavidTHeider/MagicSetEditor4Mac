package storage

// **Die Fehlerwege des Zusagenschutzes** — Werkstück `T2` von Phase 5.6
// (#1467,
// Vermerk [0120](../../docs/entscheidungen/0120-eine-zusage-bindet-den-schreibweg.md)).
//
// **Eine eigene Datei** neben `stolperreservierungen_test.go`, weil sie eine
// andere Frage beantwortet: dort das Festhalten, hier der **Schutz** und der
// **Zähler**, die seit T2 vor jedem Schreiben stehen (§2).
//
// # Warum das ein Stolpertreiber sein muss
//
// **Weil ein echter Treiber diese Zweige nicht hergibt.** Ein Scan, der
// mitten in der Zeile abbricht, eine Frist, die keine ist, ein Ende, das erst
// nach der letzten Zeile klagt — das sind die Wege, auf denen ein halber Stand
// zurückbliebe, und sie stünden ohne diesen Lauf als ungeprüfte Zusagen da
// (§10).

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Jeder Weg des Zählers reicht seinen Fehler durch.**
func TestJederWegDesZaehlersReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-fragen", "stolper-zaehler-scan", "stolper-zaehler-frist",
		"stolper-zaehler-ende",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeueReservierungsablage(db).ReserviertJeExemplar(
				probekontext(t)); fehler == nil {
				t.Error("der Zähler kam ohne Fehler zurück")
			}
		})
	}
}

// **Und ohne Profil zählt er gar nicht.**
//
// **Jede Wurzelabfrage klemmt auf ein Profil** (#1351, Vermerk 0066); ohne den
// Zweig zählte dieser Weg über alle Profile — und zeigte fremde Zusagen als
// die eigenen.
func TestDerZaehlerBrauchtEinProfil(t *testing.T) {
	db, err := sql.Open("stolper-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := NeueReservierungsablage(db).ReserviertJeExemplar(
		t.Context()); fehler == nil {
		t.Error("der Zähler kam ohne Profil durch")
	}
}

// **Das Löschen reicht jeden seiner Fehler durch.**
//
// **Seit T2 hat es eine Transaktion** — vorher war es ein einzelnes `DELETE`.
// Damit kommen drei Wege dazu: das Beginnen, die Frage nach den Zusagen und
// das Abschließen.
func TestJederWegDesLoeschensReichtSeinenFehlerDurch(t *testing.T) {
	for _, treiber := range []string{
		"stolper-zusagen-scan", "stolper-zusagen-ende", "stolper-schreiben",
		"stolper-commit",
	} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			zweiterSchreibzaehler = 0

			if fehler := NeuerSammlungsspeicher(db).Loeschen(
				probekontext(t), 1, 1); fehler == nil {
				t.Error("das Löschen kam ohne Fehler zurück")
			}
		})
	}
}

// **Und das Ändern meldet, wenn die Zusagen nicht zu lesen sind.**
//
// **Der Unterschied zum Löschen**: hier steht die Frage **hinter** der
// Druckzeile, und ein Treiber, der schon dort scheitert, prüfte etwas
// anderes.
func TestDasAendernMeldetUnlesbareZusagen(t *testing.T) {
	db, err := sql.Open("stolper-zusagen-scan", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0

	if fehler := NeuerSammlungsspeicher(db).Aendern(probekontext(t),
		&domain.Exemplar{
			ID: 1, Fassung: 1, Printing: "abc", Menge: 1,
			Zustand: domain.ZustandNearMint, Sprache: "en",
		}); fehler == nil {
		t.Error("das Ändern kam ohne Fehler zurück")
	}
}
