package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege der Setänderung, an einer geschlossenen Datenbank.**
//
// Paketintern wie `feldaenderung_intern_test.go`, und aus demselben Grund: mit
// einer funktionierenden Datenbank sind sie nicht erreichbar, weil sie
// funktioniert. Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht.
func TestDieSetaenderungMeldetEineGeschlosseneDatenbank(t *testing.T) {
	schreiber := NeuerKartenschreiber(geschlosseneDatenbank(t))

	if _, _, err := schreiber.SetwertSetzenMitVorwert(
		probekontext(t), 1, "set code", "PRB"); err == nil {
		t.Error("das Setzen meldete keinen Fehler")
	}
}

// **Jeder Weg hinter der Abfrage, mit dem Stolpertreiber.**
//
// Ändern, Zurückrollen, das Kürzen des Verlaufs und das Bestätigen — fünf
// Wege, die eine funktionierende Datenbank nicht hergibt und die im Betrieb
// vorkommen: eine volle Platte, eine abgebrochene Verbindung.
//
// **Drei Stellen sind die der Feldänderung, zwei eigene.** Der Weg sieht
// gleich aus, die Aufrufkette ist eine andere: die vorhandenen Stellen zählen
// Schreibvorgänge und treffen hier daneben. `stolper-set-kuerzen` erkennt
// seine Anweisung am Text statt an einer Zahl — das hält auch, wenn jemand
// einen Schreibvorgang dazwischenschiebt.
func TestDieSetaenderungMeldetJedenWegDahinter(t *testing.T) {
	faelle := map[string]string{
		"beim Ändern":                  "stolper-feld-schreiben",
		"beim Bestätigen":              "stolper-feld-commit",
		"beim Zurückrollen":            "stolper-zurueckrollen",
		"beim Kürzen des Verlaufs":     "stolper-set-kuerzen",
		"beim Bestätigen der Änderung": "stolper-set-commit",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			// **Der Zähler ist eine Paketvariable** und wird von
			// `database/sql` nicht zurückgesetzt: `Open` ruft der Treiber, und
			// ein Unterlauf erbte sonst die Zahl des vorigen — grün oder rot
			// je nach Reihenfolge.
			zweiterSchreibzaehler = 0

			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, _, fehler := NeuerKartenschreiber(db).
				SetwertSetzenMitVorwert(probekontext(t), 1, "set code", "PRB"); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}
