package storage_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Packarten eines Sets überleben den Bestand**
// (#1081).
func TestSetpackartenKommenMitDemSet(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probe", Game: "probe",
		Packs: []*domain.PackArt{
			{
				Name: "mein booster", Select: "all", Selectable: true,
				Items: []domain.PackItem{
					{Name: "common", Amount: 10, Weight: 1},
					{Name: "rare", Amount: 1, Weight: 0.5},
				},
			},
			{Name: "nur filter", Select: "no replace", Filter: `card.rarity == "rare"`},
		},
	}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	arten, lesefehler := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Setpackarten: %v", lesefehler)
	}

	if len(arten) != 2 {
		t.Fatalf("%d Packarten: %+v", len(arten), arten)
	}

	// **Die Reihenfolge der Datei ist die der Liste** — sie ist der Ort, an
	// den zurückgeschrieben wird.
	if arten[0].Name != "mein booster" || arten[1].Name != "nur filter" {
		t.Errorf("die Reihenfolge ist %q, %q", arten[0].Name, arten[1].Name)
	}

	erste := arten[0]
	if !erste.Selectable || erste.Select != "all" {
		t.Errorf("die erste ist %+v", erste)
	}

	if len(erste.Items) != 2 {
		t.Fatalf("%d Posten: %+v", len(erste.Items), erste.Items)
	}

	if erste.Items[0].Name != "common" || erste.Items[0].Amount != 10 {
		t.Errorf("der erste Posten ist %+v", erste.Items[0])
	}

	// **Das Gewicht ist ein REAL** — eine Gewichtung von 0,5 ist zulässig
	// (`double weight`, `src/data/pack.hpp`, Zeile 75, Fassung f9d9356 vom
	// 2021-07-06). Als INTEGER käme hier 0 heraus.
	if erste.Items[1].Weight != 0.5 {
		t.Errorf("das Gewicht des zweiten ist %v statt 0,5", erste.Items[1].Weight)
	}

	// **`selectable: false` überlebt ebenso** — ohne diesen Fall stünde nur
	// fest, dass `true` durchkommt.
	if arten[1].Selectable {
		t.Errorf("die zweite gilt als wählbar: %+v", arten[1])
	}

	// **Und ihre leere Postenliste ist keine `nil`.**
	if arten[1].Items == nil {
		t.Error("die Posten der zweiten kamen als nil")
	}
}

// **Ein Set ohne eigene Packarten gibt eine leere Liste.**
func TestEinSetOhnePackartenGibtEineLeereListe(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	arten, lesefehler := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Setpackarten: %v", lesefehler)
	}

	if arten == nil || len(arten) != 0 {
		t.Errorf("die Liste ist %+v — eine leere wäre richtig", arten)
	}
}

// **Ein Lesefehler kommt durch — an beiden Tabellen.**
//
// **Zwei Fälle und nicht einer**: die Posten hängen an einer zweiten Tabelle,
// und ein Fehler dort bliebe unbetreten, wenn nur die erste fehlt.
func TestEinLesefehlerBeiDenPackartenKommtDurch(t *testing.T) {
	t.Run("ohne die Arten", func(t *testing.T) {
		db := frischeDatenbank(t)

		// **Ein Set, das es gibt** — seit #1815 fragt die Ablage zuerst, ob es
		// dem Profil gehört, und ohne Set scheiterte der Lauf an dieser Frage
		// statt an der Tabelle.
		setID := einLeeresSet(t, db)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_types`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID); err == nil ||
			!strings.Contains(err.Error(), "Packarten lesen") {
			t.Errorf("das Lesen ohne Tabelle meldete %v", err)
		}
	})

	t.Run("ohne die Posten", func(t *testing.T) {
		db := frischeDatenbank(t)

		setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
			Name: "Probe", Game: "probe",
			Packs: []*domain.PackArt{{Name: "booster"}},
		}, nil)
		if err != nil {
			t.Fatalf("SetAnlegen: %v", err)
		}

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_items`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID); err == nil {
			t.Error("das Lesen der Posten ohne Tabelle kam durch")
		}
	})
}

// **Ein Schreibfehler kommt durch — an beiden Tabellen.**
func TestEinSchreibfehlerBeiDenPackartenKommtDurch(t *testing.T) {
	t.Run("ohne die Arten", func(t *testing.T) {
		db := frischeDatenbank(t)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_types`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		_, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
			Name: "Probe", Game: "probe",
			Packs: []*domain.PackArt{{Name: "booster"}},
		}, nil)
		if err == nil {
			t.Error("das Schreiben ohne Tabelle kam durch")
		}
	})

	t.Run("ohne die Posten", func(t *testing.T) {
		db := frischeDatenbank(t)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_items`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		_, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
			Name: "Probe", Game: "probe",
			Packs: []*domain.PackArt{{
				Name: "booster", Items: []domain.PackItem{{Name: "common", Amount: 1}},
			}},
		}, nil)
		if err == nil {
			t.Error("das Schreiben der Posten ohne Tabelle kam durch")
		}
	})
}

// **Eine Spalte in fremder Form bricht das Lesen** — an beiden Tabellen.
//
// # Warum über eine Sicht
//
// Die Tabellen sind `STRICT`; ein Text in `waehlbar` kommt dort nicht hinein.
// Der Scanfehler ist trotzdem kein toter Zweig: eine ältere Datenbank, eine
// Migration von Hand, ein Treiberwechsel — dann steht dort etwas anderes. Eine
// Sicht mit demselben Namen führt genau diesen Fall herbei, und dieselbe
// Vorrichtung steht schon in `historie_test.go`.
func TestEineFremdeSpalteBrichtDasLesenDerPackarten(t *testing.T) {
	t.Run("bei den Arten", func(t *testing.T) {
		db := frischeDatenbank(t)
		setID := einLeeresSet(t, db)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_types`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		// **Das Set aus dem Bestand und nicht als Text in der Sicht** —
		// Parameter nimmt eine Sicht nicht an, und eine zusammengesetzte
		// Abfrage wiese gosec ab. Es gibt genau eines.
		if _, err := db.ExecContext(probekontext(t), `CREATE VIEW set_pack_types AS
			SELECT 1 AS id, 0 AS reihenfolge, 'x' AS name, 'all' AS auswahl,
			       '' AS filter, 'keine Zahl' AS waehlbar,
			       (SELECT max(id) FROM sets) AS set_id`); err != nil {
			t.Fatalf("Sicht anlegen: %v", err)
		}

		if _, err := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID); err == nil ||
			!strings.Contains(err.Error(), "eine Packart lesen") {
			t.Errorf("»keine Zahl« als waehlbar meldete %v", err)
		}
	})

	t.Run("bei den Posten", func(t *testing.T) {
		db := frischeDatenbank(t)

		setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
			Name: "Probe", Game: "probe",
			Packs: []*domain.PackArt{{Name: "booster"}},
		}, nil)
		if err != nil {
			t.Fatalf("SetAnlegen: %v", err)
		}

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_pack_items`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if _, err := db.ExecContext(probekontext(t), `CREATE VIEW set_pack_items AS
			SELECT 'x' AS name, 'keine Zahl' AS anzahl, 1.0 AS gewicht,
			       0 AS reihenfolge, 1 AS pack_id`); err != nil {
			t.Fatalf("Sicht anlegen: %v", err)
		}

		if _, err := storage.NeuerSetLeser(db).Setpackarten(probekontext(t), setID); err == nil {
			t.Error("»keine Zahl« kam als anzahl durch")
		}
	})
}

// einLeeresSet legt ein Set ohne Karten und Packarten an.
func einLeeresSet(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	return setID
}
