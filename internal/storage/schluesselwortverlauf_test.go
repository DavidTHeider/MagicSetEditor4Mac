// **Die Rücknahme greift auf Schlüsselwörter**
// (#1077).
package storage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitWoertern legt ein Set mit zwei Schlüsselwörtern an.
func mitWoertern(t *testing.T) (db *sql.DB, setID int64) {
	t.Helper()

	db = frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probe", Game: "probe",
		Keywords: []*domain.Keyword{
			{Word: "Flug", Match: "Flug", Mode: "core"},
			{Word: "Beben", Match: "Beben", Mode: "custom"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	return db, setID
}

// woerter liest die Schlüsselwörter eines Sets.
func woerter(t *testing.T, db *sql.DB, setID int64) []*domain.Keyword {
	t.Helper()

	gelesen, err := storage.NeuerSetLeser(db).Schluesselwoerter(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Schluesselwoerter: %v", err)
	}

	return gelesen
}

// zurueck nimmt einen Schritt zurück.
func zurueck(t *testing.T, db *sql.DB, setID int64) (domain.Verlaufsschritt, bool) {
	t.Helper()

	schritt, bewegt, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Zuruecknehmen: %v", err)
	}

	return schritt, bewegt
}

// vor stellt einen Schritt wieder her.
func vor(t *testing.T, db *sql.DB, setID int64) (domain.Verlaufsschritt, bool) {
	t.Helper()

	schritt, bewegt, err := storage.NeuerVerlaufsleser(db).Wiederherstellen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Wiederherstellen: %v", err)
	}

	return schritt, bewegt
}

// **Ein zurückgenommenes Anlegen nimmt das Schlüsselwort wieder weg.**
func TestEinZurueckgenommenesAnlegenNimmtDasWortWeg(t *testing.T) {
	db, setID := mitWoertern(t)

	if _, err := storage.NeueSchluesselwortablage(db).
		SchluesselwortAnlegen(probekontext(t), setID, "core"); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	if gelesen := woerter(t, db, setID); len(gelesen) != 3 {
		t.Fatalf("%d Schlüsselwörter nach dem Anlegen", len(gelesen))
	}

	if _, bewegt := zurueck(t, db, setID); !bewegt {
		t.Fatal("nichts bewegt")
	}

	if gelesen := woerter(t, db, setID); len(gelesen) != 2 {
		t.Errorf("%d Schlüsselwörter nach der Rücknahme", len(gelesen))
	}

	// **Und ⌘⇧Z bringt es wieder.**
	if _, bewegt := vor(t, db, setID); !bewegt {
		t.Fatal("nichts wiederhergestellt")
	}

	if gelesen := woerter(t, db, setID); len(gelesen) != 3 {
		t.Errorf("%d Schlüsselwörter nach dem Wiederherstellen", len(gelesen))
	}
}

// **Ein zurückgenommenes Ändern stellt den vorigen Wert her.**
func TestEinZurueckgenommenesAendernStelltDenVorigenWertHer(t *testing.T) {
	db, setID := mitWoertern(t)

	erstes := woerter(t, db, setID)[0]

	if err := storage.NeueSchluesselwortablage(db).
		SchluesselwortSetzen(probekontext(t), setID, erstes.ID, "wort", "Fliegend"); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if wort := woerter(t, db, setID)[0].Word; wort != "Fliegend" {
		t.Fatalf("nach dem Setzen steht %q", wort)
	}

	zurueck(t, db, setID)

	if wort := woerter(t, db, setID)[0].Word; wort != "Flug" {
		t.Errorf("nach der Rücknahme steht %q statt »Flug«", wort)
	}

	vor(t, db, setID)

	if wort := woerter(t, db, setID)[0].Word; wort != "Fliegend" {
		t.Errorf("nach dem Wiederherstellen steht %q statt »Fliegend«", wort)
	}
}

// **Ein zurückgenommenes Löschen bringt das Wort an seine Stelle zurück.**
//
// **An seine Stelle und nicht ans Ende** — die Abnahme aus #1077. Die
// Reihenfolge entscheidet, welches Schlüsselwort zuerst auf einen Text passt
// (`src/data/keyword.cpp`, Zeilen 148–208, Fassung f9d9356 vom 2021-07-06);
// eine Rücknahme, die sie ändert, ist keine.
func TestEinZurueckgenommenesLoeschenBringtDasWortAnSeineStelle(t *testing.T) {
	db, setID := mitWoertern(t)

	vorheriges := woerter(t, db, setID)

	erstes := vorheriges[0]
	if erstes.Word != "Flug" {
		t.Fatalf("das erste ist %q", erstes.Word)
	}

	if err := storage.NeueSchluesselwortablage(db).
		SchluesselwortLoeschen(probekontext(t), setID, erstes.ID); err != nil {
		t.Fatalf("Löschen: %v", err)
	}

	if gelesen := woerter(t, db, setID); len(gelesen) != 1 || gelesen[0].Word != "Beben" {
		t.Fatalf("nach dem Löschen: %+v", gelesen)
	}

	zurueck(t, db, setID)

	nachher := woerter(t, db, setID)
	if len(nachher) != 2 {
		t.Fatalf("%d Schlüsselwörter nach der Rücknahme", len(nachher))
	}

	// **Vorn, nicht hinten** — das ist der ganze Punkt.
	if nachher[0].Word != "Flug" || nachher[1].Word != "Beben" {
		t.Errorf("die Reihenfolge ist %q, %q", nachher[0].Word, nachher[1].Word)
	}

	// **Und jedes Feld ist wieder da**, nicht nur der Name.
	if nachher[0].Match != "Flug" || nachher[0].Mode != "core" {
		t.Errorf("das zurückgebrachte ist %+v", nachher[0])
	}

	// **Und die Kennung ist dieselbe** — eine neue machte jeden Verweis
	// darauf ungültig.
	if nachher[0].ID != erstes.ID {
		t.Errorf("die Kennung ist %d statt %d", nachher[0].ID, erstes.ID)
	}
}

// **Der Urheber steht im Schritt.**
//
// Ohne ihn ist eine Rücknahme im Zwei-Schreiber-Betrieb eine Vermutung (#52).
func TestDerUrheberEinerWortaenderungStehtImSchritt(t *testing.T) {
	db, setID := mitWoertern(t)

	if _, err := storage.NeueSchluesselwortablage(db).Als(domain.UrheberServer).
		SchluesselwortAnlegen(probekontext(t), setID, "core"); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	schritt, _ := zurueck(t, db, setID)
	if schritt.Urheber != domain.UrheberServer {
		t.Errorf("der Urheber ist %q statt %q", schritt.Urheber, domain.UrheberServer)
	}
}

// **Eine gescheiterte Änderung schreibt keinen Schritt.**
//
// **Die Gegenprobe zur Transaktion**: ohne sie stünde nur fest, dass ein
// geglückter Zug einen Schritt anlegt — nicht, dass ein gescheiterter keinen
// halben zurücklässt.
func TestEineGescheiterteWortaenderungSchreibtKeinenSchritt(t *testing.T) {
	db, setID := mitWoertern(t)

	if err := storage.NeueSchluesselwortablage(db).
		SchluesselwortLoeschen(probekontext(t), setID, 4711); err == nil {
		t.Fatal("eine Kennung, die es nicht gibt, kam durch")
	}

	if schritt, bewegt := zurueck(t, db, setID); bewegt {
		t.Errorf("es wurde etwas bewegt: %+v", schritt)
	}
}

// **Ein Schreibfehler beim Verlauf kommt durch.**
func TestEinVerlaufsfehlerBeiWoerternKommtDurch(t *testing.T) {
	db, setID := mitWoertern(t)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE history`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := storage.NeueSchluesselwortablage(db).
		SchluesselwortAnlegen(probekontext(t), setID, "core"); err == nil {
		t.Error("das Anlegen ohne Verlaufstabelle kam durch")
	}

	// **Und die Änderung ist zurückgenommen** — geschrieben wird beides oder
	// nichts.
	if gelesen := woerter(t, db, setID); len(gelesen) != 2 {
		t.Errorf("%d Schlüsselwörter — der halbe Zug ist stehengeblieben", len(gelesen))
	}
}

// **Ein Lesefehler beim Stand kommt durch — beim Setzen und beim Löschen.**
//
// **Beide und nicht eines**: jedes liest den Stand davor selbst, und ein
// vergessener Fehlerweg gäbe einen leeren Vorher-Stand — der beim Zurücknehmen
// ein Schlüsselwort löschte, das dastehen soll.
func TestEinStandsfehlerBeiWoerternKommtDurch(t *testing.T) {
	t.Run("beim Setzen", func(t *testing.T) {
		db, setID := mitWoertern(t)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_keywords`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if err := storage.NeueSchluesselwortablage(db).
			SchluesselwortSetzen(probekontext(t), setID, 1, "wort", "x"); err == nil {
			t.Error("das Setzen ohne Tabelle kam durch")
		}
	})

	t.Run("beim Löschen", func(t *testing.T) {
		db, setID := mitWoertern(t)

		if _, err := db.ExecContext(probekontext(t), `DROP TABLE set_keywords`); err != nil {
			t.Fatalf("Tabelle entfernen: %v", err)
		}

		if err := storage.NeueSchluesselwortablage(db).
			SchluesselwortLoeschen(probekontext(t), setID, 1); err == nil {
			t.Error("das Löschen ohne Tabelle kam durch")
		}
	})
}

// **Und eine geschlossene Datenbank auch.**
func TestEineGeschlosseneDatenbankBeiWoerternKommtDurch(t *testing.T) {
	db, setID := mitWoertern(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if _, err := storage.NeueSchluesselwortablage(db).
		SchluesselwortAnlegen(context.Background(), setID, "core"); err == nil {
		t.Error("eine geschlossene Datenbank kam durch")
	}
}
