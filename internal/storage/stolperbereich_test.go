package storage

// **Der Stolpertreiber für den Bereichswechsel** — Befund 7 des Zweitblicks
// über Phase 5.1
// (#1405).
//
// # Warum es ihn braucht
//
// `PlatzVerschieben` fährt **vier** Anweisungen in einer Transaktion: den
// Platz lesen, die Menge im Zielbereich lesen, den alten löschen, den neuen
// setzen. Eine echte Datenbank gibt davon nur die **erste** zum Scheitern her
// — ein `DROP TABLE deckplaetze` trifft sie, und die drei danach werden nie
// erreicht. Ein Fehlerweg, den niemand geht, ist keiner (§3).

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// bereichabfrage bedient den Bereichswechsel — und lässt eine Stelle stolpern.
//
// **Der Platz muss dastehen, sonst kommt der Zug nie zur zweiten Anweisung.**
// Der erste Anlauf ließ ihn fehlen: alle drei Läufe waren grün, weil
// `platzLesenImZug` mit „gibt es nicht" abbrach — die Fehlerwege dahinter
// blieben unbetreten, und das sah aus wie ein Ergebnis (§10, die Gegenprobe).
func bereichabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	if !bereichstelle(stelle) {
		return nil, false, nil
	}

	befehl := strings.TrimSpace(abfrage)

	switch {
	case strings.HasPrefix(befehl, "SELECT id, karte_id, printing, druck_id, menge, bereich"):
		return &druckzeilen{
			spalten: []string{"id", "karte_id", "printing", "druck_id", "menge", "bereich"},
			zeilen: [][]driver.Value{{
				int64(1), nil, "aaaaaaaa-0000-4000-8000-00000000000a", nil,
				int64(4), "hauptdeck",
			}},
		}, true, nil

	case strings.HasPrefix(befehl, "SELECT menge FROM deckplaetze"):
		if stelle == beimBereichmenge {
			return nil, true, errors.New("die Menge drüben nicht lesbar")
		}

		// **Nichts drüben** — dann geht der Zug den Weg über das Setzen.
		return &druckzeilen{spalten: []string{"menge"}}, true, nil

	case strings.HasPrefix(befehl, "SELECT id FROM deckplaetze"):
		return &druckzeilen{spalten: []string{"id"}}, true, nil
	}

	return nil, false, nil
}

// bereichstelle sagt, ob diese Stelle zum Bereichswechsel gehört.
func bereichstelle(stelle stolperstelle) bool {
	return stelle == beimBereichmenge ||
		stelle == beimBereichloeschen ||
		stelle == beimBereichsetzen
}

// bereichschreiben lässt Löschen oder Setzen des Platzes scheitern.
//
// **Zwei Stellen und nicht eine**: die eine steht **vor** dem Setzen, die
// andere danach — und beide melden etwas anderes.
func bereichschreiben(stelle stolperstelle, abfrage string) (bool, error) {
	befehl := strings.TrimSpace(abfrage)

	switch {
	case stelle == beimBereichloeschen && strings.HasPrefix(befehl, "DELETE FROM deckplaetze"):
		return true, errors.New("den alten Platz nicht löschbar")
	case stelle == beimBereichsetzen && strings.HasPrefix(befehl, "INSERT INTO deckplaetze"):
		return true, errors.New("den neuen Platz nicht setzbar")
	}

	return false, nil
}
