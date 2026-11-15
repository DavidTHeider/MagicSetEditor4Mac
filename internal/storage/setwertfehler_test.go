package storage_test

// **Die Fehlerwege der Setangaben** — abgetrennt von
// `setwertaenderung_test.go`, als die Datei mit der Profilklemme (#1351) über
// die Zeilengrenze aus §2 kam.
//
// **Nach der Frage geschnitten**: nebenan steht, was glückt; hier steht, was
// dabei schiefgehen kann und wie es sich meldet.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Jeder Fehlerweg der Setänderung, an einem kaputten Schema.**
//
// # Warum so und nicht mit dem Stolpertreiber
//
// Die Stolperstellen zählen Schreibvorgänge und sind auf die Aufrufkette der
// **Feldänderung** eingestellt (`stolpertreiber_test.go`). Die Setänderung hat
// eine andere; sie nachzubilden hieße, den Treiber um Stellen zu erweitern,
// die niemand sonst braucht.
//
// Eine fehlende Tabelle ist dagegen ein Fehler, den SQLite selbst meldet — und
// einer, der im Betrieb vorkommt: eine Datenbank aus einer neueren Fassung,
// eine Migration, die halb durchlief.
func TestDieSetaenderungMeldetJedenFehlerweg(t *testing.T) {
	faelle := []struct {
		name   string
		kaputt string
		warum  string
	}{
		{
			name: "die Angaben fehlen", kaputt: "set_values",
			warum: "der vorige Wert lässt sich nicht lesen",
		},
		{
			name: "der Verlauf fehlt", kaputt: "history",
			warum: "der Eintrag lässt sich nicht schreiben",
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db := frischeDatenbank(t)
			// **Mit Profil** — seit Werkstück 2 von Phase 4.12 gehört jedes
			// Set einem (#1351). Ein von Hand angelegtes ohne eines fände
			// keine Abfrage wieder, und das sähe aus wie ein Fehler im Weg.
			ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
				VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)
			ausfuehren(t, db, `DROP TABLE `+fall.kaputt)

			if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
				probekontext(t), 1, "set code", "PRB"); err == nil {
				t.Errorf("kein Fehler, obwohl %s", fall.warum)
			}
		})
	}
}

// **Und jeder Fehlerweg der Rücknahme.**
//
// Drei Wege gehen aus `setwegGehen` hinaus — entfernen, ans Ende setzen, an
// die Stelle setzen —, und jeder kann scheitern. Ohne diese Prüfung stünde
// dort eine Behauptung darüber, was im Ernstfall geschieht.
func TestDieRuecknahmeEinerSetangabeMeldetJedenFehlerweg(t *testing.T) {
	faelle := []struct {
		name   string
		vorher string
		ohne   bool
	}{
		{name: "beim Entfernen", vorher: ""},
		{name: "beim Setzen an die Stelle", vorher: "PRB"},
		{name: "beim Anhängen", vorher: "PRB", ohne: true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			db := frischeDatenbank(t)
			// **Mit Profil** — seit Werkstück 2 von Phase 4.12 gehört jedes
			// Set einem (#1351). Ein von Hand angelegtes ohne eines fände
			// keine Abfrage wieder, und das sähe aus wie ein Fehler im Weg.
			ausfuehren(t, db, `INSERT INTO sets (id, name, spiel, geoeffnet, profil_id)
				VALUES (1, 'Probe', 'probe', '2026-08-19T10:00:00.000Z', 1)`)

			if fall.vorher != "" {
				ausfuehren(t, db,
					`INSERT INTO set_values VALUES (1, 'set code', '`+fall.vorher+`', 0)`)
			}

			if _, _, err := storage.NeuerKartenschreiber(db).SetwertSetzenMitVorwert(
				probekontext(t), 1, "set code", "XYZ"); err != nil {
				t.Fatalf("setzen: %v", err)
			}

			// **Die Stelle herausnehmen** macht aus dem Weg „an die Stelle"
			// den Weg „ans Ende" — so sah eine Zeile vor #160 aus.
			if fall.ohne {
				ausfuehren(t, db,
					`UPDATE history SET rueckweg = json_remove(rueckweg, '$.reihenfolge')`)
			}

			// **Und dann die Tabelle weg.** Der Verlauf steht noch, der Weg
			// dorthin nicht mehr.
			ausfuehren(t, db, `DROP TABLE set_values`)

			if _, _, err := storage.NeuerVerlaufsleser(db).
				Zuruecknehmen(probekontext(t), 1); err == nil {
				t.Error("kein Fehler, obwohl die Tabelle fehlt")
			}
		})
	}
}
