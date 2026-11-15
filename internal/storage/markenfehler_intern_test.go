package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein `Scan` gegen die falsche Spaltenzahl meldet sich** —
// #1208.
//
// # Warum dieser Weg überhaupt geprüft wird
//
// Er ist der, den niemand herbeiführen kann, solange Abfrage und `Scan`
// zusammenpassen — und genau deshalb fiele es nicht auf, wenn seine Meldung
// verloren ginge. Dieselbe Bauform wie bei der Sammlung
// (`sammlungsfehler_intern_test.go`): eine Abfrage mit der falschen
// Spaltenzahl an einen `Scan`, der eine andere erwartet.
func TestEinMarkenscanGegenDieFalscheSpaltenzahlMeldetSich(t *testing.T) {
	db := offeneDatenbank(t)

	marken := NeueMarken(db)

	if _, err := marken.woerter(probekontext(t), `SELECT 1, 2`); err == nil {
		t.Error("ein Scan gegen zwei Spalten kam durch")
	}

	if _, err := marken.jeKarte(probekontext(t), `SELECT 1`); err == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}

	if _, err := NeuerSetLeser(db).markenzeilen(probekontext(t), `SELECT 1`); err == nil {
		t.Error("ein Markentreffer gegen eine Spalte kam durch")
	}
}

// **Und ein Feldfilter, der keine Marken meint, fragt gar nicht erst.**
func TestEinFremderFeldfilterFragtDieMarkenNicht(t *testing.T) {
	db := offeneDatenbank(t)

	treffer, err := NeuerSetLeser(db).Markentreffer(probekontext(t),
		domain.SearchQuery{Text: "a", Field: "name"})
	if err != nil || treffer != nil {
		t.Errorf("Markentreffer = %v, %v", treffer, err)
	}
}

// **Ein leerer Suchtext findet keine Marke.**
//
// Über [SetLeser.Suche] ist der Weg nicht zu erreichen — dort hält schon das
// leere FTS5-Muster an. Hier steht er als eigene Frage: „suche nach nichts"
// gibt nichts und nicht **alles**.
func TestEinLeererSuchtextFindetKeineMarke(t *testing.T) {
	db := offeneDatenbank(t)

	treffer, err := NeuerSetLeser(db).Markentreffer(probekontext(t),
		domain.SearchQuery{Text: "   "})
	if err != nil || treffer != nil {
		t.Errorf("Markentreffer = %v, %v", treffer, err)
	}
}

// **Fehlt die Markentabelle, sagt die Suche es** — statt die Feldtreffer
// allein zurückzugeben.
//
// **Der Fall ist der Ernstfall einer halben Migration**: die Volltextsuche
// träfe weiter, und die Marken fielen still aus.
func TestEineKaputteMarkentabelleHaeltDieSucheAn(t *testing.T) {
	db := migrierteDatenbank(t)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE karten_marken`); err != nil {
		t.Fatalf("Tabelle entfernen: %v", err)
	}

	if _, err := NeuerSetLeser(db).markenzeilen(probekontext(t), markensuchen,
		"a", int64(0), int64(0), 10); err == nil {
		t.Error("die fehlende Tabelle kam durch")
	}

	if _, err := NeuerSetLeser(db).Suche(probekontext(t),
		domain.SearchQuery{Text: "feuerball"}); err == nil {
		t.Error("die Suche gab ein Ergebnis ohne ihre Marken")
	}
}

// **Und ein `RowsAffected`, das scheitert, wird gemeldet.**
//
// Der SQLite-Treiber kann das nicht; im Ernstfall entscheidet es, ob die
// Anwendung „die Marke stand nicht da" von „das Zählen ging schief"
// unterscheidet.
func TestDieMarkenablageMeldetIhrenZaehlfehler(t *testing.T) {
	db, err := sql.Open("stolper-zaehlen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	zweiterSchreibzaehler = 0

	if fehler := NeueMarken(db).Entfernen(probekontext(t), 1, "fertig"); fehler == nil {
		t.Error("der Zählfehler ging verloren")
	}
}

// migrierteDatenbank gibt einen Bestand mit allen Tabellen.
//
// **`offeneDatenbank` migriert nicht** — sie ist für die Wege gebaut, die eine
// Verbindung brauchen und keine Tabelle.
func migrierteDatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db := offeneDatenbank(t)

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	return db
}
