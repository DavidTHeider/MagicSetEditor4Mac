package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege der Einstellungsablage** (#245).
//
// Paketintern wie `feldaenderung_intern_test.go` und aus demselben Grund: mit
// einer funktionierenden Datenbank sind sie nicht erreichbar, weil sie
// funktioniert. Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht.
func TestDieEinstellungenMeldenIhreFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Abfrage scheitert": "stolper-fragen",
		"das Lesen scheitert":   "stolper-lesen",

		// **`beimUmwandeln` trifft das Zeilenende**, nicht das Lesen: der
		// Treiber gibt eine Zeile her und scheitert danach — genau der Weg,
		// den `zeilen.Err()` abfängt und den ein Test sonst nie betritt.
		"das Zeilenende scheitert": "stolper-umwandeln",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			zweiterSchreibzaehler = 0

			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeueEinstellungen(db).Alle(probekontext(t)); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}

// **Und das Schreiben an einer geschlossenen Datenbank.**
func TestDasSetzenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	if err := NeueEinstellungen(geschlosseneDatenbank(t)).
		Setzen(probekontext(t), "sprache", "de"); err == nil {
		t.Error("das Setzen meldete keinen Fehler")
	}
}
