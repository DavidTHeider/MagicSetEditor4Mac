// **Die Fehlerwege der Schlüsselwortrücknahme**
// (#1077).
//
// **Eine eigene Datei**, weil `schluesselwortverlauf_test.go` mit ihnen über
// die Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort, **dass**
// die Rücknahme greift, hier, **was geschieht, wenn sie es nicht kann**.
package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// alsSicht ersetzt `set_keywords` durch eine Sicht auf ihre Zeilen.
//
// # Warum eine Sicht
//
// Eine Sicht ohne `INSTEAD OF`-Trigger lässt sich **lesen**, aber nicht
// schreiben — und genau das trennt die beiden Hälften von [imZug]: der Stand
// davor kommt an, die Änderung scheitert. Wer stattdessen die Tabelle
// entfernte, scheiterte schon am Lesen, und die Fehlerwege der Änderung
// blieben unbetreten (§10, „eine Gegenprobe, die nichts ändert, ist keine").
func alsSicht(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.ExecContext(probekontext(t),
		`CREATE TABLE abschrift AS SELECT * FROM set_keywords`); err != nil {
		t.Fatalf("abschreiben: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_keywords`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t),
		`CREATE VIEW set_keywords AS SELECT * FROM abschrift`); err != nil {
		t.Fatalf("Sicht anlegen: %v", err)
	}
}

// **Jede der drei Änderungen meldet ihren Schreibfehler.**
func TestJedeWortaenderungMeldetIhrenSchreibfehler(t *testing.T) {
	t.Run("anlegen", func(t *testing.T) {
		db, setID := mitWoertern(t)
		alsSicht(t, db)

		if _, err := storage.NeueSchluesselwortablage(db).
			SchluesselwortAnlegen(probekontext(t), setID, "core"); err == nil {
			t.Error("das Anlegen auf eine Sicht kam durch")
		}
	})

	t.Run("setzen", func(t *testing.T) {
		db, setID := mitWoertern(t)
		erstes := woerter(t, db, setID)[0]

		alsSicht(t, db)

		if err := storage.NeueSchluesselwortablage(db).
			SchluesselwortSetzen(probekontext(t), setID, erstes.ID, "wort", "X"); err == nil {
			t.Error("das Setzen auf eine Sicht kam durch")
		}
	})

	t.Run("löschen", func(t *testing.T) {
		db, setID := mitWoertern(t)
		erstes := woerter(t, db, setID)[0]

		alsSicht(t, db)

		if err := storage.NeueSchluesselwortablage(db).
			SchluesselwortLoeschen(probekontext(t), setID, erstes.ID); err == nil {
			t.Error("das Löschen auf eine Sicht kam durch")
		}
	})
}

// **Und die Rücknahme meldet ihren.**
//
// **Beide Richtungen**: das Zurückbringen schreibt mit `INSERT OR REPLACE`,
// das Wegnehmen mit `DELETE` — zwei Anweisungen, zwei Fehlerwege.
func TestDieWortruecknahmeMeldetIhrenSchreibfehler(t *testing.T) {
	t.Run("zurückbringen", func(t *testing.T) {
		db, setID := mitWoertern(t)
		erstes := woerter(t, db, setID)[0]

		if err := storage.NeueSchluesselwortablage(db).
			SchluesselwortLoeschen(probekontext(t), setID, erstes.ID); err != nil {
			t.Fatalf("Löschen: %v", err)
		}

		alsSicht(t, db)

		if _, _, err := storage.NeuerVerlaufsleser(db).
			Zuruecknehmen(probekontext(t), setID); err == nil {
			t.Error("das Zurückbringen auf eine Sicht kam durch")
		}
	})

	t.Run("wegnehmen", func(t *testing.T) {
		db, setID := mitWoertern(t)

		if _, err := storage.NeueSchluesselwortablage(db).
			SchluesselwortAnlegen(probekontext(t), setID, "core"); err != nil {
			t.Fatalf("Anlegen: %v", err)
		}

		alsSicht(t, db)

		if _, _, err := storage.NeuerVerlaufsleser(db).
			Zuruecknehmen(probekontext(t), setID); err == nil {
			t.Error("das Wegnehmen auf eine Sicht kam durch")
		}
	})
}

// **Ein unlesbarer Verlaufseintrag wird gemeldet, nicht verschluckt.**
//
// **Beide Wege**: `hinweg` und `rueckweg` sind zwei Spalten, und ein
// vergessener Fehlerweg an einer von beiden gäbe einen leeren Stand — der beim
// Zurücknehmen ein Schlüsselwort löschte, das dastehen soll.
func TestEinUnlesbarerWortschrittWirdGemeldet(t *testing.T) {
	for name, zeile := range map[string]string{
		"hinweg":   `'{kein json', '{"gabEs":true}'`,
		"rueckweg": `'{"gabEs":true}', '{kein json'`,
	} {
		t.Run(name, func(t *testing.T) {
			db, setID := mitWoertern(t)

			// #nosec G202 -- `zeile` steht in der Tabelle darüber, nicht von außen
			anweisung := `INSERT INTO history
				        (set_id, karte_id, zeitpunkt, handlung, hinweg, rueckweg)
				 VALUES (?, 0, '2026-08-25', 'schluesselwort', ` + zeile + `)`

			if _, err := db.ExecContext(probekontext(t), anweisung, setID); err != nil {
				t.Fatalf("Eintrag setzen: %v", err)
			}

			if _, _, err := storage.NeuerVerlaufsleser(db).
				Zuruecknehmen(probekontext(t), setID); err == nil {
				t.Error("der unlesbare Eintrag kam durch")
			}
		})
	}
}

// **Jedes der fünf Felder kommt zurück.**
//
// **Alle fünf und nicht eines**: `mitFeld` rechnet den Stand danach aus, statt
// ihn noch einmal zu lesen — und eine vergessene Zeile in dieser Zuordnung
// fiele nur an dem einen Feld auf, das sie betrifft.
func TestJedesFeldEinesSchluesselwortsKommtZurueck(t *testing.T) {
	for feld, wert := range map[string]string{
		"wort":       "Flügelschlag",
		"muster":     "Flügelschlag <atom-param>zahl</atom-param>",
		"erinnerung": "Nur von Fliegern zu blocken.",
		"regeln":     "Eine Erfindung dieses Sets.",
		"modus":      "expert",
	} {
		t.Run(feld, func(t *testing.T) {
			db, setID := mitWoertern(t)
			erstes := woerter(t, db, setID)[0]

			if err := storage.NeueSchluesselwortablage(db).
				SchluesselwortSetzen(probekontext(t), setID, erstes.ID, feld, wert); err != nil {
				t.Fatalf("Setzen: %v", err)
			}

			// **Zurück und wieder vor** — der Weg vorwärts trägt den neuen
			// Stand, und nur er zeigt, ob `mitFeld` das richtige Feld gesetzt
			// hat.
			zurueck(t, db, setID)
			vor(t, db, setID)

			if gelesen := feldwert(t, db, setID, erstes.ID, feld); gelesen != wert {
				t.Errorf("nach dem Wiederherstellen steht %q statt %q", gelesen, wert)
			}
		})
	}
}

// feldwert liest ein Feld eines Schlüsselworts.
func feldwert(t *testing.T, db *sql.DB, setID, wortID int64, feld string) string {
	t.Helper()

	for _, wort := range woerter(t, db, setID) {
		if wort.ID != wortID {
			continue
		}

		switch feld {
		case "wort":
			return wort.Word
		case "muster":
			return wort.Match
		case "erinnerung":
			return wort.Reminder
		case "regeln":
			return wort.Rules
		case "modus":
			return wort.Mode
		}
	}

	t.Fatalf("das Schlüsselwort %d steht nicht in der Liste", wortID)

	return ""
}

// **Ein Löschen, das nichts trifft, ist ein Fehler.**
//
// **Neben dem Fall darüber**: jener löscht eine Kennung, die es nie gab; hier
// gibt es das Set, und die Kennung gehört einem anderen. Beide gehen durch
// `getroffen`, aber nur dieser betritt ihn nach einem geglückten Standlesen.
func TestEinLoeschenUeberDieSetgrenzeWirdAbgewiesen(t *testing.T) {
	db, setID := mitWoertern(t)
	erstes := woerter(t, db, setID)[0]

	zweites, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Anderes", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	if fehler := storage.NeueSchluesselwortablage(db).
		SchluesselwortLoeschen(probekontext(t), zweites, erstes.ID); fehler == nil {
		t.Fatal("die Kennung aus dem fremden Set kam durch")
	}

	// **Und es steht noch da** — die Setgrenze hat gehalten.
	if gelesen := woerter(t, db, setID); len(gelesen) != 2 {
		t.Errorf("%d Schlüsselwörter im ersten Set", len(gelesen))
	}
}
