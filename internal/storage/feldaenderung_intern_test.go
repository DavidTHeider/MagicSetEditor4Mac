package storage

import (
	"database/sql"
	"testing"
)

// **Die Fehlerwege der Feldänderung, an einer geschlossenen Datenbank.**
//
// Paketintern wie `fehlerpfade_test.go`, und aus demselben Grund: mit einer
// funktionierenden Datenbank sind sie nicht erreichbar, weil sie funktioniert.
// Ein Fehlerpfad, den kein Test betritt, ist eine Behauptung darüber, was im
// Ernstfall geschieht.
func TestDieFeldaenderungMeldetEineGeschlosseneDatenbank(t *testing.T) {
	schreiber := NeuerKartenschreiber(geschlosseneDatenbank(t))

	if _, _, err := schreiber.FeldSetzenMitVorwert(probekontext(t), 1, "name", "x"); err == nil {
		t.Error("das Setzen meldete keinen Fehler")
	}

	if _, _, _, err := schreiber.FeldEntfernen(probekontext(t), 1, "name"); err == nil {
		t.Error("das Entfernen meldete keinen Fehler")
	}
}

// **Jeder Weg hinter der Abfrage, mit dem Stolpertreiber.**
//
// Ändern, Bestätigen, Zurückrollen — drei Wege, die eine funktionierende
// Datenbank nicht hergibt und die im Betrieb vorkommen: eine volle Platte, eine
// abgebrochene Verbindung.
func TestDieFeldaenderungMeldetJedenWegDahinter(t *testing.T) {
	faelle := map[string]string{
		"beim Ändern":       "stolper-feld-schreiben",
		"beim Bestätigen":   "stolper-feld-commit",
		"beim Zurückrollen": "stolper-zurueckrollen",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, _, fehler := NeuerKartenschreiber(db).
				FeldSetzenMitVorwert(probekontext(t), 1, "name", "x"); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}

// Eine Zeile, die sich nicht in die Zielfelder lesen lässt.
func TestDieFeldaenderungMeldetEineKrummeZeile(t *testing.T) {
	db, err := sql.Open("stolper-umwandeln", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, _, fehler := NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), 1, "name", "x"); fehler == nil {
		t.Error("kein Fehler")
	}
}

// Und die der Setangaben.
func TestSetwerteMeldenEineGeschlosseneDatenbank(t *testing.T) {
	if _, err := NeuerSetLeser(geschlosseneDatenbank(t)).Setwerte(probekontext(t), 1); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Zeile, die sich nicht lesen lässt, und ein Fehler nach der Schleife.**
//
// Beides ist an einer echten SQLite-Datenbank nicht herbeizuführen: `set_values`
// führt `TEXT NOT NULL`, und eine vollständige Schleife meldet dort nichts mehr.
// Beides kommt im Betrieb vor — eine volle Platte, eine abgebrochene Verbindung
// —, und dafür gibt es den Stolpertreiber (`stolpertreiber_test.go`).
func TestSetwerteMeldenUnlesbareZeilen(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			if _, fehler := NeuerSetLeser(db).Setwerte(probekontext(t), 1); fehler == nil {
				t.Errorf("%s: kein Fehler", treiber)
			}
		})
	}
}
