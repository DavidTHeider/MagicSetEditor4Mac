package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege der Optionsablage** (#492).
//
// Paketintern wie `einstellungsfehler_intern_test.go` und aus demselben Grund:
// mit einer funktionierenden Datenbank sind sie nicht erreichbar, weil sie
// funktioniert.
func TestDieOptionsablageMeldetIhreFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Abfrage scheitert":    "stolper-fragen",
		"das Lesen scheitert":      "stolper-lesen",
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

			if _, fehler := NeueExportoptionen(db).Fuer(probekontext(t), "v"); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}

func TestDasSetzenEinerOptionMeldetEineGeschlosseneDatenbank(t *testing.T) {
	if err := NeueExportoptionen(geschlosseneDatenbank(t)).
		Setzen(probekontext(t), "v", "f", "w"); err == nil {
		t.Error("das Setzen meldete keinen Fehler")
	}
}
