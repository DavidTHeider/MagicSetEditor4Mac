// **Die Packnaht für die Flächen ohne Fenster** —
// #889,
// Befund 7 des Zweitblicks über Phase W.
package app_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Naht kommt zusammen und antwortet — mit den Packarten der Saat.**
//
// **Gegen die Saat und nicht gegen ein leeres Zuhause**: ohne Spielpaket wäre
// die Antwort leer, und ein Lauf über eine leere Liste prüfte die Umsetzung
// nicht. Die Saat trägt seit #889 zwei Packarten (§11).
func TestDiePacknahtReichtBeideSeitenDurch(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)

	setID, karteID := saatsetImBestand(t, db)
	_ = karteID

	weg, nahtfehler := app.PackUeber(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	antwort, lesefehler := weg.Arten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Arten: %v", lesefehler)
	}

	gefunden := false

	for _, art := range antwort.Arten {
		if art.Name == "probepack" && art.Waehlbar {
			gefunden = true
		}
	}

	if !gefunden {
		t.Fatalf("`probepack` fehlt in %+v", antwort.Arten)
	}

	ziehung, ziehfehler := weg.Ziehen(probekontext(t), setID, "alles", "42")
	if ziehfehler != nil {
		t.Fatalf("Ziehen: %v", ziehfehler)
	}

	if len(ziehung.Karten) == 0 {
		t.Error("die Ziehung ist leer — die Saatkarte müsste durch den Filter kommen")
	}

	if _, err := weg.Ziehen(probekontext(t), setID, "gibtsnicht", "42"); err == nil {
		t.Error("eine Packart, die es nicht gibt, zog etwas")
	}
}

// **Ein unbrauchbarer Same wird abgewiesen, bevor gezogen wird.**
//
// **Als Validierungsfehler und nicht als Fehler der Naht**: es ist die Eingabe,
// die nicht stimmt, und eine Fläche muss das unterscheiden können (§8).
func TestEinUnbrauchbarerSameWirdAbgewiesen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PackUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	_, fehler := weg.Ziehen(probekontext(t), 1, "alles", "keine Zahl")
	if fehler == nil {
		t.Fatal("»keine Zahl« kam als Same durch")
	}

	if !strings.Contains(fehler.Error(), domain.ErrValidation.Error()) {
		t.Errorf("der Fehler ist %v — will eine Validierung", fehler)
	}
}

// **Ohne Zuhause kommt die Naht nicht zusammen**, und ein Lesefehler kommt
// durch.
func TestDiePacknahtMeldetIhreNahtfehler(t *testing.T) {
	t.Run("ohne Zuhause", func(t *testing.T) {
		t.Setenv("HOME", "")

		if _, fehler := app.PackUeber(frischerBestand(t)); fehler == nil {
			t.Error("kein Fehler ohne Zuhause")
		}
	})

	t.Run("ein Lesefehler kommt durch", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())

		db := frischerBestand(t)

		weg, fehler := app.PackUeber(db)
		if fehler != nil {
			t.Fatalf("Naht: %v", fehler)
		}

		// **Ohne die Tabelle scheitert das Lesen** — der Weg, auf dem der
		// Fehler durch die Naht kommt.
		if _, err := db.ExecContext(probekontext(t), `DROP TABLE cards`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := weg.Arten(probekontext(t), 1); err == nil {
			t.Error("das Lesen ohne Tabelle kam durch")
		}

		if _, err := weg.Ziehen(probekontext(t), 1, "alles", "1"); err == nil {
			t.Error("das Ziehen ohne Tabelle kam durch")
		}
	})
}
