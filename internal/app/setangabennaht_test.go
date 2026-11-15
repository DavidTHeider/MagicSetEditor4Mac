// **Die Setangabennaht für die Flächen ohne Fenster**
// (#1107).
package app_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Naht kommt zusammen und reicht beide Seiten durch.**
//
// **Gegen die Saat und nicht gegen ein leeres Zuhause**: ohne Spielpaket wäre
// die Feldliste leer, und ein Lauf über eine leere Liste prüfte die Umsetzung
// nicht.
func TestDieSetangabennahtReichtBeideSeitenDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	weg, nahtfehler := app.SetangabennahtUeber(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	felder, werte, lesefehler := weg.Angaben(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Angaben: %v", lesefehler)
	}

	if len(felder) == 0 {
		t.Errorf("kein Angabenfeld — die Saat müsste welche haben: %+v", werte)
	}

	// **Zuerst schreiben, dann noch einmal**: der zweite Ruf nennt den ersten
	// Wert, und das ist der Beleg, dass geschrieben wurde.
	if _, hatte, fehler := weg.Setzen(probekontext(t), setID, "title", "Erst"); fehler != nil {
		t.Fatalf("Setzen: %v", fehler)
	} else if hatte {
		t.Error("das Feld gab es angeblich schon")
	}

	vorher, hatte, fehler := weg.Setzen(probekontext(t), setID, "title", "Dann")
	if fehler != nil {
		t.Fatalf("zweites Setzen: %v", fehler)
	}

	if !hatte || vorher != "Erst" {
		t.Errorf("der vorige Wert lautet %q (hatte: %v) statt »Erst«", vorher, hatte)
	}

	pruefeVorbehalte(t, db, weg, setID)
}

// pruefeVorbehalte legt einen Rohblock an und liest ihn zurück.
//
// **Ein Block wird eingesetzt und nicht gehofft.** Ein Set aus der Saat hat
// keinen — und ein Lauf über eine leere Liste betritt die Umsetzung nicht, die
// er zu prüfen behauptet (§10).
func pruefeVorbehalte(
	t *testing.T, db *sql.DB, weg *app.Setangabenweg, setID int64,
) {
	t.Helper()

	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO raw_blocks VALUES (?, 'set', 12, 'unbekannt: wert', 0)`,
		setID); err != nil {
		t.Fatalf("Rohblock einsetzen: %v", err)
	}

	vorbehalte, fehler := weg.Vorbehalte(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Vorbehalte: %v", fehler)
	}

	if len(vorbehalte) != 1 {
		t.Fatalf("%d Vorbehalte statt einem: %+v", len(vorbehalte), vorbehalte)
	}

	if vorbehalte[0].Datei != "set" || vorbehalte[0].Zeile != 12 ||
		vorbehalte[0].Inhalt != "unbekannt: wert" {
		t.Errorf("der Block kam verstümmelt an: %+v", vorbehalte[0])
	}
}

// **Ein Lesefehler bei den Werten kommt durch** — mit Saat, damit die Felder
// vorher gelingen.
//
// **Ohne Saat prüfte das nichts**: dann scheitert schon `Setfelder`, und der
// zweite Fehlerweg bliebe unbetreten. Genau das war der erste Anlauf.
func TestEinFehlerBeimLesenDerSetwerteKommtDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	weg, nahtfehler := app.SetangabennahtUeber(db, domain.UrheberDienst)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_values`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, _, err := weg.Angaben(probekontext(t), setID); err == nil {
		t.Error("das Lesen ohne set_values kam durch")
	}
}

// **Ohne Zuhause kommt die Naht nicht zusammen**, und ein Lesefehler kommt
// durch.
func TestDieSetangabennahtMeldetIhreNahtfehler(t *testing.T) {
	t.Run("ohne Zuhause", func(t *testing.T) {
		t.Setenv("HOME", "")

		_, fehler := app.SetangabennahtUeber(frischerBestand(t), domain.UrheberDienst)
		if fehler == nil {
			t.Error("kein Fehler ohne Zuhause")
		}
	})

	t.Run("ein Lesefehler kommt durch", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		db := frischerBestand(t)

		weg, fehler := app.SetangabennahtUeber(db, domain.UrheberDienst)
		if fehler != nil {
			t.Fatalf("Naht: %v", fehler)
		}

		// **Ohne die Tabelle scheitert das Lesen** — der Weg, auf dem der
		// Fehler durch die Naht kommt.
		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_values`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, _, err := weg.Angaben(probekontext(t), 1); err == nil {
			t.Error("das Lesen ohne Tabelle kam durch")
		}

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE raw_blocks`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := weg.Vorbehalte(probekontext(t), 1); err == nil {
			t.Error("die Vorbehalte ohne Tabelle kamen durch")
		}
	})
}

// **Die mitgelieferten Sets kommen ohne Bestand.**
//
// **Ohne Bündel gibt es keine** — und das ist kein Fehler, sondern die Lage
// jedes `go test`: `Beispielsetverzeichnis` rechnet vom Programmpfad, und ein
// Testbinär liegt in keinem `.app`. Was hier geprüft wird, ist deshalb, dass
// die Antwort eine **Liste** ist und kein `nil`.
func TestDieBeispielsetlisteBrauchtKeinenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	sets, befund := app.Beispielsetliste()
	if befund != nil {
		t.Fatalf("Befund: %+v", befund)
	}

	if sets == nil {
		t.Error("die Liste kam als nil — eine leere wäre richtig")
	}
}
